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

type ProductFilter struct {
	CategoryID string
	IsActive   *bool
	Search     string
	Page       int
	Limit      int
}

type ProductRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Product, error)
	FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int, error)
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, id string, stock int) error
	WithTx(tx *sql.Tx) ProductRepository
}

type productRepository struct {
	db sqlDB
}

func NewProductRepository(db sqlDB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) WithTx(tx *sql.Tx) ProductRepository {
	return &productRepository{db: tx}
}

func (r *productRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	query := `
		SELECT p.id, p.category_id, p.name, p.description, p.price, p.stock,
		       p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
		WHERE p.id = ? AND p.deleted_at IS NULL`

	p := &entity.Product{}
	var cat entity.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.Stock,
		&p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		&cat.ID, &cat.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Category = &cat
	return p, nil
}

func (r *productRepository) FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	where, args := buildProductWhere(filter)

	// total count
	countQuery := "SELECT COUNT(*) FROM products p" + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	dataQuery := `
		SELECT p.id, p.category_id, p.name, p.description, p.price, p.stock,
		       p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL` +
		where + ` ORDER BY p.created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, dataQuery, append(args, filter.Limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := make([]entity.Product, 0)
	for rows.Next() {
		var p entity.Product
		var cat entity.Category
		if err := rows.Scan(
			&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.Stock,
			&p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&cat.ID, &cat.Name,
		); err != nil {
			return nil, 0, err
		}
		p.Category = &cat
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	product.ID = uuid.New().String()
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	query := `INSERT INTO products (id, category_id, name, description, price, stock, image_url, is_active, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		product.ID, product.CategoryID, product.Name, product.Description,
		product.Price, product.Stock, product.ImageURL, product.IsActive,
		product.CreatedAt, product.UpdatedAt,
	)
	return err
}

func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
	product.UpdatedAt = time.Now()

	query := `UPDATE products
	          SET category_id=?, name=?, description=?, price=?, image_url=?, is_active=?, updated_at=?
	          WHERE id=? AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		product.CategoryID, product.Name, product.Description, product.Price,
		product.ImageURL, product.IsActive, product.UpdatedAt, product.ID,
	)
	return err
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE products SET deleted_at=NOW() WHERE id=? AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// UpdateStock must always be called alongside StockRepository.Create()
// within a single DB transaction — never call this directly from a handler.
func (r *productRepository) UpdateStock(ctx context.Context, id string, stock int) error {
	query := `UPDATE products SET stock=?, updated_at=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, stock, time.Now(), id)
	return err
}

func buildProductWhere(filter ProductFilter) (string, []interface{}) {
	conditions := []string{"p.deleted_at IS NULL"}
	args := make([]interface{}, 0)

	if filter.CategoryID != "" {
		conditions = append(conditions, "p.category_id = ?")
		args = append(args, filter.CategoryID)
	}
	if filter.IsActive != nil {
		conditions = append(conditions, "p.is_active = ?")
		args = append(args, *filter.IsActive)
	}
	if filter.Search != "" {
		conditions = append(conditions, "p.name LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}
