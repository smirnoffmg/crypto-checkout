package logging

import (
	"go.uber.org/fx"
)

// Module provides logging dependencies.
var Module = fx.Module("logging",
	fx.Provide(
		NewLogger,
	),
)
