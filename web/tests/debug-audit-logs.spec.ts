import { test, expect } from '@playwright/test';
import { APIRequestContext } from '@playwright/test';

test.describe('Audit Logs Debug Investigation', () => {
  let apiContext: APIRequestContext;

  test.beforeAll(async ({ playwright }) => {
    // Create API context for direct backend testing
    apiContext = await playwright.request.newContext({
      baseURL: 'http://localhost:8080',
      timeout: 30000,
    });
  });

  test.afterAll(async () => {
    await apiContext.dispose();
  });

  test('Step 1: Verify backend is accessible', async () => {
    console.log('\n=== Step 1: Verifying backend accessibility ===');

    const response = await apiContext.get('/health');
    expect(response.ok()).toBeTruthy();

    const health = await response.json();
    console.log('Backend health check:', health);
    expect(health.status).toBe('healthy');
  });

  test('Step 2: Test direct API authentication', async () => {
    console.log('\n=== Step 2: Testing direct API authentication ===');

    // Test login via API
    const loginResponse = await apiContext.post('/api/v1/auth/login', {
      data: {
        username: 'admin',
        password: 'Admin@123'
      }
    });

    console.log('Login response status:', loginResponse.status());
    console.log('Login response headers:', Object.fromEntries(loginResponse.headers()));

    if (loginResponse.ok()) {
      const loginData = await loginResponse.json();
      console.log('Login successful:', {
        access_token: loginData.access_token ? `${loginData.access_token.substring(0, 20)}...` : 'No token',
        refresh_token: loginData.refresh_token ? `${loginData.refresh_token.substring(0, 20)}...` : 'No token',
        user: loginData.user
      });

      // Test getting user profile
      const profileResponse = await apiContext.get('/api/v1/me', {
        headers: {
          'Authorization': `Bearer ${loginData.access_token}`
        }
      });

      console.log('Profile response status:', profileResponse.status());
      if (profileResponse.ok()) {
        const profileData = await profileResponse.json();
        console.log('User profile:', profileData);
        console.log('User permissions:', profileData.user?.permissions);
      } else {
        console.log('Profile failed:', await profileResponse.text());
      }

      // Test direct audit API call
      const auditResponse = await apiContext.get('/api/v1/audit', {
        headers: {
          'Authorization': `Bearer ${loginData.access_token}`
        }
      });

      console.log('Direct audit API call status:', auditResponse.status());
      console.log('Direct audit API response headers:', Object.fromEntries(auditResponse.headers()));

      if (auditResponse.ok()) {
        const auditData = await auditResponse.json();
        console.log('Direct audit API response:', auditData);
        console.log('Number of audit logs:', auditData.audit_logs?.length || 0);
        console.log('Total audit logs:', auditData.total || 0);
      } else {
        const errorText = await auditResponse.text();
        console.log('Direct audit API failed:', errorText);
        console.log('Audit API error status:', auditResponse.status());
      }
    } else {
      const errorText = await loginResponse.text();
      console.log('Login failed:', errorText);
      console.log('Login error status:', loginResponse.status());
    }
  });

  test('Step 3: Browser-based investigation', async ({ page }) => {
    console.log('\n=== Step 3: Browser-based investigation ===');

    // Enable console logging
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      const msgText = msg.text();
      consoleMessages.push(msgText);
      console.log(`Browser Console [${msg.type()}]:`, msgText);
    });

    // Enable network logging
    const networkRequests: any[] = [];
    page.on('request', request => {
      const requestInfo = {
        url: request.url(),
        method: request.method(),
        headers: request.headers(),
        postData: request.postData()
      };
      networkRequests.push({ type: 'request', ...requestInfo });
      console.log(`Network Request: ${request.method()} ${request.url()}`);
      console.log('Request headers:', request.headers());
    });

    page.on('response', response => {
      const responseInfo = {
        url: response.url(),
        status: response.status(),
        statusText: response.statusText(),
        headers: response.headers()
      };
      networkRequests.push({ type: 'response', ...responseInfo });
      console.log(`Network Response: ${response.status()} ${response.statusText()} ${response.url()}`);
      console.log('Response headers:', response.headers());
    });

    // Navigate to login page
    console.log('\nNavigating to login page...');
    await page.goto('/login');
    await page.waitForLoadState('networkidle');

    // Take screenshot of login page
    await page.screenshot({ path: 'test-results/login-page.png', fullPage: true });
    console.log('Login page screenshot saved');

    // Fill login form
    console.log('\nFilling login credentials...');
    await page.fill('input[name="username"], input[id="username"], input[placeholder*="username"]', 'admin');
    await page.fill('input[name="password"], input[id="password"], input[placeholder*="password"]', 'Admin@123');

    // Take screenshot before login
    await page.screenshot({ path: 'test-results/login-form-filled.png', fullPage: true });

    // Submit login form
    console.log('\nSubmitting login form...');
    await Promise.all([
      page.waitForNavigation({ timeout: 30000 }),
      page.click('button[type="submit"], button:has-text("Login"), .btn-primary')
    ]);

    // Wait for page to load
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'test-results/post-login-dashboard.png', fullPage: true });
    console.log('Post-login dashboard screenshot saved');

    // Check current URL and auth state
    console.log('Current URL after login:', page.url());
    console.log('Local storage after login:', await page.evaluate(() => {
      return {
        access_token: localStorage.getItem('access_token'),
        refresh_token: localStorage.getItem('refresh_token')
      };
    }));

    // Navigate to audit page
    console.log('\nNavigating to audit page...');
    await page.goto('/audit');
    await page.waitForLoadState('networkidle');

    // Take screenshot of audit page
    await page.screenshot({ path: 'test-results/audit-page-initial.png', fullPage: true });
    console.log('Initial audit page screenshot saved');

    // Wait for debug information to be visible
    await page.waitForSelector('.bg-white.shadow', { timeout: 10000 });

    // Extract debug information
    const debugInfo = await page.evaluate(() => {
      const debugSection = document.querySelector('.bg-white.shadow');
      if (!debugSection) return null;

      const text = debugSection.textContent || '';
      const lines = text.split('\n').map(line => line.trim()).filter(line => line);

      return {
        debugText: text,
        lines: lines,
        pageUrl: window.location.href,
        localStorage: {
          access_token: localStorage.getItem('access_token'),
          refresh_token: localStorage.getItem('refresh_token')
        },
        sessionStorage: {
          access_token: sessionStorage.getItem('access_token'),
          refresh_token: sessionStorage.getItem('refresh_token')
        }
      };
    });

    console.log('\n=== Debug Information from Page ===');
    console.log('Debug info:', debugInfo);

    // Click "Test Auth" button and capture output
    console.log('\n=== Testing Auth Button ===');
    await page.click('button:has-text("Test Auth")');

    // Wait a moment for console output
    await page.waitForTimeout(2000);

    // Take screenshot after auth test
    await page.screenshot({ path: 'test-results/after-test-auth.png', fullPage: true });

    // Click "Test API Call" button and capture output
    console.log('\n=== Testing API Call Button ===');
    await page.click('button:has-text("Test API Call")');

    // Wait for API call to complete
    await page.waitForTimeout(3000);

    // Take screenshot after API test
    await page.screenshot({ path: 'test-results/after-test-api-call.png', fullPage: true });

    // Click "Debug Load Audit Logs" button and capture output
    console.log('\n=== Testing Debug Load Audit Logs Button ===');
    await page.click('button:has-text("Debug Load Audit Logs")');

    // Wait for the operation to complete
    await page.waitForTimeout(5000);

    // Take final screenshot
    await page.screenshot({ path: 'test-results/after-debug-load-audit-logs.png', fullPage: true });

    // Extract final debug information
    const finalDebugInfo = await page.evaluate(() => {
      const debugSection = document.querySelector('.bg-white.shadow');
      if (!debugSection) return null;

      const text = debugSection.textContent || '';
      const lines = text.split('\n').map(line => line.trim()).filter(line => line);

      return {
        debugText: text,
        lines: lines
      };
    });

    console.log('\n=== Final Debug Information ===');
    console.log('Final debug info:', finalDebugInfo);

    // Analyze network requests
    console.log('\n=== Network Request Analysis ===');
    const auditRequests = networkRequests.filter(req =>
      req.url && req.url.includes('/audit')
    );

    console.log(`Found ${auditRequests.length} audit-related requests:`);
    auditRequests.forEach((req, index) => {
      console.log(`\nRequest ${index + 1}:`);
      console.log(`  Type: ${req.type}`);
      console.log(`  URL: ${req.url}`);
      console.log(`  Method: ${req.method}`);
      console.log(`  Status: ${req.status}`);
      if (req.headers && req.headers.authorization) {
        console.log(`  Authorization: ${req.headers.authorization.substring(0, 30)}...`);
      }
    });

    // Analyze console messages for errors
    console.log('\n=== Console Message Analysis ===');
    const errorMessages = consoleMessages.filter(msg =>
      msg.toLowerCase().includes('error') ||
      msg.toLowerCase().includes('failed') ||
      msg.toLowerCase().includes('401') ||
      msg.toLowerCase().includes('403')
    );

    if (errorMessages.length > 0) {
      console.log('Found error messages in console:');
      errorMessages.forEach(msg => console.log(`  - ${msg}`));
    } else {
      console.log('No error messages found in console');
    }

    // Save all collected data to a file
    const investigationData = {
      timestamp: new Date().toISOString(),
      debugInfo,
      finalDebugInfo,
      networkRequests,
      consoleMessages,
      errorMessages,
      auditRequests,
      pageUrl: page.url(),
      localStorage: await page.evaluate(() => ({
        access_token: localStorage.getItem('access_token'),
        refresh_token: localStorage.getItem('refresh_token')
      }))
    };

    await page.evaluate((data) => {
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'audit-debug-investigation.json';
      a.click();
      URL.revokeObjectURL(url);
    }, investigationData);

    console.log('\n=== Investigation Complete ===');
    console.log('All data saved to test-results/ and downloaded as audit-debug-investigation.json');
  });

  test('Step 4: Verify user permissions via API', async () => {
    console.log('\n=== Step 4: Verifying user permissions ===');

    // Login first
    const loginResponse = await apiContext.post('/api/v1/auth/login', {
      data: {
        username: 'admin',
        password: 'Admin@123'
      }
    });

    if (loginResponse.ok()) {
      const loginData = await loginResponse.json();
      const token = loginData.access_token;

      // Check if user has audit:read permission
      const profileResponse = await apiContext.get('/api/v1/me', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (profileResponse.ok()) {
        const profileData = await profileResponse.json();
        const permissions = profileData.user?.permissions || [];
        console.log('User permissions:', permissions);
        console.log('Has audit:read permission:', permissions.includes('audit:read'));

        // Test with missing permission
        if (!permissions.includes('audit:read')) {
          console.log('❌ User does not have audit:read permission - this is likely the issue!');
        }
      }
    }
  });
});