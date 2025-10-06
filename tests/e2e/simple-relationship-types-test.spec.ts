const { test, expect } = require('@playwright/test');

test('simple relationship types access test', async ({ page }) => {
  console.log('Testing simple relationship types access...');

  // Login first
  await page.goto('http://localhost:3000/login');
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'Admin@123');
  await page.click('button[type="submit"]');

  // Wait for login to complete
  await page.waitForLoadState('networkidle');
  console.log('Login completed, current URL:', page.url());

  // Navigate to relationship types
  await page.goto('http://localhost:3000/relationship-types');
  await page.waitForLoadState('networkidle');
  console.log('Navigated to relationship types, final URL:', page.url());

  // Check that we're on the correct page
  expect(page.url()).toContain('relationship-types');
  console.log('✅ Successfully accessed relationship types page');

  // Take screenshot for verification
  await page.screenshot({ path: 'simple-relationship-types-test-success.png' });

  console.log('🎉 Relationship types page is accessible!');
});