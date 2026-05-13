# Database Schema
## Terenjit Cafe — Point of Sale System

**Database:** MySQL 8.0+  
**Charset:** utf8mb4  
**Collation:** utf8mb4_unicode_ci

---

## 1. Table Overview

| Table | Description |
|---|---|
| `users` | All system users — both owners and cashiers. Role differentiates access. Supports soft delete. |
| `categories` | Product categories (e.g., Coffee, Food). Supports soft delete. |
| `products` | Product catalog with pricing and availability status. Supports soft delete. |
| `stock_movements` | Immutable audit log of every stock change. Current stock is derived by summing this table per product. |
| `tables` | Physical cafe tables that can be assigned to orders. |
| `promos` | Promotions with percentage or fixed-amount discounts and a validity date range. |
| `shifts` | Cashier shift records tracking opening/closing cash and total sales. |
| `orders` | Order headers linking a shift, cashier, table, and optional promo. |
| `order_items` | Line items for each order with a price snapshot at the time of sale. |
| `payments` | Midtrans payment records, one per order, storing gateway status and raw webhook data. |

---

## 2. Table Definitions

---

### 2.1 `users`

Stores owner and cashier accounts. The `role` column controls which interface and endpoints a user can access. Soft-deleted users are retained for historical order attribution.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| name | VARCHAR(100) | NOT NULL | Full display name |
| email | VARCHAR(100) | UNIQUE, NOT NULL | Login email address |
| password | VARCHAR(255) | NOT NULL | bcrypt-hashed password; plaintext never stored |
| role | ENUM('owner', 'cashier') | NOT NULL | Determines access level and UI routing |
| is_active | BOOLEAN | NOT NULL, DEFAULT TRUE | FALSE = deactivated; blocks login immediately |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Record creation time |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last modification time |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp; NULL = not deleted |

**Indexes:**
- `UNIQUE INDEX idx_users_email ON users(email)`
- `INDEX idx_users_role ON users(role)`
- `INDEX idx_users_deleted_at ON users(deleted_at)`

---

### 2.2 `categories`

Product groupings shown as filter tabs on the cashier order screen. A category with available products cannot be soft-deleted.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| name | VARCHAR(100) | NOT NULL | Category label (e.g., "Coffee", "Food") |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Record creation time |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last modification time |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp; NULL = not deleted |

**Indexes:**
- `INDEX idx_categories_deleted_at ON categories(deleted_at)`

---

### 2.3 `products`

The product catalog. Stock is not stored directly on this table — it is always computed from `stock_movements`. Soft-deleted products are hidden from all views but remain valid FK targets in `order_items`.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| category_id | VARCHAR(36) | FK → categories.id, NOT NULL | Owning category |
| name | VARCHAR(100) | NOT NULL | Product name |
| description | TEXT | NULL | Optional product description |
| price | BIGINT | NOT NULL | Unit price in IDR (whole rupiah) |
| image_url | VARCHAR(255) | NULL | Storage path or URL for product image; NULL = no image |
| is_available | BOOLEAN | NOT NULL, DEFAULT TRUE | FALSE = hidden from cashier's product list without soft delete |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Record creation time |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last modification time |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp; NULL = not deleted |

**Indexes:**
- `INDEX idx_products_category_id ON products(category_id)`
- `INDEX idx_products_is_available ON products(is_available)`
- `INDEX idx_products_deleted_at ON products(deleted_at)`
- `INDEX idx_products_available_deleted ON products(is_available, deleted_at)` — composite for the common "available, non-deleted" filter

---

### 2.4 `stock_movements`

Append-only audit trail for every stock change. Rows are never updated or deleted. Positive `quantity` adds stock; negative `quantity` reduces it. The live stock of a product is `SUM(quantity)` over all its movements.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| product_id | VARCHAR(36) | FK → products.id, NOT NULL | Product whose stock changed |
| type | ENUM('purchase', 'sale', 'adjustment', 'return') | NOT NULL | Source of the movement |
| quantity | INT | NOT NULL | Units changed; positive = stock in, negative = stock out |
| note | VARCHAR(255) | NULL | Optional human-readable note (e.g., "Restock from supplier", "Manual correction") |
| created_by | VARCHAR(36) | FK → users.id, NOT NULL | User who recorded this movement |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Time of movement; no updated_at — rows are immutable |

**Indexes:**
- `INDEX idx_stock_movements_product_id ON stock_movements(product_id)`
- `INDEX idx_stock_movements_created_by ON stock_movements(created_by)`
- `INDEX idx_stock_movements_type ON stock_movements(type)`

---

### 2.5 `tables`

