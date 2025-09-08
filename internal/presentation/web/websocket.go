package web

import (
	"crypto-checkout/internal/domain/invoice"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10

	// Time allowed to read the next pong message from the peer.
	pongWait = 60

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10 //nolint:mnd // Standard WebSocket ping period calculation

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// WebSocket buffer sizes.
	readBufferSize  = 1024
	writeBufferSize = 1024

	// Client send channel buffer size.
	sendChannelBuffer = 256
)

var upgrader = websocket.Upgrader{ //nolint:gochecknoglobals // WebSocket upgrader is a global configuration
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(_ *http.Request) bool {
		// Allow connections from any origin for now
		// In production, you should implement proper origin checking
		return true
	},
}

// Client represents a websocket client connection.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	invoiceID string
	logger    *zap.Logger
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Invoice-specific clients
	invoiceClients map[string]map[*Client]bool

	// Mutex for thread-safe access to invoiceClients
	mutex sync.RWMutex

	logger *zap.Logger
}

// NewHub creates a new WebSocket hub.
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		broadcast:      make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		invoiceClients: make(map[string]map[*Client]bool),
		logger:         logger,
	}
}

// Run starts the hub.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client.
func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true
	h.mutex.Lock()
	if h.invoiceClients[client.invoiceID] == nil {
		h.invoiceClients[client.invoiceID] = make(map[*Client]bool)
	}
	h.invoiceClients[client.invoiceID][client] = true
	h.mutex.Unlock()

	h.logger.Info("Client registered for invoice",
		zap.String("invoice_id", client.invoiceID),
		zap.Int("total_clients", len(h.clients)),
	)
}

// unregisterClient unregisters a client.
func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}

	h.mutex.Lock()
	if invoiceClients, exists := h.invoiceClients[client.invoiceID]; exists {
		delete(invoiceClients, client)
		if len(invoiceClients) == 0 {
			delete(h.invoiceClients, client.invoiceID)
		}
	}
	h.mutex.Unlock()

	h.logger.Info("Client unregistered for invoice",
		zap.String("invoice_id", client.invoiceID),
		zap.Int("total_clients", len(h.clients)),
	)
}

// broadcastMessage broadcasts a message to all clients.
func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// BroadcastToInvoice broadcasts a message to all clients subscribed to a specific invoice.
func (h *Hub) BroadcastToInvoice(invoiceID string, message []byte) {
	h.mutex.RLock()
	invoiceClients, exists := h.invoiceClients[invoiceID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	for client := range invoiceClients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			h.mutex.Lock()
			delete(invoiceClients, client)
			delete(h.clients, client)
			h.mutex.Unlock()
		}
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		if err := c.conn.Close(); err != nil {
			c.logger.Error("Failed to close WebSocket connection", zap.Error(err))
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.logger.Error("Failed to set read deadline", zap.Error(err))
		return
	}

	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			c.logger.Error("Failed to set read deadline in pong handler", zap.Error(err))
		}
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			c.logger.Error("Failed to close WebSocket connection in writePump", zap.Error(err))
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !c.writeMessage(message, ok) {
				return
			}

		case <-ticker.C:
			if !c.writePing() {
				return
			}
		}
	}
}

// writeMessage writes a message to the WebSocket connection.
func (c *Client) writeMessage(message []byte, ok bool) bool {
	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		c.logger.Error("Failed to set write deadline", zap.Error(err))
		return false
	}

	if !ok {
		if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
			c.logger.Error("Failed to write close message", zap.Error(err))
		}
		return false
	}

	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		c.logger.Error("Failed to get next writer", zap.Error(err))
		return false
	}

	if _, writeErr := w.Write(message); writeErr != nil {
		c.logger.Error("Failed to write message", zap.Error(writeErr))
		return false
	}

	// Add queued messages to the current websocket message.
	n := len(c.send)
	for range n {
		if _, writeErr := w.Write([]byte{'\n'}); writeErr != nil {
			c.logger.Error("Failed to write newline", zap.Error(writeErr))
			return false
		}
		if _, writeErr := w.Write(<-c.send); writeErr != nil {
			c.logger.Error("Failed to write queued message", zap.Error(writeErr))
			return false
		}
	}

	if closeErr := w.Close(); closeErr != nil {
		c.logger.Error("Failed to close writer", zap.Error(closeErr))
		return false
	}

	return true
}

// writePing writes a ping message to the WebSocket connection.
func (c *Client) writePing() bool {
	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		c.logger.Error("Failed to set write deadline for ping", zap.Error(err))
		return false
	}

	if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		c.logger.Error("Failed to write ping message", zap.Error(err))
		return false
	}

	return true
}

// serveWS handles websocket requests from the peer.
func (h *Handler) serveWS(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invoice ID is required"})
		return
	}

	// Verify invoice exists
	_, err := h.invoiceService.GetInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		h.Logger.Error("Failed to get invoice for WebSocket", zap.Error(err), zap.String("invoice_id", invoiceID))
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.Logger.Error("Failed to upgrade connection to WebSocket", zap.Error(err))
		return
	}

	client := &Client{
		hub:       h.hub,
		conn:      conn,
		send:      make(chan []byte, sendChannelBuffer),
		invoiceID: invoiceID,
		logger:    h.Logger,
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// BroadcastInvoiceUpdate broadcasts an invoice update to all connected clients.
func (h *Handler) BroadcastInvoiceUpdate(inv *invoice.Invoice) {
	update := InvoiceStatusUpdate{
		InvoiceID: string(inv.ID()),
		Status:    inv.Status().String(),
		Timestamp: time.Now().UTC(),
	}

	message, err := json.Marshal(update)
	if err != nil {
		h.Logger.Error("Failed to marshal invoice update", zap.Error(err))
		return
	}

	h.hub.BroadcastToInvoice(string(inv.ID()), message)
}

// InvoiceStatusUpdate represents a WebSocket message for invoice status updates.
type InvoiceStatusUpdate struct {
	InvoiceID string    `json:"invoice_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
