# API Contract
## Terenjit Cafe — Point of Sale System

**Version:** 1.0  
**Date:** 2026-05-12

---

## 1. Base URL and Versioning

```
Base URL: /api/v1
```

All endpoints are prefixed with `/api/v1`. Versioning is path-based. Future breaking changes will use `/api/v2`.

### URL Prefix Convention

| Prefix | Middleware Applied | Who Accesses |
|---|---|---|
| `/api/v1/auth/...` | None | Public |
| `/api/v1/owner/...` | Owner role JWT middleware | Owner only |
| `/api/v1/cashier/...` | Cashier role JWT middleware | Cashier only |
| `/api/v1/webhooks/...` | Midtrans signature verification | External (Midtrans) |

The `/owner/` and `/cashier/` prefixes are structural — all routes under each prefix share a single middleware group in Go, so role enforcement is applied at the router level rather than per-handler.

---

## 2. Authentication

All protected endpoints require a JWT access token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

Tokens are obtained via the login endpoint. Each token carries the user's `id`, `role`, and expiry. The server rejects requests with missing, expired, or malformed tokens with `401 Unauthorized`. Deactivated accounts receive `403 Forbidden` even with a valid token.

**Role labels used in this document:**
- **Owner** — authenticated user with `role = owner`
- **Cashier** — authenticated user with `role = cashier`
- **Public** — no authentication required

---

## 3. Standard Response Format

All responses use a consistent JSON envelope.

### Success — Single Object
```json
{
  "success": true,
  "message": "Product retrieved successfully",
  "data": { }
}
```

### Success — List with Pagination
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 85,
    "total_pages": 5
  }
}
```

### Success — No Body (e.g. DELETE)
```json
{
  "success": true,
  "message": "Category deleted successfully"
}
```

### Error — General
```json
{
  "success": false,
  "message": "Product not found"
}
```

### Error — Validation
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "price": "price must be greater than 0",
    "category_id": "category_id is required"
  }
}
```

---

## 4. HTTP Status Codes

| Code | Meaning |
|---|---|
| 200 | OK — successful GET and PUT |
| 201 | Created — successful POST (create) |
| 400 | Bad Request — malformed or invalid request |
| 401 | Unauthorized — missing or expired token |
| 403 | Forbidden — valid token but no access |
| 404 | Not Found — data not found |
| 422 | Unprocessable Entity — validation failed or business constraint violation |
| 429 | Too Many Requests — rate limit exceeded |
| 500 | Internal Server Error — unhandled server error |

---

## 5. Endpoints

---

### 5.1 Auth

---

#### `POST /api/v1/auth/login`
**Access:** Public

Authenticates a user and returns a JWT access token and a refresh token. The `role` field in the response determines which interface the frontend loads.

