import { test, expect } from '@playwright/test';

test.describe('Debug UI - Menu Visibility', () => {
  test('check sidebar relationship types link visibility', async ({ page }) => {
    // Login as admin user
    await page.goto('http://localhost:3000/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'Admin@123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');

    // Wait for page to fully load
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // Check user data in localStorage
    const localStorageData = await page.evaluate(() => {
      const token = localStorage.getItem('access_token');
      const userStr = localStorage.getItem('user');
      return {
        hasToken: !!token,
        user: userStr ? JSON.parse(userStr) : null,
        tokenLength: token?.length || 0
      };
    });
    console.log('Local storage data:', JSON.stringify(localStorageData, null, 2));

    // Check all navigation links in sidebar
    const navLinks = await page.locator('nav a').all();
    console.log('Total navigation links found:', navLinks.length);

    const linkTexts = [];
    for (let i = 0; i < navLinks.length; i++) {
      const text = await navLinks[i].textContent();
      const href = await navLinks[i].getAttribute('href');
      linkTexts.push({ text: text?.trim(), href });
    }
    console.log('Navigation links:', linkTexts);

    // Specifically look for relationship types link
    const relationshipTypesLink = page.locator('a[href="/relationship-types"]');
    const isLinkVisible = await relationshipTypesLink.isVisible();
    console.log('Relationship types link visible:', isLinkVisible);

    if (isLinkVisible) {
      const linkText = await relationshipTypesLink.textContent();
      console.log('Relationship types link text:', linkText);
    } else {
      console.log('Relationship types link not found or not visible');

      // Check if it exists but is hidden
      const linkExists = await relationshipTypesLink.count();
      console.log('Relationship types link exists in DOM:', linkExists > 0);

      if (linkExists > 0) {
        // Check its computed style to see why it's hidden
        const visibility = await relationshipTypesLink.evaluate(el => {
          const style = window.getComputedStyle(el);
          return {
            display: style.display,
            visibility: style.visibility,
            opacity: style.opacity,
            height: style.height,
            width: style.width
          };
        });
        console.log('Relationship types link computed style:', visibility);
      }
    }

    // Check if hasPermission function is working correctly
    const permissionCheck = await page.evaluate(() => {
      // Try to access the auth store to check permissions
      try {
        const userStr = localStorage.getItem('user');
        if (userStr) {
          const user = JSON.parse(userStr);
          const hasRelationshipTypePermission = user.permissions?.includes('relationship-type:read') || false;
          return {
            permissions: user.permissions,
            hasRelationshipTypePermission: hasRelationshipTypePermission
          };
        }
      } catch (e) {
        console.error('Error checking permissions:', e);
      }
      return null;
    });
    console.log('Permission check result:', permissionCheck);

    // Take a screenshot to see what the UI actually looks like
    await page.screenshot({ path: 'debug-sidebar-ui.png', fullPage: true });
    console.log('Screenshot saved as debug-sidebar-ui.png');
  });
});