package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/terenjit/Cafe-POS/internal/entity"
)

type ShiftRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Shift, error)
	FindOpenByCashierID(ctx context.Context, cashierID string) (*entity.Shift, error)
	FindAll(ctx context.Context, cashierID string, page, limit int) ([]entity.Shift, int, error)
	Create(ctx context.Context, shift *entity.Shift) error
	Close(ctx context.Context, id string, closingCash int64, notes string) error
	WithTx(tx *sql.Tx) ShiftRepository
}

type shiftRepository struct {
	db sqlDB
}

func NewShiftRepository(db sqlDB) ShiftRepository {
	return &shiftRepository{db: db}
}

func (r *shiftRepository) WithTx(tx *sql.Tx) ShiftRepository {
	return &shiftRepository{db: tx}
}

func (r *shiftRepository) FindByID(ctx context.Context, id string) (*entity.Shift, error) {
	query := `
		SELECT s.id, s.cashier_id, s.opened_at, s.closed_at, s.opening_cash, s.closing_cash,
		       s.total_transactions, s.status, s.notes, s.created_at, s.updated_at,
		       u.id, u.name
		FROM shifts s
		JOIN users u ON u.id = s.cashier_id
		WHERE s.id = ?`

	s := &entity.Shift{}
	var cashier entity.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.CashierID, &s.OpenedAt, &s.ClosedAt, &s.OpeningCash, &s.ClosingCash,
		&s.TotalTransactions, &s.Status, &s.Notes, &s.CreatedAt, &s.UpdatedAt,
		&cashier.ID, &cashier.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Cashier = &cashier
	return s, nil
}

func (r *shiftRepository) FindOpenByCashierID(ctx context.Context, cashierID string) (*entity.Shift, error) {
	query := `
		SELECT s.id, s.cashier_id, s.opened_at, s.closed_at, s.opening_cash, s.closing_cash,
		       s.total_transactions, s.status, s.notes, s.created_at, s.updated_at,
		       u.id, u.name
		FROM shifts s
		JOIN users u ON u.id = s.cashier_id
		WHERE s.cashier_id = ? AND s.status = 'open'`

	s := &entity.Shift{}
	var cashier entity.User
	err := r.db.QueryRowContext(ctx, query, cashierID).Scan(
		&s.ID, &s.CashierID, &s.OpenedAt, &s.ClosedAt, &s.OpeningCash, &s.ClosingCash,
		&s.TotalTransactions, &s.Status, &s.Notes, &s.CreatedAt, &s.UpdatedAt,
		&cashier.ID, &cashier.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Cashier = &cashier
	return s, nil
}

func (r *shiftRepository) FindAll(ctx context.Context, cashierID string, page, limit int) ([]entity.Shift, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	if cashierID != "" {
		conditions = append(conditions, "s.cashier_id = ?")
		args = append(args, cashierID)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shifts s"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	dataQuery := `
		SELECT s.id, s.cashier_id, s.opened_at, s.closed_at, s.opening_cash, s.closing_cash,
		       s.total_transactions, s.status, s.notes, s.created_at, s.updated_at,
		       u.id, u.name
		FROM shifts s
		JOIN users u ON u.id = s.cashier_id` +
		where + ` ORDER BY s.opened_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, dataQuery, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	shifts := make([]entity.Shift, 0)
	for rows.Next() {
		var s entity.Shift
		var cashier entity.User
		if err := rows.Scan(
			&s.ID, &s.CashierID, &s.OpenedAt, &s.ClosedAt, &s.OpeningCash, &s.ClosingCash,
			&s.TotalTransactions, &s.Status, &s.Notes, &s.CreatedAt, &s.UpdatedAt,
			&cashier.ID, &cashier.Name,
		); err != nil {
			return nil, 0, err
		}
		s.Cashier = &cashier
		shifts = append(shifts, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return shifts, total, nil
}

func (r *shiftRepository) Create(ctx context.Context, shift *entity.Shift) error {
	shift.ID = uuid.New().String()
	now := time.Now()
	shift.OpenedAt = now
	shift.CreatedAt = now
	shift.UpdatedAt = now
	shift.Status = entity.ShiftStatusOpen

	query := `INSERT INTO shifts
	          (id, cashier_id, opened_at, opening_cash, total_transactions, status, notes, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		shift.ID, shift.CashierID, shift.OpenedAt, shift.OpeningCash,
		shift.TotalTransactions, shift.Status, shift.Notes,
		shift.CreatedAt, shift.UpdatedAt,
	)
	return err
}

func (r *shiftRepository) Close(ctx context.Context, id string, closingCash int64, notes string) error {
	query := `UPDATE shifts
	          SET status='closed', closed_at=NOW(), closing_cash=?, notes=?, updated_at=NOW()
	          WHERE id=? AND status='open'`

	_, err := r.db.ExecContext(ctx, query, closingCash, notes, id)
	return err
}
