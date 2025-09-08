// Package main is the entry point for the crypto-checkout application.
package main

import (
	"crypto-checkout/internal/application"
	"flag"
	"os"
)

func main() {
	// Parse command line flags
	healthCheck := flag.Bool("health-check", false, "Run health check and exit")
	flag.Parse()

	if *healthCheck {
		// Simple health check - just exit with 0
		os.Exit(0)
	}

	app := application.GetApp()
	app.Run()
}
