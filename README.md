# OT UAT Monorepo

This repository is split into two master folders:

- `ot-frontend/`: static UI served by nginx in a Podman container.
- `ot-backend/`: Go API + OT engine + openGauss persistence.

## Run with Podman Compose

1) Initialize credentials (recommended interactive way):

```bash
./scripts/init-db-credentials.sh
```

This lets you type backend DB username/password and writes:

- `.env` (`DB_USER`)
- `opengauss.env` (`GS_PASSWORD`, `OT_USER_DB_USER`, `OT_USER_DB_PASSWORD`)
- `secrets/ot_db_password.txt` (backend DB password)

When to run this:

- **First-time setup** on a new machine/repo clone (before first `podman compose up`).
- **Any time you change DB username/password** and want `.env`, `opengauss.env`, and secret file to stay in sync.
- **After deleting/recreating env/secret files**.

Then start the stack:

```bash
podman compose -f podman-compose.yml up --build
```

Or copy templates manually:

```bash
cp .env.example .env
cp opengauss.env.example opengauss.env
mkdir -p secrets
cp secrets/ot_db_password.txt.example secrets/ot_db_password.txt
```

2) Edit configs:

- `opengauss.env` -> set `GS_PASSWORD`, `OT_USER_DB_USER`, `OT_USER_DB_PASSWORD`, `GAUSSLOG=/var/log/opengauss`
- `secrets/ot_db_password.txt` -> same value as `OT_USER_DB_PASSWORD`
- `.env` -> non-secret DB host/port/name/user only (`DB_USER` should match `OT_USER_DB_USER`)
- `.env` -> also controls host volume paths used by compose (`OTOPENGAUSS_DB_PATH`, `OTOPENGAUSS_LOG_PATH`)

> Note: `podman-compose` volume interpolation reads from shell/`.env`, not from `opengauss.env`.
> If you want to keep paths in `opengauss.env`, use `./scripts/compose-up.sh` (it loads `opengauss.env` and exports the path variables before calling compose).

DB init user source precedence during first boot:

1. `OT_USER_DB_USER` (from `opengauss.env`)
2. `DB_USER` (from `.env`)
3. fallback `ot_user`

3) Start services:

```bash
sudo mkdir -p /data/otopengauss/db /data/otopengauss/log
podman compose -f podman-compose.yml up --build

# or (loads path variables from opengauss.env)
./scripts/compose-up.sh up --build
```

openGauss data is persisted on host path: `/data/otopengauss/db`.

openGauss logs are persisted on host path: `/data/otopengauss/log`.

Services:

- Frontend: http://localhost:8081
- Backend: http://localhost:8080
- openGauss: localhost:5432

## Run backend only (DB already running)

If your openGauss container is already up and initialized, you can start only the backend service:

```bash
podman compose -f podman-compose.yml up --build --no-deps ot-backend
```

Notes:

- `--no-deps` prevents compose from trying to (re)start `opengauss`.
- Backend image build is configured with `network: host` in compose so `go mod download` uses host DNS (avoids common `proxy.golang.org` lookup timeout during build).
- Keep `.env` with `DB_HOST=opengauss` when backend and DB are on the same compose network.
- If you started DB outside compose/network, set `DB_HOST` to the reachable hostname/IP before starting backend.

DNS behavior note:

- For backend **image builds**, host DNS is used at build time (`build.network: host`). If host DNS changes tomorrow, the next build automatically uses the new host DNS.
- For already-running containers, DNS config is not hot-reloaded. Restart/recreate the container after DNS changes so it picks up updated resolver settings.

Optional one-time rebuild + detached run:

```bash
podman compose -f podman-compose.yml build ot-backend
podman compose -f podman-compose.yml up -d --no-deps ot-backend
```

Images are pinned with fully-qualified names (`docker.io/...`) to avoid Podman short-name resolution errors on hardened hosts.

The backend initializes schema automatically on startup from `ot-backend/internal/db/schema.sql`.

## DB Schema

The openGauss init SQL is also mounted to the DB container at:

- `deploy/opengauss-init/001_schema.sql`
- `deploy/opengauss-init/002_create_ot_user.sh`

It creates:

- `ot_uat.work_session`
- `ot_uat.time_entry`
- `ot_uat.session_result`

If these tables already exist after `podman compose ... up -d opengauss`, **do not** run `001_schema.sql` manually again.

- The entrypoint already executed it during first DB initialization.
- Re-running is usually unnecessary and can hide credential/permission issues you should fix directly.
- Only run schema SQL manually for intentional recovery/migration workflows.

and initializes a backend DB account:

- `ot_user` (password defaults to `GS_PASSWORD`, overridable with `OT_USER_DB_PASSWORD`)

When the DB volume is fresh, `002_create_ot_user.sh` automatically creates/grants this app user during initialization.

It also reassigns ownership of existing `ot_uat` tables to `ot_user` during init so backend writes are not blocked by table-owner mismatches.
It also sets `ot_user` as owner of schema `ot_uat`.

Important: DB init scripts run only on a **fresh** data directory. If `/data/otopengauss/db` already has data, changing `.env` later will not re-run user creation.

If you want permissions/username/password to be auto-ready **at container creation time** with no manual steps, create the DB container with a fresh empty `/data/otopengauss/db` so init scripts execute.

