package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/terenjit/Cafe-POS/internal/entity"
)

type PromoRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Promo, error)
	FindByCode(ctx context.Context, code string) (*entity.Promo, error)
	FindAll(ctx context.Context, page, limit int) ([]entity.Promo, int, error)
	Create(ctx context.Context, promo *entity.Promo) error
	Update(ctx context.Context, promo *entity.Promo) error
	Delete(ctx context.Context, id string) error
	IncrementUsedCount(ctx context.Context, id string) error
	WithTx(tx *sql.Tx) PromoRepository
}

type promoRepository struct {
	db sqlDB
}

func NewPromoRepository(db sqlDB) PromoRepository {
	return &promoRepository{db: db}
}

func (r *promoRepository) WithTx(tx *sql.Tx) PromoRepository {
	return &promoRepository{db: tx}
}

const promoColumns = `id, name, code, type, value, min_order, max_discount, usage_limit,
	used_count, started_at, ended_at, is_active, created_at, updated_at, deleted_at`

func scanPromo(row interface {
	Scan(...interface{}) error
}, p *entity.Promo) error {
	return row.Scan(
		&p.ID, &p.Name, &p.Code, &p.Type, &p.Value, &p.MinOrder, &p.MaxDiscount, &p.UsageLimit,
		&p.UsedCount, &p.StartedAt, &p.EndedAt, &p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
	)
}

func (r *promoRepository) FindByID(ctx context.Context, id string) (*entity.Promo, error) {
	query := "SELECT " + promoColumns + " FROM promos WHERE id = ? AND deleted_at IS NULL"

	p := &entity.Promo{}
	if err := scanPromo(r.db.QueryRowContext(ctx, query, id), p); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *promoRepository) FindByCode(ctx context.Context, code string) (*entity.Promo, error) {
	query := "SELECT " + promoColumns + " FROM promos WHERE code = ? AND deleted_at IS NULL"

	p := &entity.Promo{}
	if err := scanPromo(r.db.QueryRowContext(ctx, query, code), p); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *promoRepository) FindAll(ctx context.Context, page, limit int) ([]entity.Promo, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM promos WHERE deleted_at IS NULL").Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := "SELECT " + promoColumns + " FROM promos WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?"

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	promos := make([]entity.Promo, 0)
	for rows.Next() {
		var p entity.Promo
		if err := scanPromo(rows, &p); err != nil {
			return nil, 0, err
		}
		promos = append(promos, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return promos, total, nil
}

func (r *promoRepository) Create(ctx context.Context, promo *entity.Promo) error {
	promo.ID = uuid.New().String()
	now := time.Now()
	promo.CreatedAt = now
	promo.UpdatedAt = now

	query := `INSERT INTO promos
	          (id, name, code, type, value, min_order, max_discount, usage_limit,
	           used_count, started_at, ended_at, is_active, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		promo.ID, promo.Name, promo.Code, promo.Type, promo.Value, promo.MinOrder,
		promo.MaxDiscount, promo.UsageLimit, promo.UsedCount,
		promo.StartedAt, promo.EndedAt, promo.IsActive,
		promo.CreatedAt, promo.UpdatedAt,
	)
	return err
}

func (r *promoRepository) Update(ctx context.Context, promo *entity.Promo) error {
	promo.UpdatedAt = time.Now()

	query := `UPDATE promos
	          SET name=?, type=?, value=?, min_order=?, max_discount=?, usage_limit=?,
	              started_at=?, ended_at=?, is_active=?, updated_at=?
	          WHERE id=? AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		promo.Name, promo.Type, promo.Value, promo.MinOrder, promo.MaxDiscount, promo.UsageLimit,
		promo.StartedAt, promo.EndedAt, promo.IsActive, promo.UpdatedAt, promo.ID,
	)
	return err
}

func (r *promoRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE promos SET deleted_at=NOW() WHERE id=? AND deleted_at IS NULL", id)
	return err
}

func (r *promoRepository) IncrementUsedCount(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE promos SET used_count = used_count + 1, updated_at = NOW() WHERE id = ?",
		id,
	)
	return err
}
