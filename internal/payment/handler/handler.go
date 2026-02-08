package handler

import (
	"context"

	"github.com/fekuna/omnipos-payment-service/internal/payment/usecase"
	"github.com/fekuna/omnipos-pkg/logger"
	paymentv1 "github.com/fekuna/omnipos-proto/proto/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type PaymentHandler struct {
	paymentv1.UnimplementedPaymentServiceServer
	useCase usecase.UseCase
	logger  logger.ZapLogger
}

func NewPaymentHandler(useCase usecase.UseCase, logger logger.ZapLogger) *PaymentHandler {
	return &PaymentHandler{
		useCase: useCase,
		logger:  logger,
	}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	merchantID := getMerchantID(ctx)
	if merchantID == "" {
		h.logger.Error("Missing merchant_id in context")
		return nil, status.Error(codes.Unauthenticated, "missing merchant_id")
	}

	payment, err := h.useCase.Create(ctx, merchantID, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &paymentv1.CreatePaymentResponse{Payment: payment}, nil
}

func (h *PaymentHandler) GetPayment(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	payment, err := h.useCase.Get(ctx, req)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &paymentv1.GetPaymentResponse{Payment: payment}, nil
}

func (h *PaymentHandler) ListPayments(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	resp, err := h.useCase.List(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

// Helper to extract merchant_id from metadata
func getMerchantID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	// Gateway usually passes x-merchant-id
	if vals := md.Get("x-merchant-id"); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
