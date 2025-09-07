# State Transition Patterns: FSM, Domain Model, Repository

## Overview

This guide explains how to properly implement state transitions using three core patterns:
- **FSM (Finite State Machine)**: Validates allowed transitions
- **Domain Model**: Encapsulates business logic and state
- **Repository Pattern**: Handles persistence with transactional integrity

## Architecture Principles

- **Single Responsibility**: Each component has one clear purpose
- **Dependency Direction**: Domain → FSM, Service → Repository → Domain
- **Transactional Boundaries**: Repository handles data consistency
- **Business Logic Isolation**: Domain model owns all business rules

## 1. Finite State Machine (FSM)

**Purpose**: Pure validation logic for state transitions

```go
type StateMachine[T comparable] struct {
    transitions map[T][]T
}

func NewStateMachine[T comparable](transitions map[T][]T) *StateMachine[T] {
    return &StateMachine[T]{transitions: transitions}
}

func (sm *StateMachine[T]) CanTransition(from, to T) bool {
    allowed, exists := sm.transitions[from]
    if !exists {
        return false
    }
    return slices.Contains(allowed, to)
}

func (sm *StateMachine[T]) GetAllowedTransitions(from T) []T {
    return sm.transitions[from]
}
```

### FSM Configuration Example

```go
type InvoiceStatus string

const (
    StatusPending    InvoiceStatus = "pending"
    StatusPartial    InvoiceStatus = "partial"
    StatusConfirming InvoiceStatus = "confirming"
    StatusPaid       InvoiceStatus = "paid"
    StatusExpired    InvoiceStatus = "expired"
    StatusCancelled  InvoiceStatus = "cancelled"
)

func NewInvoiceStateMachine() *StateMachine[InvoiceStatus] {
    transitions := map[InvoiceStatus][]InvoiceStatus{
        StatusPending: {
            StatusPartial, StatusConfirming, StatusExpired, StatusCancelled,
        },
        StatusPartial: {
            StatusConfirming, StatusExpired, StatusCancelled,
        },
        StatusConfirming: {
            StatusPaid, StatusExpired, StatusCancelled,
        },
        // Terminal states have no transitions
        StatusPaid:      {},
        StatusExpired:   {},
        StatusCancelled: {},
    }
    
    return NewStateMachine(transitions)
}
```

## 2. Domain Model

**Purpose**: Encapsulate business logic, state, and invariants

```go
type Invoice struct {
    // State
    id          InvoiceID
    merchantID  MerchantID
    status      InvoiceStatus
    amount      Money
    expiresAt   time.Time
    
    // Behavior dependencies
    fsm    *StateMachine[InvoiceStatus]
    events []DomainEvent
}

func NewInvoice(id InvoiceID, merchantID MerchantID, amount Money) *Invoice {
    return &Invoice{
        id:         id,
        merchantID: merchantID,
        status:     StatusPending,
        amount:     amount,
        expiresAt:  time.Now().Add(30 * time.Minute),
        fsm:        NewInvoiceStateMachine(),
        events:     make([]DomainEvent, 0),
    }
}

// Primary business method
func (i *Invoice) ChangeStatus(newStatus InvoiceStatus, reason string) error {
    // 1. FSM validation
    if !i.fsm.CanTransition(i.status, newStatus) {
        return NewInvalidTransitionError(i.status, newStatus)
    }
    
    // 2. Business rule validation
    if err := i.validateStatusChange(newStatus); err != nil {
        return err
    }
    
    // 3. Apply change
    oldStatus := i.status
    i.status = newStatus
    
    // 4. Emit domain event
    i.emitEvent(InvoiceStatusChangedEvent{
        InvoiceID: i.id,
        From:      oldStatus,
        To:        newStatus,
        Reason:    reason,
        OccurredAt: time.Now(),
    })
    
    return nil
}

// Business rule validation
func (i *Invoice) validateStatusChange(newStatus InvoiceStatus) error {
    switch newStatus {
    case StatusExpired:
        if time.Now().Before(i.expiresAt) {
            return ErrInvoiceNotExpired
        }
    case StatusPaid:
        if i.status != StatusConfirming {
            return ErrPaymentNotConfirmed
        }
    }
    return nil
}

// Getters
func (i *Invoice) ID() InvoiceID       { return i.id }
func (i *Invoice) Status() InvoiceStatus { return i.status }
func (i *Invoice) Amount() Money       { return i.amount }

// Domain events
func (i *Invoice) emitEvent(event DomainEvent) {
    i.events = append(i.events, event)
}

func (i *Invoice) GetEvents() []DomainEvent {
    return i.events
}

func (i *Invoice) ClearEvents() {
    i.events = i.events[:0]
}
```

