#!/usr/bin/zsh
# Script to trigger manual amortization run for Pustaka CMDB
# Usage: ./run-amortization.sh [options]
#
# Options:
#   -d, --dry-run    Run in dry-run mode (no changes to database)
#   -w, --wait       Wait for completion and show results
#   -h, --help       Show this help message
#
# Environment variables:
#   API_URL       API endpoint URL (default: http://localhost:8080)
#   ADMIN_USER    Admin username (default: admin)
#   ADMIN_PASS    Admin password (default: Admin@123)

set -e

# Default values
API_URL="${API_URL:-http://localhost:8080}"
ADMIN_USER="${ADMIN_USER:-admin}"
ADMIN_PASS="${ADMIN_PASS:-Admin@123}"
DRY_RUN=false
WAIT=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -w|--wait)
            WAIT=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -d, --dry-run    Run in dry-run mode (no changes to database)"
            echo "  -w, --wait       Wait for completion and show results"
            echo "  -h, --help       Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  API_URL       API endpoint URL (default: http://localhost:8080)"
            echo "  ADMIN_USER    Admin username (default: admin)"
            echo "  ADMIN_PASS    Admin password (default: Admin@123)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Pustaka Amortization Manual Run${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Login and get access token
echo -e "${YELLOW}Step 1: Authenticating...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${API_URL}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${ADMIN_USER}\",\"password\":\"${ADMIN_PASS}\"}")

# Check if login was successful
if echo "$LOGIN_RESPONSE" | grep -q "access_token"; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
    echo -e "${GREEN}✓ Authentication successful${NC}"
else
    echo -e "${RED}✗ Authentication failed${NC}"
    echo "$LOGIN_RESPONSE"
    exit 1
fi

# Step 2: Trigger manual amortization run
echo ""
echo -e "${YELLOW}Step 2: Triggering amortization run...${NC}"

if [ "$DRY_RUN" = true ]; then
    echo -e "${BLUE}Mode: Dry-run (no database changes)${NC}"
    PAYLOAD='{"dry_run": true}'
else
    echo -e "${BLUE}Mode: Live (will create ledger entries)${NC}"
    PAYLOAD='{"dry_run": false}'
fi

RUN_RESPONSE=$(curl -s -X POST "${API_URL}/api/v1/amortization/runs" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TOKEN}" \
    -d "${PAYLOAD}")

# Check if run was initiated
if echo "$RUN_RESPONSE" | grep -q "run_id"; then
    RUN_ID=$(echo "$RUN_RESPONSE" | jq -r '.run_id')
    echo -e "${GREEN}✓ Amortization run initiated${NC}"
    echo -e "  Run ID: ${BLUE}${RUN_ID}${NC}"
else
    echo -e "${RED}✗ Failed to trigger amortization run${NC}"
    echo "$RUN_RESPONSE"
    exit 1
fi

# Step 3: Wait for completion and show results (if requested)
if [ "$WAIT" = true ]; then
    echo ""
    echo -e "${YELLOW}Step 3: Waiting for completion...${NC}"

    # Poll for completion (max 30 seconds)
    MAX_WAIT=30
    ELAPSED=0
    while [ $ELAPSED -lt $MAX_WAIT ]; do
        sleep 2
        ELAPSED=$((ELAPSED + 2))

        STATUS_RESPONSE=$(curl -s "${API_URL}/api/v1/amortization/runs/${RUN_ID}" \
            -H "Authorization: Bearer ${TOKEN}")

        RUN_STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.status')

        if [ "$RUN_STATUS" = "completed" ] || [ "$RUN_STATUS" = "failed" ] || [ "$RUN_STATUS" = "partial" ]; then
            break
        fi

        echo -ne "\r  Status: ${RUN_STATUS} (${ELAPSED}s)"
    done
    echo ""

    # Display results
    echo -e "${YELLOW}Step 4: Run Results${NC}"
    echo "$STATUS_RESPONSE" | jq '{
        id,
        status,
        is_manual,
        processing_date,
        total_amortizable_cis,
        processed_cis,
        dry_run,
        started_at,
        completed_at
    }'

    # Extract some stats
    PROCESSED=$(echo "$STATUS_RESPONSE" | jq -r '.processed_cis // 0')
    TOTAL=$(echo "$STATUS_RESPONSE" | jq -r '.total_amortizable_cis // 0')
    STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.status')

    echo ""
    if [ "$STATUS" = "completed" ]; then
        echo -e "${GREEN}✓ Run completed successfully${NC}"
        echo -e "  Processed: ${BLUE}${PROCESSED}${NC} / ${TOTAL} CIs"
    elif [ "$STATUS" = "partial" ]; then
        echo -e "${YELLOW}⚠ Run completed with partial success${NC}"
        echo -e "  Processed: ${BLUE}${PROCESSED}${NC} / ${TOTAL} CIs"
    else
        echo -e "${RED}✗ Run failed${NC}"
    fi
else
    echo ""
    echo -e "${BLUE}Run ID: ${RUN_ID}${NC}"
    echo -e "To check status, run:"
    echo -e "  curl -s \"${API_URL}/api/v1/amortization/runs/${RUN_ID}\" \\"
    echo -e "    -H \"Authorization: Bearer ${TOKEN}\" | jq ."
fi

echo ""
echo -e "${GREEN}Done!${NC}"
