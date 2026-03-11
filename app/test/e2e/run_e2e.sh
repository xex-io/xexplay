#!/bin/bash
set -euo pipefail

export API_BASE_URL="${API_BASE_URL:-https://xexplay-api.yellowfield-3543caf0.northeurope.azurecontainerapps.io/v1}"

cd /Users/nima/Developments/xexplay/app

echo "Running E2E tests against: $API_BASE_URL"
flutter test test/e2e/ --dart-define=API_BASE_URL="$API_BASE_URL"
