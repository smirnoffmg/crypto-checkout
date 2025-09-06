package shared

import (
	"errors"
	"fmt"
)

// ConfirmationCount represents the number of blockchain confirmations.
type ConfirmationCount struct {
	count int
}

// NewConfirmationCount creates a new ConfirmationCount.
func NewConfirmationCount(count int) (*ConfirmationCount, error) {
	if count < 0 {
		return nil, errors.New("confirmation count cannot be negative")
	}

	return &ConfirmationCount{
		count: count,
	}, nil
}

// Count returns the confirmation count.
func (cc *ConfirmationCount) Count() int {
	return cc.count
}

// Int returns the confirmation count as an integer.
func (cc *ConfirmationCount) Int() int {
	return cc.count
}

// Increment returns a new ConfirmationCount with count increased by 1.
func (cc *ConfirmationCount) Increment() *ConfirmationCount {
	return &ConfirmationCount{
		count: cc.count + 1,
	}
}

// IsZero returns true if the confirmation count is zero.
func (cc *ConfirmationCount) IsZero() bool {
	return cc.count == 0
}

// IsGreaterThan returns true if this count is greater than the other.
func (cc *ConfirmationCount) IsGreaterThan(other *ConfirmationCount) bool {
	if other == nil {
		return false
	}
	return cc.count > other.count
}

// IsGreaterThanOrEqual returns true if this count is greater than or equal to the other.
func (cc *ConfirmationCount) IsGreaterThanOrEqual(other *ConfirmationCount) bool {
	if other == nil {
		return false
	}
	return cc.count >= other.count
}

// String returns the string representation of the confirmation count.
func (cc *ConfirmationCount) String() string {
	return fmt.Sprintf("%d", cc.count)
}

// Equals returns true if this confirmation count equals the other.
func (cc *ConfirmationCount) Equals(other *ConfirmationCount) bool {
	if other == nil {
		return false
	}
	return cc.count == other.count
}
