package service

import (
	"go.uber.org/fx"
)

// Module provides the service layer dependencies.
var Module = fx.Module("service", //nolint:gochecknoglobals // Required by Fx module pattern
	fx.Provide(
		fx.Annotate(
			NewInvoiceService,
			fx.As(new(InvoiceService)),
		),
		fx.Annotate(
			NewPaymentService,
			fx.As(new(PaymentService)),
		),
	),
)
