# Expense Tracker API

<div align="center">

![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-336791?logo=postgresql)
![REST API](https://img.shields.io/badge/API-REST%20Service-4EAAFF)
![Status](https://img.shields.io/badge/Status-Backend%20API%20Active-2ea44f)

A lightweight backend service for tracking personal income and expenses.

</div>

## Overview

This project is a backend expense tracker built with Go and PostgreSQL. Its REST API supports user registration and login, income and expense transactions, user-defined categories, and date-filtered financial analytics. Protected endpoints use JWT bearer authentication.

This repository is focused on the API layer and data persistence. There is no frontend application or dashboard yet; the project currently provides the server-side foundation for an expense tracking system.

## Features

- Register and log in to a user account
- Create income or expense transactions and assign categories
- Retrieve a single transaction by ID
- Retrieve all transactions in reverse chronological order
- Update an existing transaction
- Delete a transaction
- Create, list, rename, and delete categories
- Retrieve date-filtered income and expense analytics
- Input validation for transaction amounts, kinds, and category names
- PostgreSQL-backed persistence
- JSON-based API responses
- Lightweight HTTP server built with Go's standard library

## Tech Stack

- Go
- PostgreSQL
- lib/pq driver
- Standard library HTTP server

## Project Structure

```text
expense-tracker/
├── README.md
├── backend/
│   ├── go.mod
│   ├── main.go
│   ├── api/
│   │   ├── analytics_handler.go
│   │   ├── category_handler.go
│   │   ├── middlewares.go
│   │   ├── server.go
│   │   ├── transaction_handler.go
│   │   └── user_handler.go
│   ├── models/
│   │   ├── analytics.go
│   │   ├── category.go
│   │   ├── transaction.go
│   │   └── users.go
│   └── storage/
│       ├── analytics_store.go
│       ├── category_store.go
│       ├── store.go
│       ├── transaction_store.go
│       └── user_store.go
```

## Current API

The service runs on port 8080. `POST /users` and `POST /login` are public; all other endpoints require a JWT in the `Authorization: Bearer <token>` header. Login returns the token in a `token` field.

### Public Routes

| Method | Route  | Description              |
| ------ | ------ | ------------------------ |
| POST   | /users | Register an account      |
| POST   | /login | Log in and receive a JWT |

### Authenticated Routes

| Method | Route                    | Description                              |
| ------ | ------------------------ | ---------------------------------------- |
| POST   | /transactions            | Create an income or expense              |
| GET    | /transactions            | Fetch transactions, newest first         |
| GET    | /transactions/{id}       | Fetch one transaction                    |
| PUT    | /transactions/{id}       | Update a transaction                     |
| DELETE | /transactions/{id}       | Delete a transaction                     |
| POST   | /categories              | Create a category                        |
| GET    | /categories              | List the authenticated user's categories |
| PUT    | /categories/{id}         | Rename a category                        |
| DELETE | /categories/{id}         | Delete a category                        |
| GET    | /analytics/summary       | Retrieve a date-filtered summary         |
| GET    | /analytics/daily-updates | Retrieve daily totals for a date range   |

Category creation and rename requests use a JSON body with a non-empty `name`. Duplicate names return `409 Conflict`; successful updates and deletions return `204 No Content`.

The analytics endpoint requires `start_date` and `end_date` query parameters in `YYYY-MM-DD` format. Both dates are inclusive, use Asia/Kolkata boundaries, and the range cannot exceed 366 calendar days. The response includes total, average, and median transaction amounts; income, expense, and overall counts; net balance and savings rate (percentage); daily totals including zero-activity dates; highest and lowest income/expense days ranked only among days with transactions of that kind; and category totals, averages, and counts split by income and expense. Category-less transactions appear under `Uncategorized`. Each peak/low day field is `null` when there are no transactions of its kind in the range.

`GET /analytics/daily-updates` accepts the same required date parameters and returns the `daily_totals` array on its own, including days with no activity.

```text
GET /analytics/summary?start_date=2026-09-01&end_date=2026-09-30
```

### Transaction Model

```json
{
  "id": 1,
  "amount": 2500,
  "kind": "expense",
  "note": "Groceries",
  "created_at": "2026-09-26T12:00:00Z",
  "category_id": 3,
  "category_name": "Food"
}
```

### Validation Rules

- Amount must be greater than or equal to 0
- Kind must be either `income` or `expense`
- Category names must not be blank

## Database Setup

This application expects a PostgreSQL database configured via the `DATABASE_URL` environment variable.

Example:

```bash
export DATABASE_URL="postgres://username:password@localhost:5432/expense_tracker?sslmode=disable"
```

The storage layer expects `users`, `categories`, and `transactions` tables. This minimal schema matches the columns queried by the backend:

```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE categories (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  user_id INTEGER NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ,
  UNIQUE (name, user_id)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    amount INTEGER NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('income', 'expense')),
    note TEXT,
  user_id INTEGER NOT NULL REFERENCES users(id),
  category_id INTEGER REFERENCES categories(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Running the Project

From the `backend` directory:

```bash
go mod tidy
export DATABASE_URL="postgres://username:password@localhost:5432/expense_tracker?sslmode=disable"
export JWT_SECRET="replace-with-a-long-random-secret"
go run .
```

The server starts on:

```text
http://localhost:8080
```

## Example Requests

Register and log in, then use the returned token for authenticated requests. Set `TOKEN` to the `token` value from the login response. Use the category ID returned by category creation when assigning a category to a transaction.

### Register and log in

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"your-password"}'

curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"your-password"}'
```

### Create a transaction

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1200,
    "kind": "expense",
    "note": "Rent",
    "category_id": 3
  }'
```

### Manage categories

```bash
curl -X POST http://localhost:8080/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Groceries"}'

curl http://localhost:8080/categories \
  -H "Authorization: Bearer $TOKEN"

curl -X PUT http://localhost:8080/categories/3 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Food"}'

curl -X DELETE http://localhost:8080/categories/3 \
  -H "Authorization: Bearer $TOKEN"
```

### Get analytics

```bash
curl "http://localhost:8080/analytics/summary?start_date=2026-09-01&end_date=2026-09-30" \
  -H "Authorization: Bearer $TOKEN"
```

### Get all transactions

```bash
curl http://localhost:8080/transactions \
  -H "Authorization: Bearer $TOKEN"
```

### Get one transaction

```bash
curl http://localhost:8080/transactions/1 \
  -H "Authorization: Bearer $TOKEN"
```

### Update a transaction

```bash
curl -X PUT http://localhost:8080/transactions/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1500,
    "kind": "expense",
    "note": "Updated rent"
  }'
```

### Delete a transaction

```bash
curl -X DELETE http://localhost:8080/transactions/1 \
  -H "Authorization: Bearer $TOKEN"
```

## Notes

- The project provides a backend API; it does not include a frontend dashboard.
- Transaction, category, and analytics data is scoped to the authenticated user.
- Error handling currently returns appropriate HTTP status codes such as `400`, `404`, and `500` based on request and persistence outcomes.

## License

This project is currently provided as a codebase for local development and learning purposes.
