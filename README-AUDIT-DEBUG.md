# Audit Logs Debug Investigation

This document outlines the comprehensive debugging approach for the audit logs issue in the Pustaka application.

## Problem Statement

The audit logs are not showing on the `/audit` page despite having a debug version with additional logging and test buttons.

## Debugging Approach

### 1. Test Files Created

- `/home/syam/dev/pustaka/web/tests/debug-audit-logs.spec.ts` - Comprehensive Playwright test suite
- `/home/syam/dev/pustaka/run-debug-test.sh` - Shell script to run the debug test
- `/home/syam/dev/pustaka/README-AUDIT-DEBUG.md` - This documentation file

### 2. Investigation Steps

The Playwright test performs the following investigations:

#### Step 1: Backend Accessibility Check
- Verifies the backend health endpoint
- Confirms the API is responding correctly

#### Step 2: Direct API Authentication Test
- Tests login via direct API call
- Verifies user profile retrieval
- Tests audit API endpoint directly with authentication
- Captures detailed response information

#### Step 3: Browser-Based Investigation
- Navigates through the application UI
- Captures console logs and network requests
- Tests each debug button on the audit page
- Takes screenshots at each step
- Analyzes authentication state in browser storage

#### Step 4: User Permissions Verification
- Verifies the user has the required `audit:read` permission
- Checks if permissions are correctly assigned

### 3. Key Areas Investigated

1. **Authentication Flow**
   - JWT token generation and storage
   - Token refresh mechanism
   - Authorization header inclusion in requests

2. **Authorization/Permissions**
   - User permission verification
   - RBAC (Role-Based Access Control) implementation
   - Route-level permission checks

3. **API Communication**
   - Request/response headers
   - HTTP status codes
   - Error handling and responses

4. **Frontend State Management**
   - Pinia store state
   - Component lifecycle and data loading
   - Debug information display

### 4. Expected Outputs

The investigation will produce:

- **Screenshots**: Visual documentation of each step
- **Console Logs**: JavaScript console output and errors
- **Network Requests**: Detailed HTTP request/response analysis
- **Debug Information**: Component state and authentication details
- **JSON Report**: Comprehensive investigation data

### 5. Running the Debug Test

#### Prerequisites

Ensure both services are running:
```bash
# Backend
make run

# Frontend (in separate terminal)
cd web && npm run dev
```

#### Execute the Test

```bash
# Make the script executable
chmod +x /home/syam/dev/pustaka/run-debug-test.sh

# Run the debug investigation
./run-debug-test.sh
```

Or run directly with Playwright:

```bash
cd /home/syam/dev/pustaka/web
npx playwright test debug-audit-logs.spec.ts --headed
```

### 6. Key Files to Examine After Test

1. **Screenshots** (`test-results/*.png`)
   - Visual evidence of each step
   - UI state at key moments

2. **HTML Report** (`test-results/html-report/index.html`)
   - Interactive test report
   - Detailed step-by-step results

3. **Investigation JSON** (`audit-debug-investigation.json`)
   - Complete investigation data
   - Network requests and console logs
   - Debug information from the page

### 7. Common Issues to Look For

1. **Authentication Issues**
   - Missing or invalid JWT tokens
   - Token refresh failures
   - Authorization header not being sent

2. **Permission Issues**
   - User lacks `audit:read` permission
   - Role not properly assigned
   - Permission check logic errors

3. **API Issues**
   - 401/403 HTTP status codes
   - Malformed requests
   - Backend errors

4. **Frontend Issues**
   - Component state management problems
   - API request failures
   - JavaScript errors in console

### 8. Next Steps Based on Findings

After running the investigation, analyze the results to identify:

1. **Root Cause**: The specific issue preventing audit logs from loading
2. **Impact Areas**: Which part of the authentication/authorization flow is failing
3. **Fix Strategy**: The appropriate solution based on the identified issue

### 9. Expected Test Duration

The complete investigation should take approximately 2-5 minutes to run, including:
- Service health checks
- Authentication flow
- UI navigation and interaction
- Data collection and analysis

### 10. Troubleshooting the Test Itself

If the test fails to run:

1. **Service Issues**: Ensure both frontend and backend are running
2. **Playwright Issues**: Install dependencies with `npm install`
3. **Browser Issues**: Ensure browser dependencies are installed with `npx playwright install`

This comprehensive investigation will provide detailed insights into why the audit logs are not displaying and help identify the exact cause of the issue.