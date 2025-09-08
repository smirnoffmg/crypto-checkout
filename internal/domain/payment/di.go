package payment

import (
	"go.uber.org/fx"
)

// Module provides the payment service layer dependencies.
var Module = fx.Module("payment-service",
	fx.Provide(
		fx.Annotate(
			NewPaymentService,
			fx.As(new(PaymentService)),
		),
	),
)
