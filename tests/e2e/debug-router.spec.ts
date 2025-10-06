import { test, expect } from '@playwright/test';

test.describe('Debug Router Guard', () => {
  test('check auth store initialization', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Wait a bit for auth store to fully initialize
    await page.waitForTimeout(2000);

    // Check auth store state directly in the browser
    const authState = await page.evaluate(() => {
      // @ts-ignore - accessing window for debugging
      const { useAuthStore } = window.Pinia?.store || {};
      if (useAuthStore) {
        const authStore = useAuthStore();
        return {
          user: authStore.user,
          isAuthenticated: authStore.isAuthenticated,
          isInitialized: authStore.isInitialized,
          hasPermission: authStore.hasPermission('relationship-type:read')
        };
      }
      return null;
    });

    console.log('Auth store state:', authState);

    // Now try to navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Check final URL
    console.log('Final URL after navigation attempt:', page.url());
  });
});