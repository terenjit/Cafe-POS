# Product Requirements Document
## Terenjit Cafe — Point of Sale System

**Version:** 1.0  
**Date:** 2026-05-12  
**Stack:** Go (backend API), MySQL (database), Midtrans Snap (payment gateway)

---

## 1. Product Description

Terenjit Cafe POS is a web-based point-of-sale system built for medium-scale coffee shop operations. It provides two distinct interfaces: an owner-facing management dashboard and a cashier-facing transaction terminal.

The system handles the full transaction lifecycle — from product catalog management and stock tracking, through order creation and payment via Midtrans Snap, to shift reconciliation and revenue reporting. It is designed to be operated on desktop browsers at the counter and on a management device used by the owner.

**Primary users:** Coffee shop owner and cashiers employed at a single-location coffee shop with multiple product categories, table service, and daily shift operations.

---

## 2. User Roles

### Owner
The Owner has full system access. They configure the product catalog, manage pricing and promotions, monitor stock levels, and review business performance through the report dashboard. The Owner can also create and deactivate cashier accounts. The Owner does not process transactions directly.

### Cashier
The Cashier has access only to transaction-related features. They open and close their own shift, create and manage orders for tables, apply active promotions, and process payments. A Cashier cannot access reports, product management, or stock settings. Cashiers can only see transactions from their own current shift.

---

## 3. Owner Features

### 3.1 Product Management

The Owner can create, read, update, and soft-delete products. Each product has a name, description, price, category, photo, stock quantity, and active/inactive status. Soft-deleted products are hidden from all views but preserved in the database for historical transaction integrity.

Setting a product to inactive removes it from the cashier's product selection without deleting it or its stock record.

**Acceptance criteria:**
- Owner can create a product with all fields including an uploaded photo.
- Owner can edit any product field, including replacing the photo.
- Owner can toggle a product between active and inactive; inactive products do not appear in the cashier's order screen.
- Owner can soft-delete a product; it disappears from all lists but its product ID remains valid in past transactions.
- Product photo is stored and served correctly; a placeholder is shown when no photo is set.

---

### 3.2 Product Category Management

Categories group products (e.g., Coffee, Non-Coffee, Food, Merchandise). The Owner can create, rename, and soft-delete categories. A category cannot be deleted if it has active products assigned to it.

**Acceptance criteria:**
- Owner can create and rename categories.
- Attempting to delete a category that contains active products returns a clear error.
- Deleting a category that has only soft-deleted or inactive products succeeds; those products retain a reference to the category for historical display.
- Categories appear as filter tabs on the cashier's product selection screen.

---

### 3.3 Stock Management

Each product has a stock quantity. The Owner can adjust stock manually (e.g., restock delivery) or see it reduced automatically when a transaction is completed. Every stock change — whether from a sale or a manual adjustment — is recorded as a stock movement entry containing: product, change amount, direction (in/out), reason, and timestamp.

**Acceptance criteria:**
- Stock decrements automatically when a transaction is marked as paid.
- Owner can perform a manual stock adjustment (add or subtract) with a required reason field.
- Stock movement history is viewable per product, sorted by most recent.
- Stock quantity cannot go below zero via a sale; the system prevents checkout if any item in the order exceeds available stock.
- Stock movement records are immutable; they cannot be edited or deleted.

---

### 3.4 Table Management

The Owner defines the tables available in the cafe. Each table has a name/number and an active/inactive status. Tables are assigned to transactions by the cashier at order creation time.

**Acceptance criteria:**
- Owner can add, rename, and deactivate tables.
- Inactive tables do not appear in the cashier's table selection dropdown.
- A table cannot be deactivated while it has an open (unpaid) transaction linked to it.

---

### 3.5 Cashier User Management

The Owner can create cashier accounts with a name, username, and password. The Owner can also deactivate a cashier account to prevent login. Cashiers cannot manage other accounts.

**Acceptance criteria:**
- Owner can create a cashier account; the new cashier can immediately log in.
- Owner can deactivate a cashier account; a deactivated cashier cannot log in and receives a clear error message.
- Owner cannot delete a cashier account with associated historical transactions; only deactivation is permitted.
- Passwords are stored hashed (bcrypt); plaintext passwords are never stored or returned.

