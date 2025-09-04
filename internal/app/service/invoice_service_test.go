package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"crypto-checkout/internal/app/service"
	"crypto-checkout/internal/domain/invoice"
)

// mockInvoiceRepository is a mock implementation of the invoice.Repository interface.
type mockInvoiceRepository struct {
	mock.Mock
}

func (m *mockInvoiceRepository) Save(ctx context.Context, inv *invoice.Invoice) error {
	args := m.Called(ctx, inv)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock save error: %w", err)
	}
	return nil
}

func (m *mockInvoiceRepository) FindByID(ctx context.Context, id string) (*invoice.Invoice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("mock find by id error: %w", err)
		}
		return nil, invoice.ErrInvoiceNotFound
	}
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock find by id error: %w", err)
	}
	if inv, ok := args.Get(0).(*invoice.Invoice); ok {
		return inv, nil
	}
	return nil, errors.New("mock find by id: invalid type returned")
}

func (m *mockInvoiceRepository) FindByPaymentAddress(
	ctx context.Context,
	address *invoice.PaymentAddress,
) (*invoice.Invoice, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("mock find by payment address error: %w", err)
		}
		return nil, invoice.ErrInvoiceNotFound
	}
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock find by payment address error: %w", err)
	}
	if inv, ok := args.Get(0).(*invoice.Invoice); ok {
		return inv, nil
	}
	return nil, errors.New("mock find by payment address: invalid type returned")
}

func (m *mockInvoiceRepository) FindByStatus(
	ctx context.Context,
	status invoice.InvoiceStatus,
) ([]*invoice.Invoice, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("mock find by status error: %w", err)
		}
		return []*invoice.Invoice{}, nil
	}
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock find by status error: %w", err)
	}
	if invoices, ok := args.Get(0).([]*invoice.Invoice); ok {
		return invoices, nil
	}
	return nil, errors.New("mock find by status: invalid type returned")
}

func (m *mockInvoiceRepository) FindActive(ctx context.Context) ([]*invoice.Invoice, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("mock find active error: %w", err)
		}
		return []*invoice.Invoice{}, nil
	}
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock find active error: %w", err)
	}
	if invoices, ok := args.Get(0).([]*invoice.Invoice); ok {
		return invoices, nil
	}
	return nil, errors.New("mock find active: invalid type returned")
}

func (m *mockInvoiceRepository) FindExpired(ctx context.Context) ([]*invoice.Invoice, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		if err := args.Error(1); err != nil {
			return nil, fmt.Errorf("mock find expired error: %w", err)
		}
		return []*invoice.Invoice{}, nil
	}
	if err := args.Error(1); err != nil {
		return nil, fmt.Errorf("mock find expired error: %w", err)
	}
	if invoices, ok := args.Get(0).([]*invoice.Invoice); ok {
		return invoices, nil
	}
	return nil, errors.New("mock find expired: invalid type returned")
}

func (m *mockInvoiceRepository) Update(ctx context.Context, inv *invoice.Invoice) error {
	args := m.Called(ctx, inv)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock update error: %w", err)
	}
	return nil
}

func (m *mockInvoiceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	if err := args.Error(0); err != nil {
		return fmt.Errorf("mock delete error: %w", err)
	}
	return nil
}

func (m *mockInvoiceRepository) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	if err := args.Error(1); err != nil {
		return false, fmt.Errorf("mock exists error: %w", err)
	}
	return args.Bool(0), nil
}

func TestInvoiceService_CreateInvoice_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	request := service.CreateInvoiceRequest{
		Items: []service.CreateInvoiceItemRequest{
			{
				Description: "Test Item",
				UnitPrice:   "10.50",
				Quantity:    "2",
			},
		},
		TaxRate: "0.10",
	}

	result, err := serviceImpl.CreateInvoice(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.ID())
	assert.Len(t, result.Items(), 1)

	expectedTaxRate, _ := decimal.NewFromString("0.10")
	assert.True(t, expectedTaxRate.Equal(result.TaxRate()))

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_CreateInvoice_EmptyItems(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	request := service.CreateInvoiceRequest{
		Items:   []service.CreateInvoiceItemRequest{},
		TaxRate: "0.10",
	}

	result, err := serviceImpl.CreateInvoice(ctx, request)

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvalidRequest)
	assert.Nil(t, result)
}

func TestInvoiceService_CreateInvoice_InvalidTaxRate(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	request := service.CreateInvoiceRequest{
		Items: []service.CreateInvoiceItemRequest{
			{
				Description: "Test Item",
				UnitPrice:   "10.50",
				Quantity:    "2",
			},
		},
		TaxRate: "invalid",
	}

	result, err := serviceImpl.CreateInvoice(ctx, request)

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvalidRequest)
	assert.Nil(t, result)
}

func TestInvoiceService_CreateInvoice_RepositoryError(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(errors.New("database error"))

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	request := service.CreateInvoiceRequest{
		Items: []service.CreateInvoiceItemRequest{
			{
				Description: "Test Item",
				UnitPrice:   "10.50",
				Quantity:    "2",
			},
		},
		TaxRate: "0.10",
	}

	result, err := serviceImpl.CreateInvoice(ctx, request)

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvoiceServiceError)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_GetInvoice_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoice(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	result, err := serviceImpl.GetInvoice(ctx, "test-id")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.ID())

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_GetInvoice_NotFound(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	mockRepo.On("FindByID", mock.Anything, "non-existent").Return(nil, invoice.ErrInvoiceNotFound)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	result, err := serviceImpl.GetInvoice(ctx, "non-existent")

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvoiceNotFound)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_GetInvoice_EmptyID(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	result, err := serviceImpl.GetInvoice(ctx, "")

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvalidRequest)
	assert.Nil(t, result)
}

