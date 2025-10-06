const { test, expect } = require('@playwright/test');

test('debug relationship types navigation', async ({ page }) => {
  console.log('Starting debug test for relationship types navigation');

  // Login first
  await page.goto('http://localhost:3000/login');
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'Admin@123');
  await page.click('button[type="submit"]');

  // Wait for login to complete
  await page.waitForLoadState('networkidle');
  console.log('Login completed, current URL:', page.url());

  // Try to navigate directly to relationship types
  console.log('Navigating directly to relationship types...');
  await page.goto('http://localhost:3000/relationship-types');

  // Wait for navigation
  await page.waitForLoadState('networkidle');
  console.log('Direct navigation completed, final URL:', page.url());

  // Check page title
  const title = await page.title();
  console.log('Page title:', title);

  // Check for any h1 elements
  const h1Elements = await page.locator('h1').all();
  console.log('Number of h1 elements found:', h1Elements.length);

  for (let i = 0; i < h1Elements.length; i++) {
    const text = await h1Elements[i].textContent();
    console.log(`h1[${i}]:`, text);
  }

  // Take screenshot
  await page.screenshot({ path: 'debug-relationship-types-final.png' });

  // Check for console errors
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      console.log('Console error:', msg.text());
    }
  });

  // Check if we're actually on the relationship types page
  const currentUrl = page.url();
  if (currentUrl.includes('relationship-types')) {
    console.log('✅ Successfully navigated to relationship types page');
  } else {
    console.log('❌ Failed to navigate to relationship types page, ended up at:', currentUrl);
  }
});