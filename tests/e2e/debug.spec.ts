import { test, expect } from '@playwright/test';

test.describe('Debug - Check actual page content', () => {
  test('debug user permissions and relationship types navigation', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Check user permissions by accessing localStorage
    const localStorageData = await page.evaluate(() => {
      const token = localStorage.getItem('access_token');
      const userStr = localStorage.getItem('user');
      return {
        hasToken: !!token,
        user: userStr ? JSON.parse(userStr) : null,
        tokenLength: token?.length || 0
      };
    });
    console.log('Local storage data:', localStorageData);

    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Wait for navigation to complete
    await page.waitForLoadState('networkidle');

    // Check current URL
    const currentUrl = page.url();
    console.log('Current URL after navigation:', currentUrl);

    // Take a screenshot to see what's actually there
    await page.screenshot({ path: 'debug-relationship-types-page.png', fullPage: true });

    // Get the page title
    const title = await page.title();
    console.log('Page title:', title);

    // Get all h1 elements first
    const h1Elements = await page.locator('h1').all();
    console.log('Number of h1 elements:', h1Elements.length);

    // Get the page title h1 (page-title class)
    const pageTitle = await page.locator('h1.page-title').textContent();
    console.log('Page title h1:', pageTitle);

    // Check if we got redirected back to dashboard
    if (currentUrl.includes('/dashboard')) {
      console.log('Got redirected back to dashboard - checking navigation links...');

      // Find navigation links
      const navLinks = await page.locator('nav a').allTextContents();
      console.log('Navigation links found:', navLinks);

      // Try to find relationship types link specifically
      const relationshipTypeLink = page.locator('a[href*="relationship-types"]');
      const isLinkVisible = await relationshipTypeLink.isVisible();
      console.log('Relationship types link visible:', isLinkVisible);

      if (isLinkVisible) {
        const linkText = await relationshipTypeLink.textContent();
        console.log('Relationship types link text:', linkText);
        await relationshipTypeLink.click();
        await page.waitForLoadState('networkidle');
        console.log('URL after clicking relationship types link:', page.url());
      }
    }
  });
});