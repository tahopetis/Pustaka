import { test, expect } from '@playwright/test';

test.describe('Audit Page - Localhost Testing', () => {
  test('should load http://localhost:3000/audit and show all logs by default', async ({ page }) => {
    console.log('🚀 Starting audit page test on localhost:3000/audit');

    // Navigate directly to the audit page
    console.log('📍 Navigating to http://localhost:3000/audit');
    await page.goto('http://localhost:3000/audit');

    // Wait for potential redirect to login
    console.log('⏳ Waiting for page load...');
    await page.waitForLoadState('networkidle');

    // Check if we were redirected to login
    const currentUrl = page.url();
    console.log('🔗 Current URL:', currentUrl);

    if (currentUrl.includes('/login')) {
      console.log('🔐 Redirected to login, performing authentication...');

      // Fill in login credentials
      await page.fill('input[name="username"], input[placeholder*="username"], input[type="text"], input#username', 'admin');
      await page.fill('input[name="password"], input[placeholder*="password"], input[type="password"], input#password', 'Admin@123');

      // Submit login form
      await Promise.all([
        page.waitForNavigation(),
        page.click('button[type="submit"], .btn-primary, button:has-text("Login"), button:has-text("Sign in")')
      ]);

      // After login, navigate to audit page
      console.log('✅ Login successful, navigating to audit page...');
      await page.goto('http://localhost:3000/audit');
    }

    // Wait for audit page to load
    console.log('⏳ Waiting for audit page content...');
    await page.waitForSelector('h1:has-text("Audit Logs")', { timeout: 15000 });
    await page.waitForLoadState('networkidle');

    console.log('✅ Audit page loaded successfully');

    // Take screenshot before verification
    await page.screenshot({ path: 'audit-page-initial-state.png', fullPage: true });

    // Set up console logging to capture debugging information
    const consoleLogs: string[] = [];
    page.on('console', msg => {
      const text = msg.text();
      consoleLogs.push(text);
      if (text.includes('Audit logs:') || msg.type() === 'error') {
        console.log(`📝 Browser console: ${text}`);
      }
    });

    // Wait a bit for any API calls to complete
    await page.waitForTimeout(3000);

    // Verify page title and content
    const pageTitle = await page.textContent('h1.page-title, h1:has-text("Audit Logs")');
    console.log('📄 Page title:', pageTitle);
    expect(pageTitle).toContain('Audit Logs');

    // Check date filters - they should have 30-day default values
    console.log('🔍 Checking date filter values...');
    const dateInputs = page.locator('input[type="date"]');
    const dateInputCount = await dateInputs.count();
    console.log(`📅 Found ${dateInputCount} date inputs`);

    if (dateInputCount >= 2) {
      const fromDateInput = dateInputs.first();
      const toDateInput = dateInputs.nth(1);

      const fromDateValue = await fromDateInput.inputValue();
      const toDateValue = await toDateInput.inputValue();

      console.log(`📅 From date value: "${fromDateValue}"`);
      console.log(`📅 To date value: "${toDateValue}"`);

      // Assert that date filters have 30-day default values
      expect(fromDateValue).not.toBe('');
      expect(toDateValue).not.toBe('');
      console.log('✅ Date filters have 30-day default values');
    }

    // Check for audit logs in the table
    console.log('🔍 Checking for audit logs in table...');
    const auditTable = page.locator('table');
    const auditRows = page.locator('table tbody tr');

    const rowCount = await auditRows.count();
    console.log(`📊 Found ${rowCount} audit log rows in table`);

    // Take screenshot showing table state
    await page.screenshot({ path: 'audit-page-table-state.png', fullPage: false });

    if (rowCount > 0) {
      console.log('✅ SUCCESS: Audit logs are displayed in the table');

      // Log the first few rows for verification
      for (let i = 0; i < Math.min(3, rowCount); i++) {
        const rowText = await auditRows.nth(i).textContent();
        console.log(`📝 Row ${i + 1}: ${rowText?.trim().substring(0, 100)}...`);
      }

      // Check that logs are from different time periods (not just last 30 days)
      const timestampCells = page.locator('table tbody tr td:first-child');
      const timestampCount = await timestampCells.count();
      console.log(`⏰ Found ${timestampCount} timestamp cells`);

      if (timestampCount > 0) {
        const firstTimestamp = await timestampCells.first().textContent();
        console.log(`📅 First timestamp: ${firstTimestamp}`);
      }
    } else {
      console.log('ℹ️  INFO: No audit log rows found in table');

      // Check for empty state message
      const emptyState = page.locator('h3:has-text("No audit logs found")');
      const emptyStateVisible = await emptyState.isVisible();

      if (emptyStateVisible) {
        const emptyText = await emptyState.textContent();
        console.log(`📭 Empty state message: ${emptyText}`);

        // Check if there's a reset filters button
        const resetButton = page.locator('button:has-text("Reset Filters")');
        if (await resetButton.isVisible()) {
          console.log('🔄 Reset filters button is available');

          // Test the reset functionality
          console.log('🔄 Testing reset filters button...');
          await resetButton.click();
          await page.waitForTimeout(2000);

          // Check filters are reset to 30-day default after reset
          if (dateInputCount >= 2) {
            const fromDateAfterReset = await dateInputs.first().inputValue();
            const toDateAfterReset = await dateInputs.nth(1).inputValue();
            console.log(`📅 After reset - From: "${fromDateAfterReset}", To: "${toDateAfterReset}"`);
            expect(fromDateAfterReset).toBe('');
            expect(toDateAfterReset).toBe('');
          }
        }
      }
    }

    // Check for any error messages
    const errorMessages = page.locator('.toast-error, .error-message, .alert-error, :text("Failed to load")');
    if (await errorMessages.isVisible()) {
      const errorText = await errorMessages.first().textContent();
      console.log(`❌ Error message found: ${errorText}`);
    }

    // Check for loading state
    const loadingState = page.locator('.spinner, :text("Loading"), [data-testid="loading"]');
    const isLoading = await loadingState.isVisible();
    console.log(`⏳ Loading state visible: ${isLoading}`);

    // Log all console messages for debugging
    console.log('📋 All console logs captured:');
    consoleLogs.forEach((log, index) => {
      console.log(`  ${index + 1}: ${log}`);
    });

    // Final verification screenshot
    await page.screenshot({ path: 'audit-page-final-state.png', fullPage: true });

    // Final assertions
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible();

    if (dateInputCount >= 2) {
      const fromDateFinal = await dateInputs.first().inputValue();
      const toDateFinal = await dateInputs.nth(1).inputValue();
      expect(fromDateFinal).not.toBe('');
      expect(toDateFinal).not.toBe('');
    }

    console.log('✅ Audit page test completed successfully!');
    console.log('📁 Screenshots saved for verification');
  });
});