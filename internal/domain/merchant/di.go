package merchant

import (
	"go.uber.org/fx"
)

// Module provides the merchant service layer dependencies.
var Module = fx.Module("merchant-service",
	fx.Provide(
		fx.Annotate(
			NewMerchantService,
			fx.As(new(MerchantService)),
		),
		fx.Annotate(
			NewAPIKeyService,
			fx.As(new(APIKeyService)),
		),
		fx.Annotate(
			NewWebhookEndpointService,
			fx.As(new(WebhookEndpointService)),
		),
	),
)
