package repository

import (
	"context"

	paymentv1 "github.com/fekuna/omnipos-proto/gen/go/omnipos/payment/v1"
)

type Repository interface {
	Create(ctx context.Context, payment *paymentv1.Payment) error
	Get(ctx context.Context, id string) (*paymentv1.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*paymentv1.Payment, error)
	List(ctx context.Context, req *paymentv1.ListPaymentsRequest) ([]*paymentv1.Payment, int32, error)
}
