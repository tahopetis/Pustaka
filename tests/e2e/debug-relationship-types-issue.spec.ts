import { test, expect } from '@playwright/test';

test.describe('Relationship Types Issue Investigation', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'Admin@123');
    await page.click('button[type="submit"]');

    // Wait for dashboard to load
    await page.waitForURL('http://localhost:3000/dashboard');
    await expect(page.locator('h1')).toContainText('Dashboard');
  });

  test('investigate relationship types page data loading issue', async ({ page }) => {
    // Navigate to relationship types page
    await page.goto('http://localhost:3000/relationship-types');

    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Take screenshot of initial state
    await page.screenshot({ path: 'relationship-types-initial-state.png' });

    // Check page title
    await expect(page.locator('h1')).toContainText('Relationship Types');

    // Check what the sidebar shows vs main content
    const sidebarCount = page.locator('nav a[href="/relationship-types"]');
    const sidebarText = await sidebarCount.textContent();
    console.log('Sidebar text:', sidebarText);

    // Check main content area
    const mainContent = page.locator('.page-container');
    const mainText = await mainContent.textContent();
    console.log('Main content text:', mainText);

    // Check for loading state
    const loadingElement = page.locator('.spinner');
    const isLoading = await loadingElement.isVisible();
    console.log('Is loading:', isLoading);

    // Wait a bit more if loading
    if (isLoading) {
      await page.waitForTimeout(5000);
      await page.screenshot({ path: 'relationship-types-after-loading.png' });
    }

    // Check for empty state message
    const emptyState = page.locator('text=No relationship types found');
    const isEmptyStateVisible = await emptyState.isVisible();
    console.log('Empty state visible:', isEmptyStateVisible);

    // Check if there's a count displayed in the card header
    const cardHeader = page.locator('.card-header h3');
    const headerText = await cardHeader.textContent();
    console.log('Card header text:', headerText);

    // Monitor network requests
    const responses = [];
    page.on('response', response => {
      if (response.url().includes('/relationship-types')) {
        responses.push({
          url: response.url(),
          status: response.status(),
          ok: response.ok()
        });
      }
    });

    // Reload the page to catch network requests
    await page.reload();
    await page.waitForLoadState('networkidle');

    console.log('Network responses for relationship-types:', responses);

    // Check specific API endpoints
    const apiResponses = await page.evaluate(() => {
      return fetch('/api/v1/relationship-types', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        }
      }).then(response => response.json()).catch(err => ({ error: err.message }));
    });

    console.log('Direct API response:', apiResponses);

    // Check stats API
    const statsResponse = await page.evaluate(() => {
      return fetch('/api/v1/relationship-types/stats', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        }
      }).then(response => response.json()).catch(err => ({ error: err.message }));
    });

    console.log('Stats API response:', statsResponse);

    // Check what data is in the Vue store
    const storeData = await page.evaluate(() => {
      return {
        relationshipTypes: window.__NUXT__?.data?.relationshipTypes || 'N/A',
        total: window.__NUXT__?.data?.total || 'N/A'
      };
    });

    console.log('Store data:', storeData);

    // Check for JavaScript errors
    const jsErrors = await page.evaluate(() => {
      const errors = [];
      window.addEventListener('error', (e) => {
        errors.push(e.message);
      });
      return errors;
    });

    console.log('JavaScript errors:', jsErrors);

    // Check if there are any system types being filtered out
    const includeSystemCheckbox = page.locator('input[type="checkbox"]');
    const isIncludeSystemChecked = await includeSystemCheckbox.isChecked();
    console.log('Include system checkbox checked:', isIncludeSystemChecked);

    // Try clicking include system checkbox
    if (!isIncludeSystemChecked) {
      await includeSystemCheckbox.click();
      await page.waitForTimeout(2000);
      await page.screenshot({ path: 'relationship-types-including-system.png' });

      // Check if data appears after including system types
      const tableRows = page.locator('table tbody tr');
      const rowCount = await tableRows.count();
      console.log('Table rows after including system types:', rowCount);
    }

    // Check browser console for any errors
    const consoleLogs = [];
    page.on('console', msg => {
      if (msg.type() === 'error' || msg.type() === 'warn') {
        consoleLogs.push({
          type: msg.type(),
          text: msg.text()
        });
      }
    });

    // Final screenshot
    await page.screenshot({ path: 'relationship-types-final-state.png' });

    console.log('Console errors/warnings:', consoleLogs);
  });

  test('check authentication and permissions', async ({ page }) => {
    // Check if user has proper permissions
    await page.goto('http://localhost:3000/relationship-types');

    const userPermissions = await page.evaluate(() => {
      const token = localStorage.getItem('access_token');
      if (!token) return 'No token found';

      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        return payload.permissions || [];
      } catch (e) {
        return 'Error parsing token';
      }
    });

    console.log('User permissions:', userPermissions);

    // Check if relationship_type:read permission is present
    const hasReadPermission = Array.isArray(userPermissions) && userPermissions.includes('relationship_type:read');
    console.log('Has relationship_type:read permission:', hasReadPermission);
  });
});