package invoice

import (
	"go.uber.org/fx"
)

// Module provides the service layer dependencies.
var Module = fx.Module("service",
	fx.Provide(
		fx.Annotate(
			NewInvoiceService,
			fx.As(new(InvoiceService)),
		),
	),
)
