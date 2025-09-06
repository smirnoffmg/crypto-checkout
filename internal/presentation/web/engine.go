package web

import (
	"crypto-checkout/pkg/config"
	"fmt"
	"html/template"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewGinEngine creates a new Gin engine with appropriate configuration.
func NewGinEngine(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	// Set Gin mode based on configuration
	if cfg.Log.Level == DebugLogLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Set up logging for Gin (stdout only, no file logging)
	setupGinLogging(cfg, logger)

	router := gin.New()
	router.Use(gin.Logger())

	// Load HTML templates using Go's embed package
	// This embeds the templates directly into the binary, making them available
	// regardless of working directory or file system layout
	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*"))
	router.SetHTMLTemplate(tmpl)
	logger.Info("HTML templates loaded from embedded filesystem", zap.String("pattern", "templates/*"))

	// Use custom recovery that integrates with our error handling
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		// Convert the recovered value to an error
		var err error
		if recoveredErr, ok := recovered.(error); ok {
			err = recoveredErr
		} else {
			err = fmt.Errorf("panic: %v", recovered)
		}

		// Add the error to the context so our ErrorHandler can process it
		if err := c.Error(err); err != nil {
			// Log the error but don't panic - we're already in a panic recovery
			// Use os.Stderr since we can't use logger in panic recovery
			_, _ = fmt.Fprintf(os.Stderr, "Failed to set error in context: %v\n", err)
		}

		// Let the ErrorHandler middleware handle the response
		c.Next()
	}))

	// Add custom error handling middleware
	router.Use(ErrorHandler(cfg, logger))

	return router
}
