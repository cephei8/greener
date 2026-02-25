#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"

log() { echo "==> $*"; }

# ─── Go ──────────────────────────────────────────────────────────────────────
GO_PROJECTS=(
  server
  reporting/greener-reporter-cli
  reporting/greener-reporter-go
  reporting/greener-reporter-junitxml
)

for project in "${GO_PROJECTS[@]}"; do
  dir="$ROOT/$project"
  if [[ -f "$dir/go.mod" ]]; then
    log "Upgrading Go deps: $project"
    (cd "$dir" && go get -u ./... && go mod tidy)
  fi
done

# ─── npm ─────────────────────────────────────────────────────────────────────
NPM_PROJECTS=(
  server
  reporting/greener-reporter-js
  reporting/jest-greener
  reporting/mocha-greener
)

for project in "${NPM_PROJECTS[@]}"; do
  dir="$ROOT/$project"
  if [[ -f "$dir/package.json" ]]; then
    log "Upgrading npm deps: $project"
    (cd "$dir" && npx --yes npm-check-updates --upgrade && npm install)
  fi
done

# ─── Python (Poetry) ─────────────────────────────────────────────────────────
POETRY_PROJECTS=(
  reporting/greener-reporter-py
  reporting/pytest-greener
)

for project in "${POETRY_PROJECTS[@]}"; do
  dir="$ROOT/$project"
  if [[ -f "$dir/pyproject.toml" ]]; then
    log "Upgrading Python deps: $project"
    (cd "$dir" && poetry update)
  fi
done

# ─── Rust (Cargo) ────────────────────────────────────────────────────────────
CARGO_PROJECTS=(
  reporting/greener-reporter
)

for project in "${CARGO_PROJECTS[@]}"; do
  dir="$ROOT/$project"
  if [[ -f "$dir/Cargo.toml" ]]; then
    log "Upgrading Cargo deps: $project"
    (cd "$dir" && cargo update)
  fi
done

log "All dependencies upgraded."
