#!/usr/bin/env bash
#
# XEX Play - Staging Environment Setup on Azure
#
# Creates a staging Container Apps environment and deploys staging versions
# of the API and admin panel. Uses the same PostgreSQL server but a separate
# database (xexplay_staging).
#
# This script is idempotent: it checks for existing resources before creating.
#
# Prerequisites:
#   - Azure CLI logged in with sufficient permissions
#   - ACR (xexregistrystd.azurecr.io) must already have :latest images pushed
#   - PostgreSQL server (xexplay-postgres) must already exist
#   - Redis cache (xexplay-redis) must already exist
#
# Usage:
#   ./setup-staging.sh
#
set -euo pipefail

# ─── Configuration ──────────────────────────────────────────────────────────
PLAY_RG="xexplay-production-rg"
LOCATION="northeurope"

# Staging Container Apps environment (separate from production xex-production-env)
STAGING_ENV="xexplay-staging-env"

# Staging container apps
API_APP="xexplay-api-staging"
ADMIN_APP="xexplay-admin-staging"

# ACR
ACR_SERVER="xexregistrystd.azurecr.io"
API_IMAGE="${ACR_SERVER}/xexplay-api:latest"
ADMIN_IMAGE="${ACR_SERVER}/xexplay-admin:latest"

# Infrastructure (shared with production -- same server, different database)
POSTGRES_SERVER="xexplay-postgres"
STAGING_DB="xexplay_staging"
REDIS_NAME="xexplay-redis"

# Staging-specific ports
API_PORT="8081"
ADMIN_PORT="3001"

# Log Analytics workspace (reuse from monitoring setup if exists)
LOG_WORKSPACE_NAME="xexplay-log-analytics"

# ─── Helpers ─────────────────────────────────────────────────────────────────
info()  { echo -e "\033[1;34m[INFO]\033[0m  $*"; }
ok()    { echo -e "\033[1;32m[OK]\033[0m    $*"; }
warn()  { echo -e "\033[1;33m[WARN]\033[0m  $*"; }
err()   { echo -e "\033[1;31m[ERROR]\033[0m $*"; }
step()  { echo -e "\n\033[1;35m━━━ $* ━━━\033[0m"; }

# ─── Preflight ───────────────────────────────────────────────────────────────
if ! command -v az &>/dev/null; then
  err "Azure CLI (az) is not installed or not in PATH."
  exit 1
fi

info "Verifying Azure login..."
ACCOUNT=$(az account show --query name -o tsv 2>/dev/null || true)
if [[ -z "$ACCOUNT" ]]; then
  err "Not logged in to Azure. Run: az login"
  exit 1
fi
ok "Logged in to Azure account: $ACCOUNT"

# ─── 1. Container Apps Environment ──────────────────────────────────────────
step "1. Container Apps Environment"

EXISTING_ENV=$(az containerapp env show \
  --name "$STAGING_ENV" \
  --resource-group "$PLAY_RG" \
  --query id -o tsv 2>/dev/null || echo "")

if [[ -n "$EXISTING_ENV" && "$EXISTING_ENV" != "None" ]]; then
  ok "Container Apps environment '$STAGING_ENV' already exists."
else
  info "Creating Container Apps environment '$STAGING_ENV'..."

  # Try to reuse existing Log Analytics workspace
  LAW_ID=$(az monitor log-analytics workspace show \
    --workspace-name "$LOG_WORKSPACE_NAME" \
    --resource-group "$PLAY_RG" \
    --query customerId -o tsv 2>/dev/null || echo "")

  LAW_KEY=""
  if [[ -n "$LAW_ID" && "$LAW_ID" != "None" ]]; then
    LAW_KEY=$(az monitor log-analytics workspace get-shared-keys \
      --workspace-name "$LOG_WORKSPACE_NAME" \
      --resource-group "$PLAY_RG" \
      --query primarySharedKey -o tsv 2>/dev/null || echo "")
  fi

  CREATE_CMD=(
    az containerapp env create
    --name "$STAGING_ENV"
    --resource-group "$PLAY_RG"
    --location "$LOCATION"
  )

  if [[ -n "$LAW_ID" && "$LAW_ID" != "None" && -n "$LAW_KEY" ]]; then
    info "Reusing Log Analytics workspace '$LOG_WORKSPACE_NAME'."
    CREATE_CMD+=(
      --logs-workspace-id "$LAW_ID"
      --logs-workspace-key "$LAW_KEY"
    )
  fi

  "${CREATE_CMD[@]}"
  ok "Created Container Apps environment '$STAGING_ENV'."