### Domain Events

```go
type DomainEvent interface {
    EventType() string
    OccurredAt() time.Time
}

type InvoiceStatusChangedEvent struct {
    InvoiceID  InvoiceID
    From       InvoiceStatus
    To         InvoiceStatus
    Reason     string
    OccurredAt time.Time
}

func (e InvoiceStatusChangedEvent) EventType() string {
    return "invoice.status_changed"
}

func (e InvoiceStatusChangedEvent) OccurredAt() time.Time {
    return e.OccurredAt
}
```

## 3. Repository Pattern

**Purpose**: Handle persistence with transactional integrity

```go
type InvoiceRepository interface {
    FindByID(ctx context.Context, id InvoiceID) (*Invoice, error)
    Save(ctx context.Context, invoice *Invoice) error
    FindByStatus(ctx context.Context, status InvoiceStatus) ([]*Invoice, error)
}

type PostgreSQLInvoiceRepository struct {
    db DB
}

func (r *PostgreSQLInvoiceRepository) FindByID(ctx context.Context, id InvoiceID) (*Invoice, error) {
    var data InvoiceData
    
    err := r.db.QueryRowContext(ctx, `
        SELECT id, merchant_id, status, amount, expires_at
        FROM invoices 
        WHERE id = $1
    `, id).Scan(&data.ID, &data.MerchantID, &data.Status, &data.Amount, &data.ExpiresAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrInvoiceNotFound
        }
        return nil, err
    }
    
    return r.toDomain(data), nil
}

func (r *PostgreSQLInvoiceRepository) Save(ctx context.Context, invoice *Invoice) error {
    // Upsert invoice data
    _, err := r.db.ExecContext(ctx, `
        INSERT INTO invoices (id, merchant_id, status, amount, expires_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW())
        ON CONFLICT (id) 
        DO UPDATE SET 
            status = EXCLUDED.status,
            updated_at = EXCLUDED.updated_at
    `, invoice.ID(), invoice.MerchantID(), invoice.Status(), 
       invoice.Amount(), invoice.ExpiresAt())
    
    return err
}

// Data transfer object
type InvoiceData struct {
    ID         InvoiceID
    MerchantID MerchantID
    Status     InvoiceStatus
    Amount     Money
    ExpiresAt  time.Time
}

func (r *PostgreSQLInvoiceRepository) toDomain(data InvoiceData) *Invoice {
    invoice := &Invoice{
        id:         data.ID,
        merchantID: data.MerchantID,
        status:     data.Status,
        amount:     data.Amount,
        expiresAt:  data.ExpiresAt,
        fsm:        NewInvoiceStateMachine(),
        events:     make([]DomainEvent, 0),
    }
    return invoice
}
```

## 4. Application Service (Orchestration)

**Purpose**: Coordinate the flow and handle transactions

