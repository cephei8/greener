#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

if [ -z "$GREENER_INGRESS_ENDPOINT" ]; then
    echo -e "${RED}Error: GREENER_INGRESS_ENDPOINT environment variable is not set${NC}"
    exit 1
fi

if [ -z "$GREENER_INGRESS_API_KEY" ]; then
    echo -e "${RED}Error: GREENER_INGRESS_API_KEY environment variable is not set${NC}"
    exit 1
fi

ENDPOINT="$GREENER_INGRESS_ENDPOINT"
API_KEY="$GREENER_INGRESS_API_KEY"

echo -e "${GREEN}==================================================================${NC}"
echo -e "${GREEN}  Greener Platform - Comprehensive Sample Data Generator${NC}"
echo -e "${GREEN}==================================================================${NC}"
echo ""
echo "Endpoint: $ENDPOINT"
echo ""

run_cli() {
    ./greener-reporter-cli --endpoint "$ENDPOINT" --api-key "$API_KEY" "$@"
}

# Test suites and their test cases
SUITE_NAMES=(
    "AuthenticationTests"
    "DatabaseTests"
    "APITests"
    "IntegrationTests"
    "PerformanceTests"
    "SecurityTests"
)

get_tests_for_suite() {
    case "$1" in
        AuthenticationTests)
            echo "test_login_success test_login_invalid_password test_login_missing_credentials test_logout test_session_timeout test_token_refresh test_password_reset"
            ;;
        DatabaseTests)
            echo "test_connection_pool test_query_users test_insert_record test_update_record test_delete_record test_transaction_rollback test_concurrent_writes test_index_performance"
            ;;
        APITests)
            echo "test_get_user test_create_user test_update_user test_delete_user test_list_users test_pagination test_rate_limiting test_api_versioning"
            ;;
        IntegrationTests)
            echo "test_end_to_end_flow test_payment_processing test_email_notification test_webhook_delivery test_third_party_integration"
            ;;
        PerformanceTests)
            echo "test_load_1000_users test_concurrent_requests test_memory_usage test_response_time test_throughput"
            ;;
        SecurityTests)
            echo "test_sql_injection test_xss_protection test_csrf_token test_auth_bypass test_data_encryption"
            ;;
    esac
}

# Session configuration matrix
VERSIONS=("1.0.0" "1.1.0" "1.2.0" "2.0.0" "2.1.0")
ARCHITECTURES=("x86_64" "arm64" "aarch64")
OPERATING_SYSTEMS=("linux" "windows" "macos")
ENVIRONMENTS=("ci" "staging" "production")
REGIONS=("us-east" "eu-west" "ap-south")

# Counter for progress tracking
TOTAL_SESSIONS=0
SESSION_COUNT=0

