package payment

// Helper functions for common event data patterns
func createPaymentEventData(payment *Payment) map[string]interface{} {
	return map[string]interface{}{
		"payment_id":       string(payment.ID()),
		"invoice_id":       string(payment.InvoiceID()),
		"amount":           payment.Amount(),
		"transaction_hash": payment.TransactionHash().String(),
		"from_address":     payment.FromAddress(),
		"to_address":       payment.ToAddress().Address(),
		"detected_at":      payment.DetectedAt(),
		"confirmations":    payment.Confirmations().Int(),
		"block_number":     payment.BlockInfo().Number(),
	}
}
