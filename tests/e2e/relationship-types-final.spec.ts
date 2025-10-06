import { test, expect } from '@playwright/test';

test.describe('Relationship Types - Final Test', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Wait a moment for auth store to fully initialize
    await page.waitForTimeout(1000);
  });

  test('should access relationship types page successfully', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Check we're on the correct page (not redirected to dashboard)
    expect(page.url()).toContain('/relationship-types');
    expect(page.url()).not.toContain('/dashboard');

    // Check page title
    await expect(page.locator('h1')).toContainText('Relationship Types');

    // Check for relationship types table
    await expect(page.locator('.table')).toBeVisible();

    // Should see system relationship types when including system types
    await page.click('input[type="checkbox"]'); // Include system types
    await expect(page.locator('text=depends_on')).toBeVisible();
    await expect(page.locator('text=hosts')).toBeVisible();
  });

  test('should create new relationship type successfully', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');
    await page.waitForLoadState('networkidle');

    // Click create new button
    await page.click('a[href="/relationship-types/new"]');
    await page.waitForURL('**/relationship-types/new');

    // Fill form fields
    await page.fill('input[placeholder="e.g., depends_on, runs_on, managed_by"]', 'test_dependency');
    await page.fill('input[placeholder="e.g., Depends On, Runs On, Managed By"]', 'Test Dependency');
    await page.fill('textarea[placeholder="Describe when this relationship type should be used..."]', 'Test dependency for e2e testing');
    await page.fill('input[name="forward_label"]', 'depends on');
    await page.fill('input[name="reverse_label"]', 'required by');

    // Select cardinality (zero or more for both)
    await page.selectOption('select[name="source_cardinality"]', '*');
    await page.selectOption('select[name="target_cardinality"]', '*');

    // Submit form
    await page.click('button[type="submit"]');

    // Should redirect back to relationship types list
    await page.waitForURL('**/relationship-types');

    // Should see the new relationship type
    await expect(page.locator('text=test_dependency')).toBeVisible();
  });

  test('should show relationship types link in sidebar', async ({ page }) => {
    // Navigate to dashboard first
    await page.goto('http://localhost:3000/dashboard');

    // Check if Relationship Types link is visible in sidebar
    const relationshipTypesLink = page.locator('a[href="/relationship-types"]');
    await expect(relationshipTypesLink).toBeVisible();

    // Click the link
    await relationshipTypesLink.click();

    // Should navigate to relationship types page
    await page.waitForURL('**/relationship-types');
    await expect(page.locator('h1')).toContainText('Relationship Types');
  });
});