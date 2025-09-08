package merchant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// MerchantServiceImpl implements the MerchantService interface.
type MerchantServiceImpl struct {
	merchantRepo MerchantRepository
	logger       *zap.Logger
}

// NewMerchantService creates a new merchant service.
func NewMerchantService(merchantRepo MerchantRepository, logger *zap.Logger) MerchantService {
	return &MerchantServiceImpl{
		merchantRepo: merchantRepo,
		logger:       logger,
	}
}

// CreateMerchant creates a new merchant account.
func (s *MerchantServiceImpl) CreateMerchant(
	ctx context.Context,
	req *CreateMerchantRequest,
) (*CreateMerchantResponse, error) {
	if req == nil {
		return nil, errors.New("create merchant request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if merchant with email already exists
	existingMerchant, err := s.merchantRepo.FindByEmail(ctx, req.ContactEmail)
	if err == nil && existingMerchant != nil {
		return nil, errors.New("merchant with this email already exists")
	}

	// Generate merchant ID
	merchantID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate merchant ID: %w", err)
	}

	// Create merchant
	merchant, err := NewMerchant(
		merchantID,
		req.BusinessName,
		req.ContactEmail,
		req.Settings,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create merchant: %w", err)
	}

	// Save to repository
	if err := s.merchantRepo.Save(ctx, merchant); err != nil {
		return nil, fmt.Errorf("failed to save merchant: %w", err)
	}

	s.logger.Info("Merchant created successfully",
		zap.String("merchant_id", merchant.ID()),
		zap.String("business_name", merchant.BusinessName()),
		zap.String("contact_email", merchant.ContactEmail()),
	)

	return &CreateMerchantResponse{
		Merchant: merchant,
	}, nil
}

// GetMerchant retrieves a merchant by ID.
func (s *MerchantServiceImpl) GetMerchant(ctx context.Context, req *GetMerchantRequest) (*GetMerchantResponse, error) {
	if req == nil {
		return nil, errors.New("get merchant request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Find merchant
	merchant, err := s.merchantRepo.FindByID(ctx, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find merchant: %w", err)
	}

	return &GetMerchantResponse{
		Merchant: merchant,
	}, nil
}

// UpdateMerchant updates an existing merchant.
func (s *MerchantServiceImpl) UpdateMerchant(
	ctx context.Context,
	req *UpdateMerchantRequest,
) (*UpdateMerchantResponse, error) {
	if req == nil {
		return nil, errors.New("update merchant request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Find existing merchant
	merchant, err := s.merchantRepo.FindByID(ctx, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find merchant: %w", err)
	}

	// Update fields if provided
	if req.BusinessName != nil {
		if err := merchant.UpdateBusinessName(*req.BusinessName); err != nil {
			return nil, fmt.Errorf("failed to update business name: %w", err)
		}
	}

	if req.ContactEmail != nil {
		// Check if email is already taken by another merchant
		existingMerchant, err := s.merchantRepo.FindByEmail(ctx, *req.ContactEmail)
		if err == nil && existingMerchant != nil && existingMerchant.ID() != merchant.ID() {
			return nil, errors.New("email is already taken by another merchant")
		}

		if err := merchant.UpdateContactEmail(*req.ContactEmail); err != nil {
			return nil, fmt.Errorf("failed to update contact email: %w", err)
		}
	}

	if req.Settings != nil {
		if err := merchant.UpdateSettings(req.Settings); err != nil {
			return nil, fmt.Errorf("failed to update settings: %w", err)
		}
	}

	// Save updated merchant
	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return nil, fmt.Errorf("failed to update merchant: %w", err)
	}

	s.logger.Info("Merchant updated successfully",
		zap.String("merchant_id", merchant.ID()),
	)

	return &UpdateMerchantResponse{
		Merchant: merchant,
	}, nil
}

// ChangeMerchantStatus changes the status of a merchant.
func (s *MerchantServiceImpl) ChangeMerchantStatus(
	ctx context.Context,
	req *ChangeMerchantStatusRequest,
) (*ChangeMerchantStatusResponse, error) {
	if req == nil {
		return nil, errors.New("change merchant status request cannot be nil")
	}

	// Validate request
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Find existing merchant
	merchant, err := s.merchantRepo.FindByID(ctx, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find merchant: %w", err)
	}

	// Change status
	if err := merchant.ChangeStatus(req.Status); err != nil {
		return nil, fmt.Errorf("failed to change merchant status: %w", err)
	}

	// Save updated merchant
	if err := s.merchantRepo.Update(ctx, merchant); err != nil {
		return nil, fmt.Errorf("failed to update merchant: %w", err)
	}

	s.logger.Info("Merchant status changed successfully",
		zap.String("merchant_id", merchant.ID()),
		zap.String("new_status", string(req.Status)),
		zap.String("reason", req.Reason),
	)

	return &ChangeMerchantStatusResponse{
		Merchant: merchant,
	}, nil
}

// ListMerchants lists merchants with filtering and pagination.
func (s *MerchantServiceImpl) ListMerchants(
	ctx context.Context,
	req *ListMerchantsRequest,
) (*ListMerchantsResponse, error) {
	if req == nil {
		return nil, errors.New("list merchants request cannot be nil")
	}

	// Set default values
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// List merchants
	response, err := s.merchantRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list merchants: %w", err)
	}

	return response, nil
}

// generateID generates a random ID.
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
