package payment

import (
	"github.com/shopspring/decimal"
)

// MustNewUSDTAmount creates a new USDTAmount from a string, panicking on error.
func MustNewUSDTAmount(value string) *USDTAmount {
	amount, err := NewUSDTAmount(value)
	if err != nil {
		panic(err)
	}
	return amount
}

// MustNewPaymentAddress creates a new PaymentAddress from a string, panicking on error.
func MustNewPaymentAddress(address string) *Address {
	addr, err := NewPaymentAddress(address)
	if err != nil {
		panic(err)
	}
	return addr
}

// MustNewTransactionHash creates a new TransactionHash from a string, panicking on error.
func MustNewTransactionHash(hash string) *TransactionHash {
	txHash, err := NewTransactionHash(hash)
	if err != nil {
		panic(err)
	}
	return txHash
}

// MustNewConfirmationCount creates a new ConfirmationCount, panicking on error.
func MustNewConfirmationCount(count int) *ConfirmationCount {
	confCount, err := NewConfirmationCount(count)
	if err != nil {
		panic(err)
	}
	return confCount
}

// MustNewDecimal creates a new decimal from a string, panicking on error.
func MustNewDecimal(value string) decimal.Decimal {
	d, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return d
}
