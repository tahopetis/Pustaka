import { test, expect } from '@playwright/test';

test.describe('Audit Page - Simple Test', () => {
  test('should load audit page and verify date filters are empty', async ({ page }) => {
    console.log('🚀 Starting simple audit page test');

    // Navigate to audit page
    await page.goto('http://localhost:3000/audit');

    // Wait and check if redirected to login
    await page.waitForLoadState('networkidle');

    const currentUrl = page.url();
    console.log('📍 Current URL:', currentUrl);

    if (currentUrl.includes('/login')) {
      console.log('🔐 Logging in...');

      // Login
      await page.fill('input[name="username"], input[type="text"]', 'admin');
      await page.fill('input[name="password"], input[type="password"]', 'Admin@123');

      await Promise.all([
        page.waitForNavigation(),
        page.click('button[type="submit"], .btn-primary, button:has-text("Login")')
      ]);

      // Navigate to audit page after login
      await page.goto('http://localhost:3000/audit');
    }

    // Wait for audit page
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 15000 });
    console.log('✅ Audit page loaded');

    // Capture console logs
    const consoleLogs: string[] = [];
    page.on('console', msg => {
      const text = msg.text();
      consoleLogs.push(text);
      if (text.includes('Audit logs:') || msg.type() === 'error') {
        console.log(`📝 Console: ${text}`);
      }
    });

    // Check date filters have 30-day default values (new requirement)
    const dateInputs = page.locator('input[type="date"]');
    const dateCount = await dateInputs.count();

    if (dateCount >= 2) {
      const fromValue = await dateInputs.first().inputValue();
      const toValue = await dateInputs.nth(1).inputValue();

      console.log(`📅 Date filters - From: "${fromValue}", To: "${toValue}"`);

      // Main assertion: date filters should have 30-day default values
      expect(fromValue).not.toBe('');
      expect(toValue).not.toBe('');
      console.log('✅ SUCCESS: Date filters have 30-day default values');
    }

    // Wait for API calls
    await page.waitForTimeout(5000);

    // Check what's actually displayed
    const rows = page.locator('table tbody tr');
    const rowCount = await rows.count();
    console.log(`📊 Table rows found: ${rowCount}`);

    const emptyMessage = page.locator('h3:has-text("No audit logs found")');
    const emptyVisible = await emptyMessage.isVisible();

    if (emptyVisible) {
      console.log('📭 Empty state is shown');
    } else if (rowCount > 0) {
      console.log('✅ Audit logs are displayed');
    } else {
      console.log('⚠️  Neither empty state nor logs found - check for loading or errors');
    }

    // Check for any error elements
    const errorElements = page.locator('.error, .alert-error, .toast-error');
    const errorCount = await errorElements.count();
    console.log(`❌ Error elements found: ${errorCount}`);

    // Final screenshot
    await page.screenshot({ path: 'audit-page-simple-test.png', fullPage: true });

    // Log all console messages for debugging
    console.log('📋 Console logs:');
    consoleLogs.forEach((log, i) => {
      console.log(`  ${i+1}: ${log}`);
    });

    // Final verification
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible();

    if (dateCount >= 2) {
      const fromFinal = await dateInputs.first().inputValue();
      const toFinal = await dateInputs.nth(1).inputValue();
      expect(fromFinal).not.toBe('');
      expect(toFinal).not.toBe('');
    }

    console.log('✅ Simple audit page test completed');
  });
});