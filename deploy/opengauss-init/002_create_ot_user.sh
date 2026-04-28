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

GRANT CONNECT, CREATE, TEMP ON DATABASE postgres TO ${APP_USER};
GRANT USAGE, CREATE ON SCHEMA staffinfo TO ${APP_USER};
GRANT USAGE, CREATE ON SCHEMA otdriverstd TO ${APP_USER};
ALTER SCHEMA staffinfo OWNER TO ${APP_USER};
ALTER SCHEMA otdriverstd OWNER TO ${APP_USER};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA staffinfo TO ${APP_USER};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA otdriverstd TO ${APP_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA staffinfo
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${APP_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA otdriverstd
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${APP_USER};

DO
\$\$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT schemaname, tablename
    FROM pg_tables
    WHERE schemaname IN ('staffinfo', 'otdriverstd')
  LOOP
    EXECUTE format('ALTER TABLE %I.%I OWNER TO %I', r.schemaname, r.tablename, '${APP_USER}');
  END LOOP;
END
\$\$;
EOSQL

echo "OT user '${APP_USER}' is ready for staffinfo/otdriverstd schemas."
