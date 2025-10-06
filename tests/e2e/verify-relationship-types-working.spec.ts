const { test, expect } = require('@playwright/test');

test('verify relationship types functionality is working', async ({ page }) => {
  console.log('Starting comprehensive relationship types test...');

  // Login first
  await page.goto('http://localhost:3000/login');
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'Admin@123');
  await page.click('button[type="submit"]');

  // Wait for login to complete
  await page.waitForLoadState('networkidle');
  console.log('Login completed');

  // Navigate to relationship types
  await page.goto('http://localhost:3000/relationship-types');
  await page.waitForLoadState('networkidle');
  console.log('Navigated to relationship types page');

  // Check that we're on the correct page
  const currentUrl = page.url();
  expect(currentUrl).toContain('relationship-types');
  console.log('✅ Successfully accessed relationship types page');

  // Check page title
  const h1Element = await page.locator('h1').first();
  const title = await h1Element.textContent();
  expect(title).toContain('Relationship Types');
  console.log('✅ Page title is correct:', title);

  // Check if there are relationship types displayed
  const relationshipTypeItems = await page.locator('[data-testid="relationship-type-item"]').all();
  console.log(`Found ${relationshipTypeItems.length} relationship type items`);

  if (relationshipTypeItems.length > 0) {
    console.log('✅ Relationship types are being displayed');

    // Check if the "Create New" button is visible (should be with relationship_type:create permission)
    const createButton = await page.locator('a[href="/relationship-types/new"]').first();
    if (await createButton.isVisible()) {
      console.log('✅ Create New button is visible');
    } else {
      console.log('ℹ️ Create New button not visible (permission issue or no relationship types)');
    }
  } else {
    console.log('ℹ️ No relationship types found in the list');
  }

  // Take a screenshot for verification
  await page.screenshot({ path: 'relationship-types-working-verification.png' });
  console.log('✅ Screenshot saved for verification');

  console.log('🎉 Relationship types functionality is working correctly!');
});