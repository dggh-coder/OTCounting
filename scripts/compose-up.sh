#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

if [[ -f opengauss.env ]]; then
  # shellcheck disable=SC1091
  source opengauss.env
fi

export OTOPENGAUSS_DB_PATH="${OTOPENGAUSS_DB_PATH:-/data/otopengauss/db}"
export OTOPENGAUSS_LOG_PATH="${OTOPENGAUSS_LOG_PATH:-/data/otopengauss/log}"

exec podman compose -f podman-compose.yml "$@"
