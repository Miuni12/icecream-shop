# Ice Cream Shop API

A REST API for managing an ice cream shop's inventory and customer records. It uses JWT authentication and role-based access control for three user roles: Owner, Manager, and Cashier.

## Tech Stack

- Go 1.21+
- Fiber v2
- GORM
- SQLite
- JWT (HS256)
- bcrypt

## Getting Started

### 1. Install Go

Install Go 1.21 or later from [the official Go website](https://go.dev/dl/).

### 2. Clone the repository

```bash
git clone https://github.com/miuni12/icecream-shop.git
cd icecream-shop
```

### 3. Configure environment variables

Create a `.env` file from the example:

```powershell
Copy-Item .env.example .env
```

Update the values in `.env` if needed.

| Variable | Description | Default |
|---|---|---|
| `PORT` | Port used by the server | `3000` |
| `DB_PATH` | SQLite database file path | `shop.db` |
| `JWT_SECRET` | Secret used to sign JWTs | Required |

### 4. Start the server

```bash
go run ./cmd/server
```

The database is created, migrated, and seeded with sample data when the server starts.

### 5. Run the tests

```bash
go test ./...
```

Check whether the server is running:

`GET http://localhost:3000/health`

## Test Accounts

All three accounts use the password `password123`.

| Role | Email |
|---|---|
| Owner | owner@shop.com |
| Manager | manager@shop.com |
| Cashier | cashier@shop.com |

The database also includes sample categories, inventory items, and a customer record.

## Role Permissions

| Action | Owner | Manager | Cashier |
|---|---|---|---|
| View items and customers | Yes | Yes | Yes |
| Add items | Yes | Yes | No |
| Edit items | Yes | Yes | Yes |
| Delete items | Yes | No | No |
| Add and edit customers | Yes | Yes | Yes |
| Delete customers | Yes | No | No |

Access permissions are checked on the backend using authentication and role-based middleware.

## API Endpoints

All endpoints require a valid bearer token except the login and health endpoints.

### Authentication

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/auth/login` | Log in and receive a JWT |
| POST | `/api/auth/logout` | Log out and revoke the current token |

### Inventory Items

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/items` | Get all items |
| GET | `/api/items?search=choc` | Search items by name |
| GET | `/api/items/:id` | Get an item by ID |
| POST | `/api/items` | Add a new item |
| PUT | `/api/items/:id` | Update an item |
| PATCH | `/api/items/:id` | Update selected fields |
| DELETE | `/api/items/:id` | Delete an item |

Example item request:

```json
{
  "name": "Mint Scoop",
  "category_id": 1,
  "price": 2.5,
  "quantity": 50,
  "minimum_stock_level": 5,
  "is_active": true
}
```

### Customers

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/customers` | Get all customers |
| GET | `/api/customers?search=jane` | Search by name or email |
| GET | `/api/customers/:id` | Get a customer by ID |
| POST | `/api/customers` | Add a customer |
| PUT | `/api/customers/:id` | Update a customer |
| PATCH | `/api/customers/:id` | Update selected fields |
| DELETE | `/api/customers/:id` | Delete a customer |

Example customer request:

```json
{
  "name": "Jane Perera",
  "email": "jane@example.com",
  "phone": "0771234567",
  "address": "Colombo"
}
```

## Error Responses

The API returns appropriate HTTP status codes for common errors:

- `400 Bad Request` for invalid input
- `401 Unauthorized` for missing or invalid authentication
- `403 Forbidden` when a user does not have permission
- `404 Not Found` when a record does not exist
- `409 Conflict` when a record conflicts with existing data

Error messages are returned in this format:

```json
{
  "error": "message"
}
```