Physical tables in the cafe. Orders reference this table; a NULL `table_id` on an order means the order is takeaway.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| name | VARCHAR(50) | NOT NULL | Table label (e.g., "Table 1", "VIP Room") |
| is_active | BOOLEAN | NOT NULL, DEFAULT TRUE | FALSE = hidden from cashier's table selector |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Record creation time |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last modification time |

**Indexes:**
- `INDEX idx_tables_is_active ON tables(is_active)`

---

### 2.6 `promos`

Discount promotions. A promo is eligible only when `is_active = TRUE` AND `CURDATE() BETWEEN valid_from AND valid_until`. Two separate discount columns are used so each type maps to the correct data type.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| name | VARCHAR(100) | NOT NULL | Promo display name |
| description | TEXT | NULL | Optional details shown to cashier |
| discount_type | ENUM('percentage', 'fixed') | NOT NULL | Determines which discount column is active |
| discount_percentage | DECIMAL(5,2) | NULL | Percentage value when discount_type = 'percentage' (e.g., 10.00 = 10%); NULL otherwise |
| discount_fixed_amount | BIGINT | NULL | Fixed discount in IDR when discount_type = 'fixed'; NULL otherwise |
| min_transaction_amount | BIGINT | NULL | Minimum order subtotal in IDR required to apply promo; NULL = no minimum |
| valid_from | DATE | NOT NULL | First calendar day this promo is valid (inclusive) |
| valid_until | DATE | NOT NULL | Last calendar day this promo is valid (inclusive) |
| is_active | BOOLEAN | NOT NULL, DEFAULT TRUE | Manual override toggle; FALSE always excludes promo regardless of dates |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Record creation time |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last modification time |

**Indexes:**
- `INDEX idx_promos_active_dates ON promos(is_active, valid_from, valid_until)` — composite for the active promo lookup

---

### 2.7 `shifts`

One row per cashier shift. Only one shift per cashier may have `status = 'open'` at any time — enforced at the application layer. `total_sales` is a running aggregate updated each time an order in this shift is paid.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| cashier_id | VARCHAR(36) | FK → users.id, NOT NULL | Cashier who owns this shift |
| opening_cash | BIGINT | NOT NULL | Cash in the drawer when shift was opened, in IDR |
| closing_cash | BIGINT | NULL | Cash in the drawer when shift was closed, in IDR; NULL while shift is open |
| total_sales | BIGINT | NOT NULL, DEFAULT 0 | Running total of all paid order amounts in this shift, in IDR |
| status | ENUM('open', 'closed') | NOT NULL, DEFAULT 'open' | Shift lifecycle status |
| opened_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | When the cashier opened the shift |
| closed_at | TIMESTAMP | NULL | When the cashier closed the shift; NULL while open |

**Indexes:**
- `INDEX idx_shifts_cashier_id ON shifts(cashier_id)`
- `INDEX idx_shifts_status ON shifts(status)`
- `INDEX idx_shifts_cashier_status ON shifts(cashier_id, status)` — composite for the "does this cashier have an open shift?" check

---

### 2.8 `orders`

Order header. `cashier_id` is denormalized from the shift for efficient reporting. `table_id` is nullable to support takeaway orders. `midtrans_order_id` uniqueness prevents duplicate Snap payment requests.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| shift_id | VARCHAR(36) | FK → shifts.id, NOT NULL | Shift this order was created under |
| table_id | VARCHAR(36) | FK → tables.id, NULL | Assigned table; NULL = takeaway order |
| cashier_id | VARCHAR(36) | FK → users.id, NOT NULL | Cashier who created the order (denormalized from shift for reporting) |
| promo_id | VARCHAR(36) | FK → promos.id, NULL | Applied promo; NULL if no promo used |
| status | ENUM('draft', 'pending_payment', 'paid', 'cancelled') | NOT NULL, DEFAULT 'draft' | Order lifecycle status |
| subtotal | BIGINT | NOT NULL | Sum of all order_items.subtotal before discount, in IDR |
| discount_amount | BIGINT | NOT NULL, DEFAULT 0 | Calculated discount applied to this order, in IDR |
| total | BIGINT | NOT NULL | subtotal − discount_amount; the amount charged to the customer, in IDR |
| payment_method | VARCHAR(50) | NULL | Payment method reported by Midtrans (e.g., 'gopay', 'bank_transfer'); NULL until paid |
| midtrans_order_id | VARCHAR(100) | UNIQUE, NULL | Order ID sent to Midtrans Snap; set when Snap request is created |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Record creation time |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last modification time |

**Indexes:**
- `INDEX idx_orders_shift_id ON orders(shift_id)`
- `INDEX idx_orders_cashier_id ON orders(cashier_id)`
- `INDEX idx_orders_status ON orders(status)`
- `INDEX idx_orders_table_id ON orders(table_id)`
- `UNIQUE INDEX idx_orders_midtrans_order_id ON orders(midtrans_order_id)`

