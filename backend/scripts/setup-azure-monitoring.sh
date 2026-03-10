#!/usr/bin/env bash
#
# XEX Play - Azure Monitoring, Alerting & Log Aggregation Setup
#
# This script is idempotent: it checks for existing resources before creating.
#
# Prerequisites:
#   - Azure CLI logged in with sufficient permissions
#   - Target resource groups and resources must already exist
#
set -euo pipefail

# ─── Configuration ──────────────────────────────────────────────────────────
PLAY_RG="xexplay-production-rg"
ENV_RG="xex-production-rg"
LOCATION="northeurope"

CONTAINER_APP_API="xexplay-api"
CONTAINER_APP_ADMIN="xexplay-admin"
CONTAINER_APP_ENV="xex-production-env"

POSTGRES_SERVER="xexplay-postgres"
REDIS_NAME="xexplay-redis"

ACTION_GROUP_NAME="xexplay-alerts-ag"
ALERT_EMAIL="golsharifi@outlook.com"

LOG_WORKSPACE_NAME="xexplay-log-analytics"

# ─── Helper ─────────────────────────────────────────────────────────────────
info()  { echo -e "\033[1;34m[INFO]\033[0m  $*"; }
ok()    { echo -e "\033[1;32m[OK]\033[0m    $*"; }
warn()  { echo -e "\033[1;33m[WARN]\033[0m  $*"; }
err()   { echo -e "\033[1;31m[ERROR]\033[0m $*"; }

# ─── 1. Log Analytics Workspace ────────────────────────────────────────────
info "Checking Log Analytics workspace on Container App Environment..."

LAW_CUSTOMER_ID=$(az containerapp env show \
  --name "$CONTAINER_APP_ENV" \
  --resource-group "$ENV_RG" \
  --query "properties.appLogsConfiguration.logAnalyticsConfiguration.customerId" \
  -o tsv 2>/dev/null || echo "")

if [[ -n "$LAW_CUSTOMER_ID" && "$LAW_CUSTOMER_ID" != "None" ]]; then
  ok "Container App Environment already has Log Analytics workspace: $LAW_CUSTOMER_ID"

  # Find the workspace resource ID by customer ID so we can use it for diagnostic settings
  LAW_RESOURCE_ID=$(az monitor log-analytics workspace list \
    --query "[?customerId=='$LAW_CUSTOMER_ID'].id | [0]" -o tsv 2>/dev/null || echo "")

  if [[ -z "$LAW_RESOURCE_ID" || "$LAW_RESOURCE_ID" == "None" ]]; then
    # Try across all subscriptions / resource groups
    LAW_RESOURCE_ID=$(az monitor log-analytics workspace list \
      --resource-group "$ENV_RG" \
      --query "[?customerId=='$LAW_CUSTOMER_ID'].id | [0]" -o tsv 2>/dev/null || echo "")
  fi

  if [[ -z "$LAW_RESOURCE_ID" || "$LAW_RESOURCE_ID" == "None" ]]; then
    warn "Could not resolve workspace resource ID from customer ID. Creating a dedicated workspace."
    LAW_RESOURCE_ID=""
  else
    ok "Resolved workspace resource ID: $LAW_RESOURCE_ID"
  fi
else
  info "No Log Analytics workspace found on environment."
  LAW_RESOURCE_ID=""
fi

# If we still don't have a workspace resource ID, create one in the play RG
if [[ -z "$LAW_RESOURCE_ID" ]]; then
  EXISTING_LAW=$(az monitor log-analytics workspace show \
    --workspace-name "$LOG_WORKSPACE_NAME" \
    --resource-group "$PLAY_RG" \
    --query id -o tsv 2>/dev/null || echo "")

  if [[ -n "$EXISTING_LAW" && "$EXISTING_LAW" != "None" ]]; then
    ok "Log Analytics workspace '$LOG_WORKSPACE_NAME' already exists."
    LAW_RESOURCE_ID="$EXISTING_LAW"
  else
    info "Creating Log Analytics workspace '$LOG_WORKSPACE_NAME'..."
    LAW_RESOURCE_ID=$(az monitor log-analytics workspace create \
      --workspace-name "$LOG_WORKSPACE_NAME" \
      --resource-group "$PLAY_RG" \
      --location "$LOCATION" \
      --retention-time 30 \
      --query id -o tsv)
    ok "Created Log Analytics workspace: $LAW_RESOURCE_ID"
  fi
fi

echo ""

# ─── 2. Action Group ───────────────────────────────────────────────────────
info "Checking action group '$ACTION_GROUP_NAME'..."

EXISTING_AG=$(az monitor action-group show \
  --name "$ACTION_GROUP_NAME" \
  --resource-group "$PLAY_RG" \
  --query id -o tsv 2>/dev/null || echo "")

if [[ -n "$EXISTING_AG" && "$EXISTING_AG" != "None" ]]; then
  ok "Action group '$ACTION_GROUP_NAME' already exists."
  AG_ID="$EXISTING_AG"
else
  info "Creating action group '$ACTION_GROUP_NAME' with email: $ALERT_EMAIL..."
  AG_ID=$(az monitor action-group create \
    --name "$ACTION_GROUP_NAME" \
    --resource-group "$PLAY_RG" \
    --short-name "xexplay-ag" \
    --action email xexplay-email "$ALERT_EMAIL" \
    --query id -o tsv)
  ok "Created action group: $AG_ID"
fi

echo ""

# ─── 3. Get Resource IDs ───────────────────────────────────────────────────
info "Resolving resource IDs..."

