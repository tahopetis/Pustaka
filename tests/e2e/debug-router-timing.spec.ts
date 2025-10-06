import { test, expect } from '@playwright/test';

test.describe('Debug Router Timing', () => {
  test('check router guard behavior with console logs', async ({ page }) => {
    // Listen for console logs
    const consoleMessages = [];
    page.on('console', msg => {
      consoleMessages.push(msg.text());
      console.log('Console:', msg.type(), msg.text());
    });

    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Wait for auth store to initialize
    await page.waitForTimeout(3000);

    // Check auth store state
    const authState = await page.evaluate(() => {
      return new Promise((resolve) => {
        // Try to get the auth store from the Vue app
        const checkAuthStore = () => {
          try {
            // Access window.__VUE__ or other global Vue instance
            const app = document.querySelector('#app').__vue_app__;
            if (app && app._context && app._context.provides) {
              // Try to find the auth store
              for (const key in app._context.provides) {
                const value = app._context.provides[key];
                if (value && typeof value === 'object' && value.hasPermission) {
                  return {
                    user: value.user,
                    isAuthenticated: value.isAuthenticated,
                    isInitialized: value.isInitialized,
                    hasPermission: value.hasPermission('relationship_type:read')
                  };
                }
              }
            }
            return null;
          } catch (e) {
            return null;
          }
        };

        // Try multiple times as Vue might be initializing
        let attempts = 0;
        const check = () => {
          attempts++;
          const result = checkAuthStore();
          if (result || attempts > 10) {
            resolve(result);
          } else {
            setTimeout(check, 100);
          }
        };
        check();
      });
    });

    console.log('Auth store state:', authState);

    // Now try to navigate to relationship types
    console.log('Navigating to relationship types...');
    await page.goto('http://localhost:3000/relationship-types');

    // Wait for navigation
    await page.waitForTimeout(2000);

    console.log('Final URL:', page.url());
    console.log('Console messages received:', consoleMessages.length);

    // Look for router guard console messages
    const routerMessages = consoleMessages.filter(msg =>
      msg.includes('Router guard') || msg.includes('Missing permission')
    );
    console.log('Router guard messages:', routerMessages);

    // Take screenshot
    await page.screenshot({ path: 'debug-router-timing.png', fullPage: true });
  });
});