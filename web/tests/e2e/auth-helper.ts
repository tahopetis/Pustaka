import { test as base, Page } from '@playwright/test';

// Extend base test to include authenticated page
export const test = base.extend({
  page: async ({ page }, use) => {
    // Navigate to login page or main page first
    await page.goto('http://localhost:3000');

    // Wait a moment for page to load
    await page.waitForLoadState('networkidle');

    // Check if we're already logged in by looking for login form or authenticated content
    try {
      // Check if login form is present
      const loginForm = page.locator('form').filter({ has: page.locator('input[type="password"]') });

      if (await loginForm.isVisible({ timeout: 5000 })) {
        // Fill in login credentials
        await page.fill('input[name="username"], input[type="text"], input[placeholder*="username"], input[placeholder*="email"]', 'admin');
        await page.fill('input[name="password"], input[type="password"], input[placeholder*="password"]', 'Admin@123');

        // Click login button
        await page.click('button[type="submit"], button:has-text("Login"), button:has-text("Sign in"), button:has-text("Sign In")');

        // Wait for login to complete
        await page.waitForLoadState('networkidle');
        await page.waitForTimeout(2000);
      }
    } catch (error) {
      console.log('Login form not found or login failed, assuming already authenticated or using different auth method');
    }

    // Use the authenticated page
    await use(page);
  },
});

export { expect } from '@playwright/test';