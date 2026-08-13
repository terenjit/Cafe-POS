package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/terenjit/Cafe-POS/internal/entity"
)

type StockFilter struct {
	ProductID string
	UserID    string
	Type      string
	Page      int
	Limit     int
}

type StockRepository interface {
	Create(ctx context.Context, movement *entity.StockMovement) error
	FindByProductID(ctx context.Context, productID string, filter StockFilter) ([]entity.StockMovement, int, error)
	FindAll(ctx context.Context, filter StockFilter) ([]entity.StockMovement, int, error)
}

type stockRepository struct {
	db *sql.DB
}

func NewStockRepository(db *sql.DB) StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) Create(ctx context.Context, movement *entity.StockMovement) error {
	movement.ID = uuid.New().String()
	now := time.Now()
	movement.CreatedAt = now
	movement.UpdatedAt = now

	query := `INSERT INTO stock_movements
	          (id, product_id, user_id, type, quantity, stock_before, stock_after, notes, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		movement.ID, movement.ProductID, movement.UserID, movement.Type,
		movement.Quantity, movement.StockBefore, movement.StockAfter,
		movement.Notes, movement.CreatedAt, movement.UpdatedAt,
	)
	return err
}

func (r *stockRepository) FindByProductID(ctx context.Context, productID string, filter StockFilter) ([]entity.StockMovement, int, error) {
	filter.ProductID = productID
	return r.query(ctx, filter)
}

func (r *stockRepository) FindAll(ctx context.Context, filter StockFilter) ([]entity.StockMovement, int, error) {
	return r.query(ctx, filter)
}

func (r *stockRepository) query(ctx context.Context, filter StockFilter) ([]entity.StockMovement, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	where, args := buildStockWhere(filter)

	var total int
	countQuery := "SELECT COUNT(*) FROM stock_movements sm" + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	dataQuery := `
		SELECT sm.id, sm.product_id, sm.user_id, sm.type, sm.quantity,
		       sm.stock_before, sm.stock_after, sm.notes, sm.created_at, sm.updated_at,
		       p.id, p.name,
		       u.id, u.name
		FROM stock_movements sm
		JOIN products p ON p.id = sm.product_id
		JOIN users u ON u.id = sm.user_id` +
		where + ` ORDER BY sm.created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, dataQuery, append(args, filter.Limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	movements := make([]entity.StockMovement, 0)
	for rows.Next() {
		var sm entity.StockMovement
		var product entity.Product
		var user entity.User
		if err := rows.Scan(
			&sm.ID, &sm.ProductID, &sm.UserID, &sm.Type, &sm.Quantity,
			&sm.StockBefore, &sm.StockAfter, &sm.Notes, &sm.CreatedAt, &sm.UpdatedAt,
			&product.ID, &product.Name,
			&user.ID, &user.Name,
		); err != nil {
			return nil, 0, err
		}
		sm.Product = &product
		sm.User = &user
		movements = append(movements, sm)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return movements, total, nil
}

func buildStockWhere(filter StockFilter) (string, []interface{}) {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)

	if filter.ProductID != "" {
		conditions = append(conditions, "sm.product_id = ?")
		args = append(args, filter.ProductID)
	}
	if filter.UserID != "" {
		conditions = append(conditions, "sm.user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.Type != "" {
		conditions = append(conditions, "sm.type = ?")
		args = append(args, filter.Type)
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
