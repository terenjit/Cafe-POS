package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/terenjit/Cafe-POS/internal/entity"
)

type CategoryRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Category, error)
	FindByName(ctx context.Context, name string) (*entity.Category, error)
	FindAll(ctx context.Context) ([]entity.Category, error)
	Create(ctx context.Context, category *entity.Category) error
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *sql.Tx) CategoryRepository
}

type categoryRepository struct {
	db sqlDB
}

func NewCategoryRepository(db sqlDB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) WithTx(tx *sql.Tx) CategoryRepository {
	return &categoryRepository{db: tx}
}

func (r *categoryRepository) FindByID(ctx context.Context, id string) (*entity.Category, error) {
	query := `SELECT id, name, created_at, updated_at, deleted_at
	          FROM categories WHERE id = ? AND deleted_at IS NULL`

	cat := &entity.Category{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cat.ID, &cat.Name, &cat.CreatedAt, &cat.UpdatedAt, &cat.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cat, nil
}

func (r *categoryRepository) FindByName(ctx context.Context, name string) (*entity.Category, error) {
	query := `SELECT id, name, created_at, updated_at, deleted_at
	          FROM categories WHERE name = ? AND deleted_at IS NULL`

	cat := &entity.Category{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&cat.ID, &cat.Name, &cat.CreatedAt, &cat.UpdatedAt, &cat.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cat, nil
}

func (r *categoryRepository) FindAll(ctx context.Context) ([]entity.Category, error) {
	query := `SELECT id, name, created_at, updated_at, deleted_at
	          FROM categories WHERE deleted_at IS NULL ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]entity.Category, 0)
	for rows.Next() {
		var cat entity.Category
		if err := rows.Scan(
			&cat.ID, &cat.Name, &cat.CreatedAt, &cat.UpdatedAt, &cat.DeletedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	category.ID = uuid.New().String()
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	query := `INSERT INTO categories (id, name, created_at, updated_at)
	          VALUES (?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.CreatedAt, category.UpdatedAt,
	)
	return err
}

func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	category.UpdatedAt = time.Now()

	query := `UPDATE categories SET name=?, updated_at=? WHERE id=? AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, category.Name, category.UpdatedAt, category.ID)
	return err
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE categories SET deleted_at=NOW() WHERE id=? AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