---

### 3.6 Report Dashboard

The Owner has access to a dashboard displaying:

- **Revenue summary:** total revenue for a selected period (daily, weekly, monthly), broken down by day.
- **Best-selling products:** top N products ranked by quantity sold in the selected period.
- **Transactions per cashier:** number of transactions and total revenue per cashier in the selected period.

All report data is read-only and derived from completed (paid) transactions only.

**Acceptance criteria:**
- Owner can switch between daily, weekly, and monthly views; data updates accordingly.
- Revenue figures match the sum of completed transaction totals within the period.
- Best-selling product list correctly ranks by units sold, not revenue.
- Transactions-per-cashier report correctly attributes each transaction to the cashier who processed it.
- Reports with no data in the period display empty states, not errors.

---

### 3.7 Promo and Discount Management

The Owner can create promotions that apply a discount to a transaction. Each promo has:
- Name and description
- Discount type: percentage (e.g., 10%) or fixed amount (e.g., Rp 10,000)
- Valid date range (start date, end date)
- Active/inactive toggle (manual override regardless of date range)
- Minimum transaction amount (optional)

A promo is considered "active" only when: its active toggle is on AND the current date falls within its valid date range.

**Acceptance criteria:**
- Owner can create, edit, and deactivate promos.
- Only promos meeting both conditions (active toggle + valid date) appear in the cashier's promo list.
- Percentage discount is applied to the transaction subtotal before tax/service charge (if any).
- Fixed discount reduces the transaction total; the total cannot go below zero.
- If a minimum transaction amount is set, the promo cannot be applied to transactions below that threshold — the cashier sees a clear error.
- Expired promos are automatically excluded from the cashier's list without requiring manual deactivation.

---

### 3.8 Export Reports to CSV

The Owner can export any report view (revenue summary, best-selling products, transactions per cashier) to a CSV file for the selected period.

**Acceptance criteria:**
- Each report type exports to a distinct CSV with appropriate column headers.
- Exported data matches exactly what is displayed on screen for the same period.
- CSV file downloads directly to the browser without requiring a separate step.
- Export works correctly for periods with zero transactions (outputs headers only, no data rows).

---

## 4. Cashier Features

### 4.1 Shift Management

A Cashier must open a shift before they can create transactions. When opening a shift, the cashier declares their starting cash amount. When closing a shift, the system presents a summary: number of transactions, total cash received, total non-cash payments, and total revenue. The cashier confirms and closes the shift.

Only one shift per cashier can be open at a time. A cashier cannot open a new shift if a previous shift is still open.

**Acceptance criteria:**
- Cashier cannot access the order screen until a shift is opened.
- Opening a shift requires a valid starting cash amount (non-negative number).
- Closing a shift displays an accurate summary of all transactions completed during that shift.
- After closing, the cashier is redirected to the shift open screen; historical shift data is preserved.
- A cashier with an open shift from a previous session (e.g., forgot to close) can still close it; the system does not auto-close shifts.

---

### 4.2 Create New Transaction

The cashier selects a table, then adds products to the order. For each product they set the quantity. The order screen shows product name, price, and current stock. Products with zero stock are shown as unavailable and cannot be added. The cashier can adjust quantities or remove items before checkout.

**Acceptance criteria:**
- Cashier must select a table before adding products.
- Products are filterable by category.
- Adding a product with zero stock is blocked with a clear message.
- Quantity can be increased, decreased, or removed from the order before payment.
- Order subtotal updates in real time as items are added or removed.
- A transaction is created with status `pending` when the cashier initiates checkout.

---

### 4.3 Apply Active Promo

Before checkout, the cashier can select one active promo to apply to the transaction. The discount is shown as a line item on the order summary. Only one promo can be applied per transaction.

**Acceptance criteria:**
- Promo selection shows only currently active promos.
- Applying a promo shows the discount amount and updated total.
- If the transaction total does not meet the promo's minimum amount, application is rejected with a clear message.
- The cashier can remove a promo and the total reverts to the original subtotal.
- The applied promo is recorded permanently on the transaction; if the promo is later deactivated, it does not affect completed transactions.