# Calculate total sessions (we'll create various combinations)
# - Core matrix: versions × architectures × OS (simplified)
# - Plus special sessions with different label/baggage patterns
TOTAL_SESSIONS=$((${#VERSIONS[@]} * ${#ARCHITECTURES[@]} * 3 + 10))

echo -e "${BLUE}Planned sessions: $TOTAL_SESSIONS${NC}"
echo ""

create_session_with_tests() {
    local version="$1"
    local arch="$2"
    local os="$3"
    local env="$4"
    local region="$5"
    local baggage_type="$6"
    local extra_labels="$7"

    SESSION_COUNT=$((SESSION_COUNT + 1))
    echo -e "${YELLOW}[$SESSION_COUNT/$TOTAL_SESSIONS] Creating session: version=$version, arch=$arch, os=$os, env=$env${NC}"

    # Build label arguments
    local label_args=(
        --label "version=$version"
        --label "arch=$arch"
        --label "os=$os"
    )

    # Add environment label
    if [ -n "$env" ]; then
        label_args+=(--label "env=$env")
    fi

    # Add region label
    if [ -n "$region" ]; then
        label_args+=(--label "region=$region")
    fi

    # Add extra labels (comma-separated)
    if [ -n "$extra_labels" ]; then
        IFS=',' read -ra EXTRA <<< "$extra_labels"
        for label in "${EXTRA[@]}"; do
            label_args+=(--label "$label")
        done
    fi

    # Build session description
    local description="Test run for version $version on $os/$arch in $env environment"
    if [ -n "$region" ]; then
        description="$description (region: $region)"
    fi

    # Build baggage argument based on type
    local baggage_arg=""
    case "$baggage_type" in
        "none")
            # No baggage argument
            ;;
        "empty")
            baggage_arg='--baggage {}'
            ;;
        "simple")
            baggage_arg='--baggage {"runner":"github-actions","build":"'$SESSION_COUNT'"}'
            ;;
        "gitlab")
            baggage_arg='--baggage {"runner":"gitlab-ci","pipeline_id":"'$((1000 + SESSION_COUNT))'","job":"test"}'
            ;;
        "jenkins")
            baggage_arg='--baggage {"runner":"jenkins","build_number":'$SESSION_COUNT',"job_name":"integration-tests"}'
            ;;
        "complex")
            baggage_arg='--baggage {"runner":"github-actions","build":'$SESSION_COUNT',"metadata":{"branch":"main","commit":"abc'$SESSION_COUNT'","pr":null},"tags":["ci","automated"],"parallel":true}'
            ;;
    esac

    # Create session with description
    local cmd_args=(create session "${label_args[@]}" --description "$description")
    if [ -n "$baggage_arg" ]; then
        cmd_args+=($baggage_arg)
    fi

    SESSION_OUTPUT=$(run_cli "${cmd_args[@]}")
    SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')

    if [ -z "$SESSION_ID" ]; then
        echo -e "${RED}Error: Could not extract session ID${NC}"
        exit 1
    fi

    echo "  Session ID: $SESSION_ID"

    # Generate test cases for this session
    local test_count=0

    for suite_name in "${SUITE_NAMES[@]}"; do
        IFS=' ' read -ra TESTS <<< "$(get_tests_for_suite "$suite_name")"

        for test_name in "${TESTS[@]}"; do
            test_count=$((test_count + 1))

            # Determine test status and output based on various conditions
            local status="pass"
            local output=""
            local classname="$suite_name"
            local file="tests/$(echo "$suite_name" | tr '[:upper:]' '[:lower:]').py"
            local testsuite="$suite_name"
            local test_baggage=""

            # Create interesting failure patterns

            # Version-specific failures
            if [ "$test_name" = "test_session_timeout" ] && [ "$version" = "1.0.0" ]; then
                status="fail"
                output="AssertionError: Session did not timeout after 30 minutes
  Expected: SessionExpired
  Got: ActiveSession
  at assert_session_expired() line 67

  Stack trace:
    File \"tests/authenticationtests.py\", line 67, in test_session_timeout
      assert session.is_expired() == True
    AssertionError: Session still active after timeout period"
            fi

            # OS-specific failures
            if [ "$test_name" = "test_transaction_rollback" ] && [ "$os" = "windows" ]; then
                status="fail"
                output="DatabaseError: Rollback failed - transaction already committed
  at connection.rollback() line 42
  Database: PostgreSQL 14.5
  Connection pool: exhausted"
            fi

            # Architecture-specific failures
            if [ "$test_name" = "test_memory_usage" ] && [ "$arch" = "arm64" ]; then
                status="error"
                output="MemoryError: Memory allocation failed
  Requested: 2048MB
  Available: 1024MB
  Architecture: arm64 (memory-constrained environment)"
            fi

            # Environment-specific failures
            if [ "$test_name" = "test_payment_processing" ] && [ "$env" = "staging" ]; then
                status="fail"
                output="PaymentError: SSL certificate verification failed
  Expected: valid certificate
  Got: self-signed certificate
  URL: https://payment-gateway.staging.example.com"
            fi

            # Region-specific issues
            if [ "$test_name" = "test_third_party_integration" ] && [ "$region" = "ap-south" ]; then
                status="error"
                output="TimeoutError: Connection to third-party API timed out
  Endpoint: api.partner.com
  Region: ap-south
  Latency: >5000ms"
            fi

            # Skip patterns
            if [ "$test_name" = "test_webhook_delivery" ] && [ "$version" = "1.0.0" ]; then
                status="skip"
                output="Skipped: Webhook service not available in version 1.0.0"
            fi

            if [ "$test_name" = "test_delete_record" ] && [ "$env" = "production" ]; then
                status="skip"
                output="Skipped: Destructive operations disabled in production environment"
            fi

            # Use modulo patterns to create varied but deterministic field inclusion
            # Different modulos create an alternating pattern that appears random
            local include_classname=$((test_count % 3))
            local include_file=$((test_count % 5))
            local include_testsuite=$((test_count % 7))
            local include_baggage=$((test_count % 11))

            # Add test baggage for every 11th test
            if [ $include_baggage -eq 0 ]; then
                test_baggage='--baggage {"duration_ms":'$((100 + (test_count * 37) % 900))',"retry_count":'$((test_count % 3))'}'
            fi

            # Build test case command with patterned fields
            local cmd_args=(
                create testcase
                --session-id "$SESSION_ID"
                --name "$test_name"
                --status "$status"
            )

            # Include optional fields based on modulo patterns
            if [ $include_classname -ne 0 ]; then
                cmd_args+=(--classname "$classname")
            fi

            if [ $include_file -ne 0 ]; then
                cmd_args+=(--file "$file")
            fi

            if [ $include_testsuite -ne 0 ]; then
                cmd_args+=(--testsuite "$testsuite")
            fi

            if [ -n "$output" ]; then
                cmd_args+=(--output "$output")
            fi

            if [ -n "$test_baggage" ]; then
                cmd_args+=($test_baggage)
            fi

            run_cli "${cmd_args[@]}" > /dev/null
        done
    done

    echo "  Created $test_count test cases"
    echo ""
}

