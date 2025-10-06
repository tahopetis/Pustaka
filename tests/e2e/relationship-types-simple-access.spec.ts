import { test, expect } from '@playwright/test';

test.describe('Relationship Types - Simple Access Test', () => {
  test('should access relationship types page directly', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Navigate to relationship types directly
    await page.goto('http://localhost:3000/relationship-types');
    await page.waitForLoadState('networkidle');

    // Check current URL
    console.log('Current URL:', page.url());

    // Check if we're on the relationship types page (not redirected)
    if (page.url().includes('/relationship-types')) {
      console.log('✅ Successfully accessed relationship types page');

      // Check page title
      const title = await page.title();
      console.log('Page title:', title);

      // Take a screenshot to see what's on the page
      await page.screenshot({ path: 'relationship-types-page-success.png', fullPage: true });
      console.log('Screenshot saved as relationship-types-page-success.png');

      // Look for any h1 elements
      const h1Elements = await page.locator('h1').all();
      console.log('Number of h1 elements:', h1Elements.length);

      for (let i = 0; i < h1Elements.length; i++) {
        const text = await h1Elements[i].textContent();
        console.log(`h1[${i}]:`, text);
      }

      // Look for any content
      const bodyText = await page.locator('body').textContent();
      if (bodyText && bodyText.includes('Relationship Types')) {
        console.log('✅ Found "Relationship Types" text on page');
      } else {
        console.log('❌ "Relationship Types" text not found on page');
      }
    } else {
      console.log('❌ Got redirected to:', page.url());
      await page.screenshot({ path: 'relationship-types-page-redirected.png', fullPage: true });
    }
  });

  test('should verify relationship types link is clickable', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Find and click the relationship types link
    const relationshipTypesLink = page.locator('a[href="/relationship-types"]');
    await expect(relationshipTypesLink).toBeVisible();

    console.log('✅ Relationship Types link is visible in sidebar');

    // Click it
    await relationshipTypesLink.click();

    // Wait a moment for navigation
    await page.waitForTimeout(2000);

    console.log('After clicking link, current URL:', page.url());

    // Take screenshot to see result
    await page.screenshot({ path: 'after-clicking-relationship-types-link.png', fullPage: true });
  });
});