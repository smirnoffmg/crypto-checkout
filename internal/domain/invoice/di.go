package invoice

import (
	"go.uber.org/fx"
)

// Module provides the invoice service layer dependencies.
var Module = fx.Module("invoice-service",
	fx.Provide(
		fx.Annotate(
			NewInvoiceService,
			fx.As(new(InvoiceService)),
		),
	),
)
