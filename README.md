# OT Calculator Website (Go)

Simple UAT overtime calculator for Employee A and Employee B.

## Run

```bash
go run ./cmd/server
```

Open http://localhost:8080.

## Architecture

- `internal/web`: HTML template rendering.
- `internal/api`: JSON API handler (`POST /api/calculate`).
- `internal/service`: input validation and orchestration.
- `internal/engine`: all OT/business rules.
- `static/app.js`: localStorage + API calls + rendering only (no OT math).

## Notes

- No login.
- No database.
- Data persistence is browser localStorage only.
