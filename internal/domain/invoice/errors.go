package invoice

import "errors"

// Common service errors.
var (
	ErrInvoiceServiceError   = errors.New("invoice service error")
	ErrInvalidRequest        = errors.New("invalid request")
	ErrInvoiceNotFound       = errors.New("invoice not found")
	ErrInvalidPaymentAddress = errors.New("invalid payment address")
	ErrInvalidTransition     = errors.New("invalid status transition")
	ErrInvoiceAlreadyExists  = errors.New("invoice already exists")
	ErrInvalidInvoice        = errors.New("invalid invoice")
	ErrRepositoryError       = errors.New("repository error")
)
