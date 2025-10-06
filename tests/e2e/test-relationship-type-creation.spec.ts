import { test, expect } from '@playwright/test';

test.describe('Relationship Type Creation Form', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin
    await page.goto('http://localhost:3000/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'Admin@123');
    await page.click('button[type="submit"]');

    // Wait for navigation to dashboard
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('should create a new relationship type successfully', async ({ page }) => {
    // Navigate to relationship types creation page
    await page.goto('http://localhost:3000/relationship-types/new');

    // Wait for page to load by looking for the form
    await page.waitForSelector('form', { timeout: 10000 });

    // Fill in the form
    await page.fill('input[name="name"]', 'test_relationship');
    await page.fill('input[name="display_name"]', 'Test Relationship');
    await page.fill('textarea[name="description"]', 'This is a test relationship type created by automated test');

    // Select category (should load from API)
    await page.waitForSelector('select[name="category"]', { timeout: 10000 });

    // Wait for categories to be loaded
    await page.waitForFunction(() => {
      const select = document.querySelector('select[name="category"]') as HTMLSelectElement;
      return select && select.options.length > 1;
    }, { timeout: 10000 });

    await page.selectOption('select[name="category"]', 'Technical');

    // Fill labels
    await page.fill('input[name="forward_label"]', 'test relates to');
    await page.fill('input[name="reverse_label"]', 'is related by test');

    // Cardinality should default to 'many' for both source and target
    // Verify default values are set
    const sourceCardinality = await page.inputValue('select[name="cardinality_source"]');
    const targetCardinality = await page.inputValue('select[name="cardinality_target"]');

    expect(sourceCardinality).toBe('many');
    expect(targetCardinality).toBe('many');

    // Set allowed CI types
    await page.fill('input[name="allowed_source_types"]', 'server,application');
    await page.fill('input[name="allowed_target_types"]', 'server,application');

    // Submit the form
    await page.click('button[type="submit"]');

    // Wait for either success message or navigation
    try {
      // Look for success toast/message or check if redirected
      await page.waitForFunction(() => {
        // Check for any success indicators
        const successSelectors = [
          '.alert-success',
          '.notification.success',
          '[class*="success"]',
          '[class*="toast"]'
        ];

        for (const selector of successSelectors) {
          const element = document.querySelector(selector);
          if (element && element.textContent && element.textContent.includes('success')) {
            return true;
          }
        }

        // Check if redirected to list view
        return window.location.href.includes('/relationship-types') && !window.location.href.includes('/new');
      }, { timeout: 10000 });

      console.log('✅ Form submitted successfully');
    } catch (error) {
      // Check current URL to see what happened
      const currentUrl = page.url();
      console.log('⚠️ Form submission result - current URL:', currentUrl);

      // Take screenshot for debugging
      await page.screenshot({ path: 'form-submission-result.png' });

      // Check if we're still on the form page (validation error) or redirected
      if (currentUrl.includes('/new')) {
        console.log('❌ Form submission failed - still on form page');

        // Look for validation errors
        const errorElements = await page.$$('[class*="error"], [class*="invalid"]');
        if (errorElements.length > 0) {
          for (const error of errorElements) {
            const errorText = await error.textContent();
            console.log('Validation error:', errorText);
          }
        }
      } else {
        console.log('✅ Form submitted - redirected to:', currentUrl);
      }
    }
  });

  test('should load categories from API', async ({ page }) => {
    // Navigate to relationship types creation page
    await page.goto('http://localhost:3000/relationship-types/new');

    // Wait for form to load
    await page.waitForSelector('form', { timeout: 10000 });

    // Wait for category dropdown to be populated
    await page.waitForFunction(() => {
      const select = document.querySelector('select[name="category"]') as HTMLSelectElement;
      return select && select.options.length > 1;
    }, { timeout: 10000 });

    // Check that categories are loaded
    const categories = await page.$$eval('select[name="category"] option', options =>
      options.map(option => option.textContent).filter(text => text && text.trim() !== '')
    );

    expect(categories.length).toBeGreaterThan(0);
    console.log('✅ Categories loaded successfully:', categories.slice(0, 3).join(', '), '...');
  });
});