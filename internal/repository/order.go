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

type OrderFilter struct {
	ShiftID   string
	CashierID string
	Status    string
	Page      int
	Limit     int
}

type OrderRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Order, error)
	FindAll(ctx context.Context, filter OrderFilter) ([]entity.Order, int, error)
	Create(ctx context.Context, order *entity.Order) error
	Update(ctx context.Context, order *entity.Order) error
	AddItem(ctx context.Context, item *entity.OrderItem) error
	UpdateItem(ctx context.Context, item *entity.OrderItem) error
	DeleteItem(ctx context.Context, itemID string) error
	FindItemByID(ctx context.Context, itemID string) (*entity.OrderItem, error)
	CountPaidByShiftID(ctx context.Context, shiftID string) (int64, error)
	RecalculateTotal(ctx context.Context, orderID string) error
	WithTx(tx *sql.Tx) OrderRepository
}

type orderRepository struct {
	db sqlDB
}

func NewOrderRepository(db sqlDB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) WithTx(tx *sql.Tx) OrderRepository {
	return &orderRepository{db: tx}
}

func (r *orderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	query := `
		SELECT o.id, o.shift_id, o.cashier_id, o.table_id, o.promo_id,
		       o.status, o.subtotal, o.discount_amount, o.total, o.notes,
		       o.created_at, o.updated_at,
		       u.id, u.name,
		       t.id, t.name
		FROM orders o
		JOIN users u ON u.id = o.cashier_id
		LEFT JOIN ` + "`tables`" + ` t ON t.id = o.table_id
		WHERE o.id = ?`

	o := &entity.Order{}
	var cashier entity.User
	var tableID, tableName sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.ShiftID, &o.CashierID, &o.TableID, &o.PromoID,
		&o.Status, &o.Subtotal, &o.DiscountAmount, &o.Total, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt,
		&cashier.ID, &cashier.Name,
		&tableID, &tableName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o.Cashier = &cashier
	if tableID.Valid {
		t := &entity.Table{ID: tableID.String, Name: tableName.String}
		o.Table = t
	}

	items, err := r.findItems(ctx, id)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return o, nil
}

func (r *orderRepository) findItems(ctx context.Context, orderID string) ([]entity.OrderItem, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.subtotal,
		       oi.notes, oi.created_at, oi.updated_at,
		       p.id, p.name, p.price AS product_price
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = ?`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entity.OrderItem, 0)
	for rows.Next() {
		var item entity.OrderItem
		var product entity.Product
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price, &item.Subtotal,
			&item.Notes, &item.CreatedAt, &item.UpdatedAt,
			&product.ID, &product.Name, &product.Price,
		); err != nil {
			return nil, err
		}
		item.Product = &product
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *orderRepository) FindAll(ctx context.Context, filter OrderFilter) ([]entity.Order, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	where, args := buildOrderWhere(filter)

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders o"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	dataQuery := `
		SELECT o.id, o.shift_id, o.cashier_id, o.table_id, o.promo_id,
		       o.status, o.subtotal, o.discount_amount, o.total, o.notes,
		       o.created_at, o.updated_at,
		       u.id, u.name
		FROM orders o
		JOIN users u ON u.id = o.cashier_id` +
		where + ` ORDER BY o.created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, dataQuery, append(args, filter.Limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	orders := make([]entity.Order, 0)
	for rows.Next() {
		var o entity.Order
		var cashier entity.User
		if err := rows.Scan(
			&o.ID, &o.ShiftID, &o.CashierID, &o.TableID, &o.PromoID,
			&o.Status, &o.Subtotal, &o.DiscountAmount, &o.Total, &o.Notes,
			&o.CreatedAt, &o.UpdatedAt,
			&cashier.ID, &cashier.Name,
		); err != nil {
			return nil, 0, err
		}
		o.Cashier = &cashier
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	order.ID = uuid.New().String()
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	query := `INSERT INTO orders
	          (id, shift_id, cashier_id, table_id, promo_id, status, subtotal, discount_amount, total, notes, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		order.ID, order.ShiftID, order.CashierID, order.TableID, order.PromoID,
		order.Status, order.Subtotal, order.DiscountAmount, order.Total, order.Notes,
		order.CreatedAt, order.UpdatedAt,
	)
	return err
}

func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	order.UpdatedAt = time.Now()

	query := `UPDATE orders
	          SET table_id=?, promo_id=?, status=?, subtotal=?, discount_amount=?, total=?, notes=?, updated_at=?
	          WHERE id=?`

	_, err := r.db.ExecContext(ctx, query,
		order.TableID, order.PromoID, order.Status, order.Subtotal,
		order.DiscountAmount, order.Total, order.Notes, order.UpdatedAt, order.ID,
	)
	return err
}

func (r *orderRepository) AddItem(ctx context.Context, item *entity.OrderItem) error {
	item.ID = uuid.New().String()
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	item.Subtotal = item.Price * int64(item.Quantity)

	query := `INSERT INTO order_items (id, order_id, product_id, quantity, price, subtotal, notes, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.OrderID, item.ProductID, item.Quantity,
		item.Price, item.Subtotal, item.Notes, item.CreatedAt, item.UpdatedAt,
	)
	return err
}

func (r *orderRepository) UpdateItem(ctx context.Context, item *entity.OrderItem) error {
	item.UpdatedAt = time.Now()

	query := `UPDATE order_items SET quantity=?, notes=?, subtotal=?, updated_at=? WHERE id=?`

	_, err := r.db.ExecContext(ctx, query,
		item.Quantity, item.Notes, item.Subtotal, item.UpdatedAt, item.ID,
	)
	return err
}

func (r *orderRepository) DeleteItem(ctx context.Context, itemID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM order_items WHERE id=?`, itemID)
	return err
}

func (r *orderRepository) FindItemByID(ctx context.Context, itemID string) (*entity.OrderItem, error) {
	query := `SELECT id, order_id, product_id, quantity, price, subtotal, notes, created_at, updated_at
	          FROM order_items WHERE id=?`

	item := &entity.OrderItem{}
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID, &item.OrderID, &item.ProductID, &item.Quantity,
		&item.Price, &item.Subtotal, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *orderRepository) RecalculateTotal(ctx context.Context, orderID string) error {
	query := `UPDATE orders SET
		subtotal   = (SELECT COALESCE(SUM(subtotal), 0) FROM order_items WHERE order_id = ?),
		total      = (SELECT COALESCE(SUM(subtotal), 0) FROM order_items WHERE order_id = ?) - discount_amount,
		updated_at = NOW()
	WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, orderID, orderID, orderID)
	return err
}

func (r *orderRepository) CountPaidByShiftID(ctx context.Context, shiftID string) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM orders WHERE shift_id = ? AND status = 'paid'",
		shiftID,
	).Scan(&count)
	return count, err
}

func buildOrderWhere(filter OrderFilter) (string, []interface{}) {
	conditions := make([]string, 0)
	args := make([]interface{}, 0)

	if filter.ShiftID != "" {
		conditions = append(conditions, "o.shift_id = ?")
		args = append(args, filter.ShiftID)
	}
	if filter.CashierID != "" {
		conditions = append(conditions, "o.cashier_id = ?")
		args = append(args, filter.CashierID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "o.status = ?")
		args = append(args, filter.Status)
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