### DB users and passwords (quick clarification)

- `omm` is the openGauss initial admin user. Its password is `GS_PASSWORD` from `opengauss.env` **at first initialization time**.
- `ot_user` is the application user created by `002_create_ot_user.sh` (password defaults to `OT_USER_DB_PASSWORD`, or falls back to `GS_PASSWORD`).
- `secrets/ot_db_password.txt` is only for backend runtime login password injection; it does not define DB users by itself.
- Credentials are not stored in this repo after startup; live auth data is inside the DB data directory (`/data/otopengauss/db`).
- If you change `opengauss.env` later, existing initialized DB credentials do not automatically change.

Check existing DB roles:

```bash
podman exec -it ot-opengauss gsql -d postgres -U omm -W 'YOUR_GS_PASSWORD' -c "\du"
```

## Fix: `FATAL: Forbid remote connection with initial user`

If backend logs show this error, your backend is still connecting as `omm` (initial admin user), which openGauss blocks for remote TCP login.

Use `DB_USER=ot_user` in `.env` and keep `secrets/ot_db_password.txt` aligned with the app user password.

Backend DSN resolves username from `DB_USER` (default `ot_user`), so it no longer falls back to `omm`.

For an **already initialized DB volume** (init scripts do not re-run), create/grant the app user once:

```bash
podman exec -it ot-opengauss gsql -d postgres -U omm -W 'YOUR_GS_PASSWORD' -c "CREATE USER ot_user WITH PASSWORD 'YOUR_APP_PASSWORD';"
podman exec -it ot-opengauss gsql -d postgres -U omm -W 'YOUR_GS_PASSWORD' -c "GRANT CONNECT,CREATE,TEMP ON DATABASE postgres TO ot_user; GRANT USAGE,CREATE ON SCHEMA ot_uat TO ot_user; ALTER SCHEMA ot_uat OWNER TO ot_user; GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA ot_uat TO ot_user; ALTER DEFAULT PRIVILEGES IN SCHEMA ot_uat GRANT SELECT,INSERT,UPDATE,DELETE ON TABLES TO ot_user;"
podman exec -it ot-opengauss gsql -d postgres -U omm -W 'YOUR_GS_PASSWORD' -c "ALTER TABLE ot_uat.work_session OWNER TO ot_user; ALTER TABLE ot_uat.time_entry OWNER TO ot_user; ALTER TABLE ot_uat.session_result OWNER TO ot_user;"
```

Or run the helper (recommended):

```bash
./scripts/fix-db-permissions.sh
```

Use this helper when backend logs show errors like `Invalid username/password` (SQLSTATE `28P01`) or `permission denied for database postgres` (SQLSTATE `42501`) on an existing DB volume.

If logs show both `002_create_ot_app_user.sh` and `002_create_ot_user.sh` running, you likely still have a legacy script file in `deploy/opengauss-init/` on disk. Remove `002_create_ot_app_user.sh` and recreate the DB container with a fresh data directory.

## Fix: `failed to connect Unknown:5432`

This usually means openGauss is not fully ready yet, or `GS_PASSWORD` does not match the initialized DB volume.

1) Ensure compose is using the correct DB image from `podman-compose.yml` (`docker.io/opengauss/opengauss-server:latest`) and not a different manual image.

2) Wait for DB readiness and check logs:

```bash
podman logs -f --tail 200 ot-opengauss
```

3) Test login from inside the DB container (explicit localhost):

```bash
podman exec -it ot-opengauss gsql -h 127.0.0.1 -p 5432 -d postgres -U omm -W 'YOUR_GS_PASSWORD' -c "SELECT 1;"
```

4) If password/auth still fails after changing `opengauss.env`, you are likely reusing an old DB data directory (`/data/otopengauss/db`) initialized with a different password. Recreate the DB volume for a fresh init:

```bash
podman compose -f podman-compose.yml down
sudo rm -rf /data/otopengauss/db/*
podman compose -f podman-compose.yml up -d opengauss
```

## Next step after validating the tables

If `SELECT tablename ...` returns the 3 expected tables (as in your output), the database bootstrap is complete. The next step is to verify end-to-end API + persistence.

1) Keep the compose stack running and check backend health:

```bash
curl -s http://localhost:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```

2) Submit one sample OT calculation:

```bash
curl -s -X POST http://localhost:8080/api/calculate \
  -H 'Content-Type: application/json' \
  -d '{
    "employee_id":"E1001",
    "work_date":"2026-04-19",
    "entries":[
      {"start":"17:30","end":"20:00","entry_type":"work"},
      {"start":"20:00","end":"20:30","entry_type":"break"},
      {"start":"20:30","end":"22:00","entry_type":"work"}
    ]
  }'
```

3) Confirm a row was persisted to `work_session`:

```bash
podman exec -it ot-opengauss gsql -d postgres -U omm -W 'YOUR_PASSWORD' \
  -c "SET search_path TO ot_uat, public; SELECT id, employee_id, work_date, created_at FROM work_session ORDER BY created_at DESC LIMIT 5;"
```

4) Open the frontend and run the same flow in UI:

- http://localhost:8081

If all four checks pass, your UAT environment is ready for test scenarios.
