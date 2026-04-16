# OT Calculator Website (Go)

Simple UAT overtime calculator for Employee A and Employee B.

## Run

```bash
go run ./cmd/server
```

Open http://localhost:8080.

## Input rules

- Use 24-hour times (`HH:MM`), e.g. `20:30`.
- Date accepts `YYYY-MM-DD` and `MM/DD/YYYY`.
- UI uses one combined entry table with a Type selector (`OT` or `Break`).
- Duplicate OT entries with the same employee and exact same start/end datetime are counted once.
- Daily rounding uses combined 1.5/2.0 minute balancing (per latest UAT rule), not independent rounding per rate bucket.

## Architecture

- `internal/web`: HTML template rendering.
- `internal/api`: JSON API handler (`POST /api/calculate`).
- `internal/service`: input validation/normalization and orchestration.
- `internal/engine`: all OT/business rules.
- `static/app.js`: localStorage + API calls + rendering only (no OT math).

## Notes

- No login.
- No database.
- Data persistence is browser localStorage only.
