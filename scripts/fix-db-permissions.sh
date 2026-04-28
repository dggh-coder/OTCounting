#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

if [[ ! -f opengauss.env ]]; then
  echo "opengauss.env not found. Copy from opengauss.env.example first."
  exit 1
fi

# shellcheck disable=SC1091
source opengauss.env

APP_USER="${OT_USER_DB_USER:-${DB_USER:-ot_user}}"
APP_PASSWORD="${OT_USER_DB_PASSWORD:-${GS_PASSWORD:-}}"

if [[ -z "${GS_PASSWORD:-}" ]]; then
  echo "GS_PASSWORD is empty in opengauss.env"
  exit 1
fi

if [[ -z "${APP_PASSWORD}" ]]; then
  echo "OT_USER_DB_PASSWORD/GS_PASSWORD is empty in opengauss.env"
  exit 1
fi

podman exec -i ot-opengauss gsql -d postgres -U omm -W "${GS_PASSWORD}" -c \
"DO \
\$\$ \
BEGIN \
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${APP_USER}') THEN \
    EXECUTE format('CREATE USER %I WITH PASSWORD %L', '${APP_USER}', '${APP_PASSWORD}'); \
  ELSE \
    RAISE NOTICE 'Role % already exists; skip password reset.', '${APP_USER}'; \
  END IF; \
END \
\$\$;"

podman exec -i ot-opengauss gsql -d postgres -U omm -W "${GS_PASSWORD}" -c \
"GRANT CONNECT,CREATE,TEMP ON DATABASE postgres TO ${APP_USER}; \
 GRANT USAGE,CREATE ON SCHEMA staffinfo TO ${APP_USER}; \
 GRANT USAGE,CREATE ON SCHEMA otdriverstd TO ${APP_USER}; \
 ALTER SCHEMA staffinfo OWNER TO ${APP_USER}; \
 ALTER SCHEMA otdriverstd OWNER TO ${APP_USER}; \
 GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA staffinfo TO ${APP_USER}; \
 GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA otdriverstd TO ${APP_USER}; \
 ALTER DEFAULT PRIVILEGES IN SCHEMA staffinfo GRANT SELECT,INSERT,UPDATE,DELETE ON TABLES TO ${APP_USER}; \
 ALTER DEFAULT PRIVILEGES IN SCHEMA otdriverstd GRANT SELECT,INSERT,UPDATE,DELETE ON TABLES TO ${APP_USER};"

podman exec -i ot-opengauss gsql -d postgres -U omm -W "${GS_PASSWORD}" -c \
"DO \
\$\$ \
DECLARE r RECORD; \
BEGIN \
  FOR r IN SELECT schemaname, tablename FROM pg_tables WHERE schemaname IN ('staffinfo','otdriverstd') LOOP \
    EXECUTE format('ALTER TABLE %I.%I OWNER TO %I', r.schemaname, r.tablename, '${APP_USER}'); \
  END LOOP; \
END \
\$\$;"

echo "Permissions refreshed for user '${APP_USER}' in schemas staffinfo and otdriverstd."
