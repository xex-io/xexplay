#!/usr/bin/env bash
#
# XEX Play - Deployment Verification Script
#
# Verifies that all XEX Play components are running and healthy after a
# deployment. This script is read-only and idempotent -- it never modifies
# any resource; it only queries current state and reports results.
#
# Prerequisites:
#   - Azure CLI logged in with at least Reader role
#   - Network access to health endpoints
#   - psql and redis-cli available (optional, skipped if missing)
#
# Usage:
#   ./verify-deployment.sh                    # uses defaults
#   API_URL=https://custom.domain ./verify-deployment.sh
#
set -euo pipefail

# ─── Configuration ──────────────────────────────────────────────────────────
PLAY_RG="${PLAY_RG:-xexplay-production-rg}"
ENV_RG="${ENV_RG:-xex-production-rg}"

CONTAINER_APP_API="${CONTAINER_APP_API:-xexplay-api}"
CONTAINER_APP_ADMIN="${CONTAINER_APP_ADMIN:-xexplay-admin}"

POSTGRES_SERVER="${POSTGRES_SERVER:-xexplay-postgres}"
REDIS_NAME="${REDIS_NAME:-xexplay-redis}"

# Health endpoint URLs -- override via env if needed
API_URL="${API_URL:-}"
ADMIN_URL="${ADMIN_URL:-}"

# ─── Helpers ────────────────────────────────────────────────────────────────
PASS=0
FAIL=0
WARN=0

info()  { echo -e "\033[1;34m[INFO]\033[0m  $*"; }
ok()    { echo -e "\033[1;32m[PASS]\033[0m  $*"; PASS=$((PASS + 1)); }
fail()  { echo -e "\033[1;31m[FAIL]\033[0m  $*"; FAIL=$((FAIL + 1)); }
warn()  { echo -e "\033[1;33m[WARN]\033[0m  $*"; WARN=$((WARN + 1)); }
sep()   { echo ""; }

# ─── Preflight ──────────────────────────────────────────────────────────────
if ! command -v az &>/dev/null; then
  fail "Azure CLI (az) is not installed or not in PATH."
  echo "Install it: https://learn.microsoft.com/en-us/cli/azure/install-azure-cli"
  exit 1
fi

if ! az account show &>/dev/null; then
  fail "Azure CLI is not logged in. Run 'az login' first."
  exit 1
fi

ACCOUNT=$(az account show --query name -o tsv 2>/dev/null)
info "Azure subscription: $ACCOUNT"
sep

# ─── 1. API Container App Status ───────────────────────────────────────────
info "--- 1. API Container App ($CONTAINER_APP_API) ---"

API_STATUS=$(az containerapp show \
  --name "$CONTAINER_APP_API" \
  --resource-group "$ENV_RG" \
  --query "properties.runningStatus" \
  -o tsv 2>/dev/null || echo "UNAVAILABLE")

if [[ "$API_STATUS" == "Running" ]]; then
  ok "API container app status: $API_STATUS"
elif [[ "$API_STATUS" == "UNAVAILABLE" ]]; then
  fail "API container app not found or not accessible in resource group '$ENV_RG'."
else
  fail "API container app status: $API_STATUS (expected: Running)"
fi

# Check provisioning state as a fallback indicator
API_PROV=$(az containerapp show \
  --name "$CONTAINER_APP_API" \
  --resource-group "$ENV_RG" \
  --query "properties.provisioningState" \
  -o tsv 2>/dev/null || echo "UNKNOWN")

if [[ "$API_PROV" == "Succeeded" ]]; then
  ok "API provisioning state: $API_PROV"
else
  warn "API provisioning state: $API_PROV"
fi
sep

# ─── 2. API Revision Status ────────────────────────────────────────────────
info "--- 2. API Active Revisions ---"

ACTIVE_REVISIONS=$(az containerapp revision list \
  --name "$CONTAINER_APP_API" \
  --resource-group "$ENV_RG" \
  --query "[?properties.active==\`true\`].{name:name, created:properties.createdTime, replicas:properties.replicas, status:properties.runningState}" \
  -o table 2>/dev/null || echo "")

if [[ -n "$ACTIVE_REVISIONS" ]]; then
  ok "Active revisions found:"
  echo "$ACTIVE_REVISIONS"
else
  fail "No active revisions found for API container app."
fi

