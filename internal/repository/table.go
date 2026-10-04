package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/terenjit/Cafe-POS/internal/entity"
)

type TableRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Table, error)
	FindByName(ctx context.Context, name string) (*entity.Table, error)
	FindAll(ctx context.Context) ([]entity.Table, error)
	Create(ctx context.Context, table *entity.Table) error
	Update(ctx context.Context, table *entity.Table) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status string) error
	WithTx(tx *sql.Tx) TableRepository
}

type tableRepository struct {
	db sqlDB
}

func NewTableRepository(db sqlDB) TableRepository {
	return &tableRepository{db: db}
}

func (r *tableRepository) WithTx(tx *sql.Tx) TableRepository {
	return &tableRepository{db: tx}
}

func (r *tableRepository) FindByID(ctx context.Context, id string) (*entity.Table, error) {
	query := "SELECT id, name, capacity, status, created_at, updated_at, deleted_at FROM `tables` WHERE id = ? AND deleted_at IS NULL"

	t := &entity.Table{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Capacity, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *tableRepository) FindByName(ctx context.Context, name string) (*entity.Table, error) {
	query := "SELECT id, name, capacity, status, created_at, updated_at, deleted_at FROM `tables` WHERE name = ? AND deleted_at IS NULL"

	t := &entity.Table{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&t.ID, &t.Name, &t.Capacity, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *tableRepository) FindAll(ctx context.Context) ([]entity.Table, error) {
	query := "SELECT id, name, capacity, status, created_at, updated_at, deleted_at FROM `tables` WHERE deleted_at IS NULL ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]entity.Table, 0)
	for rows.Next() {
		var t entity.Table
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Capacity, &t.Status, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
		); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func (r *tableRepository) Create(ctx context.Context, table *entity.Table) error {
	table.ID = uuid.New().String()
	now := time.Now()
	table.CreatedAt = now
	table.UpdatedAt = now

	query := "INSERT INTO `tables` (id, name, capacity, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"

	_, err := r.db.ExecContext(ctx, query,
		table.ID, table.Name, table.Capacity, table.Status, table.CreatedAt, table.UpdatedAt,
	)
	return err
}

func (r *tableRepository) Update(ctx context.Context, table *entity.Table) error {
	table.UpdatedAt = time.Now()

	query := "UPDATE `tables` SET name=?, capacity=?, updated_at=? WHERE id=? AND deleted_at IS NULL"

	_, err := r.db.ExecContext(ctx, query,
		table.Name, table.Capacity, table.UpdatedAt, table.ID,
	)
	return err
}

func (r *tableRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE `tables` SET deleted_at=NOW() WHERE id=? AND deleted_at IS NULL", id)
	return err
}

func (r *tableRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE `tables` SET status=?, updated_at=? WHERE id=?",
		status, time.Now(), id,
	)
	return err
}