API_ID=$(az containerapp show \
  --name "$CONTAINER_APP_API" \
  --resource-group "$ENV_RG" \
  --query id -o tsv 2>/dev/null || echo "")

if [[ -z "$API_ID" || "$API_ID" == "None" ]]; then
  err "Container app '$CONTAINER_APP_API' not found in '$ENV_RG'."
  exit 1
fi
ok "API container app: $API_ID"

PG_ID=$(az postgres flexible-server show \
  --name "$POSTGRES_SERVER" \
  --resource-group "$PLAY_RG" \
  --query id -o tsv 2>/dev/null || echo "")

if [[ -z "$PG_ID" || "$PG_ID" == "None" ]]; then
  err "PostgreSQL server '$POSTGRES_SERVER' not found in '$PLAY_RG'."
  exit 1
fi
ok "PostgreSQL server: $PG_ID"

REDIS_ID=$(az redis show \
  --name "$REDIS_NAME" \
  --resource-group "$PLAY_RG" \
  --query id -o tsv 2>/dev/null || echo "")

if [[ -z "$REDIS_ID" || "$REDIS_ID" == "None" ]]; then
  err "Redis cache '$REDIS_NAME' not found in '$PLAY_RG'."
  exit 1
fi
ok "Redis cache: $REDIS_ID"

echo ""

# ─── 4. Alert Rules ────────────────────────────────────────────────────────

# Helper: create metric alert if it doesn't already exist
create_alert_if_missing() {
  local alert_name="$1"
  local resource_group="$2"
  shift 2

  EXISTING=$(az monitor metrics alert show \
    --name "$alert_name" \
    --resource-group "$resource_group" \
    --query id -o tsv 2>/dev/null || echo "")

  if [[ -n "$EXISTING" && "$EXISTING" != "None" ]]; then
    ok "Alert rule '$alert_name' already exists. Skipping."
    return 0
  fi

  info "Creating alert rule '$alert_name'..."
  az monitor metrics alert create \
    --name "$alert_name" \
    --resource-group "$resource_group" \
    --action "$AG_ID" \
    "$@"
  ok "Created alert rule '$alert_name'."
}

# 4a. API Container App Restart Count > 3 in 5 minutes
create_alert_if_missing \
  "xexplay-api-restart-alert" \
  "$PLAY_RG" \
  --scopes "$API_ID" \
  --condition "total RestartCount > 3" \
  --window-size 5m \
  --evaluation-frequency 1m \
  --severity 1 \
  --description "XEX Play API container restart count exceeded 3 in 5 minutes"

echo ""

# 4b. API Container App Response Time P95 > 2 seconds (2000ms)
create_alert_if_missing \
  "xexplay-api-response-time-alert" \
  "$PLAY_RG" \
  --scopes "$API_ID" \
  --condition "avg RequestDuration > 2000" \
  --window-size 5m \
  --evaluation-frequency 1m \
  --severity 2 \
  --description "XEX Play API P95 response time exceeded 2 seconds"

echo ""

# 4c. PostgreSQL CPU > 80% for 5 minutes
create_alert_if_missing \
  "xexplay-pg-cpu-alert" \
  "$PLAY_RG" \
  --scopes "$PG_ID" \
  --condition "avg cpu_percent > 80" \
  --window-size 5m \
  --evaluation-frequency 1m \
  --severity 2 \
  --description "XEX Play PostgreSQL CPU usage exceeded 80% for 5 minutes"

echo ""

# 4d. Redis Memory Usage > 80%
create_alert_if_missing \
  "xexplay-redis-memory-alert" \
  "$PLAY_RG" \
  --scopes "$REDIS_ID" \
  --condition "avg usedmemorypercentage > 80" \
  --window-size 5m \
  --evaluation-frequency 1m \
  --severity 2 \
  --description "XEX Play Redis memory usage exceeded 80%"

echo ""

# ─── 5. PostgreSQL Diagnostic Settings ──────────────────────────────────────
DIAG_NAME="xexplay-pg-diagnostics"

info "Checking diagnostic settings on PostgreSQL..."

EXISTING_DIAG=$(az monitor diagnostic-settings show \
  --name "$DIAG_NAME" \
  --resource "$PG_ID" \
  --query id -o tsv 2>/dev/null || echo "")

if [[ -n "$EXISTING_DIAG" && "$EXISTING_DIAG" != "None" ]]; then
  ok "Diagnostic setting '$DIAG_NAME' already exists on PostgreSQL."
else
  info "Enabling diagnostic logs on PostgreSQL server..."
  az monitor diagnostic-settings create \
    --name "$DIAG_NAME" \
    --resource "$PG_ID" \
    --workspace "$LAW_RESOURCE_ID" \
    --logs '[
      {"categoryGroup": "allLogs", "enabled": true, "retentionPolicy": {"enabled": false, "days": 0}}
    ]' \
    --metrics '[
      {"category": "AllMetrics", "enabled": true, "retentionPolicy": {"enabled": false, "days": 0}}
    ]'
  ok "Enabled diagnostic logs and metrics on PostgreSQL."
fi

echo ""
info "=== Azure Monitoring Setup Complete ==="
echo ""
echo "Summary:"
echo "  - Log Analytics Workspace: $LAW_RESOURCE_ID"
echo "  - Action Group: $AG_ID (email: $ALERT_EMAIL)"
echo "  - Alert: API restart count > 3 / 5min"
echo "  - Alert: API response time > 2s"
echo "  - Alert: PostgreSQL CPU > 80%"
echo "  - Alert: Redis memory > 80%"
echo "  - PostgreSQL diagnostic logs -> Log Analytics"
