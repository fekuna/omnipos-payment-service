package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fekuna/omnipos-pkg/logger"
	paymentv1 "github.com/fekuna/omnipos-proto/proto/payment/v1"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type postgresRepo struct {
	db     *sqlx.DB
	logger logger.ZapLogger
}

func NewPostgresRepository(db *sqlx.DB, logger logger.ZapLogger) Repository {
	return &postgresRepo{db: db, logger: logger}
}

func (r *postgresRepo) Create(ctx context.Context, p *paymentv1.Payment) error {
	query := `
		INSERT INTO payments (
			id, merchant_id, order_id, amount, payment_method, status, reference_number, provider, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	now := time.Now()
	p.CreatedAt = timestamppb.New(now)
	p.UpdatedAt = timestamppb.New(now)

	_, err := r.db.ExecContext(ctx, query,
		p.Id,
		p.MerchantId,
		p.OrderId,
		p.Amount,
		p.PaymentMethod,
		p.Status,
		p.ReferenceNumber,
		p.Provider,
		now,
		now,
	)

	return err
}

func (r *postgresRepo) Get(ctx context.Context, id string) (*paymentv1.Payment, error) {
	query := `
		SELECT id, merchant_id, order_id, amount, payment_method, status, reference_number, provider, created_at, updated_at
		FROM payments
		WHERE id = $1
	`
	var p paymentModel
	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or custom error
		}
		return nil, err
	}

	return p.toProto(), nil
}

func (r *postgresRepo) GetByOrderID(ctx context.Context, orderID string) (*paymentv1.Payment, error) {
	query := `
		SELECT id, merchant_id, order_id, amount, payment_method, status, reference_number, provider, created_at, updated_at
		FROM payments
		WHERE order_id = $1
	`
	var p paymentModel
	err := r.db.GetContext(ctx, &p, query, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return p.toProto(), nil
}

func (r *postgresRepo) List(ctx context.Context, req *paymentv1.ListPaymentsRequest) ([]*paymentv1.Payment, int32, error) {
	query := `SELECT id, merchant_id, order_id, amount, payment_method, status, reference_number, provider, created_at, updated_at FROM payments WHERE 1=1`
	countQuery := `SELECT count(*) FROM payments WHERE 1=1`
	var args []interface{}
	argID := 1

	// Filters
	if req.OrderId != "" {
		filter := fmt.Sprintf(" AND order_id = $%d", argID)
		query += filter
		countQuery += filter
		args = append(args, req.OrderId)
		argID++
	}
	if req.PaymentMethod != paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED {
		filter := fmt.Sprintf(" AND payment_method = $%d", argID)
		query += filter
		countQuery += filter
		args = append(args, req.PaymentMethod)
		argID++
	}
	if req.Status != paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		filter := fmt.Sprintf(" AND status = $%d", argID)
		query += filter
		countQuery += filter
		args = append(args, req.Status)
		argID++
	}

	// Count
	var total int32
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	offset := (req.Page - 1) * req.PageSize
	args = append(args, req.PageSize, offset)

	var models []paymentModel
	if err := r.db.SelectContext(ctx, &models, query, args...); err != nil {
		return nil, 0, err
	}

	payments := make([]*paymentv1.Payment, len(models))
	for i, m := range models {
		payments[i] = m.toProto()
	}

	return payments, total, nil
}

// Helper struct for DB scanning
type paymentModel struct {
	ID              string    `db:"id"`
	MerchantID      string    `db:"merchant_id"`
	OrderID         string    `db:"order_id"`
	Amount          float64   `db:"amount"`
	PaymentMethod   int32     `db:"payment_method"`
	Status          int32     `db:"status"`
	ReferenceNumber *string   `db:"reference_number"` // Optional handling
	Provider        *string   `db:"provider"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

func (m *paymentModel) toProto() *paymentv1.Payment {
	var ref, prov string
	if m.ReferenceNumber != nil {
		ref = *m.ReferenceNumber
	}
	if m.Provider != nil {
		prov = *m.Provider
	}

	return &paymentv1.Payment{
		Id:              m.ID,
		MerchantId:      m.MerchantID,
		OrderId:         m.OrderID,
		Amount:          m.Amount,
		PaymentMethod:   paymentv1.PaymentMethod(m.PaymentMethod),
		Status:          paymentv1.PaymentStatus(m.Status),
		ReferenceNumber: ref,
		Provider:        prov,
		CreatedAt:       timestamppb.New(m.CreatedAt),
		UpdatedAt:       timestamppb.New(m.UpdatedAt),
	}
}