# Generate sessions with various combinations

echo -e "${BLUE}Phase 1: Core version × architecture × OS matrix${NC}"
echo ""

# Create a representative subset (not full cartesian product to keep it reasonable)
for version in "${VERSIONS[@]}"; do
    for arch in "${ARCHITECTURES[@]}"; do
        # Vary OS and environment
        os="${OPERATING_SYSTEMS[$((RANDOM % ${#OPERATING_SYSTEMS[@]}))]}"
        env="${ENVIRONMENTS[$((RANDOM % ${#ENVIRONMENTS[@]}))]}"
        region="${REGIONS[$((RANDOM % ${#REGIONS[@]}))]}"

        # Vary baggage type
        baggage_types=("none" "empty" "simple" "gitlab" "jenkins" "complex")
        baggage_type="${baggage_types[$((RANDOM % ${#baggage_types[@]}))]}"

        create_session_with_tests "$version" "$arch" "$os" "$env" "$region" "$baggage_type" ""
    done
done

echo -e "${BLUE}Phase 2: Special sessions with labels without values${NC}"
echo ""

# Session with "approved" label (no value)
SESSION_COUNT=$((SESSION_COUNT + 1))
echo -e "${YELLOW}[$SESSION_COUNT/$TOTAL_SESSIONS] Creating session with 'approved' label${NC}"
SESSION_OUTPUT=$(run_cli create session \
    --label "version=2.1.0" \
    --label "arch=x86_64" \
    --label "os=linux" \
    --label "approved" \
    --description "Production deployment approved by admin team" \
    --baggage '{"runner":"manual","approved_by":"admin"}')
SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
echo "  Session ID: $SESSION_ID (approved)"
run_cli create testcase --session-id "$SESSION_ID" --name "test_approved_deployment" --status "pass" > /dev/null
echo "  Created 1 test case"
echo ""

# Session with "denied" label (no value)
SESSION_COUNT=$((SESSION_COUNT + 1))
echo -e "${YELLOW}[$SESSION_COUNT/$TOTAL_SESSIONS] Creating session with 'denied' label${NC}"
SESSION_OUTPUT=$(run_cli create session \
    --label "version=1.0.0" \
    --label "arch=arm64" \
    --label "os=windows" \
    --label "denied" \
    --description "Deployment denied due to outdated version and security concerns" \
    --baggage '{"runner":"manual","denied_by":"security_team","reason":"outdated_version"}')
SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
echo "  Session ID: $SESSION_ID (denied)"
run_cli create testcase --session-id "$SESSION_ID" --name "test_security_check" --status "fail" --output "Security check failed: Version 1.0.0 is not approved for production" > /dev/null
echo "  Created 1 test case"
echo ""

# Session with "nightly" label (no value)
SESSION_COUNT=$((SESSION_COUNT + 1))
echo -e "${YELLOW}[$SESSION_COUNT/$TOTAL_SESSIONS] Creating session with 'nightly' label${NC}"
SESSION_OUTPUT=$(run_cli create session \
    --label "version=2.1.0" \
    --label "arch=x86_64" \
    --label "os=linux" \
    --label "nightly" \
    --label "env=ci" \
    --description "Automated nightly build and test run from main branch" \
    --baggage '{"runner":"github-actions","schedule":"0 0 * * *","branch":"main"}')
SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
echo "  Session ID: $SESSION_ID (nightly)"
run_cli create testcase --session-id "$SESSION_ID" --name "test_nightly_build" --status "pass" > /dev/null
run_cli create testcase --session-id "$SESSION_ID" --name "test_integration_all" --status "pass" > /dev/null
echo "  Created 2 test cases"
echo ""

# Session with "release" label (no value)
SESSION_COUNT=$((SESSION_COUNT + 1))
echo -e "${YELLOW}[$SESSION_COUNT/$TOTAL_SESSIONS] Creating session with 'release' label${NC}"
SESSION_OUTPUT=$(run_cli create session \
    --label "version=2.0.0" \
    --label "arch=x86_64" \
    --label "os=linux" \
    --label "release" \
    --label "env=production" \
    --description "Official release candidate v2.0.0 validation tests" \
    --baggage '{"runner":"jenkins","release_tag":"v2.0.0","artifacts":["binary","docs","checksum"]}')
SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
echo "  Session ID: $SESSION_ID (release)"
run_cli create testcase --session-id "$SESSION_ID" --name "test_release_smoke" --status "pass" > /dev/null
echo "  Created 1 test case"
echo ""

# Session with no baggage (only description)
SESSION_COUNT=$((SESSION_COUNT + 1))
echo -e "${YELLOW}[$SESSION_COUNT/$TOTAL_SESSIONS] Creating session with no baggage${NC}"
SESSION_OUTPUT=$(run_cli create session \
    --label "version=1.1.0" \
    --label "arch=aarch64" \
    --label "os=macos" \
    --description "Basic test run on macOS ARM64 without additional metadata")
SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
echo "  Session ID: $SESSION_ID (no baggage)"
run_cli create testcase --session-id "$SESSION_ID" --name "test_basic" --status "pass" > /dev/null
echo "  Created 1 test case"
echo ""

# Session with empty baggage
SESSION_COUNT=$((SESSION_COUNT + 1))
echo -e "${YELLOW}[$SESSION_COUNT/$TOTAL_SESSIONS] Creating session with empty baggage${NC}"
SESSION_OUTPUT=$(run_cli create session \
    --label "version=1.2.0" \
    --label "arch=x86_64" \
    --label "os=windows" \
    --description "Windows x86_64 compatibility test for version 1.2.0" \
    --baggage '{}')
SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
echo "  Session ID: $SESSION_ID (empty baggage)"
run_cli create testcase --session-id "$SESSION_ID" --name "test_windows_specific" --status "pass" > /dev/null
echo "  Created 1 test case"
echo ""

