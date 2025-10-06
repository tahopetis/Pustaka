import { test, expect } from '@playwright/test';

test.describe('Relationship Types - Complete Workflow Test', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');
    await page.waitForTimeout(1000);
  });

  test('should access relationship types via sidebar navigation', async ({ page }) => {
    // Click Relationship Types link in sidebar
    await page.click('a[href="/relationship-types"]');

    // Should navigate to relationship types page
    await page.waitForURL('**/relationship-types');
    await expect(page.locator('h1')).toContainText('Relationship Types');

    // Should see the relationship types table
    await expect(page.locator('.table')).toBeVisible();
  });

  test('should create a new relationship type', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');
    await page.waitForLoadState('networkidle');

    // Click create new button
    await page.click('a.btn[href="/relationship-types/new"]');
    await page.waitForURL('**/relationship-types/new');

    // Fill form fields
    await page.fill('input[placeholder="e.g., depends_on, runs_on, managed_by"]', 'test_depends');
    await page.fill('input[placeholder="e.g., Depends On, Runs On, Managed By"]', 'Test Depends');
    await page.fill('textarea[placeholder="Describe when this relationship type should be used..."]', 'Test depends relationship for e2e testing');

    // Find and fill forward and reverse labels
    const inputs = await page.locator('input.form-input').all();
    if (inputs.length >= 4) {
      await inputs[2].fill('depends on'); // Forward label
      await inputs[3].fill('required by'); // Reverse label
    }

    // Select cardinality
    await page.selectOption('select', '*'); // Source cardinality
    await page.selectOption('select', '*'); // Target cardinality

    // Submit form
    await page.click('button[type="submit"]');

    // Should redirect back to relationship types list
    await page.waitForURL('**/relationship-types');

    // Should see the new relationship type
    await expect(page.locator('text=test_depends')).toBeVisible();
  });

  test('should show system relationship types when including system types', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');
    await page.waitForLoadState('networkidle');

    // Should see default system relationship types when including system types
    await page.click('input[type="checkbox"]'); // Include system types
    await expect(page.locator('text=depends_on')).toBeVisible();
    await expect(page.locator('text=hosts')).toBeVisible();
  });

  test('should verify relationship types management features', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');
    await page.waitForLoadState('networkidle');

    // Should have search functionality
    await expect(page.locator('input[placeholder*="Search"]')).toBeVisible();

    // Should have category filter
    await expect(page.locator('select')).toBeVisible();

    // Should have create button
    await expect(page.locator('a.btn[href="/relationship-types/new"]')).toBeVisible();

    // Should have table headers
    await expect(page.locator('th')).toBeVisible();
  });
});