LATEST_REVISION=$(az containerapp revision list \
  --name "$CONTAINER_APP_API" \
  --resource-group "$ENV_RG" \
  --query "reverse(sort_by([], &properties.createdTime))[0].name" \
  -o tsv 2>/dev/null || echo "")

if [[ -n "$LATEST_REVISION" ]]; then
  LATEST_ACTIVE=$(az containerapp revision show \
    --name "$CONTAINER_APP_API" \
    --resource-group "$ENV_RG" \
    --revision "$LATEST_REVISION" \
    --query "properties.active" \
    -o tsv 2>/dev/null || echo "false")

  if [[ "$LATEST_ACTIVE" == "true" ]]; then
    ok "Latest revision '$LATEST_REVISION' is active."
  else
    warn "Latest revision '$LATEST_REVISION' is NOT active. Traffic may be on an older revision."
  fi
fi
sep

# ─── 3. API Health Endpoint ────────────────────────────────────────────────
info "--- 3. API Health Endpoint ---"

# Auto-discover FQDN if URL not provided
if [[ -z "$API_URL" ]]; then
  API_FQDN=$(az containerapp show \
    --name "$CONTAINER_APP_API" \
    --resource-group "$ENV_RG" \
    --query "properties.configuration.ingress.fqdn" \
    -o tsv 2>/dev/null || echo "")

  if [[ -n "$API_FQDN" && "$API_FQDN" != "None" ]]; then
    API_URL="https://${API_FQDN}"
    info "Discovered API URL: $API_URL"
  else
    warn "Could not discover API FQDN. Set API_URL env var to check health."
  fi
fi

if [[ -n "$API_URL" ]]; then
  HEALTH_URL="${API_URL}/health"
  HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "$HEALTH_URL" 2>/dev/null || echo "000")

  if [[ "$HTTP_CODE" == "200" ]]; then
    ok "Health endpoint $HEALTH_URL returned HTTP $HTTP_CODE."
  elif [[ "$HTTP_CODE" == "000" ]]; then
    fail "Health endpoint $HEALTH_URL is unreachable (timeout or DNS failure)."
  else
    fail "Health endpoint $HEALTH_URL returned HTTP $HTTP_CODE (expected 200)."
  fi
fi
sep

# ─── 4. Admin Panel Container App Status ───────────────────────────────────
info "--- 4. Admin Panel ($CONTAINER_APP_ADMIN) ---"

ADMIN_STATUS=$(az containerapp show \
  --name "$CONTAINER_APP_ADMIN" \
  --resource-group "$ENV_RG" \
  --query "properties.runningStatus" \
  -o tsv 2>/dev/null || echo "UNAVAILABLE")

if [[ "$ADMIN_STATUS" == "Running" ]]; then
  ok "Admin panel container app status: $ADMIN_STATUS"
elif [[ "$ADMIN_STATUS" == "UNAVAILABLE" ]]; then
  fail "Admin panel container app not found or not accessible."
else
  fail "Admin panel container app status: $ADMIN_STATUS (expected: Running)"
fi

ADMIN_PROV=$(az containerapp show \
  --name "$CONTAINER_APP_ADMIN" \
  --resource-group "$ENV_RG" \
  --query "properties.provisioningState" \
  -o tsv 2>/dev/null || echo "UNKNOWN")

if [[ "$ADMIN_PROV" == "Succeeded" ]]; then
  ok "Admin provisioning state: $ADMIN_PROV"
else
  warn "Admin provisioning state: $ADMIN_PROV"
fi

# Admin health check
if [[ -z "$ADMIN_URL" ]]; then
  ADMIN_FQDN=$(az containerapp show \
    --name "$CONTAINER_APP_ADMIN" \
    --resource-group "$ENV_RG" \
    --query "properties.configuration.ingress.fqdn" \
    -o tsv 2>/dev/null || echo "")

  if [[ -n "$ADMIN_FQDN" && "$ADMIN_FQDN" != "None" ]]; then
    ADMIN_URL="https://${ADMIN_FQDN}"
    info "Discovered Admin URL: $ADMIN_URL"
  fi
fi

if [[ -n "$ADMIN_URL" ]]; then
  ADMIN_HTTP=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "$ADMIN_URL" 2>/dev/null || echo "000")

  if [[ "$ADMIN_HTTP" == "200" || "$ADMIN_HTTP" == "301" || "$ADMIN_HTTP" == "302" ]]; then
    ok "Admin panel responded with HTTP $ADMIN_HTTP."
  elif [[ "$ADMIN_HTTP" == "000" ]]; then
    fail "Admin panel is unreachable."
  else
    fail "Admin panel returned HTTP $ADMIN_HTTP."
  fi
