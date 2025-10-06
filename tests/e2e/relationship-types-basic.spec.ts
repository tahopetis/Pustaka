import { test, expect } from '@playwright/test';

test.describe('Relationship Types Basic Functionality', () => {
  test('should access relationship types page', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Check page loads by looking for specific text
    await expect(page.locator('h1.page-title')).toHaveText('Relationship Types');
    await expect(page.locator('p.page-subtitle')).toHaveText('Manage relationship types for configuration items');
  });

  test('should show create relationship type button', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Check for create button
    await expect(page.locator('a.btn:has-text("Add Relationship Type")')).toBeVisible();
  });

  test('should navigate to create relationship type page', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Navigate to relationship types
    await page.goto('http://localhost:3000/relationship-types');

    // Click create button
    await page.click('a.btn:has-text("Add Relationship Type")');
    await page.waitForURL('**/relationship-types/new');

    // Check we're on the create page
    await expect(page.locator('h1')).toContainText('Create Relationship Type');
    await expect(page.locator('p.text-gray-600')).toContainText('Define a new type of relationship between configuration items');
  });

  test('should show relationship type form fields', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Go to create page directly
    await page.goto('http://localhost:3000/relationship-types/new');

    // Check for form fields
    await expect(page.locator('label:has-text("Name *")')).toBeVisible();
    await expect(page.locator('label:has-text("Display Name")')).toBeVisible();
    await expect(page.locator('label:has-text("Forward Label *")')).toBeVisible();
    await expect(page.locator('label:has-text("Reverse Label *")')).toBeVisible();
    await expect(page.locator('label:has-text("Category")')).toBeVisible();
    await expect(page.locator('label:has-text("Description")')).toBeVisible();
    await expect(page.locator('label:has-text("Source Cardinality")')).toBeVisible();
    await expect(page.locator('label:has-text("Target Cardinality")')).toBeVisible();
  });

  test('should fill and submit relationship type form', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Go to create page
    await page.goto('http://localhost:3000/relationship-types/new');

    // Fill the name field
    const nameInput = page.locator('input[placeholder="e.g., depends_on, runs_on, managed_by"]');
    await nameInput.fill('test_depends_on');

    // Fill the forward label field
    const forwardLabelInput = page.locator('input[placeholder="e.g., Depends on, Runs on, Managed by"]');
    await forwardLabelInput.fill('depends on');

    // Fill the reverse label field
    const reverseLabelInput = page.locator('input[placeholder*=""]');
    // Find the reverse label input by looking for inputs after the forward label
    const inputs = await page.locator('input.form-input').all();
    if (inputs.length >= 4) {
      await inputs[3].fill('required by');
    }

    // Submit the form
    await page.click('button[type="submit"]');

    // Should redirect back to relationship types list
    await page.waitForURL('**/relationship-types');

    // Should show the new relationship type
    await expect(page.locator('text=test_depends_on')).toBeVisible();
  });
});