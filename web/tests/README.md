# User Management E2E Test Suite

This directory contains comprehensive end-to-end tests for the Pustaka CMDB user management functionality, specifically designed to verify that previously reported issues have been resolved.

## Test Coverage

### Issues Being Verified
1. **Issue #1:** `/users` page showing "No users found" despite having an admin user
2. **Issue #2:** Add user page returning 404 error

### Test Scenarios
- ✅ Admin login functionality (admin/Admin@123)
- ✅ Users page displays admin user correctly
- ✅ Add User button navigation works without 404
- ✅ User creation form functionality
- ✅ User list updates after creating new user
- ✅ User details page access
- ✅ User edit page access
- ✅ Authentication requirements enforcement
- ✅ Invalid login credentials handling

## Prerequisites

1. **Application Running:** Ensure the Pustaka CMDB application is running on `http://localhost:3000`
   ```bash
   cd /home/syam/dev/pustaka
   make dev
   ```

2. **Admin User:** Ensure the admin user exists with credentials:
   - Username: `admin`
   - Password: `Admin@123`

3. **Dependencies:** Playwright should be installed (included in package.json)

## Running Tests

### Quick Run
```bash
# From the web directory
npm run test:e2e:playwright

# Or run specific test file
npx playwright test tests/e2e/user-management.spec.ts
```

### Automated Test Execution
```bash
# Use the provided test script
./run-user-tests.sh
```

### Headed Mode (with browser UI)
```bash
npx playwright test tests/e2e/user-management.spec.ts --headed
```

### Debug Mode
```bash
npx playwright test tests/e2e/user-management.spec.ts --debug
```

## Test Results

### Output Locations
- **HTML Report:** `test-results/html-report/index.html`
- **JSON Results:** `test-results/results.json`
- **JUnit Results:** `test-results/results.xml`
- **Screenshots:** `test-results/`
- **Videos:** `test-results/videos/` (on failure)
- **Traces:** `test-results/traces/` (on failure)

### Key Screenshots
- `test-results/users-page-with-admin.png` - Verifies Issue #1 fix
- `test-results/add-user-page-loaded.png` - Verifies Issue #2 fix
- `test-results/user-created-successfully.png` - User creation success
- `test-results/user-details-page.png` - User details view
- `test-results/edit-user-page.png` - User edit functionality

## Test Structure

### File Organization
```
tests/
├── e2e/
│   └── user-management.spec.ts    # Main test suite
├── setup.ts                       # Test setup and configuration
└── README.md                      # This file
```

### Test Categories

#### Authentication Tests
- Validates admin login functionality
- Tests authentication requirements for protected pages
- Verifies error handling for invalid credentials

#### User Management Tests
- **Users List Test:** Verifies that admin user is displayed (Issue #1 fix)
- **Add User Navigation:** Confirms navigation to /users/new works (Issue #2 fix)
- **User Creation:** Tests complete user creation workflow
- **User Details:** Validates user details page access
- **User Edit:** Tests user edit functionality

## Configuration

### Playwright Configuration
The test suite uses the following configuration (playwright.config.ts):
- **Base URL:** `http://localhost:3000`
- **Timeout:** 60 seconds
- **Browsers:** Chromium, Firefox, WebKit
- **Reporters:** HTML, JSON, JUnit, List
- **Screenshots:** On failure
- **Videos:** On failure
- **Traces:** On failure

### Test Timeouts
- **Action Timeout:** 30 seconds
- **Navigation Timeout:** 60 seconds
- **Test Timeout:** 60 seconds
- **Expect Timeout:** 10 seconds

## Issue Resolution Verification

### Issue #1: /users page showing "No users found"
**Test Method:** The test logs in as admin and navigates to `/users`, then verifies that the admin user appears in the list instead of "No users found".

**Expected Result:** Admin user should be visible in the users table with correct details.

**Verification Screenshot:** `test-results/users-page-with-admin.png`

### Issue #2: Add user page returning 404
**Test Method:** The test navigates to `/users/new` and verifies that the page loads correctly without a 404 error.

**Expected Result:** User creation form should load successfully with all required fields.

**Verification Screenshot:** `test-results/add-user-page-loaded.png`

## Troubleshooting

### Common Issues

1. **Application Not Running**
   ```
   Error: connect ECONNREFUSED 127.0.0.1:3000
   ```
   **Solution:** Start the application with `make dev`

2. **Admin User Not Found**
   ```
   Error: Invalid credentials
   ```
   **Solution:** Create admin user with `make create-admin`

3. **Timeout Issues**
   ```
   Error: Test timeout of 60000ms exceeded
   ```
   **Solution:** Check application performance and network connectivity

### Debug Mode
Run tests in debug mode for step-by-step execution:
```bash
npx playwright test tests/e2e/user-management.spec.ts --debug
```

### Trace Viewer
View detailed traces for failed tests:
```bash
npx playwright show-trace test-results/trace.zip
```

## Best Practices

1. **Test Isolation:** Each test starts from a clean login state
2. **Wait Strategies:** Tests wait for network idle and proper element visibility
3. **Error Handling:** Tests include proper error verification and timeout handling
4. **Documentation:** Each test case is clearly documented with objectives and expected results
5. **Screenshot Coverage:** Key test steps include screenshots for verification

## Generating Reports

### Automated Report Generation
```bash
node generate-test-report.js
```

This generates a comprehensive markdown report at `test-results/test-report.md` including:
- Executive summary
- Detailed test results
- Issue resolution confirmation
- Screenshots and evidence
- Recommendations

## Contributing

When adding new tests:
1. Follow the existing test structure and naming conventions
2. Include proper documentation for test objectives
3. Add appropriate assertions and verifications
4. Include screenshots for key test steps
5. Update this README with new test coverage

## Support

For test-related issues:
1. Check the HTML report for detailed failure information
2. Review console logs and network requests
3. Use debug mode for step-by-step execution
4. Verify application is running and accessible
5. Check admin user credentials and permissions