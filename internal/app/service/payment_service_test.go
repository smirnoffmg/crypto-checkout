package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/app/service"
	"crypto-checkout/internal/domain/payment"
)

// mockPaymentRepository is a mock implementation of payment.Repository.
type mockPaymentRepository struct {
	mock.Mock
}

func (m *mockPaymentRepository) Save(ctx context.Context, payment *payment.Payment) error {
	args := m.Called(ctx, payment)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock save error: %w", err)
	}
	return nil
}

func (m *mockPaymentRepository) FindByID(ctx context.Context, id string) (*payment.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("mock find by id error: %w", err)
		}
		return nil, payment.ErrPaymentNotFound
	}
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock find by id error: %w", err)
	}
	if p, ok := args.Get(0).(*payment.Payment); ok {
		return p, nil
	}
	return nil, errors.New("mock find by id: invalid type returned")
}

func (m *mockPaymentRepository) FindByTransactionHash(
	ctx context.Context,
	hash *payment.TransactionHash,
) (*payment.Payment, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("mock find by transaction hash error: %w", err)
		}
		return nil, payment.ErrPaymentNotFound
	}
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock find by transaction hash error: %w", err)
	}
	if p, ok := args.Get(0).(*payment.Payment); ok {
		return p, nil
	}
	return nil, errors.New("mock find by transaction hash: invalid type returned")
}

func (m *mockPaymentRepository) FindByAddress(
	ctx context.Context,
	address *payment.Address,
) ([]*payment.Payment, error) {
	args := m.Called(ctx, address)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock find by address error: %w", err)
	}
	if payments, ok := args.Get(0).([]*payment.Payment); ok {
		return payments, nil
	}
	return []*payment.Payment{}, nil
}

func (m *mockPaymentRepository) FindByStatus(
	ctx context.Context,
	status payment.PaymentStatus,
) ([]*payment.Payment, error) {
	args := m.Called(ctx, status)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock find by status error: %w", err)
	}
	if payments, ok := args.Get(0).([]*payment.Payment); ok {
		return payments, nil
	}
	return []*payment.Payment{}, nil
}

func (m *mockPaymentRepository) FindPending(ctx context.Context) ([]*payment.Payment, error) {
	args := m.Called(ctx)
	if err := args.Error(0); err != nil {
		return nil, fmt.Errorf("mock find pending error: %w", err)
	}
	if payments, ok := args.Get(0).([]*payment.Payment); ok {
		return payments, nil
	}
	return []*payment.Payment{}, nil
}

func (m *mockPaymentRepository) FindConfirmed(ctx context.Context) ([]*payment.Payment, error) {
	args := m.Called(ctx)
	if err := args.Error(0); err != nil {
		return nil, fmt.Errorf("mock find confirmed error: %w", err)
	}
	if payments, ok := args.Get(0).([]*payment.Payment); ok {
		return payments, nil
	}
	return []*payment.Payment{}, nil
}

func (m *mockPaymentRepository) FindFailed(ctx context.Context) ([]*payment.Payment, error) {
	args := m.Called(ctx)
	if err := args.Error(0); err != nil {
		return nil, fmt.Errorf("mock find failed error: %w", err)
	}
	if payments, ok := args.Get(0).([]*payment.Payment); ok {
		return payments, nil
	}
	return []*payment.Payment{}, nil
}

func (m *mockPaymentRepository) FindOrphaned(ctx context.Context) ([]*payment.Payment, error) {
	args := m.Called(ctx)
	if err := args.Error(0); err != nil {
		return nil, fmt.Errorf("mock find orphaned error: %w", err)
	}
	if payments, ok := args.Get(0).([]*payment.Payment); ok {
		return payments, nil
	}
	return []*payment.Payment{}, nil
}

func (m *mockPaymentRepository) Update(ctx context.Context, payment *payment.Payment) error {
	args := m.Called(ctx, payment)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock update error: %w", err)
	}
	return nil
}

func (m *mockPaymentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock delete error: %w", err)
	}
	return nil
}

func (m *mockPaymentRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	if err := args.Error(1); err != nil {
		return false, fmt.Errorf("mock exists error: %w", err)
	}
	return args.Bool(0), nil
}

func (m *mockPaymentRepository) CountByStatus(ctx context.Context) (map[payment.PaymentStatus]int, error) {
	args := m.Called(ctx)
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock count by status error: %w", err)
	}
	if stats, ok := args.Get(0).(map[payment.PaymentStatus]int); ok {
		return stats, nil
	}
	return map[payment.PaymentStatus]int{}, nil
}

