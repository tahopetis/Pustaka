import { test, expect } from '@playwright/test';

test.describe('Quick Health Check', () => {
  test('should check if services are running', async ({ page }) => {
    try {
      // Check frontend
      const frontendResponse = await page.goto('http://localhost:3000');
      expect(frontendResponse.status()).toBe(200);

      // Check API health endpoint
      const apiResponse = await page.request.get('http://localhost:8080/health');
      console.log('API Health Status:', apiResponse.status());
      console.log('API Health Response:', await apiResponse.text());

      expect(apiResponse.status()).toBe(200);

      console.log('✅ Both frontend and API are running');
    } catch (error) {
      console.log('❌ Service check failed:', error.message);
      throw error;
    }
  });

  test('should test basic login flow', async ({ page }) => {
    try {
      await page.goto('http://localhost:3000/login');
      await page.waitForSelector('#username', { timeout: 10000 });

      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');

      // Wait for either dashboard or error
      await Promise.race([
        page.waitForURL('**/dashboard', { timeout: 15000 }),
        page.waitForSelector('.error, .alert', { timeout: 15000 })
      ]);

      const currentUrl = page.url();
      console.log('Current URL after login:', currentUrl);

      if (currentUrl.includes('/dashboard')) {
        console.log('✅ Login successful');

        // Check if relationship types link is present
        const relationshipTypesLink = page.locator('a[href="/relationship-types"]');
        if (await relationshipTypesLink.isVisible({ timeout: 5000 })) {
          console.log('✅ Relationship Types link is visible in sidebar');
        } else {
          console.log('❌ Relationship Types link is NOT visible in sidebar');
        }

        // Try to navigate to relationship types
        await page.goto('http://localhost:3000/relationship-types');
        await page.waitForTimeout(3000);

        const finalUrl = page.url();
        console.log('URL after navigating to relationship-types:', finalUrl);

        if (finalUrl.includes('/relationship-types')) {
          console.log('✅ Successfully navigated to relationship types page');

          // Check page content
          const pageTitle = await page.locator('h1, h2').first().textContent();
          console.log('Page title:', pageTitle);

          // Check for table or list
          const hasTable = await page.locator('table, .table').isVisible();
          console.log('Has table/list:', hasTable);

        } else if (finalUrl.includes('/dashboard')) {
          console.log('❌ Redirected to dashboard - likely missing permissions');
        } else {
          console.log('❌ Unexpected redirect');
        }
      } else {
        console.log('❌ Login failed or redirected unexpectedly');

        // Check for error messages
        const errorMessage = await page.locator('.error, .alert, [role="alert"]').first().textContent();
        if (errorMessage) {
          console.log('Error message:', errorMessage);
        }
      }
    } catch (error) {
      console.log('❌ Login test failed:', error.message);
      throw error;
    }
  });
});