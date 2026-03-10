#!/usr/bin/env bash
#
# XEX Play - Disaster Recovery (DR) Test Script
#
# Validates that disaster recovery mechanisms are in place and functioning.
# This script is strictly read-only -- it never kills, restarts, or modifies
# any resource. It verifies that the infrastructure is configured to handle
# failures automatically.
#
# What this script checks:
#   1. Container auto-restart configuration (Azure Container Apps built-in)
#   2. Redis connectivity and persistence configuration
#   3. PostgreSQL connectivity, HA, and backup configuration
#   4. Container app replica health across all active revisions
#   5. Azure Monitor alert rules are configured and enabled
#
# Prerequisites:
#   - Azure CLI logged in with at least Reader role
#
# Usage:
#   ./disaster-recovery-test.sh
#
set -euo pipefail

# ─── Configuration ──────────────────────────────────────────────────────────
PLAY_RG="${PLAY_RG:-xexplay-production-rg}"
ENV_RG="${ENV_RG:-xex-production-rg}"

CONTAINER_APP_API="${CONTAINER_APP_API:-xexplay-api}"
CONTAINER_APP_ADMIN="${CONTAINER_APP_ADMIN:-xexplay-admin}"
CONTAINER_APP_ENV="${CONTAINER_APP_ENV:-xex-production-env}"

POSTGRES_SERVER="${POSTGRES_SERVER:-xexplay-postgres}"
REDIS_NAME="${REDIS_NAME:-xexplay-redis}"

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
  exit 1
fi

if ! az account show &>/dev/null; then
  fail "Azure CLI is not logged in. Run 'az login' first."
  exit 1
fi

ACCOUNT=$(az account show --query name -o tsv 2>/dev/null)
info "Azure subscription: $ACCOUNT"
info "DR test started at: $(date -u '+%Y-%m-%d %H:%M:%S UTC')"
sep

# ═══════════════════════════════════════════════════════════════════════════
# 1. Container Auto-Restart / Self-Healing Configuration
# ═══════════════════════════════════════════════════════════════════════════
echo "============================================"
echo " 1. Container Auto-Restart / Self-Healing"
echo "============================================"
info "Azure Container Apps automatically restarts crashed containers."
info "Verifying that health probes are configured to detect failures."
sep

for APP_NAME in "$CONTAINER_APP_API" "$CONTAINER_APP_ADMIN"; do
  info "--- Container App: $APP_NAME ---"

  # Check if the container app exists
  APP_EXISTS=$(az containerapp show \
    --name "$APP_NAME" \
    --resource-group "$ENV_RG" \
    --query "id" \
    -o tsv 2>/dev/null || echo "")

  if [[ -z "$APP_EXISTS" || "$APP_EXISTS" == "None" ]]; then
    fail "Container app '$APP_NAME' not found in '$ENV_RG'. Cannot verify DR config."
    sep
    continue
  fi

  # Check restart policy
  # Azure Container Apps always restarts crashed containers (built-in behavior).
  # We verify the container is configured and running.
  RUNNING_STATE=$(az containerapp show \
    --name "$APP_NAME" \
    --resource-group "$ENV_RG" \
    --query "properties.runningStatus" \
    -o tsv 2>/dev/null || echo "UNKNOWN")

  if [[ "$RUNNING_STATE" == "Running" ]]; then
    ok "$APP_NAME is running. Azure will auto-restart on crash."
  else
    warn "$APP_NAME current state: $RUNNING_STATE"
  fi

  # Check min/max replicas (scaling config)
  MIN_REPLICAS=$(az containerapp show \
    --name "$APP_NAME" \
    --resource-group "$ENV_RG" \
    --query "properties.template.scale.minReplicas" \
    -o tsv 2>/dev/null || echo "0")

  MAX_REPLICAS=$(az containerapp show \
    --name "$APP_NAME" \
    --resource-group "$ENV_RG" \
    --query "properties.template.scale.maxReplicas" \
    -o tsv 2>/dev/null || echo "0")

  info "$APP_NAME replicas: min=$MIN_REPLICAS, max=$MAX_REPLICAS"

  if [[ "$MIN_REPLICAS" -ge 1 ]]; then
    ok "$APP_NAME has minReplicas >= 1, ensuring at least one instance is always running."
  else
    warn "$APP_NAME has minReplicas=$MIN_REPLICAS. Container may scale to zero and be cold-started."
  fi

  # Check health probes
  LIVENESS_PROBE=$(az containerapp show \
    --name "$APP_NAME" \
    --resource-group "$ENV_RG" \
    --query "properties.template.containers[0].probes[?type=='Liveness'] | length(@)" \
    -o tsv 2>/dev/null || echo "0")

  READINESS_PROBE=$(az containerapp show \
    --name "$APP_NAME" \
    --resource-group "$ENV_RG" \
    --query "properties.template.containers[0].probes[?type=='Readiness'] | length(@)" \
    -o tsv 2>/dev/null || echo "0")

  if [[ "$LIVENESS_PROBE" -gt 0 ]]; then
    ok "$APP_NAME has a liveness probe configured."
  else
    warn "$APP_NAME has NO liveness probe. Azure will only restart on process exit, not on hangs."
  fi

  if [[ "$READINESS_PROBE" -gt 0 ]]; then
    ok "$APP_NAME has a readiness probe configured."
  else
    warn "$APP_NAME has NO readiness probe. Traffic may be sent to unhealthy instances."
  fi

  sep
