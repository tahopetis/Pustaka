import { test, expect } from '@playwright/test';

test.describe('Relationship Types Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');
  });

  test('should display relationship types list page', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Check page loads
    await expect(page.locator('h1')).toContainText('Relationship Types');

    // Check that the table is visible
    await expect(page.locator('.table')).toBeVisible();

    // Should see default system relationship types (include system types to see them)
    await page.click('input[type="checkbox"]'); // Include system types
    await expect(page.locator('text=depends_on')).toBeVisible();
    await expect(page.locator('text=hosts')).toBeVisible();
  });

  test('should create a new relationship type', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Click create new button
    await page.click('a.btn[href="/relationship-types/new"]');
    await page.waitForURL('**/relationship-types/new');

    // Fill form - Name field
    await page.fill('input[placeholder="e.g., depends_on, runs_on, managed_by"]', 'test_dependency');

    // Fill form - Display Name field
    await page.fill('input[placeholder="e.g., Depends On, Runs On, Managed By"]', 'Test Dependency');

    // Fill form - Description field
    await page.fill('textarea[placeholder="Describe when this relationship type should be used..."]', 'Test dependency relationship for e2e testing');

    // Fill form - Forward Label field
    await page.fill('input[placeholder="e.g., Depends on, Runs on, Managed by"]', 'depends on');

    // Fill form - Reverse Label field
    await page.fill('input[placeholder="e.g., Depended by, Runs, Manages"]', 'required by');

    // Select cardinality (use the actual values from the form)
    await page.selectOption('select:has-text("No restriction")', '*'); // Source cardinality - Zero or more
    await page.selectOption('select:has-text("No restriction")', '*'); // Target cardinality - Zero or more

    // Submit form
    await page.click('button[type="submit"]');

    // Should redirect to list
    await page.waitForURL('**/relationship-types');

    // Should see the new relationship type in the list
    await expect(page.locator('text=test_dependency')).toBeVisible();
  });

  test('should view relationship type details', async ({ page }) => {
    // First create a relationship type to view
    await createTestRelationshipType(page);

    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Click on the first relationship type view link
    await page.click('table a[href^="/relationship-types/"]:first-child');

    // Should show details page
    await expect(page.locator('h1')).toContainText('Relationship Type Details');
  });

  test('should edit an existing relationship type', async ({ page }) => {
    // First create a relationship type to edit
    await createTestRelationshipType(page);

    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Click edit button for the first non-system relationship type
    await page.click('table a[href*="/edit"]:first-child');

    // Should be on edit page
    await expect(page.locator('h1')).toContainText('Edit Relationship Type');

    // Update display name
    await page.fill('input[placeholder="e.g., Depends On, Runs On, Managed By"]', 'Updated Test Dependency');

    // Update description
    await page.fill('textarea[placeholder="Describe when this relationship type should be used..."]', 'Updated description for testing');

    // Save changes
    await page.click('button[type="submit"]');

    // Should redirect to details page
    await expect(page.locator('text=Updated Test Dependency')).toBeVisible();
  });

  test('should delete a relationship type', async ({ page }) => {
    // First create a relationship type to delete
    await createTestRelationshipType(page);

    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Click delete button for the first non-system relationship type
    page.on('dialog', dialog => dialog.accept());
    await page.click('table button:has-text("Delete"):first-child');

    // Should show success message and remove item from list
    // Verify it's no longer in the list
    await expect(page.locator('text=test_dependency')).not.toBeVisible();
  });

  test('should filter and search relationship types', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Test search functionality
    await page.fill('input[placeholder*="Search"]', 'depends');
    await expect(page.locator('text=depends_on')).toBeVisible();

    // Clear search
    await page.fill('input[placeholder*="Search"]', '');

    // Test category filter
    await page.selectOption('select', 'Dependency');

    // Should only show dependency relationship types
    await expect(page.locator('text=depends_on')).toBeVisible();

    // Test active/inactive filter
    await page.selectOption('select:has-text("All Types")', 'Active');

    // Should not show any inactive types (all defaults are active)
    await expect(page.locator('.table')).toBeVisible();
  });

  test('should prevent editing system relationship types', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Include system types to see them
    await page.click('input[type="checkbox"]');

    // Find a system relationship type (like depends_on)
    const systemRow = page.locator('text=System').first();
    await expect(systemRow).toBeVisible();

    // Edit button should not be present for system types
    const editButton = systemRow.locator('a[href*="/edit"]');
    await expect(editButton).not.toBeVisible();

    // Delete button should not be present for system types
    const deleteButton = systemRow.locator('button:has-text("Delete")');
    await expect(deleteButton).not.toBeVisible();
  });
});

// Helper function to create a test relationship type
async function createTestRelationshipType(page) {
  await page.goto('http://localhost:3000/relationship-types/new');

  // Fill form with test data
  await page.fill('input[placeholder="e.g., depends_on, runs_on, managed_by"]', 'test_dependency');
  await page.fill('input[placeholder="e.g., Depends On, Runs On, Managed By"]', 'Test Dependency');
  await page.fill('textarea[placeholder="Describe when this relationship type should be used..."]', 'Test dependency relationship for testing');
  await page.fill('input[placeholder="e.g., Depends on, Runs on, Managed by"]', 'depends on');
  await page.fill('input[placeholder="e.g., Depended by, Runs, Manages"]', 'required by');

  // Select cardinality
  await page.selectOption('select:has-text("No restriction")', '*'); // Source cardinality
  await page.selectOption('select:has-text("No restriction")', '*'); // Target cardinality

  // Submit form
  await page.click('button[type="submit"]');

  // Wait for redirect
  await page.waitForURL('**/relationship-types');

  // Wait a moment for the page to load
  await page.waitForTimeout(1000);
}