```go
type InvoiceService struct {
    repo      InvoiceRepository
    db        DB
    eventBus  EventBus
}

func (s *InvoiceService) ProcessPaymentConfirmation(ctx context.Context, invoiceID InvoiceID) error {
    return s.db.WithTx(ctx, func(tx Tx) error {
        // 1. Load aggregate
        invoice, err := s.repo.FindByID(ctx, invoiceID)
        if err != nil {
            return err
        }
        
        // 2. Apply business logic (domain model handles FSM validation)
        if err := invoice.ChangeStatus(StatusPaid, "payment_confirmed"); err != nil {
            return err
        }
        
        // 3. Persist changes
        if err := s.repo.Save(ctx, invoice); err != nil {
            return err
        }
        
        // 4. Publish domain events
        return s.publishEvents(ctx, invoice.GetEvents())
    })
}

func (s *InvoiceService) ExpireInvoice(ctx context.Context, invoiceID InvoiceID) error {
    return s.db.WithTx(ctx, func(tx Tx) error {
        invoice, err := s.repo.FindByID(ctx, invoiceID)
        if err != nil {
            return err
        }
        
        if err := invoice.ChangeStatus(StatusExpired, "timeout_reached"); err != nil {
            return err
        }
        
        if err := s.repo.Save(ctx, invoice); err != nil {
            return err
        }
        
        return s.publishEvents(ctx, invoice.GetEvents())
    })
}

func (s *InvoiceService) publishEvents(ctx context.Context, events []DomainEvent) error {
    for _, event := range events {
        if err := s.eventBus.Publish(ctx, event); err != nil {
            return err
        }
    }
    return nil
}
```

## 5. Error Handling

```go
// Domain errors
type InvalidTransitionError struct {
    From InvoiceStatus
    To   InvoiceStatus
}

func (e InvalidTransitionError) Error() string {
    return fmt.Sprintf("invalid transition from %s to %s", e.From, e.To)
}

func NewInvalidTransitionError(from, to InvoiceStatus) error {
    return InvalidTransitionError{From: from, To: to}
}

// Standard errors
var (
    ErrInvoiceNotFound      = errors.New("invoice not found")
    ErrInvoiceNotExpired    = errors.New("invoice has not expired")
    ErrPaymentNotConfirmed  = errors.New("payment not confirmed")
)
```

## 6. Testing Strategy

### Unit Tests for FSM

```go
func TestInvoiceStateMachine_CanTransition(t *testing.T) {
    fsm := NewInvoiceStateMachine()
    
    tests := []struct {
        name     string
        from     InvoiceStatus
        to       InvoiceStatus
        expected bool
    }{
        {"pending to paid", StatusPending, StatusPaid, false},
        {"pending to partial", StatusPending, StatusPartial, true},
        {"confirming to paid", StatusConfirming, StatusPaid, true},
        {"paid to cancelled", StatusPaid, StatusCancelled, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := fsm.CanTransition(tt.from, tt.to)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Unit Tests for Domain Model

```go
func TestInvoice_ChangeStatus(t *testing.T) {
    invoice := NewInvoice(InvoiceID("123"), MerchantID("456"), Money{Amount: 100})
    
    t.Run("valid transition", func(t *testing.T) {
        err := invoice.ChangeStatus(StatusPartial, "partial_payment_received")
        assert.NoError(t, err)
        assert.Equal(t, StatusPartial, invoice.Status())
        assert.Len(t, invoice.GetEvents(), 1)
    })
    
    t.Run("invalid transition", func(t *testing.T) {
        err := invoice.ChangeStatus(StatusPaid, "invalid_transition")
        assert.Error(t, err)
        assert.IsType(t, InvalidTransitionError{}, err)
    })
}
```

### Service Tests with Mocks (No Database)

```go
func TestInvoiceService_ProcessPaymentConfirmation(t *testing.T) {
    // Setup mocks - no real database needed
    mockRepo := &MockInvoiceRepository{}
    mockPublisher := &MockEventPublisher{}
    service := NewInvoiceService(mockRepo, mockPublisher)
    
    t.Run("successful payment confirmation", func(t *testing.T) {
        // Setup: invoice in confirming state
        invoice := NewInvoice("test-123", "merchant-456", NewMoney(100.0, "USD"))
        invoice.ProcessFullPayment("payment_detected")
        invoice.ClearEvents() // Clear creation events
        
        // Mock expectations
        mockRepo.On("WithinTransaction", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("func(InvoiceRepository) error")).
            Return(nil).
            Run(func(args mock.Arguments) {
                // Execute the transaction function with mock repo
                txFn := args[1].(func(InvoiceRepository) error)
                
                // Setup transaction mock behavior
                mockRepo.On("FindByID", mock.Anything, "test-123").Return(invoice, nil).Once()
                mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*Invoice")).Return(nil).Once()
                
                // Execute transaction
                err := txFn(mockRepo)
                assert.NoError(t, err)
            })
        
        mockPublisher.On("Publish", mock.Anything, mock.AnythingOfType("InvoiceStatusChangedEvent")).Return(nil)
        
        // Execute
        err := service.ProcessPaymentConfirmation(context.Background(), "test-123")
        
        // Verify
        assert.NoError(t, err)
        mockRepo.AssertExpectations(t)
        mockPublisher.AssertExpectations(t)
    })
    
    t.Run("invalid transition error", func(t *testing.T) {
        // Setup: invoice already paid
        paidInvoice := NewInvoice("test-456", "merchant-789", NewMoney(100.0, "USD"))
        // Force to paid status
        paidInvoice.fsm = NewInvoiceFSM("paid", fsm.Callbacks{})
        paidInvoice.status = StatusPaid
        
        mockRepo.On("WithinTransaction", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("func(InvoiceRepository) error")).
            Return(InvalidTransitionError{From: StatusPaid, To: StatusPaid}).
            Run(func(args mock.Arguments) {
                txFn := args[1].(func(InvoiceRepository) error)
                mockRepo.On("FindByID", mock.Anything, "test-456").Return(paidInvoice, nil).Once()
                err := txFn(mockRepo)
                assert.Error(t, err)
            })
        
        // Execute
        err := service.ProcessPaymentConfirmation(context.Background(), "test-456")
        
        // Verify
        assert.Error(t, err)
        assert.IsType(t, InvalidTransitionError{}, err)
        mockRepo.AssertExpectations(t)
        // Publisher should not be called
        mockPublisher.AssertNotCalled(t, "Publish")
    })
}

