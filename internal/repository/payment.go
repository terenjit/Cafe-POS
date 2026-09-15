package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/terenjit/Cafe-POS/internal/entity"
)

type PaymentRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Payment, error)
	FindByOrderID(ctx context.Context, orderID string) (*entity.Payment, error)
	FindByMidtransOrderID(ctx context.Context, midtransOrderID string) (*entity.Payment, error)
	Create(ctx context.Context, payment *entity.Payment) error
	UpdateStatus(ctx context.Context, id string, status string, paidAt *time.Time) error
	UpdateMidtransData(ctx context.Context, id string, token string, url string, midtransOrderID string) error
	SaveRawNotification(ctx context.Context, id string, raw string) error
	WithTx(tx *sql.Tx) PaymentRepository
}

type paymentRepository struct {
	db sqlDB
}

func NewPaymentRepository(db sqlDB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) WithTx(tx *sql.Tx) PaymentRepository {
	return &paymentRepository{db: tx}
}

const paymentColumns = `id, order_id, method, status, amount,
	midtrans_order_id, midtrans_token, midtrans_url, raw_notification, paid_at,
	created_at, updated_at`

func scanPayment(row interface {
	Scan(...interface{}) error
}, p *entity.Payment) error {
	return row.Scan(
		&p.ID, &p.OrderID, &p.Method, &p.Status, &p.Amount,
		&p.MidtransOrderID, &p.MidtransToken, &p.MidtransURL, &p.RawNotification, &p.PaidAt,
		&p.CreatedAt, &p.UpdatedAt,
	)
}

func (r *paymentRepository) FindByID(ctx context.Context, id string) (*entity.Payment, error) {
	query := "SELECT " + paymentColumns + " FROM payments WHERE id = ?"

	p := &entity.Payment{}
	if err := scanPayment(r.db.QueryRowContext(ctx, query, id), p); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID string) (*entity.Payment, error) {
	query := "SELECT " + paymentColumns + " FROM payments WHERE order_id = ?"

	p := &entity.Payment{}
	if err := scanPayment(r.db.QueryRowContext(ctx, query, orderID), p); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *paymentRepository) FindByMidtransOrderID(ctx context.Context, midtransOrderID string) (*entity.Payment, error) {
	query := "SELECT " + paymentColumns + " FROM payments WHERE midtrans_order_id = ?"

	p := &entity.Payment{}
	if err := scanPayment(r.db.QueryRowContext(ctx, query, midtransOrderID), p); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *paymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	payment.ID = uuid.New().String()
	now := time.Now()
	payment.CreatedAt = now
	payment.UpdatedAt = now
	payment.Status = entity.PaymentStatusPending

	query := `INSERT INTO payments
	          (id, order_id, method, status, amount,
	           midtrans_order_id, midtrans_token, midtrans_url, paid_at,
	           created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		payment.ID, payment.OrderID, payment.Method, payment.Status, payment.Amount,
		payment.MidtransOrderID, payment.MidtransToken, payment.MidtransURL, payment.PaidAt,
		payment.CreatedAt, payment.UpdatedAt,
	)
	return err
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id string, status string, paidAt *time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE payments SET status=?, paid_at=?, updated_at=NOW() WHERE id=?",
		status, paidAt, id,
	)
	return err
}

func (r *paymentRepository) UpdateMidtransData(ctx context.Context, id string, token string, url string, midtransOrderID string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE payments SET midtrans_token=?, midtrans_url=?, midtrans_order_id=?, updated_at=NOW() WHERE id=?",
		token, url, midtransOrderID, id,
	)
	return err
}

func (r *paymentRepository) SaveRawNotification(ctx context.Context, id string, raw string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE payments SET raw_notification=?, updated_at=NOW() WHERE id=?",
		raw, id,
	)
	return err
}
