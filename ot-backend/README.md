# OT Backend (Go)

Go API backend for OT UAT.

## Environment

- `DATABASE_URL` (required)
  - Example: `postgres://omm:StrongPassw0rd!@localhost:5432/postgres?sslmode=disable`

## Run

```bash
go run ./cmd/server
```

Server listens on `:8080` and exposes:

- `POST /api/calculate`
- `POST /api/ot/input`
- `GET /api/ot/entries`
- `GET /api/ot/monthly`
- `GET /api/staff`
- `GET /healthz`

On startup it initializes openGauss schema from embedded `internal/db/schema.sql`.