---

### 2.9 `order_items`

Line items for an order. `price` is the product's unit price at the moment the item was added — it does not change if the product's price is later updated. Items are immutable once the order is paid.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| order_id | VARCHAR(36) | FK → orders.id, NOT NULL | Parent order |
| product_id | VARCHAR(36) | FK → products.id, NOT NULL | Reference to the product ordered |
| quantity | INT | NOT NULL | Units ordered; minimum 1 |
| price | BIGINT | NOT NULL | Snapshot of products.price at time of order, in IDR |
| subtotal | BIGINT | NOT NULL | price × quantity, in IDR |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Record creation time |

**Indexes:**
- `INDEX idx_order_items_order_id ON order_items(order_id)`
- `INDEX idx_order_items_product_id ON order_items(product_id)` — for best-selling product aggregate queries

---

### 2.10 `payments`

One payment record per order. Isolates Midtrans-specific data from order logic. `raw_notification` stores the full webhook payload for debugging and reconciliation.

| column_name | data_type | constraint | description |
|---|---|---|---|
| id | VARCHAR(36) | PK, NOT NULL | UUID v4 |
| order_id | VARCHAR(36) | FK → orders.id, NOT NULL, UNIQUE | Parent order; 1:1 relationship |
| midtrans_order_id | VARCHAR(100) | NOT NULL | Order ID sent to Midtrans; matches orders.midtrans_order_id |
| midtrans_transaction_id | VARCHAR(100) | UNIQUE, NULL | Midtrans server-side transaction ID from webhook; NULL until webhook arrives |
| status | ENUM('pending', 'paid', 'failed', 'expired') | NOT NULL, DEFAULT 'pending' | Payment lifecycle status |
| amount | BIGINT | NOT NULL | Total amount charged to customer, in IDR |
| payment_type | VARCHAR(50) | NULL | Payment channel reported by Midtrans (e.g., 'gopay', 'bank_transfer'); NULL until webhook arrives |
| raw_notification | JSON | NULL | Full Midtrans webhook payload stored for debugging and reconciliation |
| paid_at | TIMESTAMP | NULL | Timestamp from Midtrans settlement notification; NULL until paid |
| created_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP | Record creation time (Snap request time) |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | Last modification time |

**Indexes:**
- `UNIQUE INDEX idx_payments_order_id ON payments(order_id)`
- `UNIQUE INDEX idx_payments_midtrans_transaction_id ON payments(midtrans_transaction_id)`

---

## 3. Relationships (Foreign Keys)

```
users           → orders         (cashier_id → users.id)
users           → shifts         (cashier_id → users.id)
users           → stock_movements (created_by → users.id)

categories      → products       (category_id → categories.id)

products        → order_items    (product_id → products.id)
products        → stock_movements (product_id → products.id)

orders          → order_items    (order_id → orders.id)
orders          → payments       (order_id → orders.id)  [1:1]

shifts          → orders         (shift_id → shifts.id)

tables          → orders         (table_id → orders.table_id)  [nullable]

promos          → orders         (promo_id → orders.promo_id)  [nullable]
```

### FK Behaviour Summary

| FK | ON DELETE | ON UPDATE | Reason |
|---|---|---|---|
| products.category_id → categories.id | RESTRICT | CASCADE | Prevent orphaned products; deletion guard enforced at app layer |
| stock_movements.product_id → products.id | RESTRICT | CASCADE | Movements are immutable; product cannot be hard-deleted (soft delete used) |
| stock_movements.created_by → users.id | RESTRICT | CASCADE | Every movement must be attributed to a user; users are soft-deleted, not hard-deleted |
| shifts.cashier_id → users.id | RESTRICT | CASCADE | Historical shifts must retain cashier reference |
| orders.shift_id → shifts.id | RESTRICT | CASCADE | Orders must stay linked to their shift |
| orders.cashier_id → users.id | RESTRICT | CASCADE | Historical attribution must be preserved |
| orders.table_id → tables.id | SET NULL | CASCADE | Nullable FK; if a table were ever removed, order history is not invalidated |
| orders.promo_id → promos.id | SET NULL | CASCADE | Nullable FK; promo removal must not break order history |
| order_items.order_id → orders.id | CASCADE | CASCADE | Items are owned by the order; deleting a draft order removes its items |
| order_items.product_id → products.id | RESTRICT | CASCADE | Soft delete ensures product rows always remain; hard delete blocked |
| payments.order_id → orders.id | CASCADE | CASCADE | Payment record is owned by the order |

---

## 4. Key Design Decisions

