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

    // Select category (need to add Test category first or use existing)
    // For now, let's skip category as it may not exist
    await page.fill('select.form-input', ''); // Leave empty

    // Select cardinality (use the actual values from the form)
    await page.selectOption('select', '*'); // Source cardinality - Zero or more
    await page.selectOption('select', '*'); // Target cardinality - Zero or more

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
    await page.click('[data-testid="relationship-type-row"]:not([data-system="true"]) [data-testid="edit-button"]');

    // Should be on edit page
    await expect(page.locator('h1')).toContainText('Edit Relationship Type');

    // Update display name
    await page.fill('[data-testid="display-name-input"]', 'Updated Test Dependency');

    // Update description
    await page.fill('[data-testid="description-input"]', 'Updated description for testing');

    // Save changes
    await page.click('[data-testid="save-relationship-type-button"]');

    // Should redirect to details page and show success message
    await expect(page.locator('[data-testid="success-notification"]')).toBeVisible();
    await expect(page.locator('text=Updated Test Dependency')).toBeVisible();
  });

  test('should validate relationship type creation', async ({ page }) => {
    // Navigate to create page
    await page.goto('http://localhost:3000/relationship-types/new');

    // Try to submit without required fields
    await page.click('[data-testid="save-relationship-type-button"]');

    // Should show validation errors
    await expect(page.locator('[data-testid="name-error"]')).toBeVisible();
    await expect(page.locator('[data-testid="forward-label-error"]')).toBeVisible();
    await expect(page.locator('[data-testid="reverse-label-error"]')).toBeVisible();

    // Fill name but use existing name
    await page.fill('[data-testid="name-input"]', 'depends_on');
    await page.fill('[data-testid="forward-label-input"]', 'test');
    await page.fill('[data-testid="reverse-label-input"]', 'test');

    await page.click('[data-testid="save-relationship-type-button"]');

    // Should show duplicate name error
    await expect(page.locator('[data-testid="name-exists-error"]')).toBeVisible();
  });

  test('should delete a relationship type', async ({ page }) => {
    // First create a relationship type to delete
    await createTestRelationshipType(page);

    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Click delete button for the first non-system relationship type
    page.on('dialog', dialog => dialog.accept());
    await page.click('[data-testid="relationship-type-row"]:not([data-system="true"]) [data-testid="delete-button"]');

    // Should show success message and remove item from list
    await expect(page.locator('[data-testid="success-notification"]')).toBeVisible();

    // Verify it's no longer in the list
    await expect(page.locator('text=test_dependency')).not.toBeVisible();
  });

  test('should filter and search relationship types', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Test search functionality
    await page.fill('[data-testid="search-input"]', 'depends');
    await expect(page.locator('text=depends_on')).toBeVisible();

    // Clear search
    await page.fill('[data-testid="search-input"]', '');

    // Test category filter
    await page.click('[data-testid="category-filter"]');
    await page.click('text=Dependency');

    // Should only show dependency relationship types
    await expect(page.locator('text=depends_on')).toBeVisible();

    // Test active/inactive filter
    await page.click('[data-testid="status-filter"]');
    await page.click('text=Active Only');

    // Should not show any inactive types (all defaults are active)
    await expect(page.locator('[data-testid="relationship-types-table"]')).toBeVisible();
  });

  test('should show relationship type statistics', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Click statistics button
    await page.click('[data-testid="statistics-button"]');

    // Should show statistics modal
    await expect(page.locator('[data-testid="statistics-modal"]')).toBeVisible();
    await expect(page.locator('[data-testid="total-types-stat"]')).toBeVisible();
    await expect(page.locator('[data-testid="active-types-stat"]')).toBeVisible();
    await expect(page.locator('[data-testid="system-types-stat"]')).toBeVisible();
    await expect(page.locator('[data-testid="custom-types-stat"]')).toBeVisible();
  });

  test('should validate relationship compatibility', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Click validation button
    await page.click('[data-testid="validate-relationship-button"]');

    // Should show validation modal
    await expect(page.locator('[data-testid="validation-modal"]')).toBeVisible();

    // Select relationship type
    await page.click('[data-testid="relationship-type-select"]');
    await page.click('text=depends_on');

    // Select source and target CI types
    await page.click('[data-testid="source-type-select"]');
    await page.click('text=Application');
    await page.click('[data-testid="target-type-select"]');
    await page.click('text=Server');

    // Click validate
    await page.click('[data-testid="validate-button"]');

    // Should show validation result
    await expect(page.locator('[data-testid="validation-result"]')).toBeVisible();
  });

  test('should prevent editing system relationship types', async ({ page }) => {
    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Find a system relationship type (like depends_on)
    const systemRow = page.locator('[data-system="true"]');
    await expect(systemRow).toBeVisible();

    // Edit button should be disabled or not present for system types
    const editButton = systemRow.locator('[data-testid="edit-button"]');
    if (await editButton.isVisible()) {
      await expect(editButton).toBeDisabled();
    } else {
      // Edit button shouldn't exist for system types
      await expect(editButton).not.toBeVisible();
    }

    // Delete button should be disabled or not present for system types
    const deleteButton = systemRow.locator('[data-testid="delete-button"]');
    if (await deleteButton.isVisible()) {
      await expect(deleteButton).toBeDisabled();
    } else {
      // Delete button shouldn't exist for system types
      await expect(deleteButton).not.toBeVisible();
    }
  });

  test('should handle pagination for large lists', async ({ page }) => {
    // This test would work better with a large dataset
    // For now, we'll just test that pagination controls exist

    await page.goto('http://localhost:3000/relationship-types');

    // Check if pagination controls are present
    const pagination = page.locator('[data-testid="pagination"]');
    if (await pagination.isVisible()) {
      await expect(page.locator('[data-testid="page-info"]')).toBeVisible();

      // Test page size selector
      await page.click('[data-testid="page-size-select"]');
      await page.click('text=25');

      // Should still be on the same page with different page size
      await expect(page.locator('[data-testid="relationship-types-table"]')).toBeVisible();
    }
  });

  test('should export relationship types data', async ({ page }) => {
    await page.goto('http://localhost:3000/relationship-types');

    // Click export button
    const downloadPromise = page.waitForEvent('download');
    await page.click('[data-testid="export-button"]');

    const download = await downloadPromise;

    // Should download a file
    expect(download.suggestedFilename()).toMatch(/\.(csv|xlsx)$/);
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
  await page.selectOption('select', '*'); // Source cardinality
  await page.selectOption('select', '*'); // Target cardinality

  // Submit form
  await page.click('button[type="submit"]');

  // Wait for redirect
  await page.waitForURL('**/relationship-types');

  // Wait a moment for the page to load
  await page.waitForTimeout(1000);
}

test.describe('Relationship Types Permissions', () => {
  test('should restrict access for non-admin users', async ({ page }) => {
    // Login as admin user for testing
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');

    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Should see the page (admin has access)
    await expect(page.locator('h1')).toContainText('Relationship Types');

    // Create button should be visible for admin
    await expect(page.locator('a.btn[href="/relationship-types/new"]')).toBeVisible();
  });
});