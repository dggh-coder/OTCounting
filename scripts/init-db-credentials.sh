#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

mkdir -p secrets

if [[ ! -f .env ]]; then
  cp .env.example .env
fi

if [[ ! -f opengauss.env ]]; then
  cp opengauss.env.example opengauss.env
fi

read -r -p "Backend DB username [ot_user]: " APP_USER
APP_USER="${APP_USER:-ot_user}"

while true; do
  read -r -s -p "Backend DB password: " APP_PASSWORD
  echo
  read -r -s -p "Confirm backend DB password: " APP_PASSWORD_CONFIRM
  echo
  [[ "${APP_PASSWORD}" == "${APP_PASSWORD_CONFIRM}" ]] && break
  echo "Passwords do not match. Please try again."
done

read -r -s -p "openGauss GS_PASSWORD (leave blank to reuse backend password): " GS_PASSWORD
echo
GS_PASSWORD="${GS_PASSWORD:-${APP_PASSWORD}}"

cat > opengauss.env <<ENV
GS_PASSWORD=${GS_PASSWORD}
OT_USER_DB_USER=${APP_USER}
OT_USER_DB_PASSWORD=${APP_PASSWORD}
GS_DB=postgres
GAUSSHOME=/usr/local/opengauss
PATH=/usr/local/opengauss/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
LD_LIBRARY_PATH=/usr/local/opengauss/lib
ENV

cat > .env <<ENV
DB_HOST=opengauss
DB_PORT=5432
DB_NAME=postgres
DB_USER=${APP_USER}
ENV

printf '%s\n' "${APP_PASSWORD}" > secrets/ot_db_password.txt
chmod 600 secrets/ot_db_password.txt

echo "Wrote: .env, opengauss.env, secrets/ot_db_password.txt"
echo "If this is a fresh DB volume, the init script will create '${APP_USER}' automatically."
