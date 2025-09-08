// Package web provides HTTP handlers for the crypto-checkout API.
// @title Crypto Checkout API
// @version 1.0
// @description A comprehensive cryptocurrency payment processing API for merchants and customers.
// @termsOfService https://thecryptocheckout.com/terms

// @contact.name API Support
// @contact.url https://thecryptocheckout.com/support
// @contact.email support@thecryptocheckout.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description API Key authentication. Use format: Bearer sk_live_xxx or Bearer sk_test_xxx

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token authentication. Use format: Bearer <jwt_token>

package web