fi
sep

# ─── 5. PostgreSQL Accessibility ───────────────────────────────────────────
info "--- 5. PostgreSQL Server ($POSTGRES_SERVER) ---"

PG_STATE=$(az postgres flexible-server show \
  --name "$POSTGRES_SERVER" \
  --resource-group "$PLAY_RG" \
  --query "state" \
  -o tsv 2>/dev/null || echo "UNAVAILABLE")

if [[ "$PG_STATE" == "Ready" ]]; then
  ok "PostgreSQL server state: $PG_STATE"
elif [[ "$PG_STATE" == "UNAVAILABLE" ]]; then
  fail "PostgreSQL server '$POSTGRES_SERVER' not found in '$PLAY_RG'."
else
  fail "PostgreSQL server state: $PG_STATE (expected: Ready)"
fi

PG_FQDN=$(az postgres flexible-server show \
  --name "$POSTGRES_SERVER" \
  --resource-group "$PLAY_RG" \
  --query "fullyQualifiedDomainName" \
  -o tsv 2>/dev/null || echo "")

if [[ -n "$PG_FQDN" && "$PG_FQDN" != "None" ]]; then
  info "PostgreSQL FQDN: $PG_FQDN"

  # Optional: attempt TCP connectivity check on port 5432
  if command -v nc &>/dev/null; then
    if nc -z -w 5 "$PG_FQDN" 5432 2>/dev/null; then
      ok "PostgreSQL port 5432 is reachable on $PG_FQDN."
    else
      warn "PostgreSQL port 5432 is not reachable from this machine (may be VNet-restricted)."
    fi
  else
    info "Skipping TCP connectivity check (nc not available)."
  fi
fi
sep

# ─── 6. Redis Accessibility ────────────────────────────────────────────────
info "--- 6. Redis Cache ($REDIS_NAME) ---"

REDIS_STATE=$(az redis show \
  --name "$REDIS_NAME" \
  --resource-group "$PLAY_RG" \
  --query "provisioningState" \
  -o tsv 2>/dev/null || echo "UNAVAILABLE")

if [[ "$REDIS_STATE" == "Succeeded" ]]; then
  ok "Redis provisioning state: $REDIS_STATE"
elif [[ "$REDIS_STATE" == "UNAVAILABLE" ]]; then
  fail "Redis cache '$REDIS_NAME' not found in '$PLAY_RG'."
else
  fail "Redis provisioning state: $REDIS_STATE (expected: Succeeded)"
fi

REDIS_HOST=$(az redis show \
  --name "$REDIS_NAME" \
  --resource-group "$PLAY_RG" \
  --query "hostName" \
  -o tsv 2>/dev/null || echo "")

if [[ -n "$REDIS_HOST" && "$REDIS_HOST" != "None" ]]; then
  info "Redis hostname: $REDIS_HOST"

  REDIS_SSL_PORT=$(az redis show \
    --name "$REDIS_NAME" \
    --resource-group "$PLAY_RG" \
    --query "sslPort" \
    -o tsv 2>/dev/null || echo "6380")

  if command -v nc &>/dev/null; then
    if nc -z -w 5 "$REDIS_HOST" "$REDIS_SSL_PORT" 2>/dev/null; then
      ok "Redis SSL port $REDIS_SSL_PORT is reachable on $REDIS_HOST."
    else
      warn "Redis SSL port $REDIS_SSL_PORT is not reachable from this machine (may be VNet-restricted)."
    fi
  else
    info "Skipping TCP connectivity check (nc not available)."
  fi
fi
sep

# ─── Summary ───────────────────────────────────────────────────────────────
echo "============================================"
echo " Deployment Verification Summary"
echo "============================================"
echo -e " \033[1;32mPassed:\033[0m $PASS"
echo -e " \033[1;31mFailed:\033[0m $FAIL"
echo -e " \033[1;33mWarnings:\033[0m $WARN"
echo "============================================"

if [[ $FAIL -gt 0 ]]; then
  echo ""
  fail "Deployment verification completed with $FAIL failure(s). Investigate before proceeding."
  exit 1
else
  echo ""
  ok "All checks passed. Deployment looks healthy."
  exit 0
fi