done

# ═══════════════════════════════════════════════════════════════════════════
# 2. Redis Connectivity and Persistence
# ═══════════════════════════════════════════════════════════════════════════
echo "============================================"
echo " 2. Redis Connectivity & Persistence"
echo "============================================"

REDIS_STATE=$(az redis show \
  --name "$REDIS_NAME" \
  --resource-group "$PLAY_RG" \
  --query "provisioningState" \
  -o tsv 2>/dev/null || echo "UNAVAILABLE")

if [[ "$REDIS_STATE" == "Succeeded" ]]; then
  ok "Redis cache '$REDIS_NAME' is provisioned and available."
else
  fail "Redis cache state: $REDIS_STATE (expected: Succeeded)."
fi

REDIS_SKU=$(az redis show \
  --name "$REDIS_NAME" \
  --resource-group "$PLAY_RG" \
  --query "sku.name" \
  -o tsv 2>/dev/null || echo "UNKNOWN")

REDIS_FAMILY=$(az redis show \
  --name "$REDIS_NAME" \
  --resource-group "$PLAY_RG" \
  --query "sku.family" \
  -o tsv 2>/dev/null || echo "UNKNOWN")

REDIS_CAPACITY=$(az redis show \
  --name "$REDIS_NAME" \
  --resource-group "$PLAY_RG" \
  --query "sku.capacity" \
  -o tsv 2>/dev/null || echo "UNKNOWN")

info "Redis SKU: $REDIS_SKU, Family: $REDIS_FAMILY, Capacity: $REDIS_CAPACITY"

if [[ "$REDIS_SKU" == "Premium" || "$REDIS_SKU" == "Standard" ]]; then
  ok "Redis tier ($REDIS_SKU) supports data persistence and/or replication."
else
  warn "Redis tier is '$REDIS_SKU'. Basic tier has no replication or SLA. Consider upgrading for DR."
fi

# Check if Redis data persistence (RDB) is enabled (Premium tier only)
if [[ "$REDIS_SKU" == "Premium" ]]; then
  RDB_ENABLED=$(az redis show \
    --name "$REDIS_NAME" \
    --resource-group "$PLAY_RG" \
    --query "redisConfiguration.\"rdb-backup-enabled\"" \
    -o tsv 2>/dev/null || echo "unknown")

  if [[ "$RDB_ENABLED" == "true" ]]; then
    ok "Redis RDB persistence is enabled."
  else
    warn "Redis RDB persistence is NOT enabled. Data will be lost on full cache failure."
  fi
fi

# TCP connectivity check
REDIS_HOST=$(az redis show \
  --name "$REDIS_NAME" \
  --resource-group "$PLAY_RG" \
  --query "hostName" \
  -o tsv 2>/dev/null || echo "")

if [[ -n "$REDIS_HOST" && "$REDIS_HOST" != "None" ]]; then
  REDIS_SSL_PORT=$(az redis show \
    --name "$REDIS_NAME" \
    --resource-group "$PLAY_RG" \
    --query "sslPort" \
    -o tsv 2>/dev/null || echo "6380")

  if command -v nc &>/dev/null; then
    if nc -z -w 5 "$REDIS_HOST" "$REDIS_SSL_PORT" 2>/dev/null; then
      ok "Redis is reachable at $REDIS_HOST:$REDIS_SSL_PORT."
    else
      warn "Redis is not reachable from this machine at $REDIS_HOST:$REDIS_SSL_PORT (may be VNet-restricted)."
    fi
  else
    info "Skipping TCP check (nc not available)."
  fi
fi
sep