func TestInvoiceService_GetInvoiceByPaymentAddress_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoice(t)
	mockRepo.On("FindByPaymentAddress", mock.Anything, mock.AnythingOfType("*invoice.PaymentAddress")).Return(inv, nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	result, err := serviceImpl.GetInvoiceByPaymentAddress(ctx, "TTestAddress123456789012345678901234567890")

	require.NoError(t, err)
	require.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_GetInvoiceByPaymentAddress_InvalidAddress(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	result, err := serviceImpl.GetInvoiceByPaymentAddress(ctx, "invalid-address")

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvalidPaymentAddress)
	assert.Nil(t, result)
}

func TestInvoiceService_AssignPaymentAddress_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoice(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.AssignPaymentAddress(ctx, "test-id", "TTestAddress123456789012345678901234567890")

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_AssignPaymentAddress_InvalidAddress(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoice(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.AssignPaymentAddress(ctx, "test-id", "invalid-address")

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvalidPaymentAddress)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_MarkInvoiceAsViewed_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoice(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.MarkInvoiceAsViewed(ctx, "test-id")

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_MarkInvoiceAsPartial_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoiceInPendingState(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.MarkInvoiceAsPartial(ctx, "test-id")

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_MarkInvoiceAsCompleted_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoiceInPendingState(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.MarkInvoiceAsCompleted(ctx, "test-id")

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_MarkInvoiceAsConfirmed_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoiceInConfirmingState(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.MarkInvoiceAsConfirmed(ctx, "test-id")

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_ExpireInvoice_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoice(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.ExpireInvoice(ctx, "test-id")

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_CancelInvoice_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoice(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.CancelInvoice(ctx, "test-id")

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_RefundInvoice_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoiceInPaidState(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.RefundInvoice(ctx, "test-id")

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_HandleReorg_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	inv := createTestInvoiceInConfirmingState(t)
	mockRepo.On("FindByID", mock.Anything, "test-id").Return(inv, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*invoice.Invoice")).Return(nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.HandleReorg(ctx, "test-id")

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_EmptyInvoiceID(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	err := serviceImpl.MarkInvoiceAsViewed(ctx, "")

	require.Error(t, err)
	require.ErrorIs(t, err, service.ErrInvalidRequest)
}

func TestInvoiceService_ListInvoicesByStatus_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	invoices := []*invoice.Invoice{createTestInvoice(t)}
	mockRepo.On("FindByStatus", mock.Anything, invoice.StatusPending).Return(invoices, nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	result, err := serviceImpl.ListInvoicesByStatus(ctx, invoice.StatusPending)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result, 1)

	mockRepo.AssertExpectations(t)
}

func TestInvoiceService_ListActiveInvoices_Success(t *testing.T) {
	t.Parallel()

	mockRepo := new(mockInvoiceRepository)
	invoices := []*invoice.Invoice{createTestInvoice(t)}
	mockRepo.On("FindActive", mock.Anything).Return(invoices, nil)

	serviceImpl := service.NewInvoiceService(mockRepo)
	ctx := context.Background()

	result, err := serviceImpl.ListActiveInvoices(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result, 1)

	mockRepo.AssertExpectations(t)
}

// createTestInvoice creates a test invoice for use in tests.
func createTestInvoice(t *testing.T) *invoice.Invoice {
	t.Helper()

	item, err := invoice.NewInvoiceItem(
		"Test Item",
		mustNewUSDTAmount(t, "10.50"),
		decimal.NewFromInt(2),
	)
	require.NoError(t, err)

	inv, err := invoice.NewInvoice([]*invoice.InvoiceItem{item}, decimal.NewFromFloat(0.10))
	require.NoError(t, err)

	return inv
}

// createTestInvoiceInPendingState creates a test invoice in pending state.
func createTestInvoiceInPendingState(t *testing.T) *invoice.Invoice {
	t.Helper()

	inv := createTestInvoice(t)
	err := inv.MarkAsViewed()
	require.NoError(t, err)
	return inv
}

// createTestInvoiceInConfirmingState creates a test invoice in confirming state.
func createTestInvoiceInConfirmingState(t *testing.T) *invoice.Invoice {
	t.Helper()

	inv := createTestInvoiceInPendingState(t)
	err := inv.MarkAsCompleted()
	require.NoError(t, err)
	return inv
}

// createTestInvoiceInPaidState creates a test invoice in paid state.
func createTestInvoiceInPaidState(t *testing.T) *invoice.Invoice {
	t.Helper()

	inv := createTestInvoiceInConfirmingState(t)
	err := inv.MarkAsConfirmed()
	require.NoError(t, err)
	return inv
}

// mustNewUSDTAmount creates a USDTAmount or fails the test.
func mustNewUSDTAmount(t *testing.T, amount string) *invoice.USDTAmount {
	t.Helper()

	usdtAmount, err := invoice.NewUSDTAmount(amount)
	require.NoError(t, err)
	return usdtAmount
}
