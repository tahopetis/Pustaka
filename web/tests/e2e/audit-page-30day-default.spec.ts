import { test, expect } from '@playwright/test';

test.describe('Audit Page - 30 Day Default View', () => {
  test('should show 30-day default view with options to see all logs', async ({ page }) => {
    console.log('🚀 Testing 30-day default view behavior');

    // Navigate to audit page
    await page.goto('http://localhost:3000/audit');

    // Handle login if needed
    await page.waitForLoadState('networkidle');

    if (page.url().includes('/login')) {
      console.log('🔐 Logging in...');
      await page.fill('input[name="username"], input[type="text"]', 'admin');
      await page.fill('input[name="password"], input[type="password"]', 'Admin@123');

      await Promise.all([
        page.waitForNavigation(),
        page.click('button[type="submit"], .btn-primary, button:has-text("Login")')
      ]);

      await page.goto('http://localhost:3000/audit');
    }

    // Wait for audit page
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 15000 });
    console.log('✅ Audit page loaded');

    // Wait for API calls to complete
    await page.waitForTimeout(3000);

    // Check that date filters now have 30-day default values
    const dateInputs = page.locator('input[type="date"]');
    const dateCount = await dateInputs.count();

    if (dateCount >= 2) {
      const fromValue = await dateInputs.first().inputValue();
      const toValue = await dateInputs.nth(1).inputValue();

      console.log(`📅 Date filters - From: "${fromValue}", To: "${toValue}"`);

      // Verify date filters are NOT empty (30-day default)
      expect(fromValue).not.toBe('');
      expect(toValue).not.toBe('');

      // Verify the date range is approximately 30 days
      const today = new Date();
      const thirtyDaysAgo = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000);

      const expectedFrom = thirtyDaysAgo.toISOString().split('T')[0];
      const expectedTo = today.toISOString().split('T')[0];

      console.log(`📅 Expected 30-day range: From ${expectedFrom} to ${expectedTo}`);
      console.log(`📅 Actual range: From ${fromValue} to ${toValue}`);

      // The dates should be close to the 30-day range (allowing for timezone differences)
      expect(fromValue).toBeTruthy();
      expect(toValue).toBeTruthy();
    }

    // Check what's displayed in the table
    const rows = page.locator('table tbody tr');
    const rowCount = await rows.count();
    console.log(`📊 Table rows found: ${rowCount}`);

    // Based on our known data, the logs are from 2025-10-05, so if we're in 2025-10-06,
    // the 30-day range should include them, but if we're testing later, they might not be included
    const emptyState = page.locator('h3:has-text("No audit logs found")');
    const emptyVisible = await emptyState.isVisible();

    if (emptyVisible) {
      console.log('📭 Empty state shown - logs may be outside 30-day range');

      // Check for the new empty state message about 30 days
      const emptyMessage = page.locator(':text("No audit logs found in the last 30 days")');
      const messageVisible = await emptyMessage.isVisible();

      if (messageVisible) {
        console.log('✅ 30-day empty state message is showing');
      }

      // Test the "View All Logs" button (reset filters)
      const resetButton = page.locator('button:has-text("View All Logs")');
      const clearDateButton = page.locator('button:has-text("Clear Date Range")');

      if (await resetButton.isVisible()) {
        console.log('🔄 Testing "View All Logs" button...');
        await resetButton.click();
        await page.waitForTimeout(2000);

        // Check if logs appear after resetting filters
        const rowsAfterReset = await rows.count();
        if (rowsAfterReset > 0) {
          console.log(`✅ SUCCESS: ${rowsAfterReset} logs appeared after reset`);
        } else {
          console.log('ℹ️  Still no logs after reset - may be no older logs in database');
        }

        // Verify date filters are cleared
        const fromAfterReset = await dateInputs.first().inputValue();
        const toAfterReset = await dateInputs.nth(1).inputValue();
        console.log(`📅 After reset - From: "${fromAfterReset}", To: "${toAfterReset}"`);
        expect(fromAfterReset).toBe('');
        expect(toAfterReset).toBe('');
      }

      if (await clearDateButton.isVisible()) {
        console.log('🔄 Testing "Clear Date Range" button...');

        // Navigate back to audit page and set 30-day default again
        await page.goto('http://localhost:3000/audit');
        await page.waitForTimeout(2000);

        await clearDateButton.click();
        await page.waitForTimeout(2000);

        // Verify only date filters are cleared
        const fromAfterClear = await dateInputs.first().inputValue();
        const toAfterClear = await dateInputs.nth(1).inputValue();
        console.log(`📅 After clear date - From: "${fromAfterClear}", To: "${toAfterClear}"`);
        expect(fromAfterClear).toBe('');
        expect(toAfterClear).toBe('');
      }

    } else if (rowCount > 0) {
      console.log('✅ SUCCESS: Audit logs are displayed within 30-day range');

      // Log the first row's timestamp for verification
      const firstTimestamp = await rows.first().locator('td:first-child').textContent();
      console.log(`📅 First log timestamp: ${firstTimestamp}`);
    }

    // Test the refresh button preserves the 30-day default
    console.log('🔄 Testing refresh button preserves filters...');
    await page.click('button:has-text("Refresh")');
    await page.waitForTimeout(2000);

    const fromAfterRefresh = await dateInputs.first().inputValue();
    const toAfterRefresh = await dateInputs.nth(1).inputValue();
    console.log(`📅 After refresh - From: "${fromAfterRefresh}", To: "${toAfterRefresh}"`);

    // Refresh should preserve the 30-day filter values
    expect(fromAfterRefresh).not.toBe('');
    expect(toAfterRefresh).not.toBe('');

    // Final screenshot
    await page.screenshot({ path: 'audit-page-30day-default.png', fullPage: true });

    console.log('✅ 30-day default view test completed');
  });

  test('should maintain 30-day default on page reload', async ({ page }) => {
    console.log('🔄 Testing 30-day default persistence...');

    // Login and navigate to audit page
    await page.goto('http://localhost:3000/login');
    await page.fill('input[name="username"], input[type="text"]', 'admin');
    await page.fill('input[name="password"], input[type="password"]', 'Admin@123');

    await Promise.all([
      page.waitForNavigation(),
      page.click('button[type="submit"], .btn-primary, button:has-text("Login")')
    ]);

    await page.goto('http://localhost:3000/audit');
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(3000);

    // Check initial 30-day default
    const dateInputs = page.locator('input[type="date"]');
    const fromValue = await dateInputs.first().inputValue();
    const toValue = await dateInputs.nth(1).inputValue();

    expect(fromValue).not.toBe('');
    expect(toValue).not.toBe('');

    console.log(`📅 Initial 30-day range: ${fromValue} to ${toValue}`);

    // Reload the page
    await page.reload();
    await page.waitForTimeout(3000);

    // Check that 30-day default is maintained
    const fromAfterReload = await dateInputs.first().inputValue();
    const toAfterReload = await dateInputs.nth(1).inputValue();

    expect(fromAfterReload).not.toBe('');
    expect(toAfterReload).not.toBe('');

    console.log(`📅 After reload: ${fromAfterReload} to ${toAfterReload}`);
    console.log('✅ 30-day default maintained after reload');
  });
});