# ═══════════════════════════════════════════════════════════════════════════
# 3. PostgreSQL Connectivity, HA & Backups
# ═══════════════════════════════════════════════════════════════════════════
echo "============================================"
echo " 3. PostgreSQL Connectivity, HA & Backups"
echo "============================================"

PG_STATE=$(az postgres flexible-server show \
  --name "$POSTGRES_SERVER" \
  --resource-group "$PLAY_RG" \
  --query "state" \
  -o tsv 2>/dev/null || echo "UNAVAILABLE")

if [[ "$PG_STATE" == "Ready" ]]; then
  ok "PostgreSQL server '$POSTGRES_SERVER' state: $PG_STATE"
else
  fail "PostgreSQL server state: $PG_STATE (expected: Ready)."
fi

# High Availability mode
HA_MODE=$(az postgres flexible-server show \
  --name "$POSTGRES_SERVER" \
  --resource-group "$PLAY_RG" \
  --query "highAvailability.mode" \
  -o tsv 2>/dev/null || echo "Disabled")

if [[ "$HA_MODE" == "ZoneRedundant" || "$HA_MODE" == "SameZone" ]]; then
  ok "PostgreSQL HA is enabled (mode: $HA_MODE). Automatic failover is available."
else
  warn "PostgreSQL HA mode: $HA_MODE. No automatic failover. Consider enabling HA for production."
fi

# Backup retention
BACKUP_RETENTION=$(az postgres flexible-server show \
  --name "$POSTGRES_SERVER" \
  --resource-group "$PLAY_RG" \
  --query "backup.backupRetentionDays" \
  -o tsv 2>/dev/null || echo "UNKNOWN")

info "PostgreSQL backup retention: $BACKUP_RETENTION days"

if [[ "$BACKUP_RETENTION" != "UNKNOWN" && "$BACKUP_RETENTION" -ge 7 ]]; then
  ok "Backup retention is $BACKUP_RETENTION days (>= 7 day minimum recommended)."
else
  warn "Backup retention is $BACKUP_RETENTION days. Consider increasing to at least 7."
fi

# Geo-redundant backup
GEO_BACKUP=$(az postgres flexible-server show \
  --name "$POSTGRES_SERVER" \
  --resource-group "$PLAY_RG" \
  --query "backup.geoRedundantBackup" \
  -o tsv 2>/dev/null || echo "Disabled")

if [[ "$GEO_BACKUP" == "Enabled" ]]; then
  ok "Geo-redundant backup is enabled."
else
  info "Geo-redundant backup: $GEO_BACKUP (enable for cross-region DR)."
fi

# TCP connectivity check
PG_FQDN=$(az postgres flexible-server show \
  --name "$POSTGRES_SERVER" \
  --resource-group "$PLAY_RG" \
  --query "fullyQualifiedDomainName" \
  -o tsv 2>/dev/null || echo "")

if [[ -n "$PG_FQDN" && "$PG_FQDN" != "None" ]]; then
  if command -v nc &>/dev/null; then
    if nc -z -w 5 "$PG_FQDN" 5432 2>/dev/null; then
      ok "PostgreSQL is reachable at $PG_FQDN:5432."
    else
      warn "PostgreSQL is not reachable from this machine (may be VNet-restricted)."
    fi
  else
    info "Skipping TCP check (nc not available)."
  fi
fi
sep

# ═══════════════════════════════════════════════════════════════════════════
# 4. Container App Replica Health
# ═══════════════════════════════════════════════════════════════════════════
echo "============================================"
echo " 4. Container App Replica Health"
echo "============================================"

