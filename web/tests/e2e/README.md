# Pustaka Graph Functionality Tests

This directory contains Playwright end-to-end tests for the Pustaka graph page functionality.

## Test Coverage

The tests verify the following functionality:

### 1. Search Suggestions Dropdown
- ✅ Search input visibility and functionality
- ✅ Search suggestions appear when typing
- ✅ Debounced search with "Searching..." indicator
- ✅ Search result selection and input update
- ✅ Autocomplete dropdown hiding after selection

### 2. Graph Node Expansion
- ✅ Right-click context menu on graph nodes
- ✅ "Expand Node" option in context menu
- ✅ Node expansion functionality
- ✅ Loading states during expansion
- ✅ Graph data updates after expansion

### 3. Additional Functionality
- ✅ Page layout and element visibility
- ✅ Control buttons (Center, Fit, Clear)
- ✅ Empty state display
- ✅ No results state
- ✅ Graph loading states
- ✅ Authentication handling

## Test Files

- `graph.spec.ts` - Basic graph functionality tests (no authentication)
- `graph-authenticated.spec.ts` - Comprehensive tests with authentication
- `auth-helper.ts` - Authentication helper for tests

## Running Tests

### Prerequisites

1. Docker services must be running:
   ```bash
   cd /home/syam/dev/pustaka
   docker compose up -d
   ```

2. Frontend should be accessible at http://localhost:3000
3. API should be accessible at http://localhost:8080

### Install Dependencies

```bash
cd /home/syam/dev/pustaka/web
npm install
npx playwright install
```

### Run Tests

#### Headless mode (default)
```bash
npx playwright test
```

#### Headed mode (show browser)
```bash
npx playwright test --headed
```

#### Debug mode
```bash
npx playwright test --debug
```

#### Run specific test file
```bash
npx playwright test graph-authenticated.spec.ts
```

#### Run with specific browser
```bash
npx playwright test --project=chromium
npx playwright test --project=firefox
npx playwright test --project=webkit
```

## Test Results

After running tests, view the results:
```bash
npx playwright show-report
```

## Test Configuration

The tests are configured in `playwright.config.ts`:
- Base URL: http://localhost:3000
- Browsers: Chromium, Firefox, WebKit
- Reports: HTML format
- Screenshots: On failure
- Videos: On failure
- Traces: On first retry

## Authentication

Tests use the following credentials:
- Username: admin
- Password: Admin@123

The `auth-helper.ts` file handles automatic login before running tests.

## Troubleshooting

### Issues with Docker Services
If tests fail due to services not being available:
```bash
cd /home/syam/dev/pustaka
docker compose down
docker compose up -d
# Wait for services to be ready (30-60 seconds)
```

### Issues with Search Results
Search functionality may return no results in test environment if:
- Database is empty
- API endpoints are not accessible
- Authentication fails

The tests handle these scenarios gracefully and provide appropriate logging.

### Issues with Graph Node Interaction
Node interaction may fail if:
- Graph has no data/nodes
- Nodes are in different positions than expected
- Graph rendering is slow

Tests try multiple positions and handle cases where no nodes are found.

## Debug Tips

1. **Run tests in headed mode** to see what's happening
2. **Use debug mode** to pause execution and inspect state
3. **Check test logs** for detailed information
4. **View HTML report** for screenshots and videos of failures
5. **Check browser console** for JavaScript errors
6. **Verify API responses** using browser dev tools

## Expected Test Outcomes

### Passing Tests
- All page elements load correctly
- Search suggestions appear and work
- Graph loads with data
- Context menu appears on right-click
- Node expansion works
- Control buttons function properly

### Common Failure Reasons
- Docker services not running
- Authentication issues
- No data in database
- Network connectivity issues
- Element locator changes
- Timing issues (slow loading)

## Future Improvements

- Add data setup scripts to ensure test data availability
- Implement API mocking for more reliable tests
- Add visual regression testing
- Include performance testing
- Add mobile device testing
- Implement cross-browser compatibility testing