fi

# ─── 2. Resolve shared infrastructure ───────────────────────────────────────
step "2. Shared Infrastructure"

# PostgreSQL connection info
info "Resolving PostgreSQL server '$POSTGRES_SERVER'..."
PG_FQDN=$(az postgres flexible-server show \
  --name "$POSTGRES_SERVER" \
  --resource-group "$PLAY_RG" \
  --query fullyQualifiedDomainName -o tsv 2>/dev/null || echo "")

if [[ -z "$PG_FQDN" || "$PG_FQDN" == "None" ]]; then
  err "PostgreSQL server '$POSTGRES_SERVER' not found in '$PLAY_RG'."
  err "The staging environment shares the same PostgreSQL server as production."
  exit 1
fi
ok "PostgreSQL server: $PG_FQDN"

# Redis connection info
info "Resolving Redis cache '$REDIS_NAME'..."
REDIS_HOST=$(az redis show \
  --name "$REDIS_NAME" \
  --resource-group "$PLAY_RG" \
  --query hostName -o tsv 2>/dev/null || echo "")

if [[ -z "$REDIS_HOST" || "$REDIS_HOST" == "None" ]]; then
  err "Redis cache '$REDIS_NAME' not found in '$PLAY_RG'."
  exit 1
fi
ok "Redis cache: $REDIS_HOST"

# ─── 3. Deploy API Container App ────────────────────────────────────────────
step "3. API Container App (${API_APP})"

# NOTE: DATABASE_URL, REDIS_URL, and JWT_SECRET must be set as secrets on the
# container app after creation. This script sets placeholder values that MUST
# be updated before the app is functional.
#
# Staging DATABASE_URL should point to the staging database:
#   postgres://<user>:<password>@<PG_FQDN>:5432/xexplay_staging?sslmode=require
#
# The staging database (xexplay_staging) should be created on the same
# PostgreSQL server used by production (xexplay-postgres).

EXISTING_API=$(az containerapp show \
  --name "$API_APP" \
  --resource-group "$PLAY_RG" \
  --query id -o tsv 2>/dev/null || echo "")

if [[ -n "$EXISTING_API" && "$EXISTING_API" != "None" ]]; then
  ok "Container app '$API_APP' already exists. Updating image..."
  az containerapp update \
    --name "$API_APP" \
    --resource-group "$PLAY_RG" \
    --image "$API_IMAGE"
  ok "Updated '$API_APP' to image: $API_IMAGE"
else
  info "Creating container app '$API_APP'..."
  az containerapp create \
    --name "$API_APP" \
    --resource-group "$PLAY_RG" \
    --environment "$STAGING_ENV" \
    --image "$API_IMAGE" \
    --registry-server "$ACR_SERVER" \
    --target-port 8081 \
    --ingress external \
    --min-replicas 0 \
    --max-replicas 2 \
    --cpu 0.25 \
    --memory 0.5Gi \
    --env-vars \
      "PORT=${API_PORT}" \
      "ENVIRONMENT=staging" \
      "LOG_LEVEL=debug" \
      "CORS_ORIGINS=*"
  ok "Created container app '$API_APP'."
  warn "IMPORTANT: Set secrets on '$API_APP' before use:"
  warn "  az containerapp secret set --name $API_APP --resource-group $PLAY_RG \\"
  warn "    --secrets database-url=<STAGING_DATABASE_URL> jwt-secret=<JWT_SECRET> redis-url=<REDIS_URL>"
  warn "  az containerapp update --name $API_APP --resource-group $PLAY_RG \\"
  warn "    --set-env-vars DATABASE_URL=secretref:database-url JWT_SECRET=secretref:jwt-secret REDIS_URL=secretref:redis-url"
fi

# ─── 4. Deploy Admin Container App ──────────────────────────────────────────
step "4. Admin Container App (${ADMIN_APP})"

EXISTING_ADMIN=$(az containerapp show \
  --name "$ADMIN_APP" \
  --resource-group "$PLAY_RG" \
  --query id -o tsv 2>/dev/null || echo "")

if [[ -n "$EXISTING_ADMIN" && "$EXISTING_ADMIN" != "None" ]]; then
  ok "Container app '$ADMIN_APP' already exists. Updating image..."
  az containerapp update \
    --name "$ADMIN_APP" \
    --resource-group "$PLAY_RG" \
    --image "$ADMIN_IMAGE"
  ok "Updated '$ADMIN_APP' to image: $ADMIN_IMAGE"
