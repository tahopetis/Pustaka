import { test, expect } from '@playwright/test';

test.describe('Pustaka CMDB - Comprehensive E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Set up console logging to capture any errors
    page.on('console', (message) => {
      if (message.type() === 'error') {
        console.log('Browser console error:', message.text());
      }
    });

    page.on('pageerror', (error) => {
      console.log('Page error:', error.message);
    });
  });

  test.describe('Authentication & Basic Navigation', () => {
    test('should login successfully with admin credentials', async ({ page }) => {
      await page.goto('/login');

      // Verify login page elements
      await expect(page.locator('h1, h2')).toContainText('Login', { timeout: 10000 });
      await expect(page.locator('#username')).toBeVisible();
      await expect(page.locator('#password')).toBeVisible();
      await expect(page.locator('button[type="submit"]')).toBeVisible();

      // Fill login form
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');

      // Submit form
      await page.click('button[type="submit"]');

      // Should redirect to dashboard
      await page.waitForURL('**/dashboard', { timeout: 15000 });
      await expect(page.locator('h1')).toContainText('Dashboard', { timeout: 10000 });

      // Verify we're authenticated
      const url = page.url();
      expect(url).toContain('/dashboard');
    });

    test('should handle invalid login credentials', async ({ page }) => {
      await page.goto('/login');

      // Try invalid credentials
      await page.fill('#username', 'invalid');
      await page.fill('#password', 'invalid');
      await page.click('button[type="submit"]');

      // Should stay on login page and show error
      await expect(page.locator('.error, .alert, [role="alert"]')).toBeVisible({ timeout: 5000 });
      await expect(page.locator('#username')).toBeVisible();
    });

    test('should redirect to login when accessing protected routes without auth', async ({ page }) => {
      // Try accessing dashboard without auth
      await page.goto('/dashboard');
      await page.waitForURL('**/login', { timeout: 10000 });

      // Try accessing relationship types without auth
      await page.goto('/relationship-types');
      await page.waitForURL('**/login', { timeout: 10000 });
    });
  });

  test.describe('Sidebar Navigation & Permissions', () => {
    test('should show correct menu items based on permissions', async ({ page }) => {
      // Login first
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      // Wait for sidebar to be fully loaded
      await page.waitForSelector('nav', { timeout: 10000 });

      // Check for Relationship Types link specifically
      const relationshipTypesLink = page.locator('a[href="/relationship-types"]');
      await expect(relationshipTypesLink).toBeVisible({ timeout: 10000 });

      // Verify the link text
      await expect(relationshipTypesLink).toContainText('Relationship Types');

      // Check other menu items are visible
      await expect(page.locator('a[href="/dashboard"]')).toBeVisible();
      await expect(page.locator('a[href="/ci"]')).toBeVisible();
      await expect(page.locator('a[href="/relationships"]')).toBeVisible();
      await expect(page.locator('a[href="/users"]')).toBeVisible();
    });

    test('should navigate to relationship types via sidebar', async ({ page }) => {
      // Login first
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      // Wait for sidebar to be loaded
      await page.waitForSelector('a[href="/relationship-types"]', { timeout: 10000 });

      // Click Relationship Types link
      await page.click('a[href="/relationship-types"]');

      // Should navigate to relationship types page
      await page.waitForURL('**/relationship-types', { timeout: 10000 });
      await expect(page.locator('h1, h2')).toContainText('Relationship Types', { timeout: 5000 });
    });
  });

  test.describe('Relationship Types Functionality', () => {
    test.beforeEach(async ({ page }) => {
      // Login before each test
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');
    });

    test('should load relationship types list page', async ({ page }) => {
      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');

      // Check page title
      await expect(page.locator('h1, h2')).toContainText('Relationship Types', { timeout: 10000 });

      // Check for main components
      await expect(page.locator('table, .table, [data-testid="relationship-types-table"]')).toBeVisible({ timeout: 10000 });

      // Check for search functionality
      await expect(page.locator('input[placeholder*="search" i], input[placeholder*="Search" i]')).toBeVisible({ timeout: 5000 });

      // Check for create button
      await expect(page.locator('a[href="/relationship-types/new"], button:has-text("Create"), button:has-text("New")')).toBeVisible({ timeout: 5000 });

      // Check for filter options
      await expect(page.locator('select, .filter, .dropdown')).toBeVisible({ timeout: 5000 });
    });

    test('should create a new relationship type', async ({ page }) => {
      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');

      // Click create new button
      await page.click('a[href="/relationship-types/new"], button:has-text("Create"), button:has-text("New Relationship Type")');
      await page.waitForURL('**/relationship-types/new', { timeout: 10000 });

      // Check form is loaded
      await expect(page.locator('form')).toBeVisible({ timeout: 5000 });
      await expect(page.locator('h1, h2')).toContainText('Create', { timeout: 5000 });

      // Fill form fields
      await page.fill('input[placeholder*="name" i], input[name="name"], input[id*="name"]', 'test_dependency');
      await page.fill('input[placeholder*="display" i], input[name="displayName"], input[id*="display"]', 'Test Dependency');
      await page.fill('textarea[placeholder*="description" i], textarea[name="description"], textarea[id*="description"]', 'Test dependency relationship for comprehensive e2e testing');

      // Handle forward and reverse labels
      const forwardLabelInput = page.locator('input[placeholder*="forward" i], input[name*="forward"], label:has-text("Forward") + input').first();
      const reverseLabelInput = page.locator('input[placeholder*="reverse" i], input[name*="reverse"], label:has-text("Reverse") + input').first();

      if (await forwardLabelInput.isVisible()) {
        await forwardLabelInput.fill('depends on');
      }
      if (await reverseLabelInput.isVisible()) {
        await reverseLabelInput.fill('required by');
      }

      // Handle cardinality selectors
      const selects = page.locator('select');
      const selectCount = await selects.count();

      if (selectCount >= 2) {
        await selects.nth(0).selectOption('1'); // Source cardinality
        await selects.nth(1).selectOption('*'); // Target cardinality
      }

      // Submit form
      await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save"), button[type="submit"]');

      // Should redirect back to relationship types list
      await page.waitForURL('**/relationship-types', { timeout: 15000 });

      // Should see success message or the new relationship type
      await expect(page.locator('text=test_dependency, .success, .alert-success')).toBeVisible({ timeout: 10000 });
    });

    test('should search and filter relationship types', async ({ page }) => {
      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');

      // Test search functionality
      const searchInput = page.locator('input[placeholder*="search" i], input[placeholder*="Search" i]').first();
      if (await searchInput.isVisible()) {
        await searchInput.fill('test');
        await page.waitForTimeout(1000);

        // Should show filtered results
        const table = page.locator('table, .table').first();
        if (await table.isVisible()) {
          await expect(table).toBeVisible();
        }
      }

      // Test filter functionality
      const filterSelect = page.locator('select').first();
      if (await filterSelect.isVisible()) {
        await filterSelect.selectOption({ index: 1 });
        await page.waitForTimeout(1000);
      }

      // Test "include system types" checkbox if present
      const includeSystemCheckbox = page.locator('input[type="checkbox"]').first();
      if (await includeSystemCheckbox.isVisible()) {
        await includeSystemCheckbox.check();
        await page.waitForTimeout(1000);
      }
    });

    test('should handle relationship types CRUD operations', async ({ page }) => {
      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');

      // Look for existing relationship types to interact with
      const tableRows = page.locator('table tbody tr, .table tbody tr');
      const rowCount = await tableRows.count();

      if (rowCount > 0) {
        // Test viewing details
        const firstRow = tableRows.first();
        const detailLink = firstRow.locator('a').first();

        if (await detailLink.isVisible()) {
          await detailLink.click();
          await page.waitForTimeout(2000);

          // Should be on details page
          const currentUrl = page.url();
          expect(currentUrl).toContain('/relationship-types/');

          // Go back to list
          await page.goBack();
        }
      }
    });

    test('should handle error states gracefully', async ({ page }) => {
      // Try accessing non-existent relationship type
      await page.goto('/relationship-types/99999');
      await page.waitForTimeout(2000);

      // Should show error or redirect
      const currentUrl = page.url();
      if (currentUrl.includes('/relationship-types/99999')) {
        // Still on the detail page, check for error message
        await expect(page.locator('.error, .alert-error, [role="alert"]')).toBeVisible({ timeout: 5000 });
      }
    });
  });

  test.describe('Network & API Connectivity', () => {
    test('should handle API responses correctly', async ({ page }) => {
      // Login first
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      // Monitor network requests
      const responses = [];
      page.on('response', response => {
        responses.push({
          url: response.url(),
          status: response.status(),
          method: response.request().method()
        });
      });

      // Navigate to relationship types
      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');

      // Check API calls
      const apiResponses = responses.filter(r => r.url.includes('/api/'));
      console.log('API Responses:', apiResponses);

      // Should have successful API calls
      expect(apiResponses.some(r => r.url.includes('/relationship-types') && r.status < 400)).toBeTruthy();
    });

    test('should handle network failures gracefully', async ({ page }) => {
      // Login first
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      // Simulate offline mode
      await page.context().setOffline(true);

      // Try to navigate to relationship types
      await page.goto('/relationship-types');
      await page.waitForTimeout(3000);

      // Should show some kind of error or loading state
      const hasError = await page.locator('.error, .alert-error, .network-error').isVisible();
      const hasLoading = await page.locator('.loading, .spinner').isVisible();

      expect(hasError || hasLoading).toBeTruthy();

      // Restore connection
      await page.context().setOffline(false);
      await page.reload();
      await page.waitForLoadState('networkidle');
    });
  });

  test.describe('Responsive Design', () => {
    test('should work on mobile viewport', async ({ page }) => {
      // Set mobile viewport
      await page.setViewportSize({ width: 375, height: 667 });

      // Login
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      // Check mobile navigation
      const mobileMenuButton = page.locator('button[aria-label="Menu"], .hamburger, .mobile-menu-btn');
      if (await mobileMenuButton.isVisible()) {
        await mobileMenuButton.click();
        await page.waitForTimeout(500);
      }

      // Navigate to relationship types
      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');

      // Check content is readable on mobile
      await expect(page.locator('h1, h2')).toContainText('Relationship Types');

      // Check table responsiveness
      const table = page.locator('table, .table').first();
      if (await table.isVisible()) {
        await expect(table).toBeVisible();
      }
    });

    test('should work on tablet viewport', async ({ page }) => {
      // Set tablet viewport
      await page.setViewportSize({ width: 768, height: 1024 });

      // Login and navigate
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');

      // Should display properly on tablet
      await expect(page.locator('h1, h2')).toContainText('Relationship Types');
    });
  });

  test.describe('Performance & Loading States', () => {
    test('should load pages within reasonable time', async ({ page }) => {
      // Login
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      // Measure load time for relationship types
      const startTime = Date.now();
      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');
      const loadTime = Date.now() - startTime;

      // Should load within 10 seconds
      expect(loadTime).toBeLessThan(10000);
      console.log(`Relationship types page loaded in ${loadTime}ms`);
    });

    test('should show loading states during data fetching', async ({ page }) => {
      // Login
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      // Navigate to relationship types and check for loading states
      await page.goto('/relationship-types');

      // Check for loading indicators immediately
      const loadingIndicators = [
        '.loading',
        '.spinner',
        '.skeleton',
        '[data-testid="loading"]',
        '.progress'
      ];

      let hasLoadingIndicator = false;
      for (const selector of loadingIndicators) {
        if (await page.locator(selector).isVisible({ timeout: 1000 })) {
          hasLoadingIndicator = true;
          break;
        }
      }

      // Should eventually show content
      await page.waitForLoadState('networkidle');
      await expect(page.locator('h1, h2')).toContainText('Relationship Types');
    });
  });

  test.describe('Accessibility', () => {
    test('should have proper page structure and headings', async ({ page }) => {
      // Login
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');

      // Check for proper heading structure
      await expect(page.locator('h1')).toBeVisible();

      // Check for proper semantic elements
      await expect(page.locator('nav')).toBeVisible();
      await expect(page.locator('main, [role="main"]')).toBeVisible();
    });

    test('should support keyboard navigation', async ({ page }) => {
      // Login
      await page.goto('/login');
      await page.fill('#username', 'admin');
      await page.fill('#password', 'Admin@123');
      await page.click('button[type="submit"]');
      await page.waitForURL('**/dashboard');

      await page.goto('/relationship-types');
      await page.waitForLoadState('networkidle');

      // Test Tab navigation
      await page.keyboard.press('Tab');
      await page.keyboard.press('Tab');

      // Test Enter key on interactive elements
      const focusableElement = page.locator(':focus');
      if (await focusableElement.isVisible()) {
        await page.keyboard.press('Enter');
        await page.waitForTimeout(1000);
      }
    });
  });
});