import { test, expect } from '@playwright/test';

test.describe('Relationship Type Creation Form Debug', () => {
  test('should debug form submission issues', async ({ page }) => {
    console.log('🔍 Starting relationship type form debug test...');

    // Navigate to login
    await page.goto('http://localhost:3000');
    await page.waitForURL('**/login');

    // Login with admin credentials
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'Admin@123');
    await page.click('button[type="submit"]');

    // Wait for dashboard
    await page.waitForURL('**/dashboard');
    console.log('✅ Login completed');

    // Navigate to relationship types new form
    await page.click('a[href="/relationship-types"]');
    await page.waitForURL('**/relationship-types');
    console.log('✅ Navigated to relationship types list');

    // Click "Add Relationship Type" button
    await page.click('a[href="/relationship-types/new"]');
    await page.waitForURL('**/relationship-types/new');
    console.log('✅ Navigated to relationship type creation form');

    // Take screenshot before form fill
    await page.screenshot({ path: 'debug-form-before-fill.png', fullPage: true });

    // Fill in the form with test data
    console.log('📝 Filling form fields...');

    // Wait for form to be ready
    await page.waitForSelector('form', { state: 'visible' });

    // Fill basic fields
    await page.fill('input[placeholder="e.g., scope, hosts, supports, managed_by"]', 'test_relationship');
    await page.fill('input[placeholder="e.g., Scope, Hosts, Supports, Managed By"]', 'Test Relationship');
    await page.fill('textarea[placeholder="Describe when this relationship type should be used..."]', 'Test relationship for debugging form submission');

    // Fill forward and reverse labels
    await page.fill('input[placeholder="e.g., scopes, hosts, supports, managed_by"]', 'relates to');
    await page.fill('input[placeholder="e.g., scoped by, hosted by, supported by, manages"]', 'related by');

    // Fill color
    await page.fill('input[placeholder="#3B82F6"]', '#ff5733');

    // Fill icon
    await page.fill('input[placeholder="e.g., arrow-right, link, database"]', 'link');

    // Fill category (let me look for this field)
    await page.fill('input[placeholder*="Category"]', 'Test Category');

    // Take screenshot after form fill
    await page.screenshot({ path: 'debug-form-after-fill.png', fullPage: true });

    // Console log monitoring
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      console.log(`🖥️  Console ${msg.type()}: ${msg.text()}`);
      consoleMessages.push(msg.text());
    });

    // Network request monitoring
    const apiRequests: string[] = [];
    page.on('request', request => {
      if (request.url().includes('/api/v1/')) {
        console.log(`🌐 API Request: ${request.method()} ${request.url()}`);
        apiRequests.push(`${request.method()} ${request.url()}`);
      }
    });

    page.on('response', async response => {
      if (response.url().includes('/api/v1/')) {
        const status = response.status();
        const body = await response.text();
        console.log(`🌐 API Response: ${status} ${response.url()}`);
        if (body) {
          console.log(`📄 Response body: ${body.substring(0, 500)}...`);
        }
      }
    });

    // Check if submit button exists and is enabled
    const submitButton = page.locator('button[type="submit"]').first();
    const isVisible = await submitButton.isVisible();
    const isEnabled = await submitButton.isEnabled();

    console.log(`🔘 Submit button visible: ${isVisible}`);
    console.log(`🔘 Submit button enabled: ${isEnabled}`);

    if (!isVisible) {
      console.log('❌ Submit button not found or not visible');
      await page.screenshot({ path: 'debug-no-submit-button.png', fullPage: true });
      return;
    }

    // Try to submit the form
    console.log('🚀 Attempting to submit form...');

    // Take screenshot just before submission
    await page.screenshot({ path: 'debug-before-submit.png', fullPage: true });

    // Click submit button
    await Promise.all([
      page.waitForNavigation({ timeout: 10000 }).catch(() => console.log('⚠️  Navigation timeout - form submission may have failed')),
      submitButton.click()
    ]);

    // Wait a bit to see what happens
    await page.waitForTimeout(3000);

    // Take screenshot after submission attempt
    await page.screenshot({ path: 'debug-after-submit.png', fullPage: true });

    // Check current URL
    const currentUrl = page.url();
    console.log(`📍 Current URL after submission: ${currentUrl}`);

    // Check for any error messages
    const errorMessages = await page.locator('.error, .alert-error, .text-red-600, [role="alert"]').all();
    console.log(`📊 Found ${errorMessages.length} potential error elements`);

    for (let i = 0; i < errorMessages.length; i++) {
      const text = await errorMessages[i].textContent();
      console.log(`❌ Error message ${i + 1}: ${text}`);
    }

    // Check for success messages
    const successMessages = await page.locator('.success, .alert-success, .text-green-600').all();
    console.log(`📊 Found ${successMessages.length} potential success elements`);

    for (let i = 0; i < successMessages.length; i++) {
      const text = await successMessages[i].textContent();
      console.log(`✅ Success message ${i + 1}: ${text}`);
    }

    // Log form field values for debugging
    const formFields = {
      name: await page.locator('input[placeholder="e.g., scope, hosts, supports, managed_by"]').inputValue(),
      displayName: await page.locator('input[placeholder="e.g., Scope, Hosts, Supports, Managed By"]').inputValue(),
      description: await page.locator('textarea[placeholder="Describe when this relationship type should be used..."]').inputValue(),
      forwardLabel: await page.locator('input[placeholder="e.g., scopes, hosts, supports, managed_by"]').inputValue(),
      reverseLabel: await page.locator('input[placeholder="e.g., scoped by, hosted by, supported by, manages"]').inputValue(),
      color: await page.locator('input[placeholder="#3B82F6"]').inputValue(),
      icon: await page.locator('input[placeholder="e.g., arrow-right, link, database"]').inputValue(),
      category: await page.locator('input[placeholder*="Category"]').inputValue().catch(() => 'Not found'),
    };

    console.log('📋 Form field values:', formFields);

    // Console and API summary
    console.log(`🖥️  Total console messages: ${consoleMessages.length}`);
    console.log(`🌐 Total API requests: ${apiRequests.length}`);

    if (apiRequests.length === 0) {
      console.log('⚠️  No API requests detected - form submission may not be working');
    }

    // Check if we're still on the form page
    if (currentUrl.includes('/relationship-types/new')) {
      console.log('❌ Still on form page - submission likely failed');

      // Check form validation
      const formValidation = await page.locator('[required], [aria-required="true"]').all();
      console.log(`📋 Found ${formValidation.length} required fields`);

      // Check for any validation messages
      const validationMessages = await page.locator('.validation-message, .invalid-feedback, .text-red-500').all();
      console.log(`📋 Found ${validationMessages.length} validation messages`);

      for (let i = 0; i < validationMessages.length; i++) {
        const text = await validationMessages[i].textContent();
        console.log(`⚠️  Validation message ${i + 1}: ${text}`);
      }
    } else {
      console.log('✅ Form submission appears to have succeeded');
    }

    console.log('🔍 Debug test completed');
  });
});