// Helper function to create a test payment.
func createTestPayment() *payment.Payment {
	amount, _ := payment.NewUSDTAmount("100.50")
	address, _ := payment.NewPaymentAddress("TTestAddress123456789012345678901234567890")
	hash, _ := payment.NewTransactionHash("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	paymentEntity, _ := payment.NewPayment("test-payment-id", amount, address, hash)
	return paymentEntity
}

// Helper function to create a test payment in detected state.
func createTestPaymentInDetectedState() *payment.Payment {
	// Payments start in detected state by default
	return createTestPayment()
}

// Helper function to create a test payment in confirming state.
func createTestPaymentInConfirmingState() *payment.Payment {
	p := createTestPayment()
	_ = p.TransitionToConfirming(context.Background())
	return p
}

// Helper function to create a test payment in confirmed state.
func createTestPaymentInConfirmedState() *payment.Payment {
	p := createTestPayment()
	_ = p.TransitionToConfirming(context.Background())
	_ = p.TransitionToConfirmed(context.Background())
	return p
}

func TestPaymentService_CreatePayment(t *testing.T) {
	t.Parallel()
	runCreatePaymentTests(t)
}

func runCreatePaymentTests(t *testing.T) {
	validTests := getValidCreatePaymentTests()
	invalidTests := getInvalidCreatePaymentTests()

	allTests := make([]struct {
		name        string
		request     service.CreatePaymentRequest
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}, 0, len(validTests)+len(invalidTests))
	allTests = append(allTests, validTests...)
	allTests = append(allTests, invalidTests...)

	for _, tt := range allTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := &mockPaymentRepository{}
			tt.setupMock(mockRepo)

			paymentService := service.NewPaymentService(mockRepo)
			result, err := paymentService.CreatePayment(context.Background(), tt.request)

			if tt.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errorType)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.ID())
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func getValidCreatePaymentTests() []struct {
	name        string
	request     service.CreatePaymentRequest
	setupMock   func(*mockPaymentRepository)
	expectError bool
	errorType   error
} {
	return []struct {
		name        string
		request     service.CreatePaymentRequest
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}{
		{
			name: "successful payment creation",
			request: service.CreatePaymentRequest{
				Amount:          "100.50",
				Address:         "TTestAddress123456789012345678901234567890",
				TransactionHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			setupMock: func(m *mockPaymentRepository) {
				m.On("Save", mock.Anything, mock.AnythingOfType("*payment.Payment")).Return(nil)
			},
			expectError: false,
		},
	}
}

func getInvalidCreatePaymentTests() []struct {
	name        string
	request     service.CreatePaymentRequest
	setupMock   func(*mockPaymentRepository)
	expectError bool
	errorType   error
} {
	validationTests := getValidationErrorTests()
	repositoryTests := getRepositoryErrorTests()

	allTests := make([]struct {
		name        string
		request     service.CreatePaymentRequest
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}, 0, len(validationTests)+len(repositoryTests))
	allTests = append(allTests, validationTests...)
	allTests = append(allTests, repositoryTests...)

	return allTests
}

func getValidationErrorTests() []struct {
	name        string
	request     service.CreatePaymentRequest
	setupMock   func(*mockPaymentRepository)
	expectError bool
	errorType   error
} {
	emptyFieldTests := getEmptyFieldTests()
	invalidFormatTests := getInvalidFormatTests()

	allTests := make([]struct {
		name        string
		request     service.CreatePaymentRequest
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}, 0, len(emptyFieldTests)+len(invalidFormatTests))
	allTests = append(allTests, emptyFieldTests...)
	allTests = append(allTests, invalidFormatTests...)

	return allTests
}

func getEmptyFieldTests() []struct {
	name        string
	request     service.CreatePaymentRequest
	setupMock   func(*mockPaymentRepository)
	expectError bool
	errorType   error
} {
	return []struct {
		name        string
		request     service.CreatePaymentRequest
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}{
		{
			name: "empty amount",
			request: service.CreatePaymentRequest{
				Amount:          "",
				Address:         "TTestAddress123456789012345678901234567890",
				TransactionHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidRequest,
		},
		{
			name: "empty address",
			request: service.CreatePaymentRequest{
				Amount:          "100.50",
				Address:         "",
				TransactionHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidRequest,
		},
		{
			name: "empty transaction hash",
			request: service.CreatePaymentRequest{
				Amount:          "100.50",
				Address:         "TTestAddress123456789012345678901234567890",
				TransactionHash: "",
			},
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidRequest,
		},
	}
}

func getInvalidFormatTests() []struct {
	name        string
	request     service.CreatePaymentRequest
	setupMock   func(*mockPaymentRepository)
	expectError bool
	errorType   error
} {
	return []struct {
		name        string
		request     service.CreatePaymentRequest
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}{
		{
			name: "invalid amount format",
			request: service.CreatePaymentRequest{
				Amount:          "invalid",
				Address:         "TTestAddress123456789012345678901234567890",
				TransactionHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidAmount,
		},
		{
			name: "negative amount",
			request: service.CreatePaymentRequest{
				Amount:          "-100.50",
				Address:         "TTestAddress123456789012345678901234567890",
				TransactionHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidAmount,
		},
		{
			name: "invalid address",
			request: service.CreatePaymentRequest{
				Amount:          "100.50",
				Address:         "invalid-address",
				TransactionHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidPaymentAddress,
		},
		{
			name: "invalid transaction hash",
			request: service.CreatePaymentRequest{
				Amount:          "100.50",
				Address:         "TTestAddress123456789012345678901234567890",
				TransactionHash: "invalid-hash",
			},
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidTransactionHash,
		},
	}
}

func getRepositoryErrorTests() []struct {
	name        string
	request     service.CreatePaymentRequest
	setupMock   func(*mockPaymentRepository)
	expectError bool
	errorType   error
} {
	return []struct {
		name        string
		request     service.CreatePaymentRequest
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}{
		{
			name: "repository save error",
			request: service.CreatePaymentRequest{
				Amount:          "100.50",
				Address:         "TTestAddress123456789012345678901234567890",
				TransactionHash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			},
			setupMock: func(m *mockPaymentRepository) {
				m.On("Save", mock.Anything, mock.AnythingOfType("*payment.Payment")).Return(errors.New("save failed"))
			},
			expectError: true,
			errorType:   service.ErrPaymentServiceError,
		},
	}
}

func TestPaymentService_GetPayment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		paymentID   string
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}{
		{
			name:      "successful retrieval",
			paymentID: "test-payment-id",
			setupMock: func(m *mockPaymentRepository) {
				payment := createTestPayment()
				m.On("FindByID", mock.Anything, "test-payment-id").Return(payment, nil)
			},
			expectError: false,
		},
		{
			name:        "empty payment ID",
			paymentID:   "",
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidRequest,
		},
		{
			name:      "payment not found",
			paymentID: "non-existent-id",
			setupMock: func(m *mockPaymentRepository) {
				m.On("FindByID", mock.Anything, "non-existent-id").Return(nil, payment.ErrPaymentNotFound)
			},
			expectError: true,
			errorType:   service.ErrPaymentNotFound,
		},
		{
			name:      "repository error",
			paymentID: "test-payment-id",
			setupMock: func(m *mockPaymentRepository) {
				m.On("FindByID", mock.Anything, "test-payment-id").Return(nil, errors.New("repository error"))
			},
			expectError: true,
			errorType:   service.ErrPaymentServiceError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := &mockPaymentRepository{}
			tt.setupMock(mockRepo)

			service := service.NewPaymentService(mockRepo)
			result, err := service.GetPayment(context.Background(), tt.paymentID)

			if tt.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errorType)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.paymentID, result.ID())
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_GetPaymentByTransactionHash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		hash        string
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}{
		{
			name: "successful retrieval",
			hash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			setupMock: func(m *mockPaymentRepository) {
				payment := createTestPayment()
				m.On("FindByTransactionHash", mock.Anything, mock.AnythingOfType("*payment.TransactionHash")).
					Return(payment, nil)
			},
			expectError: false,
		},
		{
			name:        "empty transaction hash",
			hash:        "",
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidRequest,
		},
		{
			name:        "invalid transaction hash",
			hash:        "invalid-hash",
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidTransactionHash,
		},
		{
			name: "payment not found",
			hash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			setupMock: func(m *mockPaymentRepository) {
				m.On("FindByTransactionHash", mock.Anything, mock.AnythingOfType("*payment.TransactionHash")).
					Return(nil, payment.ErrPaymentNotFound)
			},
			expectError: true,
			errorType:   service.ErrPaymentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := &mockPaymentRepository{}
			tt.setupMock(mockRepo)

			service := service.NewPaymentService(mockRepo)
			result, err := service.GetPaymentByTransactionHash(context.Background(), tt.hash)

			if tt.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errorType)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_ListPaymentsByAddress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		address     string
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}{
		{
			name:    "successful retrieval",
			address: "TTestAddress123456789012345678901234567890",
			setupMock: func(m *mockPaymentRepository) {
				payments := []*payment.Payment{createTestPayment()}
				m.On("FindByAddress", mock.Anything, mock.AnythingOfType("*payment.Address")).Return(payments, nil)
			},
			expectError: false,
		},
		{
			name:        "empty address",
			address:     "",
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidRequest,
		},
		{
			name:        "invalid address",
			address:     "invalid-address",
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidPaymentAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := &mockPaymentRepository{}
			tt.setupMock(mockRepo)

			service := service.NewPaymentService(mockRepo)
			result, err := service.ListPaymentsByAddress(context.Background(), tt.address)

			if tt.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errorType)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, 1)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_ListPaymentsByStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		status      payment.PaymentStatus
		setupMock   func(*mockPaymentRepository)
		expectError bool
	}{
		{
			name:   "successful retrieval",
			status: payment.StatusDetected,
			setupMock: func(m *mockPaymentRepository) {
				payments := []*payment.Payment{createTestPaymentInDetectedState()}
				m.On("FindByStatus", mock.Anything, payment.StatusDetected).Return(payments, nil)
			},
			expectError: false,
		},
		{
			name:   "repository error",
			status: payment.StatusDetected,
			setupMock: func(m *mockPaymentRepository) {
				m.On("FindByStatus", mock.Anything, payment.StatusDetected).Return(nil, errors.New("repository error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := &mockPaymentRepository{}
			tt.setupMock(mockRepo)

			service := service.NewPaymentService(mockRepo)
			result, err := service.ListPaymentsByStatus(context.Background(), tt.status)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, 1)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_UpdatePaymentConfirmations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		paymentID   string
		request     service.UpdatePaymentConfirmationsRequest
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}{
		{
			name:      "successful update",
			paymentID: "test-payment-id",
			request: service.UpdatePaymentConfirmationsRequest{
				Confirmations: 5,
			},
			setupMock: func(m *mockPaymentRepository) {
				payment := createTestPaymentInDetectedState()
				m.On("FindByID", mock.Anything, "test-payment-id").Return(payment, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*payment.Payment")).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "empty payment ID",
			paymentID: "",
			request: service.UpdatePaymentConfirmationsRequest{
				Confirmations: 5,
			},
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidRequest,
		},
		{
			name:      "negative confirmations",
			paymentID: "test-payment-id",
			request: service.UpdatePaymentConfirmationsRequest{
				Confirmations: -1,
			},
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidRequest,
		},
		{
			name:      "payment not found",
			paymentID: "non-existent-id",
			request: service.UpdatePaymentConfirmationsRequest{
				Confirmations: 5,
			},
			setupMock: func(m *mockPaymentRepository) {
				m.On("FindByID", mock.Anything, "non-existent-id").Return(nil, payment.ErrPaymentNotFound)
			},
			expectError: true,
			errorType:   service.ErrPaymentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := &mockPaymentRepository{}
			tt.setupMock(mockRepo)

			service := service.NewPaymentService(mockRepo)
			err := service.UpdatePaymentConfirmations(context.Background(), tt.paymentID, tt.request)

			if tt.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_MarkPaymentAsDetected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		paymentID   string
		setupMock   func(*mockPaymentRepository)
		expectError bool
		errorType   error
	}{
		{
			name:      "successful transition from orphaned state",
			paymentID: "test-payment-id",
			setupMock: func(m *mockPaymentRepository) {
				payment := createTestPayment()
				_ = payment.TransitionToConfirming(context.Background())
				_ = payment.TransitionToOrphaned(context.Background())
				m.On("FindByID", mock.Anything, "test-payment-id").Return(payment, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*payment.Payment")).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "empty payment ID",
			paymentID:   "",
			setupMock:   func(_ *mockPaymentRepository) {},
			expectError: true,
			errorType:   service.ErrInvalidRequest,
		},
		{
			name:      "payment not found",
			paymentID: "non-existent-id",
			setupMock: func(m *mockPaymentRepository) {
				m.On("FindByID", mock.Anything, "non-existent-id").Return(nil, payment.ErrPaymentNotFound)
			},
			expectError: true,
			errorType:   service.ErrPaymentNotFound,
		},
		{
			name:      "invalid transition from confirmed state",
			paymentID: "test-payment-id",
			setupMock: func(m *mockPaymentRepository) {
				payment := createTestPaymentInConfirmedState()
				m.On("FindByID", mock.Anything, "test-payment-id").Return(payment, nil)
			},
			expectError: true,
			errorType:   service.ErrInvalidTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := &mockPaymentRepository{}
			tt.setupMock(mockRepo)

			service := service.NewPaymentService(mockRepo)
			err := service.MarkPaymentAsDetected(context.Background(), tt.paymentID)

			if tt.expectError {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errorType)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Helper function to run transition method tests with common test cases.
func runTransitionMethodTests(
	t *testing.T,
	_ string, // methodName is unused but kept for clarity
	method func(service.PaymentService, context.Context, string) error,
	validPayment func() *payment.Payment,
	invalidPayment func() *payment.Payment,
) {
	t.Run("successful transition", func(t *testing.T) {
		t.Parallel()
		mockRepo := &mockPaymentRepository{}
		payment := validPayment()
		mockRepo.On("FindByID", mock.Anything, "test-payment-id").Return(payment, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*payment.Payment")).Return(nil)

		paymentService := service.NewPaymentService(mockRepo)
		err := method(paymentService, context.Background(), "test-payment-id")

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty payment ID", func(t *testing.T) {
		t.Parallel()
		mockRepo := &mockPaymentRepository{}
		paymentService := service.NewPaymentService(mockRepo)
		err := method(paymentService, context.Background(), "")

		require.Error(t, err)
		require.ErrorIs(t, err, service.ErrInvalidRequest)
		mockRepo.AssertExpectations(t)
	})

	t.Run("payment not found", func(t *testing.T) {
		t.Parallel()
		mockRepo := &mockPaymentRepository{}
		mockRepo.On("FindByID", mock.Anything, "non-existent-id").Return(nil, payment.ErrPaymentNotFound)

		paymentService := service.NewPaymentService(mockRepo)
		err := method(paymentService, context.Background(), "non-existent-id")

		require.Error(t, err)
		require.ErrorIs(t, err, service.ErrPaymentNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid transition", func(t *testing.T) {
		t.Parallel()
		mockRepo := &mockPaymentRepository{}
		payment := invalidPayment()
		mockRepo.On("FindByID", mock.Anything, "test-payment-id").Return(payment, nil)

		paymentService := service.NewPaymentService(mockRepo)
		err := method(paymentService, context.Background(), "test-payment-id")

		require.Error(t, err)
		require.ErrorIs(t, err, service.ErrInvalidTransition)
		mockRepo.AssertExpectations(t)
	})
}

func TestPaymentService_MarkPaymentAsIncluded(t *testing.T) {
	t.Parallel()
	runTransitionMethodTests(
		t,
		"MarkPaymentAsIncluded",
		func(ps service.PaymentService, ctx context.Context, id string) error {
			return ps.MarkPaymentAsIncluded(ctx, id)
		},
		createTestPaymentInDetectedState,
		createTestPaymentInConfirmedState,
	)
}

func TestPaymentService_MarkPaymentAsConfirmed(t *testing.T) {
	t.Parallel()
	runTransitionMethodTests(
		t,
		"MarkPaymentAsConfirmed",
		func(ps service.PaymentService, ctx context.Context, id string) error {
			return ps.MarkPaymentAsConfirmed(ctx, id)
		},
		createTestPaymentInConfirmingState,
		createTestPaymentInDetectedState,
	)
}

func TestPaymentService_GetPaymentStatistics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		setupMock   func(*mockPaymentRepository)
		expectError bool
	}{
		{
			name: "successful retrieval",
			setupMock: func(m *mockPaymentRepository) {
				stats := map[payment.PaymentStatus]int{
					payment.StatusDetected:   5,
					payment.StatusConfirming: 3,
					payment.StatusConfirmed:  10,
					payment.StatusFailed:     2,
					payment.StatusOrphaned:   1,
				}
				m.On("CountByStatus", mock.Anything).Return(stats, nil)
			},
			expectError: false,
		},
		{
			name: "repository error",
			setupMock: func(m *mockPaymentRepository) {
				m.On("CountByStatus", mock.Anything).Return(nil, errors.New("repository error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := &mockPaymentRepository{}
			tt.setupMock(mockRepo)

			service := service.NewPaymentService(mockRepo)
			result, err := service.GetPaymentStatistics(context.Background())

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, 5, result[payment.StatusDetected])
				assert.Equal(t, 3, result[payment.StatusConfirming])
				assert.Equal(t, 10, result[payment.StatusConfirmed])
				assert.Equal(t, 2, result[payment.StatusFailed])
				assert.Equal(t, 1, result[payment.StatusOrphaned])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
