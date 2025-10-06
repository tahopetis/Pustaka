import { test, expect } from '@playwright/test';

test.describe('Simple Form Debug', () => {
  test('should debug form submission step by step', async ({ page }) => {
    console.log('🔍 Starting simple form debug...');

    // Set up console monitoring
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      console.log(`🖥️  Console ${msg.type()}: ${msg.text()}`);
      consoleMessages.push(msg.text());
    });

    // Set up network monitoring
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
        console.log(`🌐 API Response: ${status} ${response.url()}`);
        if (status >= 400) {
          const body = await response.text();
          console.log(`❌ Error response: ${body}`);
        }
      }
    });

    // Navigate to form
    await page.goto('http://localhost:3000');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    await page.goto('http://localhost:3000/relationship-types/new');
    await page.waitForSelector('form', { state: 'visible' });

    console.log('✅ Form loaded');

    // Check if form elements exist
    const formSelectors = [
      'input[placeholder="e.g., scope, hosts, supports, managed_by"]',
      'input[placeholder="e.g., Scope, Hosts, Supports, Managed By"]',
      'textarea[placeholder="Describe when this relationship type should be used..."]',
      'input[placeholder="e.g., scopes, hosts, supports, managed_by"]',
      'input[placeholder="e.g., scoped by, hosted by, supported by, manages"]',
      'select',
      'button[type="submit"]'
    ];

    for (const selector of formSelectors) {
      const exists = await page.locator(selector).isVisible().catch(() => false);
      console.log(`🔍 Element ${selector}: ${exists ? '✅' : '❌'}`);
    }

    // Try minimal form fill
    console.log('📝 Filling minimal required fields...');

    try {
      await page.fill('input[placeholder="e.g., scope, hosts, supports, managed_by"]', 'test_relationship');
      console.log('✅ Name field filled');
    } catch (error) {
      console.log('❌ Failed to fill name field:', error);
      return;
    }

    try {
      await page.fill('input[placeholder="e.g., scopes, hosts, supports, managed_by"]', 'relates to');
      console.log('✅ Forward label filled');
    } catch (error) {
      console.log('❌ Failed to fill forward label:', error);
      return;
    }

    try {
      await page.fill('input[placeholder="e.g., scoped by, hosted by, supported by, manages"]', 'related by');
      console.log('✅ Reverse label filled');
    } catch (error) {
      console.log('❌ Failed to fill reverse label:', error);
      return;
    }

    // Check submit button
    const submitButton = page.locator('button[type="submit"]');
    const isDisabled = await submitButton.isDisabled();
    console.log(`🔘 Submit button disabled: ${isDisabled}`);

    if (isDisabled) {
      console.log('⚠️  Submit button is disabled - checking form validation');

      // Check for required fields
      const requiredFields = await page.locator('[required]').all();
      console.log(`📋 Found ${requiredFields.length} required fields`);

      for (let i = 0; i < requiredFields.length; i++) {
        const field = requiredFields[i];
        const value = await field.inputValue().catch(() => '');
        const placeholder = await field.getAttribute('placeholder').catch(() => '');
        console.log(`📝 Required field ${i + 1}: "${placeholder}" = "${value}"`);

        if (!value.trim()) {
          console.log(`⚠️  Required field ${i + 1} is empty`);
        }
      }
    }

    // Take screenshot before submission
    await page.screenshot({ path: 'debug-simple-form-before-submit.png', fullPage: true });

    // Try submission
    console.log('🚀 Attempting form submission...');

    try {
      await Promise.all([
        page.waitForNavigation({ timeout: 5000 }).catch(() => console.log('⚠️  Navigation timeout')),
        submitButton.click()
      ]);
    } catch (error) {
      console.log('❌ Submit failed:', error);
    }

    // Wait a bit and check results
    await page.waitForTimeout(2000);

    const currentUrl = page.url();
    console.log(`📍 Current URL: ${currentUrl}`);

    if (currentUrl.includes('/relationship-types/new')) {
      console.log('❌ Still on form page - submission failed');

      // Check for validation errors
      const errorElements = await page.locator('.error, .text-red-600, [role="alert"]').all();
      console.log(`📊 Found ${errorElements.length} error elements`);

      for (let i = 0; i < errorElements.length; i++) {
        const text = await errorElements[i].textContent();
        console.log(`❌ Error ${i + 1}: ${text}`);
      }
    } else {
      console.log('✅ Form submission succeeded');
    }

    console.log(`🖥️  Total console messages: ${consoleMessages.length}`);
    console.log(`🌐 Total API requests: ${apiRequests.length}`);

    if (apiRequests.length > 0) {
      console.log('🌐 API requests made:');
      apiRequests.forEach(req => console.log(`  - ${req}`));
    }
  });
});