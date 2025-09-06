package shared

// ID represents a generic identifier type.
type ID string

// String returns the string representation of the ID.
func (id ID) String() string {
	return string(id)
}

// MerchantID represents a unique merchant identifier.
type MerchantID ID

// String returns the string representation of the merchant ID.
func (id MerchantID) String() string {
	return string(id)
}

// InvoiceID represents a unique invoice identifier.
type InvoiceID ID

// String returns the string representation of the invoice ID.
func (id InvoiceID) String() string {
	return string(id)
}

// PaymentID represents a unique payment identifier.
type PaymentID ID

// String returns the string representation of the payment ID.
func (id PaymentID) String() string {
	return string(id)
}

// CustomerID represents a unique customer identifier.
type CustomerID ID

// String returns the string representation of the customer ID.
func (id CustomerID) String() string {
	return string(id)
}