for APP_NAME in "$CONTAINER_APP_API" "$CONTAINER_APP_ADMIN"; do
  info "--- $APP_NAME ---"

  REVISIONS_JSON=$(az containerapp revision list \
    --name "$APP_NAME" \
    --resource-group "$ENV_RG" \
    --query "[?properties.active==\`true\`].{name:name, replicas:properties.replicas, runningState:properties.runningState, healthState:properties.healthState, trafficWeight:properties.trafficWeight}" \
    -o json 2>/dev/null || echo "[]")

  ACTIVE_COUNT=$(echo "$REVISIONS_JSON" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")

  if [[ "$ACTIVE_COUNT" -gt 0 ]]; then
    ok "$APP_NAME has $ACTIVE_COUNT active revision(s)."

    # Check each revision's running state
    echo "$REVISIONS_JSON" | python3 -c "
import sys, json
revisions = json.load(sys.stdin)
for r in revisions:
    name = r.get('name', 'unknown')
    state = r.get('runningState', 'unknown')
    replicas = r.get('replicas', 0)
    health = r.get('healthState', 'unknown')
    print(f'  Revision: {name} | State: {state} | Replicas: {replicas} | Health: {health}')
" 2>/dev/null || true

    # Check for unhealthy revisions
    UNHEALTHY=$(echo "$REVISIONS_JSON" | python3 -c "
import sys, json
revisions = json.load(sys.stdin)
unhealthy = [r['name'] for r in revisions if r.get('runningState') not in ('Running', 'Processing')]
print(','.join(unhealthy) if unhealthy else '')
" 2>/dev/null || echo "")

    if [[ -z "$UNHEALTHY" ]]; then
      ok "All active revisions of $APP_NAME are in a healthy state."
    else
      fail "Unhealthy revisions detected for $APP_NAME: $UNHEALTHY"
    fi
  else
    fail "$APP_NAME has no active revisions."
  fi
  sep
done

# ═══════════════════════════════════════════════════════════════════════════
# 5. Azure Monitor Alert Rules
# ═══════════════════════════════════════════════════════════════════════════
echo "============================================"
echo " 5. Azure Monitor Alert Rules"
echo "============================================"

ALERT_RULES=$(az monitor metrics alert list \
  --resource-group "$PLAY_RG" \
  --query "[].{name:name, severity:severity, enabled:enabled, description:description}" \
  -o json 2>/dev/null || echo "[]")

ALERT_COUNT=$(echo "$ALERT_RULES" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")

if [[ "$ALERT_COUNT" -gt 0 ]]; then
  ok "Found $ALERT_COUNT metric alert rule(s) in '$PLAY_RG'."

  echo "$ALERT_RULES" | python3 -c "
import sys, json
alerts = json.load(sys.stdin)
for a in alerts:
    name = a.get('name', 'unknown')
    sev = a.get('severity', '?')
    enabled = a.get('enabled', False)
    desc = a.get('description', '')
    status = 'ENABLED' if enabled else 'DISABLED'
    print(f'  [{status}] Sev {sev}: {name}')
    if desc:
        print(f'           {desc}')
" 2>/dev/null || true

  # Check for disabled alerts
  DISABLED_COUNT=$(echo "$ALERT_RULES" | python3 -c "
import sys, json
alerts = json.load(sys.stdin)
print(len([a for a in alerts if not a.get('enabled', False)]))
" 2>/dev/null || echo "0")

  if [[ "$DISABLED_COUNT" -gt 0 ]]; then
    warn "$DISABLED_COUNT alert rule(s) are DISABLED. Re-enable them for production monitoring."
  else
    ok "All alert rules are enabled."
  fi
else
  fail "No metric alert rules found in '$PLAY_RG'. Alerting is not configured."
  info "Run setup-azure-monitoring.sh to create alert rules."
fi

sep

# Check for action groups (notification targets)
ACTION_GROUPS=$(az monitor action-group list \
  --resource-group "$PLAY_RG" \
  --query "[].{name:name, enabled:enabled}" \
  -o json 2>/dev/null || echo "[]")

AG_COUNT=$(echo "$ACTION_GROUPS" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")

if [[ "$AG_COUNT" -gt 0 ]]; then
  ok "Found $AG_COUNT action group(s) for alert notifications."
else
  fail "No action groups found. Alerts will fire but nobody will be notified."
fi
sep

# ═══════════════════════════════════════════════════════════════════════════
# Summary
# ═══════════════════════════════════════════════════════════════════════════
echo "============================================"
echo " Disaster Recovery Test Summary"
echo "============================================"
echo -e " \033[1;32mPassed:\033[0m  $PASS"
echo -e " \033[1;31mFailed:\033[0m  $FAIL"
echo -e " \033[1;33mWarnings:\033[0m $WARN"
echo "============================================"
echo ""
echo "DR Expectations (documented):"
echo "  - Container crash: Azure Container Apps auto-restarts the container."
echo "    No manual intervention required. Health probes detect hangs."
echo "  - Redis failure: Standard/Premium tier has replication. Basic does not."
echo "    XEX Play should treat Redis as a cache layer (rebuild from DB)."
echo "  - PostgreSQL failure: If HA is enabled, automatic failover occurs."
echo "    Backups allow point-in-time restore within retention window."
echo "  - Full region outage: Requires geo-redundant backup + manual failover"
echo "    to a secondary region."
echo ""

if [[ $FAIL -gt 0 ]]; then
  fail "DR test completed with $FAIL failure(s). Review and remediate before go-live."
  exit 1
else
  ok "All DR checks passed. Infrastructure is configured for resilience."
  exit 0
fi
