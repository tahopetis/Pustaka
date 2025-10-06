const { test, expect } = require('@playwright/test');

test('debug auth timing for relationship types', async ({ page }) => {
  console.log('Starting auth timing debug test');

  // Set up console logging to capture router guard logs
  const consoleMessages = [];
  page.on('console', msg => {
    consoleMessages.push({
      type: msg.type(),
      text: msg.text(),
      location: msg.location()
    });
    if (msg.type() === 'error' || msg.text().includes('Router guard')) {
      console.log(`[${msg.type()}] ${msg.text()}`);
    }
  });

  // Login first
  await page.goto('http://localhost:3000/login');
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'Admin@123');
  await page.click('button[type="submit"]');

  // Wait for login and dashboard to load
  await page.waitForLoadState('networkidle');
  console.log('Login completed, current URL:', page.url());

  // Wait a bit longer to ensure auth store is fully initialized
  await page.waitForTimeout(2000);

  // Now try to navigate to relationship types
  console.log('Navigating to relationship types after auth store should be ready...');
  await page.goto('http://localhost:3000/relationship-types');

  // Wait for navigation
  await page.waitForLoadState('networkidle');
  console.log('Navigation completed, final URL:', page.url());

  // Check if we see router guard logs
  const routerGuardLogs = consoleMessages.filter(msg =>
    msg.text.includes('Router guard') || msg.text.includes('Missing permission')
  );
  console.log('Router guard logs found:', routerGuardLogs.length);
  routerGuardLogs.forEach(log => console.log('  -', log.text));

  // Check page content
  const h1Elements = await page.locator('h1').all();
  console.log('h1 elements on final page:');
  for (let i = 0; i < h1Elements.length; i++) {
    const text = await h1Elements[i].textContent();
    console.log(`  h1[${i}]:`, text);
  }

  // Take screenshot
  await page.screenshot({ path: 'debug-auth-timing-result.png' });

  // Success check
  const currentUrl = page.url();
  if (currentUrl.includes('relationship-types')) {
    console.log('✅ SUCCESS: Navigated to relationship types page');
  } else {
    console.log('❌ FAILED: Redirected to:', currentUrl);
  }
});