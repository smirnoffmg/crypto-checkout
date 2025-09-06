package payment

import "errors"

var (
	ErrInvalidAmount          = errors.New("invalid amount")
	ErrInvalidPayment         = errors.New("invalid payment")
	ErrInvalidPaymentAddress  = errors.New("invalid payment address")
	ErrInvalidRequest         = errors.New("invalid request")
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
	ErrInvalidTransition      = errors.New("invalid status transition")
	ErrPaymentAlreadyExists   = errors.New("payment already exists")
	ErrPaymentNotFound        = errors.New("payment not found")
	ErrPaymentServiceError    = errors.New("payment service error")
	ErrRepositoryError        = errors.New("repository error")
)