// Mock implementations
type MockInvoiceRepository struct {
    mock.Mock
}

func (m *MockInvoiceRepository) FindByID(ctx context.Context, id string) (*Invoice, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) Save(ctx context.Context, invoice *Invoice) error {
    args := m.Called(ctx, invoice)
    return args.Error(0)
}

func (m *MockInvoiceRepository) FindByStatus(ctx context.Context, status InvoiceStatus) ([]*Invoice, error) {
    args := m.Called(ctx, status)
    return args.Get(0).([]*Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) FindExpiredInvoices(ctx context.Context) ([]*Invoice, error) {
    args := m.Called(ctx)
    return args.Get(0).([]*Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) WithinTransaction(ctx context.Context, fn func(repo InvoiceRepository) error) error {
    args := m.Called(ctx, fn)
    return args.Error(0)
}
```
```

## Key Principles Summary

1. **FSM (looplab/fsm)**: Event-driven transitions with callbacks, validates and triggers state changes
2. **Domain Model**: Owns business logic, uses FSM for transitions, emits domain events
3. **Repository (GORM)**: Handles persistence with transactions, converts between domain and data models
4. **Kafka Publisher**: Provides append-only event log for audit trail and integration
5. **Service Layer**: Orchestrates flow with GORM transactions and reliable event publishing

## Flow Sequence

```
1. Service.Method(id)
2. GORM.Transaction.Begin()
3. Repository.FindByID(id) → InvoiceModel → Domain Object
4. Domain.BusinessMethod() → FSM.Event() → Callbacks → State Change → Emit Events
5. Repository.Save(domain) → InvoiceModel → GORM.Save()
6. GORM.Transaction.Commit()
7. KafkaPublisher.Publish(events) → Append-only Event Log
```

## Technology Integration Benefits

**looplab/fsm**:
- Event-driven transitions with clear naming
- Built-in callback system for validation and side effects
- Thread-safe state management
- Detailed error reporting

**GORM**:
- Automatic transaction management
- Model-based approach with clear separation
- Built-in upsert capabilities with Save()
- Connection pooling and optimization

**Kafka**:
- Guaranteed event ordering per aggregate (partitioned by ID)
- Append-only audit trail for compliance
- Horizontal scaling for high-throughput events
- Reliable delivery with acknowledgments

This pattern ensures:
- **State Consistency**: FSM prevents invalid transitions within domain logic
- **Data Consistency**: GORM transactions maintain database integrity  
- **Event Consistency**: Kafka provides reliable, ordered event delivery
- **Auditability**: Complete append-only log of all state changes
- **Testability**: Each component can be tested independently with mocks