**Request Body:**
```json
{
  "email": "budi@terenjitcafe.com",
  "password": "secret123"
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4...",
    "user": {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "Budi Santoso",
      "email": "budi@terenjitcafe.com",
      "role": "cashier"
    }
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Invalid email or password |
| 403 | Account is deactivated |

---

#### `POST /api/v1/auth/refresh`
**Access:** Public

Issues a new access token using a valid refresh token. Used when the access token expires without requiring the user to log in again.

**Request Body:**
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "bmV3UmVmcmVzaFRva2Vu..."
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Refresh token is invalid or expired |
| 403 | Account is deactivated |

---

#### `POST /api/v1/auth/logout`
**Access:** Owner, Cashier

Revokes the current access and refresh tokens server-side.

**Request Body:** _(none)_

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |

---

### 5.2 Categories

All routes under `/api/v1/owner/` are protected by the owner role middleware.

---

#### `GET /api/v1/owner/categories`
**Access:** Owner

Returns all non-deleted categories.

**Query Params:**
| Param | Type | Required | Description |
|---|---|---|---|
| page | integer | No | Default 1 |
| per_page | integer | No | Default 20, max 100 |

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Categories retrieved successfully",
  "data": [
    {
      "id": "cat-uuid-001",
      "name": "Coffee",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": "cat-uuid-002",
      "name": "Food",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 4,
    "total_pages": 1
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `POST /api/v1/owner/categories`
**Access:** Owner

Creates a new product category.

**Request Body:**
```json
{
  "name": "Merchandise"
}
```

**Success Response `201`:**
```json
{
  "success": true,
  "message": "Category created successfully",
  "data": {
    "id": "cat-uuid-005",
    "name": "Merchandise",
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `PUT /api/v1/owner/categories/:id`
**Access:** Owner

Renames an existing category.

**Request Body:**
```json
{
  "name": "Non-Coffee"
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Category updated successfully",
  "data": {
    "id": "cat-uuid-002",
    "name": "Non-Coffee",
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-05-12T10:05:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Category not found |

---

#### `DELETE /api/v1/owner/categories/:id`
**Access:** Owner

Soft-deletes a category. Fails if the category has any available (`is_available = true`, `deleted_at = null`) products assigned to it.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Category deleted successfully"
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Category not found |
| 422 | Cannot delete category with active products |

---

### 5.3 Products

---

#### `GET /api/v1/owner/products`
**Access:** Owner

Returns all non-deleted products including unavailable ones. Supports filtering and search.

**Query Params:**
| Param | Type | Required | Description |
|---|---|---|---|
| page | integer | No | Default 1 |
| per_page | integer | No | Default 20, max 100 |
| category_id | string | No | Filter by category UUID |
| search | string | No | Partial match on product name |
| is_available | boolean | No | Filter by availability status |

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": "prod-uuid-001",
      "category_id": "cat-uuid-001",
      "category_name": "Coffee",
      "name": "Kopi Susu",
      "description": "Espresso with fresh milk",
      "price": 30000,
      "image_url": "https://storage.example.com/products/kopi-susu.jpg",
      "is_available": true,
      "current_stock": 50,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-05-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 35,
    "total_pages": 2
  }
}
```

> `current_stock` is computed as `SUM(quantity)` from `stock_movements` for the product.

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `GET /api/v1/owner/products/:id`
**Access:** Owner

Returns a single product by ID including unavailable and soft-deleted ones.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Product retrieved successfully",
  "data": {
    "id": "prod-uuid-001",
    "category_id": "cat-uuid-001",
    "category_name": "Coffee",
    "name": "Kopi Susu",
    "description": "Espresso with fresh milk",
    "price": 30000,
    "image_url": "https://storage.example.com/products/kopi-susu.jpg",
    "is_available": true,
    "current_stock": 50,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-05-01T00:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Product not found |

---

#### `POST /api/v1/owner/products`
**Access:** Owner

Creates a new product. `image_url` is the URL of a file already uploaded to storage; the client uploads to storage directly and passes back the URL.

**Request Body:**
```json
{
  "category_id": "cat-uuid-001",
  "name": "Kopi Susu",
  "description": "Espresso with fresh milk",
  "price": 30000,
  "image_url": "https://storage.example.com/products/kopi-susu.jpg",
  "is_available": true
}
```

**Success Response `201`:**
```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": "prod-uuid-001",
    "category_id": "cat-uuid-001",
    "category_name": "Coffee",
    "name": "Kopi Susu",
    "description": "Espresso with fresh milk",
    "price": 30000,
    "image_url": "https://storage.example.com/products/kopi-susu.jpg",
    "is_available": true,
    "current_stock": 0,
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Category not found |

---

#### `PUT /api/v1/owner/products/:id`
**Access:** Owner

Fully updates a product. All editable fields must be supplied, including `is_available` to toggle product visibility.

**Request Body:**
```json
{
  "category_id": "cat-uuid-001",
  "name": "Kopi Susu Segar",
  "description": "Double shot espresso with fresh full-cream milk",
  "price": 32000,
  "image_url": "https://storage.example.com/products/kopi-susu-v2.jpg",
  "is_available": false
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Product updated successfully",
  "data": {
    "id": "prod-uuid-001",
    "category_id": "cat-uuid-001",
    "category_name": "Coffee",
    "name": "Kopi Susu Segar",
    "description": "Double shot espresso with fresh full-cream milk",
    "price": 32000,
    "image_url": "https://storage.example.com/products/kopi-susu-v2.jpg",
    "is_available": false,
    "current_stock": 50,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-05-12T10:10:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Product not found |
| 404 | Category not found |

---

#### `DELETE /api/v1/owner/products/:id`
**Access:** Owner

Soft-deletes a product. The product disappears from all views; existing `order_items` referencing it remain intact.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Product deleted successfully"
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Product not found |

---

### 5.4 Stock

---

#### `GET /api/v1/owner/products/:id/stock`
**Access:** Owner

Returns the current computed stock level for a product.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Stock retrieved successfully",
  "data": {
    "product_id": "prod-uuid-001",
    "product_name": "Kopi Susu",
    "current_stock": 50
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Product not found |

---

#### `POST /api/v1/owner/products/:id/stock/adjustment`
**Access:** Owner

Records a manual stock adjustment for the product. Positive `quantity` adds stock; negative `quantity` reduces it. The `note` field is required.

**Request Body:**
```json
{
  "quantity": -5,
  "note": "Expired stock removed after daily check"
}
```

**Success Response `201`:**
```json
{
  "success": true,
  "message": "Stock adjustment recorded successfully",
  "data": {
    "id": "mov-uuid-003",
    "product_id": "prod-uuid-001",
    "product_name": "Kopi Susu",
    "type": "adjustment",
    "quantity": -5,
    "note": "Expired stock removed after daily check",
    "created_by": {
      "id": "user-uuid-001",
      "name": "Owner Name"
    },
    "current_stock": 45,
    "created_at": "2026-05-12T11:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Product not found |
| 422 | Adjustment would bring stock below zero |

---

#### `GET /api/v1/owner/products/:id/stock/movements`
**Access:** Owner

Returns the stock movement history for a product, newest first.

**Query Params:**
| Param | Type | Required | Description |
|---|---|---|---|
| page | integer | No | Default 1 |
| per_page | integer | No | Default 20, max 100 |
| type | string | No | Filter by type: `purchase`, `sale`, `adjustment`, `return` |

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Stock movements retrieved successfully",
  "data": [
    {
      "id": "mov-uuid-001",
      "product_id": "prod-uuid-001",
      "type": "purchase",
      "quantity": 100,
      "note": "Restock from supplier",
      "created_by": {
        "id": "user-uuid-001",
        "name": "Owner Name"
      },
      "created_at": "2026-05-10T09:00:00Z"
    },
    {
      "id": "mov-uuid-002",
      "product_id": "prod-uuid-001",
      "type": "sale",
      "quantity": -2,
      "note": "Sale — order TCAFE-20260512-001",
      "created_by": {
        "id": "user-uuid-002",
        "name": "Budi Santoso"
      },
      "created_at": "2026-05-12T08:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 12,
    "total_pages": 1
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Product not found |

---

### 5.5 Tables

---

#### `GET /api/v1/owner/tables`
**Access:** Owner

Returns all tables including inactive ones.

**Query Params:**
| Param | Type | Required | Description |
|---|---|---|---|
| is_active | boolean | No | Filter by active status |

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Tables retrieved successfully",
  "data": [
    {
      "id": "tbl-uuid-001",
      "name": "Table 1",
      "is_active": true,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": "tbl-uuid-002",
      "name": "Table 2",
      "is_active": false,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-03-10T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `POST /api/v1/owner/tables`
**Access:** Owner

Adds a new table.

**Request Body:**
```json
{
  "name": "VIP Room"
}
```

**Success Response `201`:**
```json
{
  "success": true,
  "message": "Table created successfully",
  "data": {
    "id": "tbl-uuid-010",
    "name": "VIP Room",
    "is_active": true,
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `PUT /api/v1/owner/tables/:id`
**Access:** Owner

Renames a table and/or changes its active status.

**Request Body:**
```json
{
  "name": "Table 1 — Window Seat",
  "is_active": true
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Table updated successfully",
  "data": {
    "id": "tbl-uuid-001",
    "name": "Table 1 — Window Seat",
    "is_active": true,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-05-12T10:05:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Table not found |
| 422 | Cannot deactivate table with an open order |

---

#### `DELETE /api/v1/owner/tables/:id`
**Access:** Owner

Deactivates a table, removing it from the cashier's table selector. Cannot be performed while the table has an open (non-paid, non-cancelled) order assigned to it.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Table deactivated successfully"
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Table not found |
| 422 | Cannot deactivate table with an open order |

---

### 5.6 Cashier Management

---

#### `GET /api/v1/owner/cashiers`
**Access:** Owner

Lists all cashier accounts. Includes deactivated cashiers.

**Query Params:**
| Param | Type | Required | Description |
|---|---|---|---|
| page | integer | No | Default 1 |
| per_page | integer | No | Default 20 |
| is_active | boolean | No | Filter by active status |

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Cashiers retrieved successfully",
  "data": [
    {
      "id": "user-uuid-002",
      "name": "Budi Santoso",
      "email": "budi@terenjitcafe.com",
      "is_active": true,
      "created_at": "2026-01-15T00:00:00Z",
      "updated_at": "2026-01-15T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 3,
    "total_pages": 1
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `POST /api/v1/owner/cashiers`
**Access:** Owner

Creates a new cashier account. The account is immediately active and ready to log in.

**Request Body:**
```json
{
  "name": "Siti Rahayu",
  "email": "siti@terenjitcafe.com",
  "password": "initialpassword123"
}
```

**Success Response `201`:**
```json
{
  "success": true,
  "message": "Cashier account created successfully",
  "data": {
    "id": "user-uuid-005",
    "name": "Siti Rahayu",
    "email": "siti@terenjitcafe.com",
    "is_active": true,
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |
| 422 | Email already in use |

---

#### `PUT /api/v1/owner/cashiers/:id`
**Access:** Owner

Updates a cashier's name and/or email. Password changes are not handled here.

**Request Body:**
```json
{
  "name": "Siti Rahayu Putri",
  "email": "siti.putri@terenjitcafe.com"
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Cashier updated successfully",
  "data": {
    "id": "user-uuid-005",
    "name": "Siti Rahayu Putri",
    "email": "siti.putri@terenjitcafe.com",
    "is_active": true,
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T11:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Cashier not found |
| 422 | Email already in use |

---

#### `PATCH /api/v1/owner/cashiers/:id/toggle-status`
**Access:** Owner

Toggles the cashier's `is_active` status. Activating restores login access; deactivating blocks it immediately. Accounts cannot be hard-deleted.

**Request Body:** _(none)_

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Cashier status updated successfully",
  "data": {
    "id": "user-uuid-005",
    "name": "Siti Rahayu Putri",
    "is_active": false,
    "updated_at": "2026-05-12T12:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Cashier not found |

---

### 5.7 Promotions

---

#### `GET /api/v1/owner/promos`
**Access:** Owner

Returns all promotions. Supports filtering by active status and date.

**Query Params:**
| Param | Type | Required | Description |
|---|---|---|---|
| page | integer | No | Default 1 |
| per_page | integer | No | Default 20 |
| is_active | boolean | No | Filter by the `is_active` toggle |

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Promotions retrieved successfully",
  "data": [
    {
      "id": "promo-uuid-001",
      "name": "Happy Hour",
      "description": "10% off all drinks",
      "discount_type": "percentage",
      "discount_percentage": 10.00,
      "discount_fixed_amount": null,
      "min_transaction_amount": null,
      "valid_from": "2026-05-01",
      "valid_until": "2026-05-31",
      "is_active": true,
      "created_at": "2026-04-25T00:00:00Z",
      "updated_at": "2026-04-25T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 3,
    "total_pages": 1
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `POST /api/v1/owner/promos`
**Access:** Owner

Creates a new promotion. For `discount_type = percentage`, provide `discount_percentage` and omit `discount_fixed_amount`, and vice versa.

**Request Body (percentage):**
```json
{
  "name": "Happy Hour",
  "description": "10% off all drinks",
  "discount_type": "percentage",
  "discount_percentage": 10.00,
  "min_transaction_amount": null,
  "valid_from": "2026-06-01",
  "valid_until": "2026-06-30",
  "is_active": true
}
```

**Request Body (fixed amount):**
```json
{
  "name": "Weekend Special",
  "description": "Rp 10,000 off orders above Rp 50,000",
  "discount_type": "fixed",
  "discount_fixed_amount": 10000,
  "min_transaction_amount": 50000,
  "valid_from": "2026-06-01",
  "valid_until": "2026-06-30",
  "is_active": true
}
```

**Success Response `201`:**
```json
{
  "success": true,
  "message": "Promotion created successfully",
  "data": {
    "id": "promo-uuid-002",
    "name": "Happy Hour",
    "description": "10% off all drinks",
    "discount_type": "percentage",
    "discount_percentage": 10.00,
    "discount_fixed_amount": null,
    "min_transaction_amount": null,
    "valid_from": "2026-06-01",
    "valid_until": "2026-06-30",
    "is_active": true,
    "created_at": "2026-05-12T10:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |
| 422 | valid_until must be after valid_from |
| 422 | discount_percentage is required when discount_type is percentage |

---

#### `PUT /api/v1/owner/promos/:id`
**Access:** Owner

Fully updates an existing promotion.

**Request Body:** _(same structure as POST)_

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Promotion updated successfully",
  "data": {
    "id": "promo-uuid-001",
    "name": "Happy Hour Extended",
    "discount_type": "percentage",
    "discount_percentage": 15.00,
    "discount_fixed_amount": null,
    "min_transaction_amount": null,
    "valid_from": "2026-06-01",
    "valid_until": "2026-07-31",
    "is_active": true,
    "created_at": "2026-04-25T00:00:00Z",
    "updated_at": "2026-05-12T10:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Promotion not found |
| 422 | valid_until must be after valid_from |

---

#### `DELETE /api/v1/owner/promos/:id`
**Access:** Owner

Deletes a promotion. Past orders that used this promo retain their recorded `promo_id` and `discount_amount` for historical accuracy.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Promotion deleted successfully"
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Promotion not found |

---

### 5.8 Reports

All report endpoints operate on `paid` orders only within the requested period.

**Common Query Params for all report endpoints:**

| Param | Type | Required | Description |
|---|---|---|---|
| period | string | Yes | `daily`, `weekly`, or `monthly` |
| date | string | Yes | `daily`: `YYYY-MM-DD` · `weekly`: any date within the target week · `monthly`: `YYYY-MM` |

---

#### `GET /api/v1/owner/reports/revenue`
**Access:** Owner

Returns total revenue and order count for the period, broken down by day.

**Example:** `GET /api/v1/owner/reports/revenue?period=weekly&date=2026-05-12`

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Revenue report retrieved successfully",
  "data": {
    "period": "weekly",
    "start_date": "2026-05-11",
    "end_date": "2026-05-17",
    "total_revenue": 14500000,
    "total_orders": 265,
    "breakdown": [
      { "date": "2026-05-11", "revenue": 1800000, "orders": 33 },
      { "date": "2026-05-12", "revenue": 2500000, "orders": 45 },
      { "date": "2026-05-13", "revenue": 2200000, "orders": 40 },
      { "date": "2026-05-14", "revenue": 1900000, "orders": 35 },
      { "date": "2026-05-15", "revenue": 2100000, "orders": 38 },
      { "date": "2026-05-16", "revenue": 2500000, "orders": 46 },
      { "date": "2026-05-17", "revenue": 1500000, "orders": 28 }
    ]
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Invalid period or date format |
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `GET /api/v1/owner/reports/top-products`
**Access:** Owner

Returns products ranked by total units sold in the period.

**Query Params (in addition to `period` and `date`):**

| Param | Type | Required | Description |
|---|---|---|---|
| limit | integer | No | Number of products to return, default 10, max 50 |

**Example:** `GET /api/v1/owner/reports/top-products?period=monthly&date=2026-05&limit=5`

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Top products report retrieved successfully",
  "data": {
    "period": "monthly",
    "start_date": "2026-05-01",
    "end_date": "2026-05-31",
    "products": [
      {
        "rank": 1,
        "product_id": "prod-uuid-001",
        "product_name": "Kopi Susu",
        "category_name": "Coffee",
        "total_qty_sold": 480,
        "total_revenue": 14400000
      },
      {
        "rank": 2,
        "product_id": "prod-uuid-003",
        "product_name": "Croissant",
        "category_name": "Food",
        "total_qty_sold": 210,
        "total_revenue": 6300000
      }
    ]
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Invalid period or date format |
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `GET /api/v1/owner/reports/cashier-summary`
**Access:** Owner

Returns order count and total revenue per cashier for the period.

**Example:** `GET /api/v1/owner/reports/cashier-summary?period=daily&date=2026-05-12`

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Cashier summary report retrieved successfully",
  "data": {
    "period": "daily",
    "date": "2026-05-12",
    "cashiers": [
      {
        "cashier_id": "user-uuid-002",
        "cashier_name": "Budi Santoso",
        "total_orders": 28,
        "total_revenue": 1540000
      },
      {
        "cashier_id": "user-uuid-003",
        "cashier_name": "Siti Rahayu",
        "total_orders": 17,
        "total_revenue": 960000
      }
    ]
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Invalid period or date format |
| 401 | Unauthorized |
| 403 | Forbidden |

---

#### `GET /api/v1/owner/reports/export`
**Access:** Owner

Downloads a report as a CSV file. The `type` param selects which report to export. Uses the same `period` and `date` params as the JSON report endpoints.

**Query Params:**

| Param | Type | Required | Description |
|---|---|---|---|
| type | string | Yes | `revenue`, `top-products`, or `cashier-summary` |
| period | string | Yes | `daily`, `weekly`, or `monthly` |
| date | string | Yes | Same format as the corresponding JSON report |
| limit | integer | No | For `type=top-products` only; default 10, max 50 |

**Response Headers:**
```
Content-Type: text/csv; charset=utf-8
Content-Disposition: attachment; filename="revenue_report_2026-05.csv"
```

**CSV Format — `type=revenue`:**
```
Date,Revenue (IDR),Total Orders
2026-05-01,1800000,33
2026-05-02,2100000,39
```

**CSV Format — `type=top-products`:**
```
Rank,Product Name,Category,Total Qty Sold,Total Revenue (IDR)
1,Kopi Susu,Coffee,480,14400000
2,Croissant,Food,210,6300000
```

**CSV Format — `type=cashier-summary`:**
```
Cashier Name,Total Orders,Total Revenue (IDR)
Budi Santoso,28,1540000
Siti Rahayu,17,960000
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Invalid type, period, or date format |
| 401 | Unauthorized |
| 403 | Forbidden |

---

### 5.9 Shifts

All routes under `/api/v1/cashier/` are protected by the cashier role middleware.

---

#### `POST /api/v1/cashier/shifts/open`
**Access:** Cashier

Opens a new shift. Fails if the cashier already has an open shift.

**Request Body:**
```json
{
  "opening_cash": 500000
}
```

**Success Response `201`:**
```json
{
  "success": true,
  "message": "Shift opened successfully",
  "data": {
    "id": "shift-uuid-001",
    "cashier_id": "user-uuid-002",
    "cashier_name": "Budi Santoso",
    "opening_cash": 500000,
    "closing_cash": null,
    "total_sales": 0,
    "status": "open",
    "opened_at": "2026-05-12T07:00:00Z",
    "closed_at": null
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed — opening_cash must be a non-negative number |
| 401 | Unauthorized |
| 403 | Forbidden |
| 422 | Cashier already has an open shift |

---

#### `POST /api/v1/cashier/shifts/close`
**Access:** Cashier

Closes the cashier's current open shift and returns the full shift summary.

**Request Body:**
```json
{
  "closing_cash": 1750000
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Shift closed successfully",
  "data": {
    "id": "shift-uuid-001",
    "cashier_id": "user-uuid-002",
    "cashier_name": "Budi Santoso",
    "opening_cash": 500000,
    "closing_cash": 1750000,
    "total_sales": 2850000,
    "total_orders": 52,
    "status": "closed",
    "opened_at": "2026-05-12T07:00:00Z",
    "closed_at": "2026-05-12T15:00:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed — closing_cash is required |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | No open shift found |
| 422 | Shift is already closed |

---

#### `GET /api/v1/cashier/shifts/current`
**Access:** Cashier

Returns the cashier's currently open shift. Used on login to check if a shift is already in progress before showing the order screen.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Current shift retrieved successfully",
  "data": {
    "id": "shift-uuid-001",
    "cashier_id": "user-uuid-002",
    "cashier_name": "Budi Santoso",
    "opening_cash": 500000,
    "closing_cash": null,
    "total_sales": 1250000,
    "status": "open",
    "opened_at": "2026-05-12T07:00:00Z",
    "closed_at": null
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | No active shift found |

---

### 5.10 Orders

> **Routing note:** `GET /api/v1/cashier/orders/history` must be registered before `GET /api/v1/cashier/orders/:id` in the Go router so that `history` is not interpreted as an `:id` path parameter.

---

#### `POST /api/v1/cashier/orders`
**Access:** Cashier

Creates a new order in `draft` status. The cashier must have an open shift. `table_id` is optional — omit it for takeaway orders. Items can be included in the creation request or added separately via `POST /items`.

**Request Body:**
```json
{
  "table_id": "tbl-uuid-003",
  "items": [
    { "product_id": "prod-uuid-001", "quantity": 2 },
    { "product_id": "prod-uuid-005", "quantity": 1 }
  ]
}
```

**Success Response `201`:**
```json
{
  "success": true,
  "message": "Order created successfully",
  "data": {
    "id": "order-uuid-001",
    "shift_id": "shift-uuid-001",
    "cashier_id": "user-uuid-002",
    "cashier_name": "Budi Santoso",
    "table_id": "tbl-uuid-003",
    "table_name": "Table 3",
    "promo_id": null,
    "status": "draft",
    "items": [
      {
        "id": "item-uuid-001",
        "product_id": "prod-uuid-001",
        "product_name": "Kopi Susu",
        "quantity": 2,
        "price": 30000,
        "subtotal": 60000
      },
      {
        "id": "item-uuid-002",
        "product_id": "prod-uuid-005",
        "product_name": "Croissant",
        "quantity": 1,
        "price": 25000,
        "subtotal": 25000
      }
    ],
    "subtotal": 85000,
    "discount_amount": 0,
    "total": 85000,
    "payment_method": null,
    "midtrans_order_id": null,
    "created_at": "2026-05-12T08:30:00Z",
    "updated_at": "2026-05-12T08:30:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | No active shift — open a shift first |
| 404 | Table not found or inactive |
| 404 | Product not found or unavailable |
| 422 | Insufficient stock for product: Kopi Susu |

---

#### `GET /api/v1/cashier/orders/:id`
**Access:** Cashier

Returns a single order with full item detail. Cashiers can only retrieve orders from their own current shift.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Order retrieved successfully",
  "data": {
    "id": "order-uuid-001",
    "shift_id": "shift-uuid-001",
    "cashier_id": "user-uuid-002",
    "cashier_name": "Budi Santoso",
    "table_id": "tbl-uuid-003",
    "table_name": "Table 3",
    "promo_id": "promo-uuid-001",
    "promo_name": "Happy Hour",
    "status": "paid",
    "items": [
      {
        "id": "item-uuid-001",
        "product_id": "prod-uuid-001",
        "product_name": "Kopi Susu",
        "quantity": 2,
        "price": 30000,
        "subtotal": 60000
      }
    ],
    "subtotal": 85000,
    "discount_amount": 8500,
    "total": 76500,
    "payment_method": "gopay",
    "midtrans_order_id": "TCAFE-20260512-001",
    "created_at": "2026-05-12T08:30:00Z",
    "updated_at": "2026-05-12T08:45:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden — order does not belong to current shift |
| 404 | Order not found |

---

#### `POST /api/v1/cashier/orders/:id/items`
**Access:** Cashier

Adds one or more new items to an existing `draft` order. If a product already exists in the order, its quantity is incremented.

**Request Body:**
```json
{
  "items": [
    { "product_id": "prod-uuid-007", "quantity": 2 }
  ]
}
```

**Success Response `201`:**
```json
{
  "success": true,
  "message": "Items added successfully",
  "data": {
    "id": "order-uuid-001",
    "items": [
      {
        "id": "item-uuid-001",
        "product_id": "prod-uuid-001",
        "product_name": "Kopi Susu",
        "quantity": 2,
        "price": 30000,
        "subtotal": 60000
      },
      {
        "id": "item-uuid-003",
        "product_id": "prod-uuid-007",
        "product_name": "Matcha Latte",
        "quantity": 2,
        "price": 35000,
        "subtotal": 70000
      }
    ],
    "subtotal": 130000,
    "discount_amount": 0,
    "total": 130000,
    "updated_at": "2026-05-12T08:33:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed — items list cannot be empty |
| 401 | Unauthorized |
| 403 | Forbidden — order does not belong to current shift |
| 404 | Order not found |
| 404 | Product not found or unavailable |
| 422 | Order cannot be modified — current status: paid |
| 422 | Insufficient stock for product: Matcha Latte |

---

#### `PUT /api/v1/cashier/orders/:id/items/:item_id`
**Access:** Cashier

Updates the quantity of a single item on a `draft` order. To remove an item entirely, use the DELETE endpoint.

**Request Body:**
```json
{
  "quantity": 4
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Item updated successfully",
  "data": {
    "id": "order-uuid-001",
    "items": [
      {
        "id": "item-uuid-001",
        "product_id": "prod-uuid-001",
        "product_name": "Kopi Susu",
        "quantity": 4,
        "price": 30000,
        "subtotal": 120000
      }
    ],
    "subtotal": 120000,
    "discount_amount": 0,
    "total": 120000,
    "updated_at": "2026-05-12T08:34:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed — quantity must be at least 1 |
| 401 | Unauthorized |
| 403 | Forbidden — order does not belong to current shift |
| 404 | Order not found |
| 404 | Item not found in this order |
| 422 | Order cannot be modified — current status: paid |
| 422 | Insufficient stock for product: Kopi Susu |

---

#### `DELETE /api/v1/cashier/orders/:id/items/:item_id`
**Access:** Cashier

Removes a single item from a `draft` order.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Item removed successfully",
  "data": {
    "id": "order-uuid-001",
    "items": [],
    "subtotal": 0,
    "discount_amount": 0,
    "total": 0,
    "updated_at": "2026-05-12T08:35:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden — order does not belong to current shift |
| 404 | Order not found |
| 404 | Item not found in this order |
| 422 | Order cannot be modified — current status: paid |

---

#### `POST /api/v1/cashier/orders/:id/promo`
**Access:** Cashier

Applies an active promo to a `draft` order. Replaces any previously applied promo. The server validates eligibility at apply time.

**Request Body:**
```json
{
  "promo_id": "promo-uuid-001"
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Promo applied successfully",
  "data": {
    "id": "order-uuid-001",
    "promo_id": "promo-uuid-001",
    "promo_name": "Happy Hour",
    "subtotal": 130000,
    "discount_amount": 13000,
    "total": 117000,
    "updated_at": "2026-05-12T08:36:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Validation failed |
| 401 | Unauthorized |
| 403 | Forbidden — order does not belong to current shift |
| 404 | Order not found |
| 404 | Promo not found or not currently active |
| 422 | Order cannot be modified — current status: paid |
| 422 | Order subtotal does not meet minimum transaction amount of Rp 200,000 |

---

#### `DELETE /api/v1/cashier/orders/:id/promo`
**Access:** Cashier

Removes the applied promo from a `draft` order and recalculates the total.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Promo removed successfully",
  "data": {
    "id": "order-uuid-001",
    "promo_id": null,
    "promo_name": null,
    "subtotal": 130000,
    "discount_amount": 0,
    "total": 130000,
    "updated_at": "2026-05-12T08:37:00Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden — order does not belong to current shift |
| 404 | Order not found |
| 422 | No promo is applied to this order |
| 422 | Order cannot be modified — current status: paid |

---

#### `GET /api/v1/cashier/orders/history`
**Access:** Cashier

Returns all orders from the cashier's current open shift, newest first. Includes all statuses.

**Query Params:**
| Param | Type | Required | Description |
|---|---|---|---|
| status | string | No | Filter by status: `draft`, `pending_payment`, `paid`, `cancelled` |
| page | integer | No | Default 1 |
| per_page | integer | No | Default 20 |

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Order history retrieved successfully",
  "data": [
    {
      "id": "order-uuid-001",
      "table_id": "tbl-uuid-003",
      "table_name": "Table 3",
      "status": "paid",
      "subtotal": 130000,
      "discount_amount": 13000,
      "total": 117000,
      "payment_method": "gopay",
      "created_at": "2026-05-12T08:30:00Z",
      "updated_at": "2026-05-12T08:45:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 17,
    "total_pages": 1
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | No active shift |

---

### 5.11 Payments

---

#### `POST /api/v1/cashier/orders/:id/checkout`
**Access:** Cashier

Transitions the order from `draft` to `pending_payment`, creates the payment record, calls Midtrans Snap API, and returns the Snap token. Stock is **not** decremented here — decrement happens only on webhook confirmation.

A final stock availability check is performed before creating the Snap request.

**Request Body:** _(none)_

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Checkout initiated successfully",
  "data": {
    "order_id": "order-uuid-001",
    "midtrans_order_id": "TCAFE-20260512-001",
    "snap_token": "66e4fa55-fdac-4ef9-91a5-d283ea454f14",
    "total": 117000
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden — order does not belong to current shift |
| 404 | Order not found |
| 422 | Order has no items |
| 422 | Order is already in payment or paid |
| 422 | Insufficient stock for product: Kopi Susu |
| 500 | Failed to create Midtrans Snap payment |

---

#### `POST /api/v1/webhooks/midtrans`
**Access:** Public (Midtrans)

Receives payment status notifications from Midtrans. The server verifies the `signature_key` before processing. This endpoint is idempotent — if the order is already `paid`, the notification is acknowledged and ignored.

On successful `settlement`:
1. Payment status → `paid`, `paid_at` set
2. Order status → `paid`, `payment_method` set
3. Stock decremented for each item (one `stock_movements` record per product, `type = sale`)
4. `shifts.total_sales` incremented by `order.total`

**Request Body (sent by Midtrans):**
```json
{
  "transaction_time": "2026-05-12 08:44:32",
  "transaction_status": "settlement",
  "transaction_id": "c7d1f2a0-3b45-4e89-a12c-f3d456789012",
  "order_id": "TCAFE-20260512-001",
  "payment_type": "gopay",
  "gross_amount": "117000.00",
  "signature_key": "abc123hashedvalue..."
}
```

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Webhook received"
}
```

> Midtrans retries on any non-`2xx` response. Always return `200` — even for already-processed or unrecognised events.

**Error Responses:**
| Status | Message |
|---|---|
| 400 | Invalid signature |
| 404 | Order not found |

---

#### `GET /api/v1/cashier/orders/:id/payment-status`
**Access:** Cashier

Returns the current payment status for an order. Used for polling after the Snap popup closes.

**Success Response `200`:**
```json
{
  "success": true,
  "message": "Payment status retrieved successfully",
  "data": {
    "id": "pay-uuid-001",
    "order_id": "order-uuid-001",
    "midtrans_order_id": "TCAFE-20260512-001",
    "midtrans_transaction_id": "c7d1f2a0-3b45-4e89-a12c-f3d456789012",
    "status": "paid",
    "amount": 117000,
    "payment_type": "gopay",
    "paid_at": "2026-05-12T08:44:32Z",
    "created_at": "2026-05-12T08:38:00Z",
    "updated_at": "2026-05-12T08:44:32Z"
  }
}
```

**Error Responses:**
| Status | Message |
|---|---|
| 401 | Unauthorized |
| 403 | Forbidden — order does not belong to current shift |
| 404 | Payment not found for this order |

---

## 6. Error Reference

| Scenario | HTTP Status | Message |
|---|---|---|
| Missing or expired token | 401 | Unauthorized |
| Invalid refresh token | 401 | Refresh token is invalid or expired |
| Valid token, wrong role | 403 | Forbidden |
| Deactivated account, valid token | 403 | Account is deactivated |
| No open shift (cashier action) | 403 | No active shift — open a shift first |
| Resource does not exist | 404 | `<Resource>` not found |
| Duplicate unique field | 422 | `<Field>` already in use |
| Cashier already has open shift | 422 | Cashier already has an open shift |
| Rate limit exceeded | 429 | Too many requests |
| Stock too low | 422 | Insufficient stock for product: `<name>` |
| Modifying a non-draft order | 422 | Order cannot be modified — current status: `<status>` |
| Deleting category with active products | 422 | Cannot delete category with active products |
| Deactivating table with open order | 422 | Cannot deactivate table with an open order |
| Promo minimum not met | 422 | Order subtotal does not meet minimum transaction amount |
| Midtrans Snap creation failure | 500 | Failed to create Midtrans Snap payment |
