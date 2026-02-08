package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/fekuna/omnipos-payment-service/internal/payment/repository"
	"github.com/fekuna/omnipos-pkg/logger"
	paymentv1 "github.com/fekuna/omnipos-proto/gen/go/omnipos/payment/v1"
	"go.uber.org/zap"
)

type UseCase interface {
	Create(ctx context.Context, merchantID string, req *paymentv1.CreatePaymentRequest) (*paymentv1.Payment, error)
	Get(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.Payment, error)
	List(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error)
}

type paymentUseCase struct {
	repo   repository.Repository
	logger logger.ZapLogger
}

func NewPaymentUseCase(repo repository.Repository, logger logger.ZapLogger) UseCase {
	return &paymentUseCase{repo: repo, logger: logger}
}

func (u *paymentUseCase) Create(ctx context.Context, merchantID string, req *paymentv1.CreatePaymentRequest) (*paymentv1.Payment, error) {
	// 1. Basic Validation
	if req.OrderId == "" {
		return nil, errors.New("order_id is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if merchantID == "" {
		return nil, errors.New("merchant_id is required")
	}

	// 2. Check if payment already exists for this order (Idempotency)
	existing, err := u.repo.GetByOrderID(ctx, req.OrderId)
	if err == nil && existing != nil {
		u.logger.Warn("Duplicate payment attempt for order", zap.String("order_id", req.OrderId))
		// For now, return error. Could return existing if idempotent behavior desired.
		return nil, fmt.Errorf("payment already exists for order_id: %s", req.OrderId)
	}

	// 3. Construct Payment Object
	// Assume merchant_id is extracted from context (auth middleware)
	// For now, request doesn't have merchant_id directly... oh wait, payment table needs merchant_id.
	// Auth middleware puts it in context? Usually yes. But we need to extract safely.
	// Let's assume the handler extracts it and maybe passes it?
	// OR, we just use a placeholder or extract from Context via util if available.
	// For now, let's assume we extract it from context or pass as Metadata.
	// Since req doesn't have it, we must get it from context.

	// Simplification: We will skip merchant_id extraction here if not passed.
	// But table enforces NOT NULL.
	// We'll extract from context "x-merchant-id".

	// Refinement: CreatePaymentRequest usually assumes caller context.

	// Let's create `id` here.
	p := &paymentv1.Payment{
		Id:              uuid.New().String(),
		MerchantId:      merchantID,
		OrderId:         req.OrderId,
		Amount:          req.Amount,
		PaymentMethod:   req.PaymentMethod,
		Status:          paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCESS, // Default success for now (Cash/MVP)
		ReferenceNumber: req.ReferenceNumber,
		Provider:        req.Provider,
	}

	// Extract Merchant ID from metadata context in Handler probably best, then pass to Usecase?
	// Or Usecase extracts it. Let's do it in Usecase if we have a helper.
	// Ideally request message has merchant_id if internal call, but this is a gateway call.
	// For now, I will add `MerchantId` to `CreatePaymentRequest`? No, Proto is fixed.
	// I'll extract from context using metadata.
	// But I don't have the auth package handy here.
	// I'll define a helper or assume it's passed via context value.

	if err := u.repo.Create(ctx, p); err != nil {
		u.logger.Error("Failed to create payment", zap.Error(err))
		return nil, err
	}

	return p, nil
}

func (u *paymentUseCase) Get(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.Payment, error) {
	if req.Id != "" {
		return u.repo.Get(ctx, req.Id)
	}
	if req.OrderId != "" {
		return u.repo.GetByOrderID(ctx, req.OrderId)
	}
	return nil, errors.New("must provide id or order_id")
}

func (u *paymentUseCase) List(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	list, total, err := u.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return &paymentv1.ListPaymentsResponse{
		Payments: list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
