import { test, expect } from '@playwright/test';

test.describe('Audit Page Quick Test', () => {
  test('should load audit page and check default behavior', async ({ page }) => {
    // Login first
    await page.goto('http://localhost:3000/login');

    // Fill login form
    await page.fill('input[name="username"], input[placeholder*="username"], input[type="text"]', 'admin');
    await page.fill('input[name="password"], input[placeholder*="password"], input[type="password"]', 'Admin@123');

    // Submit login
    await Promise.all([
      page.waitForNavigation(),
      page.click('button[type="submit"], .btn-primary, button:has-text("Login")')
    ]);

    // Navigate to audit page
    await page.goto('http://localhost:3000/audit');

    // Wait for page to load
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 10000 });

    // Check if date filters are empty (should be the fix)
    // Look for date inputs with the correct selectors
    const fromDateInput = page.locator('input[type="date"]').first();
    const toDateInput = page.locator('input[type="date"]').nth(1);

    // Wait for inputs to be available
    await fromDateInput.waitFor({ state: 'attached', timeout: 5000 });
    await toDateInput.waitFor({ state: 'attached', timeout: 5000 });

    const fromDateValue = await fromDateInput.inputValue();
    const toDateValue = await toDateInput.inputValue();

    console.log('Date filter values - From:', fromDateValue, 'To:', toDateValue);

    // Assert that date filters are empty (showing all logs by default)
    expect(fromDateValue).toBe('');
    expect(toDateValue).toBe('');

    // Take screenshot for verification
    await page.screenshot({ path: 'audit-page-verification.png', fullPage: true });

    // Check if there are any audit logs or empty state
    const auditLogRows = page.locator('table tbody tr');
    const emptyState = page.locator(':text("No audit logs found")');

    if (await auditLogRows.count() > 0) {
      console.log('✅ SUCCESS: Audit logs are displayed');
      console.log('Number of rows:', await auditLogRows.count());
    } else if (await emptyState.isVisible()) {
      console.log('ℹ️  INFO: Empty state shown (no logs in database)');
    } else {
      console.log('⚠️  WARNING: Unexpected state on audit page');
    }

    // Test reset filters button
    await page.click('button:has-text("Reset Filters")');
    await page.waitForTimeout(1000);

    // Verify filters are still empty after reset
    const fromResetValue = await fromDateInput.inputValue();
    const toResetValue = await toDateInput.inputValue();

    expect(fromResetValue).toBe('');
    expect(toResetValue).toBe('');

    console.log('✅ SUCCESS: Date filters remain empty after reset');
  });
});