### 4.1 UUID (VARCHAR 36) as Primary Key
UUIDs prevent sequential ID enumeration through the API (an attacker cannot iterate `/orders/1`, `/orders/2`). They also allow the application layer to generate IDs before inserting, which simplifies insert-then-reference patterns. The performance trade-off versus AUTO_INCREMENT is accepted given the expected transaction volume of a single cafe.

### 4.2 BIGINT for Currency in Whole Rupiah
Indonesian Rupiah (IDR) has no active fractional subunit — the Indonesian sen (1/100 IDR) has been out of circulation since 1964. Storing amounts as BIGINT in whole rupiah avoids floating-point precision errors that arise with DOUBLE or DECIMAL in arithmetic-heavy queries (summing thousands of orders). BIGINT handles values up to ~9.2 quadrillion rupiah, far exceeding any realistic transaction amount.

### 4.3 ENUM for Status Columns
ENUM constrains valid values at the database layer, not just the application layer. It documents the full lifecycle directly in the schema and prevents garbage data from reaching status columns. MySQL stores ENUMs as 1–2 byte integers internally, making them space-efficient. The trade-off — schema migrations are required to add values — is acceptable for status fields that change infrequently.

### 4.4 Separate `discount_percentage` and `discount_fixed_amount` Columns in `promos`
A single polymorphic `discount_value BIGINT` column would require the application to interpret the integer differently based on `discount_type`. Separate columns make the type contract explicit: `discount_percentage` is DECIMAL(5,2) (a ratio, supports values like 7.50%), while `discount_fixed_amount` is BIGINT (a currency amount in IDR). Column nullability enforces the constraint — exactly one is non-NULL per row.

### 4.5 DATE (not TIMESTAMP) for Promo Validity
`valid_from` and `valid_until` use DATE because promo validity is a business calendar concept ("this promo runs all of May 2026"). Using TIMESTAMP would introduce timezone ambiguity — midnight in WIB (UTC+7) is the previous evening in UTC. DATE columns are compared directly against `CURDATE()` with no timezone risk.

### 4.6 Price Snapshot in `order_items` (No Name Snapshot)
`price` is copied from the product at the time the item is added to the order. This ensures that changing a product's price later does not silently alter historical order records. The product name is intentionally not snapshotted — `product_id` retains the FK reference to look up the original name if needed, and soft delete on products guarantees the row always exists.

### 4.7 Stock Derived from `stock_movements`, Not Stored on `products`
Stock is not stored as a column on `products`. Instead, the live stock of a product is `SELECT SUM(quantity) FROM stock_movements WHERE product_id = ?`. This keeps stock_movements as the single source of truth and eliminates the risk of the two values diverging. For products with high movement volume, a periodic snapshot or a materialised value can be introduced in a future version. Negative `quantity` values in stock_movements represent outflows (type `sale`, `adjustment`); positive values represent inflows (`purchase`, `return`).

### 4.8 `cashier_id` Denormalized in `orders`
`orders.cashier_id` duplicates what could be derived via `orders.shift_id → shifts.cashier_id`. This denormalization is intentional: report queries that aggregate revenue per cashier are frequent, and avoiding the extra join on every row reduces complexity and query time. The value is stable — a shift's cashier never changes — so there is no risk of the denormalization going stale.

### 4.9 `payment_method` on `orders` and `payment_type` on `payments`
`orders.payment_method` stores the payment channel confirmed by Midtrans so that order-level queries can surface the payment method without joining to `payments`. `payments.payment_type` is the canonical Midtrans-reported value. Both are populated from the same webhook; the redundancy is intentional for query convenience.

### 4.10 Separate `payments` Table
Midtrans-specific fields (`midtrans_transaction_id`, `raw_notification`, `payment_type`) belong to the payment gateway domain, not the order domain. Separating them keeps `orders` clean and focused on the business transaction. It also simplifies the idempotency check: a webhook handler queries `payments` by `midtrans_transaction_id` before acting, without locking the `orders` row.

### 4.11 `total_sales` on `shifts`
`shifts.total_sales` is a denormalized running total incremented each time an order in the shift reaches `paid` status. This avoids a `SUM` query over `orders` every time the shift close summary is displayed. The value is always reconcilable by querying `SUM(orders.total) WHERE shift_id = ? AND status = 'paid'`.

### 4.12 Nullable `table_id` on `orders`
`orders.table_id` is nullable to support takeaway orders that are not assigned to a physical table. A NULL value explicitly means "no table / takeaway." The FK is defined with `ON DELETE SET NULL` so that removing a table record (if ever required) does not break existing order history.

### 4.13 `tables` Table Name
Using `tables` as the table name is straightforward in MySQL when queries use backtick quoting (`` `tables` ``), which Go's database drivers handle by default. Applications using raw SQL must quote the identifier; ORM-generated queries handle this automatically. The name was chosen to match the domain language directly over an alternative like `cafe_tables`.