# Session with complex nested baggage
SESSION_COUNT=$((SESSION_COUNT + 1))
echo -e "${YELLOW}[$SESSION_COUNT/$TOTAL_SESSIONS] Creating session with complex nested baggage${NC}"
SESSION_OUTPUT=$(run_cli create session \
    --label "version=2.1.0" \
    --label "arch=arm64" \
    --label "os=linux" \
    --label "env=staging" \
    --description "Multi-matrix CI/CD pipeline test with Python 3.9-3.11 and Node 16-20" \
    --baggage '{"runner":"github-actions","workflow":{"id":"123456","name":"CI/CD Pipeline","trigger":"pull_request"},"matrix":{"python":["3.9","3.10","3.11"],"node":["16","18","20"]},"artifacts":{"enabled":true,"retention_days":30},"secrets_masked":true}')
SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
echo "  Session ID: $SESSION_ID (complex baggage)"
run_cli create testcase --session-id "$SESSION_ID" --name "test_matrix_combination" --status "pass" > /dev/null
echo "  Created 1 test case"
echo ""

# Session for grouping demonstration (multiple sessions with same version, different arch)
echo -e "${BLUE}Phase 3: Additional sessions for grouping demonstration${NC}"
echo ""

for i in {1..3}; do
    SESSION_COUNT=$((SESSION_COUNT + 1))
    echo -e "${YELLOW}[$SESSION_COUNT/$TOTAL_SESSIONS] Creating session for grouping demo${NC}"
    arch="${ARCHITECTURES[$((i % ${#ARCHITECTURES[@]}))]}"
    os="${OPERATING_SYSTEMS[$((i % ${#OPERATING_SYSTEMS[@]}))]}"

    SESSION_OUTPUT=$(run_cli create session \
        --label "version=2.0.0" \
        --label "arch=$arch" \
        --label "os=$os" \
        --label "env=ci" \
        --label "region=us-east" \
        --description "CI test run #$i for version 2.0.0 on $os/$arch (us-east region)" \
        --baggage '{"runner":"github-actions","build":'$i'}')
    SESSION_ID=$(echo "$SESSION_OUTPUT" | grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}')
    echo "  Session ID: $SESSION_ID"

    # Create a few test cases with different statuses
    run_cli create testcase --session-id "$SESSION_ID" --name "test_pass_example" --status "pass" > /dev/null
    run_cli create testcase --session-id "$SESSION_ID" --name "test_fail_example" --status "fail" --output "Assertion failed" > /dev/null
    run_cli create testcase --session-id "$SESSION_ID" --name "test_skip_example" --status "skip" --output "Not implemented yet" > /dev/null
    echo "  Created 3 test cases"
    echo ""
done

# Add one final test case with large detailed error output
echo -e "${BLUE}Creating final test with comprehensive error output${NC}"
LAST_SESSION_ID="$SESSION_ID"
LARGE_ERROR_OUTPUT="2024-01-15 14:23:41.123 INFO  [pool-manager] Starting connection pool health check
2024-01-15 14:23:41.245 INFO  [pool-manager] Pool status: active=45, idle=5, max=50
2024-01-15 14:23:42.567 DEBUG [connection-12] Acquired connection for query: SELECT * FROM users WHERE status='active'
2024-01-15 14:23:42.890 DEBUG [connection-34] Acquired connection for query: UPDATE sessions SET last_active=NOW()
2024-01-15 14:23:43.123 WARN  [pool-manager] Pool utilization at 92% (46/50 connections active)
2024-01-15 14:23:43.456 DEBUG [connection-56] Acquired connection for query: INSERT INTO events (type, data) VALUES (?, ?)
2024-01-15 14:23:44.234 WARN  [pool-manager] Pool utilization at 96% (48/50 connections active)
2024-01-15 14:23:44.567 DEBUG [connection-78] Waiting for available connection (queue position: 1)
2024-01-15 14:23:45.123 ERROR [pool-manager] Pool utilization at 100% (50/50 connections active)
2024-01-15 14:23:45.456 WARN  [connection-91] Waiting for available connection (queue position: 2)
2024-01-15 14:23:46.789 WARN  [connection-103] Waiting for available connection (queue position: 5)
2024-01-15 14:23:48.234 ERROR [connection-78] Connection wait timeout exceeded (5.0s elapsed)
2024-01-15 14:23:50.567 ERROR [connection-91] Connection wait timeout exceeded (5.1s elapsed)
2024-01-15 14:23:52.890 ERROR [pool-manager] Pool exhaustion detected - 12 requests waiting, 0 idle connections
2024-01-15 14:23:55.123 CRITICAL [pool-manager] POOL EXHAUSTION CRITICAL - 27 requests queued
2024-01-15 14:24:00.456 ERROR [connection-156] Connection acquisition failed after 30.2s

