package shared

import "errors"

// TransactionHash represents a blockchain transaction hash.
type TransactionHash struct {
	hash string
}

// NewTransactionHash creates a new TransactionHash.
func NewTransactionHash(hash string) (*TransactionHash, error) {
	if hash == "" {
		return nil, errors.New("transaction hash cannot be empty")
	}

	// Basic hash format validation
	if len(hash) < 32 {
		return nil, errors.New("transaction hash format is too short")
	}

	return &TransactionHash{
		hash: hash,
	}, nil
}

// Hash returns the transaction hash.
func (th *TransactionHash) Hash() string {
	return th.hash
}

// String returns the string representation of the transaction hash.
func (th *TransactionHash) String() string {
	return th.hash
}

// Equals returns true if this transaction hash equals the other.
func (th *TransactionHash) Equals(other *TransactionHash) bool {
	if other == nil {
		return false
	}
	return th.hash == other.hash
}