---

### 4.4 Checkout with Midtrans Snap

When the cashier initiates payment, the system creates a Midtrans Snap payment request and displays the Snap payment popup to the customer. The system polls or uses a callback to detect payment completion. Once Midtrans confirms payment:

- Transaction status is updated to `paid`.
- Stock quantities are decremented for all items in the order.
- The transaction is associated with the cashier's current shift.

If payment is cancelled or fails, the transaction remains in `pending` status and can be retried.

**Acceptance criteria:**
- Midtrans Snap popup opens with the correct transaction amount.
- Successful payment transitions the transaction to `paid`; stock is decremented exactly once.
- Failed or cancelled payment leaves the transaction in `pending`; no stock decrement occurs.
- Duplicate payment notifications from Midtrans (e.g., webhook retry) do not result in double stock deduction or double status update (idempotent handling).
- A paid transaction cannot be modified.

---

### 4.5 View Today's Shift Transaction History

The cashier can view a list of all transactions completed during their current open shift. Each entry shows: transaction ID, table, total amount, promo applied (if any), payment time, and status.

**Acceptance criteria:**
- Only transactions from the cashier's current shift are shown.
- Transactions from other cashiers or previous shifts are not visible.
- The list updates to include new transactions without requiring a page reload (or with a manual refresh button at minimum).
- Pending (unpaid) transactions are also shown with their status clearly labeled.

---

## 5. Business Rules

1. **Authentication:** All users must be authenticated. Sessions expire after a configurable idle timeout. Cashiers and Owners have separate login endpoints and are redirected to their respective interfaces on login.

2. **Role enforcement:** API endpoints are protected by role-based middleware. A cashier token cannot access any owner-only endpoint, and vice versa.

3. **Shift gate:** Any transaction creation or checkout API call from a cashier without an open shift is rejected with HTTP 403.

4. **Stock integrity:** Stock cannot go negative. The system checks stock availability at the time of checkout (not at item selection) and rejects the transaction if stock is insufficient for any line item.

5. **Idempotent payments:** Midtrans webhook handlers must check transaction status before processing. If a transaction is already `paid`, subsequent webhook calls for the same transaction are acknowledged and ignored without side effects.

6. **Soft deletes:** Products, categories, and cashier accounts are never hard-deleted. Soft-deleted records are excluded from all application queries but remain in the database to preserve referential integrity for historical transactions and reports.

7. **Single promo per transaction:** Only one promo may be applied per transaction. The promo is validated at the time of application; a promo that becomes inactive after being applied but before payment is still honoured for that transaction.

8. **Transaction immutability:** Once a transaction reaches `paid` status, no fields (items, quantities, promo, table) can be modified.

9. **Report data source:** All revenue and sales reports are computed from `paid` transactions only. Cancelled or pending transactions are excluded.

10. **Cashier account deactivation:** A deactivated cashier is immediately blocked from logging in. If they are currently logged in with a valid session, their session is invalidated on the next API call.

---

## 6. Out of Scope (v1.0)

The following features are explicitly not part of this version:

- **Multi-location support** — the system operates for a single cafe location only.
- **Kitchen display system (KDS)** — no order routing to kitchen screens.
- **Loyalty / membership program** — no customer accounts, points, or membership tiers.
- **Inventory ingredient tracking** — stock is tracked at the product level, not at the ingredient level (no recipe-based deduction).
- **Online ordering / delivery integration** — no third-party delivery platform connectors (GrabFood, GoFood, etc.).
- **Split payments** — each transaction has a single payment method via Midtrans.
- **Refunds and voids** — post-payment reversals are not handled in this version.
- **Tax and service charge configuration** — no configurable tax rate or service charge applied to transactions.
- **Printer / receipt integration** — no thermal printer support or receipt generation.
- **Mobile app** — the system is web-browser only; no iOS or Android native app.
- **Multi-branch reporting** — all reports are scoped to the single location.
