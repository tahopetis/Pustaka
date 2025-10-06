import { test, expect } from '@playwright/test';

test.describe('Debug Authentication & Permissions', () => {
  test('debug: test login and check relationship types permissions', async ({ page }) => {
    // Console logging for debugging
    console.log('=== Starting Debug Test ===');

    // Capture all console messages
    page.on('console', msg => {
      console.log(`[${msg.type()}] ${msg.text()}`);
    });

    page.on('pageerror', error => {
      console.log(`[PAGE ERROR] ${error.message}`);
    });

    // Step 1: Check if services are accessible
    console.log('Step 1: Checking service health...');

    try {
      const frontendResponse = await page.goto('http://localhost:3000');
      console.log('Frontend status:', frontendResponse?.status());

      const apiResponse = await page.request.get('http://localhost:8080/health');
      console.log('API health status:', apiResponse.status());
      console.log('API health response:', await apiResponse.text());
    } catch (error) {
      console.log('Service health check failed:', error.message);
      throw error;
    }

    // Step 2: Attempt login
    console.log('Step 2: Attempting login...');
    await page.goto('http://localhost:3000/login');

    // Wait for login form
    await page.waitForSelector('#username', { timeout: 10000 });
    console.log('✅ Login form loaded');

    // Fill credentials
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    console.log('✅ Credentials filled');

    // Submit login
    await page.click('button[type="submit"]');
    console.log('✅ Login form submitted');

    // Wait for navigation
    await page.waitForTimeout(3000);
    const currentUrl = page.url();
    console.log('Current URL after login:', currentUrl);

    // Step 3: Check authentication state
    if (currentUrl.includes('/dashboard')) {
      console.log('✅ Login successful - redirected to dashboard');

      // Wait for page to fully load
      await page.waitForLoadState('networkidle');

      // Check for Relationship Types link in sidebar
      console.log('Step 3: Checking for Relationship Types link in sidebar...');

      // Try different selectors for the sidebar
      const relationshipTypesSelectors = [
        'a[href="/relationship-types"]',
        'nav a[href="/relationship-types"]',
        '.sidebar a[href="/relationship-types"]',
        'text="Relationship Types"',
        'a:has-text("Relationship Types")'
      ];

      let relationshipTypesLink = null;
      for (const selector of relationshipTypesSelectors) {
        try {
          const element = page.locator(selector);
          if (await element.isVisible({ timeout: 2000 })) {
            relationshipTypesLink = element;
            console.log(`✅ Relationship Types link found with selector: ${selector}`);
            break;
          }
        } catch (e) {
          // Selector not found, try next one
        }
      }

      if (relationshipTypesLink) {
        console.log('✅ Relationship Types link is visible');

        // Click the link
        await relationshipTypesLink.click();
        await page.waitForTimeout(3000);

        const relationshipTypesUrl = page.url();
        console.log('URL after clicking Relationship Types link:', relationshipTypesUrl);

        if (relationshipTypesUrl.includes('/relationship-types')) {
          console.log('✅ Successfully navigated to Relationship Types page');

          // Check page content
          await page.waitForLoadState('networkidle');

          // Check for page title
          const pageTitle = await page.locator('h1, h2').first().textContent();
          console.log('Page title:', pageTitle);

          // Check for table or list
          const tableLocator = page.locator('table, .table, [data-testid*="table"]');
          const hasTable = await tableLocator.isVisible();
          console.log('Has table/list:', hasTable);

          // Check for create button
          const createButton = page.locator('a[href="/relationship-types/new"], button:has-text("Create"), button:has-text("New")');
          const hasCreateButton = await createButton.isVisible();
          console.log('Has create button:', hasCreateButton);

          // Check for search/filter elements
          const searchInput = page.locator('input[placeholder*="search" i], input[placeholder*="Search" i]');
          const hasSearch = await searchInput.isVisible();
          console.log('Has search input:', hasSearch);

          // Take screenshot for visual debugging
          await page.screenshot({ path: '/home/syam/dev/Pustaka/relationship-types-page-debug.png' });
          console.log('✅ Screenshot saved');

        } else if (relationshipTypesUrl.includes('/dashboard')) {
          console.log('❌ Redirected back to dashboard - likely missing permissions');

          // Check for any error messages
          const errorMessage = await page.locator('.error, .alert, [role="alert"], .toast-error').first().textContent();
          if (errorMessage) {
            console.log('Error message found:', errorMessage);
          }

          // Check authentication state from browser
          const authToken = await page.evaluate(() => {
            return localStorage.getItem('access_token');
          });
          console.log('Access token in localStorage:', authToken ? 'Present' : 'Missing');

          const userData = await page.evaluate(() => {
            const userData = localStorage.getItem('user');
            return userData ? JSON.parse(userData) : null;
          });
          console.log('User data in localStorage:', userData);

        } else {
          console.log('❌ Unexpected redirect after clicking Relationship Types link');
        }

      } else {
        console.log('❌ Relationship Types link NOT found in sidebar');

        // Take screenshot to see what's actually in the sidebar
        await page.screenshot({ path: '/home/syam/dev/Pustaka/sidebar-debug.png' });
        console.log('✅ Sidebar screenshot saved');

        // List all visible links in sidebar
        const sidebarLinks = await page.locator('nav a, .sidebar a, .nav-link').allTextContents();
        console.log('Available sidebar links:', sidebarLinks);

        // Check if sidebar is even visible
        const sidebarVisible = await page.locator('nav, .sidebar, [role="navigation"]').isVisible();
        console.log('Sidebar visible:', sidebarVisible);
      }

    } else {
      console.log('❌ Login failed - current URL:', currentUrl);

      // Check for login error messages
      const loginError = await page.locator('.error, .alert, [role="alert"]').first().textContent();
      if (loginError) {
        console.log('Login error message:', loginError);
      }

      // Take screenshot to see login state
      await page.screenshot({ path: '/home/syam/dev/Pustaka/login-debug.png' });
      console.log('✅ Login screenshot saved');
    }

    console.log('=== Debug Test Complete ===');
  });

  test('debug: check relationship types permissions directly', async ({ page }) => {
    console.log('=== Testing Direct Navigation to Relationship Types ===');

    // Try to navigate directly to relationship types without login
    await page.goto('http://localhost:3000/relationship-types');
    await page.waitForTimeout(2000);

    const directNavUrl = page.url();
    console.log('Direct navigation result:', directNavUrl);

    if (directNavUrl.includes('/login')) {
      console.log('✅ Correctly redirected to login for protected route');
    } else {
      console.log('❌ Should have redirected to login');
    }

    // Now try with valid authentication
    console.log('Login and try direct navigation...');

    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');

    await page.waitForURL('**/dashboard', { timeout: 15000 });
    console.log('✅ Login successful');

    // Now try direct navigation to relationship types
    await page.goto('http://localhost:3000/relationship-types');
    await page.waitForTimeout(3000);

    const authDirectNavUrl = page.url();
    console.log('Authenticated direct navigation result:', authDirectNavUrl);

    if (authDirectNavUrl.includes('/relationship-types')) {
      console.log('✅ Successfully accessed relationship types page with authentication');
    } else if (authDirectNavUrl.includes('/dashboard')) {
      console.log('❌ Redirected to dashboard - permission issue');

      // Check permissions in localStorage
      const permissions = await page.evaluate(() => {
        const userData = localStorage.getItem('user');
        return userData ? JSON.parse(userData).permissions : [];
      });
      console.log('User permissions:', permissions);

      // Check if relationship_type:read permission is present
      const hasPermission = permissions.includes('relationship_type:read');
      console.log('Has relationship_type:read permission:', hasPermission);
    }

    console.log('=== Direct Navigation Test Complete ===');
  });
});