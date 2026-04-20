#!/usr/bin/env bash
set -euo pipefail

APP_USER="${OT_USER_DB_USER:-${DB_USER:-ot_user}}"
APP_PASSWORD="${OT_USER_DB_PASSWORD:-${GS_PASSWORD:-}}"

if [[ -z "${APP_PASSWORD}" ]]; then
  echo "OT user setup skipped: OT_USER_DB_PASSWORD/GS_PASSWORD is empty"
  exit 0
fi

gsql -v ON_ERROR_STOP=1 -d postgres -U omm <<EOSQL
DO
\$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${APP_USER}') THEN
    EXECUTE format('CREATE USER %I WITH PASSWORD %L', '${APP_USER}', '${APP_PASSWORD}');
  ELSE
    RAISE NOTICE 'Role % already exists; skip password reset in init script.', '${APP_USER}';
  END IF;
END
\$\$;

GRANT USAGE ON SCHEMA ot_uat TO ${APP_USER};
GRANT CREATE ON SCHEMA ot_uat TO ${APP_USER};
GRANT CONNECT, CREATE, TEMP ON DATABASE postgres TO ${APP_USER};
ALTER SCHEMA ot_uat OWNER TO ${APP_USER};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA ot_uat TO ${APP_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA ot_uat
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${APP_USER};

DO
\$\$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT tablename
    FROM pg_tables
    WHERE schemaname = 'ot_uat'
  LOOP
    EXECUTE format('ALTER TABLE ot_uat.%I OWNER TO %I', r.tablename, '${APP_USER}');
  END LOOP;
END
\$\$;
EOSQL

echo "OT user '${APP_USER}' is ready."
