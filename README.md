# OT UAT Monorepo

This repository is split into two master folders:

- `ot-frontend/`: static UI served by nginx in a Podman container.
- `ot-backend/`: Go API + OT engine + openGauss persistence.

## Run with Podman Compose

1) Copy env templates and set password:

```bash
cp .env.example .env
cp opengauss.env.example opengauss.env
```

2) Edit both files and set the same password value:

- `.env` -> `OT_DB_PASSWORD=...`
- `opengauss.env` -> `GS_PASSWORD=...`

3) Start services:

```bash
sudo mkdir -p /data/otopengaussdb
podman compose -f podman-compose.yml up --build
```

openGauss data is persisted on host path: `/data/otopengaussdb`.

Services:

- Frontend: http://localhost:8081
- Backend: http://localhost:8080
- openGauss: localhost:5432

Images are pinned with fully-qualified names (`docker.io/...`) to avoid Podman short-name resolution errors on hardened hosts.

The backend initializes schema automatically on startup from `ot-backend/internal/db/schema.sql`.

## DB Schema

The openGauss init SQL is also mounted to the DB container at:

- `deploy/opengauss-init/001_schema.sql`

It creates:

- `ot_uat.work_session`
- `ot_uat.time_entry`
- `ot_uat.session_result`
