package web

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"go.uber.org/zap"

	"crypto-checkout/internal/domain/invoice"
)

const (
	// QR code generation constants.
	qrCodeWidth     = 8
	qrCodeBorder    = 2
	taxRateDecimals = 2
)

// createInvoice handles POST /api/v1/invoices requests.
func (h *Handler) createInvoice(c *gin.Context) {
	h.logger.Debug("createInvoice handler called",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("remote_addr", c.ClientIP()),
	)
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind create invoice request", zap.Error(err))
		// Wrap Gin binding errors with service error type for proper HTTP status mapping
		wrappedErr := fmt.Errorf("%w: %w", invoice.ErrInvalidRequest, err)
		if err := c.Error(wrappedErr); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	// Additional validation
	if err := validateCreateInvoiceRequest(req); err != nil {
		h.logger.Error("Invalid create invoice request", zap.Error(err))
		if err := c.Error(err); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	// Convert API request to service request
	serviceReq := convertToServiceCreateInvoiceRequest(req)

	invoice, err := h.invoiceService.CreateInvoice(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to create invoice", zap.Error(err))
		h.logger.Debug("Adding error to context", zap.Error(err))
		if err := c.Error(err); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		h.logger.Debug("Error added to context, error count", zap.Int("count", len(c.Errors)))
		return
	}

	response := ToCreateInvoiceResponse(invoice)

	// Generate the invoice URL for the user
	response.InvoiceURL = "/api/v1/invoices/" + invoice.ID()

	c.JSON(http.StatusCreated, response)
}

// convertToServiceCreateInvoiceRequest converts API request to service request.
func convertToServiceCreateInvoiceRequest(req CreateInvoiceRequest) invoice.CreateInvoiceRequest {
	items := make([]invoice.CreateInvoiceItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = invoice.CreateInvoiceItemRequest{
			Description: item.Description,
			UnitPrice:   item.UnitPrice,
			Quantity:    item.Quantity,
		}
	}

	return invoice.CreateInvoiceRequest{
		Items:   items,
		TaxRate: req.TaxRate,
	}
}

// validateCreateInvoiceRequest performs additional validation on the request.
func validateCreateInvoiceRequest(req CreateInvoiceRequest) error {
	// Validate tax rate is not negative
	if req.TaxRate == "" {
		return fmt.Errorf("%w: tax rate is required", invoice.ErrInvalidRequest)
	}

	// Parse tax rate to check if it's negative
	taxRate, err := decimal.NewFromString(req.TaxRate)
	if err != nil {
		return fmt.Errorf("%w: invalid tax rate format", invoice.ErrInvalidRequest)
	}

	if taxRate.IsNegative() {
		return fmt.Errorf("%w: tax rate cannot be negative", invoice.ErrInvalidRequest)
	}

	return nil
}

// getInvoice handles GET /api/v1/invoices/:id requests.
func (h *Handler) getInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		if err := c.Error(errors.New("invalid invoice ID")); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get invoice", zap.Error(err), zap.String("invoice_id", id))
		if err := c.Error(err); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	// Convert invoice to DTO for JSON response
	response := ToCreateInvoiceResponse(invoice)
	c.JSON(http.StatusOK, response)
}

// GenerateQRCodeImage generates a QR code image for the given content and returns the image data.
func (h *Handler) GenerateQRCodeImage(content string) ([]byte, error) {
	// Generate QR code
	qrc, err := qrcode.New(content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Create a temporary file to hold the QR code image
	tmpFile, err := os.CreateTemp("", "qrcode-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up temp file
	defer tmpFile.Close()

	// Create writer that writes to temp file
	writer, err := standard.New(tmpFile.Name(),
		standard.WithQRWidth(qrCodeWidth),
		standard.WithBorderWidth(qrCodeBorder, qrCodeBorder, qrCodeBorder, qrCodeBorder),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code writer: %w", err)
	}

	// Save QR code to temp file
	if saveErr := qrc.Save(writer); saveErr != nil {
		return nil, fmt.Errorf("failed to save QR code to temp file: %w", saveErr)
	}

	// Read the temp file and return data
	fileData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read QR code file: %w", err)
	}

	return fileData, nil
}

// getInvoiceQR handles GET /invoices/:id/qr requests.
func (h *Handler) getInvoiceQR(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		if err := c.Error(errors.New("invalid invoice ID")); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get invoice for QR code", zap.Error(err), zap.String("invoice_id", id))
		if err := c.Error(err); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	// Generate QR code content - Tron USDT payment URI
	paymentAddress := invoice.PaymentAddress()
	if paymentAddress == nil {
		if err := c.Error(errors.New("invoice has no payment address assigned")); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	// Create Tron USDT payment URI
	qrContent := fmt.Sprintf("tron:%s?amount=%s&token=USDT",
		paymentAddress.String(),
		invoice.CalculateTotal().String())

	// Generate QR code image
	imageData, err := h.GenerateQRCodeImage(qrContent)
	if err != nil {
		h.logger.Error("Failed to generate QR code image", zap.Error(err))
		if err := c.Error(errors.New("failed to generate QR code image")); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	// Set response headers for PNG image
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=300") // Cache for 5 minutes

	// Write the image data to response
	c.Data(http.StatusOK, "image/png", imageData)
}

// getPublicInvoice handles GET /invoices/:id requests (public invoice page).
func (h *Handler) getPublicInvoice(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		if err := c.Error(errors.New("invalid invoice ID")); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	invoice, err := h.invoiceService.GetInvoice(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get invoice", zap.Error(err), zap.String("invoice_id", id))
		if err := c.Error(err); err != nil {
			h.logger.Error("Failed to set error in context", zap.Error(err))
		}
		return
	}

	// Mark invoice as viewed (created → pending transition)
	if markErr := h.invoiceService.MarkInvoiceAsViewed(c.Request.Context(), id); markErr != nil {
		h.logger.Warn("Failed to mark invoice as viewed", zap.Error(markErr), zap.String("invoice_id", id))
		// Don't fail the request, just log the warning
	}

	// Prepare template data with real invoice data
	templateData := gin.H{
		"Invoice":        invoice,
		"Title":          "Invoice #" + invoice.ID(),
		"QRCodeURL":      fmt.Sprintf("/invoices/%s/qr", invoice.ID()),
		"PaymentAddress": invoice.PaymentAddress(),
		"TotalAmount":    invoice.CalculateTotal().String(),
		"SubtotalAmount": invoice.CalculateSubtotal().String(),
		"TaxAmount":      invoice.CalculateTax().String(),
		"TaxRate":        invoice.TaxRate().StringFixed(taxRateDecimals),
	}

	// Use Gin's HTML rendering
	c.HTML(http.StatusOK, "crypto_invoice_page.html", templateData)
}