Traceback (most recent call last):
  File \"/app/src/api/handlers/user_handler.py\", line 89, in get_user_profile
    user = await db.query_one(\"SELECT * FROM users WHERE id = ?\", user_id)
  File \"/app/src/database/client.py\", line 234, in query_one
    async with self.get_connection() as conn:
  File \"/app/src/database/client.py\", line 156, in get_connection
    conn = await self._pool.acquire(timeout=30.0)
  File \"/usr/local/lib/python3.11/site-packages/asyncpg/pool.py\", line 489, in acquire
    return await self._acquire(timeout)
  File \"/usr/local/lib/python3.11/site-packages/asyncpg/pool.py\", line 512, in _acquire
    raise asyncpg.exceptions.PoolTimeoutError(
asyncpg.exceptions.PoolTimeoutError: timed out waiting for a connection from the pool after 30.0 seconds

During handling of the above exception, another exception occurred:

Traceback (most recent call last):
  File \"/app/src/api/middleware/error_handler.py\", line 45, in __call__
    response = await self.app(scope, receive, send)
  File \"/usr/local/lib/python3.11/site-packages/starlette/routing.py\", line 677, in __call__
    await route.handle(scope, receive, send)
  File \"/usr/local/lib/python3.11/site-packages/starlette/routing.py\", line 261, in handle
    await self.app(scope, receive, send)
  File \"/usr/local/lib/python3.11/site-packages/starlette/middleware/base.py\", line 159, in __call__
    await self.dispatch(request, call_next)
  File \"/app/src/api/middleware/metrics.py\", line 78, in dispatch
    response = await call_next(request)
  File \"/app/src/api/handlers/user_handler.py\", line 89, in get_user_profile
    user = await db.query_one(\"SELECT * FROM users WHERE id = ?\", user_id)
DatabaseConnectionError: Failed to acquire database connection from pool

Connection Pool State at Error:
  Pool size: 50 (min=10, max=50)
  Active connections: 50
  Idle connections: 0
  Waiting requests: 127
  Total acquisitions: 45,892
  Failed acquisitions: 156
  Average wait time: 8.3s
  Max wait time: 30.2s
  Pool created at: 2024-01-15 12:15:23.456

Active Connection Details:
  Connection #1:  age=125.3s, query=\"SELECT u.*, p.* FROM users u LEFT JOIN profiles p ON u.id=p.user_id WHERE u.status=?\", state=EXECUTING
  Connection #2:  age=118.7s, query=\"UPDATE user_analytics SET page_views=page_views+1 WHERE user_id IN (?, ?, ...)\", state=EXECUTING
  Connection #3:  age=95.4s, query=\"SELECT * FROM events WHERE timestamp > ? ORDER BY timestamp DESC LIMIT 1000\", state=EXECUTING
  Connection #4:  age=87.2s, query=\"DELETE FROM sessions WHERE last_activity < ?\", state=WAITING_FOR_LOCK
  Connection #5:  age=78.9s, query=\"INSERT INTO audit_log (user_id, action, timestamp) VALUES (?, ?, ?)\", state=EXECUTING
  ... (45 more connections in EXECUTING or WAITING_FOR_LOCK state)

2024-01-15 14:24:01.234 ERROR [pool-manager] Pool metrics - CPU: 89%, Memory: 12.4GB/16GB, Connections: 50/50
2024-01-15 14:24:01.567 WARN  [health-check] Database health check failed - response time: 28.3s (threshold: 5.0s)
2024-01-15 14:24:02.890 ERROR [pool-manager] Emergency pool expansion attempted but failed - already at max capacity
2024-01-15 14:24:03.123 CRITICAL [system] Service degradation detected - 89% of requests timing out"

run_cli create testcase \
    --session-id "$LAST_SESSION_ID" \
    --name "test_database_pool_exhaustion" \
    --classname "DatabaseTests" \
    --file "tests/databasetests.py" \
    --testsuite "DatabaseTests" \
    --status "fail" \
    --output "$LARGE_ERROR_OUTPUT" \
    --baggage '{"duration_ms":30147,"retry_count":3,"memory_mb":12450,"cpu_percent":89}' > /dev/null

echo "  Created comprehensive failure test case"
echo ""

echo -e "${GREEN}==================================================================${NC}"
echo -e "${GREEN}  Sample data generation complete!${NC}"
echo -e "${GREEN}==================================================================${NC}"
echo ""
echo "Summary:"
echo "  Total sessions created: $SESSION_COUNT"
echo ""
echo "Coverage:"
echo "  - Versions: ${VERSIONS[*]}"
echo "  - Architectures: ${ARCHITECTURES[*]}"
echo "  - Operating Systems: ${OPERATING_SYSTEMS[*]}"
echo "  - Environments: ${ENVIRONMENTS[*]}"
echo "  - Regions: ${REGIONS[*]}"
echo "  - Test suites: ${SUITE_NAMES[*]}"
echo ""
echo "Session metadata:"
echo "  - All sessions include meaningful descriptions"
echo "  - No baggage (description only)"
echo "  - Empty baggage ({})"
echo "  - Simple baggage (runner + build info)"
echo "  - Complex nested baggage (workflows, matrices, artifacts)"
echo "  - Runner-specific baggage (GitHub Actions, GitLab CI, Jenkins)"
echo ""
echo "Label variations:"
echo "  - Labels with values (version=X, arch=X, os=X, env=X, region=X)"
echo "  - Labels without values (approved, denied, nightly, release)"
echo ""
echo "Test case variations:"
echo "  - All statuses: pass, fail, error, skip"
echo "  - Randomized optional fields (classname, file, testsuite)"
echo "  - Random baggage inclusion (25% of tests have performance metrics)"
echo "  - Realistic error outputs and stack traces"
echo "  - Version-specific failures"
echo "  - OS-specific failures"
echo "  - Architecture-specific failures"
echo "  - Environment-specific failures"
echo "  - Region-specific failures"
echo ""
echo -e "${GREEN}==================================================================${NC}"
echo -e "${GREEN}  Sample Grouping Queries${NC}"
echo -e "${GREEN}==================================================================${NC}"
echo ""
echo "Single-label grouping:"
echo "  - Group by version (shows results across all 5 versions)"
echo "  - Group by arch (x86_64, arm64, aarch64)"
echo "  - Group by os (linux, windows, macos)"
echo "  - Group by env (ci, staging, production)"
echo "  - Group by region (us-east, eu-west, ap-south)"
echo ""
echo "Multi-label grouping:"
echo "  - Group by version + arch (shows version performance per architecture)"
echo "  - Group by version + os (shows version compatibility per OS)"
echo "  - Group by os + arch (shows platform-specific issues)"
echo "  - Group by env + region (shows environment performance by region)"
echo "  - Group by version + env (shows version stability per environment)"
echo ""
echo "Label filtering:"
echo "  - Filter by approved (sessions marked as approved)"
echo "  - Filter by denied (sessions marked as denied)"
echo "  - Filter by nightly (nightly build sessions)"
echo "  - Filter by release (official release sessions)"
echo "  - Filter by version=2.0.0 (specific version)"
echo "  - Filter by env=production (production environment only)"
echo ""
echo "Combined queries:"
echo "  - Group by arch where version=2.0.0"
echo "  - Group by os where env=ci"
echo "  - Group by region where approved is set"
echo "  - Group by version where os=linux AND arch=x86_64"
echo ""
echo -e "${GREEN}Done.${NC}"
