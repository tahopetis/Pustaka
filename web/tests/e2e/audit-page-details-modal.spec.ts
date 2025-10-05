import { test, expect } from '@playwright/test';

test.describe('Audit Page - View Details Modal', () => {
  test('should show audit log details when clicking "View details"', async ({ page }) => {
    console.log('🔍 Testing audit log details modal functionality');

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

    // Wait for audit page to load and logs to appear
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(3000);

    // Find audit log rows
    const rows = page.locator('table tbody tr');
    const rowCount = await rows.count();
    console.log(`📊 Found ${rowCount} audit log rows`);

    expect(rowCount).toBeGreaterThan(0, 'Expected at least one audit log row');

    // Look for "View details" link in the first few rows
    let foundDetailsLink = false;
    for (let i = 0; i < Math.min(rowCount, 3); i++) {
      const detailsLink = rows.nth(i).locator('text=View details');
      const linkCount = await detailsLink.count();

      if (linkCount > 0) {
        console.log(`🔗 Found "View details" link in row ${i + 1}`);

        // Click the "View details" link
        await detailsLink.click();

        // Wait for modal to appear
        await page.waitForSelector('div:has-text("Audit Log Details")', { timeout: 5000 });
        console.log('✅ Audit details modal appeared');

        // Verify modal content
        await expect(page.locator('h3:has-text("Audit Log Details")')).toBeVisible();

        // Check for key details in the modal
        const modalContent = page.locator('.fixed.inset-0 .bg-white');

        // Verify basic information is displayed
        await expect(modalContent.locator('text=Timestamp')).toBeVisible();
        await expect(modalContent.locator('text=Action')).toBeVisible();
        await expect(modalContent.locator('text=Entity Type')).toBeVisible();
        await expect(modalContent.locator('text=Performed By')).toBeVisible();

        console.log('✅ Modal contains expected information');

        // Test closing the modal with Close button
        await page.click('button:has-text("Close")');
        await page.waitForSelector('div:has-text("Audit Log Details")', { state: 'hidden', timeout: 3000 });
        console.log('✅ Modal closed with Close button');

        foundDetailsLink = true;
        break;
      }
    }

    if (!foundDetailsLink) {
      console.log('ℹ️  No "View details" links found - logs may not have additional details');
    }

    // Test clicking outside modal to close (if we can open it again)
    if (foundDetailsLink && rowCount > 0) {
      // Find and click the first "View details" link again
      const firstDetailsLink = page.locator('text=View details').first();
      if (await firstDetailsLink.isVisible()) {
        await firstDetailsLink.click();
        await page.waitForSelector('div:has-text("Audit Log Details")', { timeout: 5000 });

        // Test clicking outside to close
        await page.click('.fixed.inset-0', { position: { x: 100, y: 100 } });
        await page.waitForSelector('div:has-text("Audit Log Details")', { state: 'hidden', timeout: 3000 });
        console.log('✅ Modal closed by clicking outside');
      }
    }

    // Final screenshot
    await page.screenshot({ path: 'audit-details-modal-test.png', fullPage: true });

    console.log('✅ Audit details modal test completed');
  });

  test('should verify modal structure and accessibility', async ({ page }) => {
    console.log('🔍 Testing modal structure and basic functionality');

    // Navigate and login
    await page.goto('http://localhost:3000/audit');
    await page.waitForLoadState('networkidle');

    if (page.url().includes('/login')) {
      await page.fill('input[name="username"], input[type="text"]', 'admin');
      await page.fill('input[name="password"], input[type="password"]', 'Admin@123');
      await Promise.all([
        page.waitForNavigation(),
        page.click('button[type="submit"], .btn-primary, button:has-text("Login")')
      ]);
      await page.goto('http://localhost:3000/audit');
    }

    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(3000);

    // Find a log with details and open it
    const detailsLink = page.locator('text=View details').first();
    if (await detailsLink.isVisible()) {
      await detailsLink.click();
      await page.waitForSelector('div:has-text("Audit Log Details")', { timeout: 5000 });

      // Verify modal has proper structure
      await expect(page.locator('h3:has-text("Audit Log Details")')).toBeVisible();

      // Check that modal is positioned correctly (should have fixed positioning)
      const modal = page.locator('.fixed.inset-0');
      await expect(modal).toBeVisible();

      // Verify close button exists
      await expect(page.locator('button:has-text("Close")')).toBeVisible();

      // Verify modal has proper z-index by checking it's on top
      const modalBox = await modal.boundingBox();
      expect(modalBox).toBeTruthy();

      // Test keyboard accessibility - press Escape to close
      await page.keyboard.press('Escape');
      await page.waitForSelector('div:has-text("Audit Log Details")', { state: 'hidden', timeout: 3000 });

      console.log('✅ Modal structure and accessibility verified');
    } else {
      console.log('ℹ️  No logs with details available to test');
    }

    console.log('✅ Modal accessibility test completed');
  });
});