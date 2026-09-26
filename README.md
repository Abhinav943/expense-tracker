# Expense Tracker API

<div align="center">

![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-336791?logo=postgresql)
![REST API](https://img.shields.io/badge/API-REST%20Service-4EAAFF)
![Status](https://img.shields.io/badge/Status-Backend%20API%20Active-2ea44f)

A lightweight backend service for tracking personal income and expenses.

</div>

## Overview

This project is a backend expense tracker built with Go and PostgreSQL. It currently exposes a REST API for managing financial transactions, including creating, reading, updating, and deleting entries. The service validates transaction data before persisting it and supports basic CRUD operations for an expense ledger.

This repository is focused on the API layer and data persistence. There is no frontend application or dashboard yet; the project currently provides the server-side foundation for an expense tracking system.

## Features

- Create new income or expense transactions
- Retrieve a single transaction by ID
- Retrieve all transactions in reverse chronological order
- Update an existing transaction
- Delete a transaction
- Input validation for amount and transaction kind
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
│   │   └── handlers.go
│   ├── models/
│   │   └── transaction.go
│   └── storage/
│       └── postgres.go
└── .env.example (if added later)
```

## Current API

The service runs on port 8080 and exposes the following endpoints:

### Transaction Routes

| Method | Route              | Description                |
| ------ | ------------------ | -------------------------- |
| POST   | /transactions      | Create a new transaction   |
| GET    | /transactions      | Fetch all transactions     |
| GET    | /transactions/{id} | Fetch a single transaction |
| PUT    | /transactions/{id} | Update a transaction       |
| DELETE | /transactions/{id} | Delete a transaction       |

### Transaction Model

```json
{
  "id": 1,
  "amount": 2500,
  "kind": "expense",
  "note": "Groceries",
  "created_at": "2026-09-26T12:00:00Z"
}
```

### Validation Rules

- Amount must be greater than or equal to 0
- Kind must be either `income` or `expense`

## Database Setup

This application expects a PostgreSQL database configured via the `DATABASE_URL` environment variable.

Example:

```bash
export DATABASE_URL="postgres://username:password@localhost:5432/expense_tracker?sslmode=disable"
```

The application uses a `transactions` table with data shaped around the following fields:

```sql
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    amount INTEGER NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('income', 'expense')),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Running the Project

From the `backend` directory:

```bash
go mod tidy
export DATABASE_URL="postgres://username:password@localhost:5432/expense_tracker?sslmode=disable"
go run .
```

The server starts on:

```text
http://localhost:8080
```

## Example Requests

### Create a transaction

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1200,
    "kind": "expense",
    "note": "Rent"
  }'
```

### Get all transactions

```bash
curl http://localhost:8080/transactions
```

### Get one transaction

```bash
curl http://localhost:8080/transactions/1
```

### Update a transaction

```bash
curl -X PUT http://localhost:8080/transactions/1 \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 1500,
    "kind": "expense",
    "note": "Updated rent"
  }'
```

### Delete a transaction

```bash
curl -X DELETE http://localhost:8080/transactions/1
```

## Notes

- The project currently focuses on backend functionality for transaction management.
- The API is designed to be extended with authentication, reporting, category management, and a frontend UI in future iterations.
- Error handling currently returns appropriate HTTP status codes such as `400`, `404`, and `500` based on request and persistence outcomes.

## License

This project is currently provided as a codebase for local development and learning purposes. 
