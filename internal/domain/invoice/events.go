package invoice

// Helper functions for common event data patterns
func createInvoiceEventData(invoice *Invoice) map[string]interface{} {
	cryptoAmount, _ := invoice.GetCryptoAmount()

	return map[string]interface{}{
		"invoice_id":    invoice.ID(),
		"merchant_id":   invoice.MerchantID(),
		"total_amount":  invoice.Pricing().Total(),
		"crypto_amount": cryptoAmount,
		"currency":      invoice.CryptoCurrency(),
		"status":        invoice.Status().String(),
		"expires_at":    invoice.Expiration().ExpiresAt(),
		"description":   invoice.Description(),
	}
}
