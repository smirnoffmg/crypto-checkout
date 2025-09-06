package payment

import (
	"go.uber.org/fx"
)

// Module provides the service layer dependencies.
var Module = fx.Module("service",
	fx.Provide(
		fx.Annotate(
			NewPaymentService,
			fx.As(new(PaymentService)),
		),
	),
)
