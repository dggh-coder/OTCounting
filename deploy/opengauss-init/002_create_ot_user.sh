#!/usr/bin/env bash
set -euo pipefail

APP_USER="${OT_USER_DB_USER:-${DB_USER:-ot_user}}"
APP_PASSWORD="${OT_USER_DB_PASSWORD:-${GS_PASSWORD:-StrongPassw0rd!}}"


gsql -v ON_ERROR_STOP=1 -d postgres -U omm <<EOSQL
DO
\$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '${APP_USER}') THEN
    EXECUTE format('CREATE USER %I WITH PASSWORD %L', '${APP_USER}', '${APP_PASSWORD}');
  ELSE
    EXECUTE format('ALTER USER %I WITH PASSWORD %L', '${APP_USER}', '${APP_PASSWORD}');
  END IF;
END
\$\$;

GRANT CONNECT, CREATE, TEMP ON DATABASE postgres TO ${APP_USER};
GRANT USAGE, CREATE ON SCHEMA ot_staffinfo TO ${APP_USER};
GRANT USAGE, CREATE ON SCHEMA ot_driverstd TO ${APP_USER};
ALTER SCHEMA ot_staffinfo OWNER TO ${APP_USER};
ALTER SCHEMA ot_driverstd OWNER TO ${APP_USER};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA ot_staffinfo TO ${APP_USER};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA ot_driverstd TO ${APP_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA ot_staffinfo
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${APP_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA ot_driverstd
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${APP_USER};

DO
\$\$
DECLARE
  has_objects boolean;
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = '${APP_USER}') THEN
    SELECT EXISTS (
      SELECT 1
      FROM pg_class c
      JOIN pg_namespace n ON n.oid = c.relnamespace
      WHERE n.nspname = '${APP_USER}'
    ) INTO has_objects;

    IF has_objects THEN
      RAISE NOTICE 'Schema % has objects; skip dropping it.', '${APP_USER}';
    ELSE
      EXECUTE format('DROP SCHEMA %I', '${APP_USER}');
    END IF;
  END IF;
END
\$\$;

DO
\$\$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT schemaname, tablename
    FROM pg_tables
    WHERE schemaname IN ('ot_staffinfo', 'ot_driverstd')
  LOOP
    EXECUTE format('ALTER TABLE %I.%I OWNER TO %I', r.schemaname, r.tablename, '${APP_USER}');
  END LOOP;
END
\$\$;
EOSQL

echo "OT user '${APP_USER}' is ready for ot_staffinfo/ot_driverstd schemas."
