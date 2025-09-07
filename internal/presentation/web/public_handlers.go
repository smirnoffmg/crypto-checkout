package web

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/shared"
)

// GetPublicInvoiceData handles GET /api/v1/public/invoice/:id requests.
// @Summary Get public invoice data
// @Description Retrieve public invoice information for customers (no authentication required)
// @Tags Public API
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} PublicInvoiceResponse "Invoice data retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid invoice ID"
// @Failure 404 {object} ErrorResponse "Invoice not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/public/invoice/{id} [get]
func (h *Handler) GetPublicInvoiceData(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Debug("Empty invoice ID in public request")
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("invoice ID is required", nil))
		return
	}

	// Get invoice from service
	inv, err := h.invoiceService.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.Logger.Error("Failed to get invoice for public view", zap.Error(err), zap.String("invoice_id", id))
		if errors.Is(err, shared.ErrNotFound) {
			c.JSON(http.StatusNotFound, createNotFoundErrorResponse("invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, createValidationErrorResponse("failed to retrieve invoice", err))
		return
	}

	// Convert to public response
	response := h.toPublicInvoiceResponse(inv)
	c.JSON(http.StatusOK, response)
}

// GetPublicInvoiceStatus handles GET /api/v1/public/invoice/:id/status requests.
// @Summary Get invoice status
// @Description Get the current status of an invoice (no authentication required)
// @Tags Public API
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} PublicInvoiceStatusResponse "Invoice status retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid invoice ID"
// @Failure 404 {object} ErrorResponse "Invoice not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/public/invoice/{id}/status [get]
func (h *Handler) GetPublicInvoiceStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Debug("Empty invoice ID in status request")
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("invoice ID is required", nil))
		return
	}

	// Get invoice status from service
	status, err := h.invoiceService.GetInvoiceStatus(c.Request.Context(), id)
	if err != nil {
		h.Logger.Error("Failed to get invoice status", zap.Error(err), zap.String("invoice_id", id))
		if errors.Is(err, shared.ErrNotFound) {
			c.JSON(http.StatusNotFound, createNotFoundErrorResponse("invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, createValidationErrorResponse("failed to retrieve invoice status", err))
		return
	}

	response := PublicInvoiceStatusResponse{
		ID:        id,
		Status:    status.String(),
		Timestamp: time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// GetPublicInvoiceEvents handles GET /api/v1/public/invoice/:id/events requests (Server-Sent Events).
// @Summary Get invoice events stream
// @Description Stream real-time invoice events using Server-Sent Events (no authentication required)
// @Tags Public API
// @Accept json
// @Produce text/event-stream
// @Param id path string true "Invoice ID"
// @Success 200 {string} string "Event stream started"
// @Failure 400 {object} ErrorResponse "Invalid invoice ID"
// @Failure 404 {object} ErrorResponse "Invoice not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/public/invoice/{id}/events [get]
func (h *Handler) GetPublicInvoiceEvents(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Debug("Empty invoice ID in events request")
		c.JSON(http.StatusBadRequest, createValidationErrorResponse("invoice ID is required", nil))
		return
	}

	// Verify invoice exists
	_, err := h.invoiceService.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.Logger.Error("Failed to get invoice for events", zap.Error(err), zap.String("invoice_id", id))
		if errors.Is(err, shared.ErrNotFound) {
			c.JSON(http.StatusNotFound, createNotFoundErrorResponse("invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, createValidationErrorResponse("failed to retrieve invoice", err))
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Send initial event
	event := fmt.Sprintf("data: {\"event\": \"connected\", \"invoice_id\": \"%s\", \"timestamp\": \"%s\"}\n\n",
		id, time.Now().UTC().Format(time.RFC3339))
	c.SSEvent("", event)
	c.Writer.Flush()

	// Keep connection alive with periodic heartbeats
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			// Send heartbeat
			heartbeat := fmt.Sprintf("data: {\"event\": \"heartbeat\", \"timestamp\": \"%s\"}\n\n",
				time.Now().UTC().Format(time.RFC3339))
			c.SSEvent("", heartbeat)
			c.Writer.Flush()
		}
	}
}

// toPublicInvoiceResponse converts a domain invoice to a public response.
func (h *Handler) toPublicInvoiceResponse(inv *invoice.Invoice) PublicInvoiceResponse {
	// Convert items
	items := make([]InvoiceItemResponse, len(inv.Items()))
	for i, item := range inv.Items() {
		items[i] = InvoiceItemResponse{
			Description: item.Description(),
			UnitPrice:   item.UnitPrice().Amount().String(),
			Quantity:    item.Quantity().String(),
			Total:       item.TotalPrice().Amount().String(),
		}
	}

	// Get payment address
	var address string
	if addr := inv.PaymentAddress(); addr != nil {
		address = addr.String()
	}

	// Get expiration time
	var expiresAt time.Time
	if exp := inv.Expiration(); exp != nil {
		expiresAt = exp.ExpiresAt()
	}

	// Calculate time remaining
	var timeRemaining int64
	if !expiresAt.IsZero() {
		remaining := time.Until(expiresAt)
		if remaining > 0 {
			timeRemaining = int64(remaining.Seconds())
		}
	}

	// TODO: Get payments from payment service
	// For now, return empty payments
	payments := []PublicPaymentResponse{}

	// TODO: Calculate payment progress
	// For now, return nil
	var paymentProgress *PaymentProgressResponse

	// TODO: Get return/cancel URLs from invoice metadata
	// For now, return nil
	var returnURL, cancelURL *string

	return PublicInvoiceResponse{
		ID:              inv.ID(),
		Title:           inv.Title(),
		Description:     inv.Description(),
		Items:           items,
		Subtotal:        inv.Pricing().Subtotal().Amount().String(),
		TaxAmount:       inv.Pricing().Tax().Amount().String(),
		Total:           inv.Pricing().Total().Amount().String(),
		Currency:        inv.Pricing().Total().Currency(),
		CryptoCurrency:  inv.CryptoCurrency().String(),
		USDTAmount:      inv.Pricing().Total().Amount().String(), // 1:1 USD to USDT for now
		Address:         address,
		Status:          inv.Status().String(),
		ExpiresAt:       expiresAt,
		CreatedAt:       inv.CreatedAt(),
		PaidAt:          inv.PaidAt(),
		Payments:        payments,
		PaymentProgress: paymentProgress,
		ReturnURL:       returnURL,
		CancelURL:       cancelURL,
		TimeRemaining:   timeRemaining,
	}
}
