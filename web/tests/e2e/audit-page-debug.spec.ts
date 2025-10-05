import { test, expect } from '@playwright/test';

test.describe('Audit Page - Debug Network Calls', () => {
  test('should debug why audit logs are not loading', async ({ page }) => {
    console.log('🐛 Starting debug test for audit page network calls');

    // Track network requests
    const apiRequests: any[] = [];
    const apiResponses: any[] = [];

    page.on('request', request => {
      if (request.url().includes('/api/v1/audit')) {
        const requestData = {
          url: request.url(),
          method: request.method(),
          headers: request.headers(),
          postData: request.postData()
        };
        apiRequests.push(requestData);
        console.log(`📤 API Request: ${request.method()} ${request.url()}`);
        if (request.postData()) {
          console.log(`   Post data: ${request.postData()}`);
        }
      }
    });

    page.on('response', response => {
      if (response.url().includes('/api/v1/audit')) {
        response.text().then(text => {
          const responseData = {
            url: response.url(),
            status: response.status(),
            statusText: response.statusText(),
            body: text
          };
          apiResponses.push(responseData);
          console.log(`📥 API Response: ${response.status()} ${response.statusText()}`);
          console.log(`   Body: ${text.substring(0, 500)}${text.length > 500 ? '...' : ''}`);
        });
      }
    });

    // Track console messages
    const consoleLogs: string[] = [];
    page.on('console', msg => {
      const text = msg.text();
      consoleLogs.push(`[${msg.type()}] ${text}`);
      if (msg.type() === 'error' || text.includes('Audit logs:')) {
        console.log(`📝 Console [${msg.type()}]: ${text}`);
      }
    });

    // Navigate to audit page
    console.log('📍 Navigating to audit page...');
    await page.goto('http://localhost:3000/audit');

    // Handle login if needed
    await page.waitForLoadState('networkidle');

    if (page.url().includes('/login')) {
      console.log('🔐 Performing login...');

      await page.fill('input[name="username"], input[type="text"]', 'admin');
      await page.fill('input[name="password"], input[type="password"]', 'Admin@123');

      await Promise.all([
        page.waitForNavigation(),
        page.click('button[type="submit"], .btn-primary, button:has-text("Login")')
      ]);

      console.log('✅ Login successful, navigating to audit page...');
      await page.goto('http://localhost:3000/audit');
    }

    // Wait for audit page to load
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 15000 });
    console.log('✅ Audit page loaded');

    // Wait for API calls to complete
    console.log('⏳ Waiting for API calls to complete...');
    await page.waitForTimeout(10000); // Give extra time for API calls

    // Analyze network requests
    console.log(`📊 Found ${apiRequests.length} audit API requests`);
    console.log(`📊 Found ${apiResponses.length} audit API responses`);

    if (apiRequests.length === 0) {
      console.log('⚠️  WARNING: No audit API requests were made!');
      console.log('   This suggests the frontend is not calling the audit API');
    } else {
      console.log('✅ Audit API requests were made:');
      apiRequests.forEach((req, i) => {
        console.log(`  ${i+1}. ${req.method} ${req.url}`);
      });
    }

    if (apiResponses.length === 0) {
      console.log('⚠️  WARNING: No audit API responses received!');
    } else {
      console.log('✅ Audit API responses received:');
      apiResponses.forEach((res, i) => {
        console.log(`  ${i+1}. Status: ${res.status} ${res.statusText}`);

        try {
          const parsed = JSON.parse(res.body);
          if (parsed.audit_logs) {
            console.log(`     Audit logs in response: ${parsed.audit_logs.length}`);
            console.log(`     Total count: ${parsed.total}`);
          }
        } catch (e) {
          console.log(`     Response body is not valid JSON`);
        }
      });
    }

    // Check current page state
    const rows = page.locator('table tbody tr');
    const rowCount = await rows.count();
    console.log(`📊 Current table rows: ${rowCount}`);

    const emptyState = page.locator('h3:has-text("No audit logs found")');
    const emptyVisible = await emptyState.isVisible();
    console.log(`📭 Empty state visible: ${emptyVisible}`);

    // Check date filters have 30-day default values (new requirement)
    const dateInputs = page.locator('input[type="date"]');
    const dateCount = await dateInputs.count();

    if (dateCount >= 2) {
      const fromValue = await dateInputs.first().inputValue();
      const toValue = await dateInputs.nth(1).inputValue();
      console.log(`📅 Date filters - From: "${fromValue}", To: "${toValue}"`);
      expect(fromValue).not.toBe('');
      expect(toValue).not.toBe('');
      console.log('✅ SUCCESS: Date filters have 30-day default values');
    }

    // Check for permission issues in console logs
    const permissionLogs = consoleLogs.filter(log =>
      log.includes('permission') || log.includes('authorized') || log.includes('403')
    );

    if (permissionLogs.length > 0) {
      console.log('🚨 Permission-related logs found:');
      permissionLogs.forEach(log => console.log(`   ${log}`));
    }

    // Check for error logs
    const errorLogs = consoleLogs.filter(log =>
      log.includes('[error]') || log.includes('Failed to load')
    );

    if (errorLogs.length > 0) {
      console.log('🚨 Error logs found:');
      errorLogs.forEach(log => console.log(`   ${log}`));
    }

    // Take final screenshot
    await page.screenshot({ path: 'audit-page-debug-final.png', fullPage: true });

    console.log('📋 All console logs:');
    consoleLogs.forEach((log, i) => {
      console.log(`  ${i+1}: ${log}`);
    });

    console.log('✅ Debug test completed');

    // Summary report
    console.log('\n📊 DEBUG SUMMARY:');
    console.log(`   API Requests: ${apiRequests.length}`);
    console.log(`   API Responses: ${apiResponses.length}`);
    console.log(`   Table Rows: ${rowCount}`);
    console.log(`   Empty State: ${emptyVisible}`);
    console.log(`   Date Filters Have 30-day Default: ${dateCount >= 2 ? 'YES' : 'N/A'}`);
  });
});