else
  info "Creating container app '$ADMIN_APP'..."
  az containerapp create \
    --name "$ADMIN_APP" \
    --resource-group "$PLAY_RG" \
    --environment "$STAGING_ENV" \
    --image "$ADMIN_IMAGE" \
    --registry-server "$ACR_SERVER" \
    --target-port 3001 \
    --ingress external \
    --min-replicas 0 \
    --max-replicas 2 \
    --cpu 0.25 \
    --memory 0.5Gi \
    --env-vars \
      "PORT=${ADMIN_PORT}" \
      "NODE_ENV=staging"
  ok "Created container app '$ADMIN_APP'."
  warn "IMPORTANT: Set the NEXT_PUBLIC_API_URL env var to point at the staging API URL."
fi

# ─── 5. Staging Database Setup ───────────────────────────────────────────────
step "5. Staging Database (${STAGING_DB})"

echo ""
info "The staging apps use a separate database '${STAGING_DB}' on the same"
info "PostgreSQL server '${POSTGRES_SERVER}'. To create and seed it:"
echo ""
echo "  # 1. Create the staging database (if it doesn't exist)"
echo "  az postgres flexible-server db create \\"
echo "    --resource-group ${PLAY_RG} \\"
echo "    --server-name ${POSTGRES_SERVER} \\"
echo "    --database-name ${STAGING_DB}"
echo ""
echo "  # 2. Run migrations against the staging database"
echo "  export DATABASE_URL=\"postgres://<user>:<password>@${PG_FQDN}:5432/${STAGING_DB}?sslmode=require\""
echo "  migrate -path backend/migrations -database \"\$DATABASE_URL\" up"
echo ""
echo "  # 3. Seed the staging database"
echo "  psql \"\$DATABASE_URL\" -f backend/migrations/seed.sql"
echo "  psql \"\$DATABASE_URL\" -f backend/migrations/seed_achievements.sql"
echo ""

# Attempt to create the database automatically
info "Attempting to create database '${STAGING_DB}' on '${POSTGRES_SERVER}'..."
EXISTING_DB=$(az postgres flexible-server db show \
  --resource-group "$PLAY_RG" \
  --server-name "$POSTGRES_SERVER" \
  --database-name "$STAGING_DB" \
  --query name -o tsv 2>/dev/null || echo "")

if [[ -n "$EXISTING_DB" && "$EXISTING_DB" != "None" ]]; then
  ok "Database '${STAGING_DB}' already exists on '${POSTGRES_SERVER}'."
else
  info "Creating database '${STAGING_DB}'..."
  az postgres flexible-server db create \
    --resource-group "$PLAY_RG" \
    --server-name "$POSTGRES_SERVER" \
    --database-name "$STAGING_DB" 2>/dev/null && \
    ok "Created database '${STAGING_DB}'." || \
    warn "Could not create database '${STAGING_DB}'. Create it manually (see instructions above)."
fi

# ─── 6. Summary ──────────────────────────────────────────────────────────────
step "Summary"

API_FQDN=$(az containerapp show \
  --name "$API_APP" \
  --resource-group "$PLAY_RG" \
  --query "properties.configuration.ingress.fqdn" -o tsv 2>/dev/null || echo "<pending>")

ADMIN_FQDN=$(az containerapp show \
  --name "$ADMIN_APP" \
  --resource-group "$PLAY_RG" \
  --query "properties.configuration.ingress.fqdn" -o tsv 2>/dev/null || echo "<pending>")

echo ""
echo "  Environment:   $STAGING_ENV"
echo "  Resource Group: $PLAY_RG"
echo "  Location:       $LOCATION"
echo ""
echo "  API App:        $API_APP"
echo "  API URL:        https://${API_FQDN}"
echo "  API Image:      $API_IMAGE"
echo ""
echo "  Admin App:      $ADMIN_APP"
echo "  Admin URL:      https://${ADMIN_FQDN}"
echo "  Admin Image:    $ADMIN_IMAGE"
echo ""
echo "  Database:       $STAGING_DB (on $POSTGRES_SERVER)"
echo "  PostgreSQL:     $PG_FQDN"
echo "  Redis:          $REDIS_HOST"
echo ""
info "Next steps:"
echo "  1. Set secrets (DATABASE_URL, JWT_SECRET, REDIS_URL) on the staging container apps"
echo "  2. Run migrations and seed data against '${STAGING_DB}'"
echo "  3. Verify: curl https://${API_FQDN}/health"
echo ""
ok "Staging environment setup complete."
