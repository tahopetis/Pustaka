import { test, expect } from '@playwright/test';

test.describe('User Management Functionality', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to login page first
    await page.goto('/login');
  });

  test('should login successfully with admin credentials', async ({ page }) => {
    // Fill in login form
    await page.fill('input[name="username"], input[id="username"], input[placeholder*="username"], input[placeholder*="Username"]', 'admin');
    await page.fill('input[name="password"], input[id="password"], input[placeholder*="password"], input[placeholder*="Password"]', 'Admin@123');

    // Submit login form
    await page.click('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")');

    // Wait for navigation and verify we're logged in
    await page.waitForURL('**/dashboard');
    await expect(page).toHaveURL(/.*dashboard/);

    // Verify authentication by checking for user-related elements
    await expect(page.locator('body')).toContainText('Dashboard');
  });

  test('should navigate to users page and display admin user', async ({ page }) => {
    // Login first
    await page.fill('input[name="username"], input[id="username"], input[placeholder*="username"], input[placeholder*="Username"]', 'admin');
    await page.fill('input[name="password"], input[id="password"], input[placeholder*="password"], input[placeholder*="Password"]', 'Admin@123');
    await page.click('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")');
    await page.waitForURL('**/dashboard');

    // Navigate to users page
    await page.goto('/users');
    await page.waitForLoadState('networkidle');

    // Verify we're on the users page
    await expect(page.locator('h1, h2')).toContainText('Users');

    // Verify that the users list loads and is not empty
    await expect(page.locator('body')).not.toContainText('No users found');

    // Look for admin user in the table
    const adminUserRow = page.locator('table tbody tr').filter({ hasText: 'admin' });
    await expect(adminUserRow).toHaveCount(1);

    // Verify admin user details
    await expect(adminUserRow).toContainText('admin');
    await expect(adminUserRow.locator('td:nth-child(1)')).toContainText('admin'); // Username
    await expect(adminUserRow.locator('td:nth-child(3)')).toContainText('admin'); // Role

    // Verify user is active
    await expect(adminUserRow.locator('td:nth-child(4)')).toContainText('Active');

    // Take screenshot for verification
    await page.screenshot({ path: 'test-results/users-page-with-admin.png' });
  });

  test('should navigate to add user page without 404 error', async ({ page }) => {
    // Login first
    await page.fill('input[name="username"], input[id="username"], input[placeholder*="username"], input[placeholder*="Username"]', 'admin');
    await page.fill('input[name="password"], input[id="password"], input[placeholder*="password"], input[placeholder*="Password"]', 'Admin@123');
    await page.click('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")');
    await page.waitForURL('**/dashboard');

    // Navigate to users page first
    await page.goto('/users');
    await page.waitForLoadState('networkidle');

    // Click "Add User" button
    await page.click('a[href="/users/new"], button:has-text("Add User"), a:has-text("Add User")');

    // Verify we're on the create user page (not 404)
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1, h2')).toContainText('Create User');

    // Verify page content is loaded properly
    await expect(page.locator('body')).not.toContainText('404');
    await expect(page.locator('body')).not.toContainText('Not Found');
    await expect(page.locator('body')).not.toContainText('Page not found');

    // Verify form elements are present
    await expect(page.locator('input[name="username"], input[id="username"]')).toBeVisible();
    await expect(page.locator('input[name="email"], input[id="email"]')).toBeVisible();
    await expect(page.locator('input[name="password"], input[id="password"]')).toBeVisible();
    await expect(page.locator('input[name="confirmPassword"], input[id="confirmPassword"]')).toBeVisible();

    // Take screenshot for verification
    await page.screenshot({ path: 'test-results/add-user-page-loaded.png' });
  });

  test('should create a new user successfully', async ({ page }) => {
    // Login first
    await page.fill('input[name="username"], input[id="username"], input[placeholder*="username"], input[placeholder*="Username"]', 'admin');
    await page.fill('input[name="password"], input[id="password"], input[placeholder*="password"], input[placeholder*="Password"]', 'Admin@123');
    await page.click('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")');
    await page.waitForURL('**/dashboard');

    // Navigate to add user page
    await page.goto('/users/new');
    await page.waitForLoadState('networkidle');

    // Fill in user creation form
    const timestamp = Date.now();
    const testUsername = `testuser${timestamp}`;
    const testEmail = `testuser${timestamp}@example.com`;

    await page.fill('input[name="username"], input[id="username"]', testUsername);
    await page.fill('input[name="email"], input[id="email"]', testEmail);
    await page.fill('input[name="password"], input[id="password"]', 'TestPassword123!');
    await page.fill('input[name="confirmPassword"], input[id="confirmPassword"]', 'TestPassword123!');

    // Select a role (viewer role should be available)
    const viewerRoleCheckbox = page.locator('input[type="checkbox"][value="viewer"]');
    if (await viewerRoleCheckbox.isVisible()) {
      await viewerRoleCheckbox.check();
    }

    // Ensure user is active
    const activeCheckbox = page.locator('input[type="checkbox"][name="is_active"], input[type="checkbox"]:has-text("Active")');
    if (await activeCheckbox.isVisible() && !(await activeCheckbox.isChecked())) {
      await activeCheckbox.check();
    }

    // Submit form
    await page.click('button[type="submit"]:has-text("Create User"), button:has-text("Create User")');

    // Wait for navigation back to users list
    await page.waitForURL('**/users');
    await page.waitForLoadState('networkidle');

    // Verify success message (if toast notifications are used)
    await expect(page.locator('body')).toContainText('User created successfully', { timeout: 5000 });

    // Verify new user appears in the list
    const newUserRow = page.locator('table tbody tr').filter({ hasText: testUsername });
    await expect(newUserRow).toHaveCount(1);
    await expect(newUserRow).toContainText(testUsername);
    await expect(newUserRow).toContainText(testEmail);

    // Take screenshot for verification
    await page.screenshot({ path: 'test-results/user-created-successfully.png' });
  });

  test('should display user details when clicking on user', async ({ page }) => {
    // Login first
    await page.fill('input[name="username"], input[id="username"], input[placeholder*="username"], input[placeholder*="Username"]', 'admin');
    await page.fill('input[name="password"], input[id="password"], input[placeholder*="password"], input[placeholder*="Password"]', 'Admin@123');
    await page.click('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")');
    await page.waitForURL('**/dashboard');

    // Navigate to users page
    await page.goto('/users');
    await page.waitForLoadState('networkidle');

    // Find admin user row and click view button
    const adminUserRow = page.locator('table tbody tr').filter({ hasText: 'admin' });
    const viewButton = adminUserRow.locator('a[title="View"], a:has-text("View"), button[title="View"]');
    await viewButton.click();

    // Verify we're on user details page
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1, h2')).toContainText('User Details');

    // Verify user details are displayed
    await expect(page.locator('body')).toContainText('admin');

    // Take screenshot for verification
    await page.screenshot({ path: 'test-results/user-details-page.png' });
  });

  test('should allow editing user details', async ({ page }) => {
    // Login first
    await page.fill('input[name="username"], input[id="username"], input[placeholder*="username"], input[placeholder*="Username"]', 'admin');
    await page.fill('input[name="password"], input[id="password"], input[placeholder*="password"], input[placeholder*="Password"]', 'Admin@123');
    await page.click('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")');
    await page.waitForURL('**/dashboard');

    // Navigate to users page
    await page.goto('/users');
    await page.waitForLoadState('networkidle');

    // Find admin user row and click edit button
    const adminUserRow = page.locator('table tbody tr').filter({ hasText: 'admin' });
    const editButton = adminUserRow.locator('a[title="Edit"], a:has-text("Edit"), button[title="Edit"]');
    await editButton.click();

    // Verify we're on edit user page
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1, h2')).toContainText('Edit User');

    // Verify edit form is loaded
    await expect(page.locator('input[name="username"], input[id="username"]')).toBeVisible();
    await expect(page.locator('input[name="email"], input[id="email"]')).toBeVisible();

    // Take screenshot for verification
    await page.screenshot({ path: 'test-results/edit-user-page.png' });
  });

  test('should handle invalid login credentials gracefully', async ({ page }) => {
    // Try to login with invalid credentials
    await page.fill('input[name="username"], input[id="username"], input[placeholder*="username"], input[placeholder*="Username"]', 'invaliduser');
    await page.fill('input[name="password"], input[id="password"], input[placeholder*="password"], input[placeholder*="Password"]', 'wrongpassword');
    await page.click('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")');

    // Should stay on login page or show error
    await page.waitForTimeout(2000);

    // Verify error message is shown
    await expect(page.locator('body')).toContainText('Invalid', { timeout: 5000 });

    // Should not be redirected to dashboard
    await expect(page).toHaveURL(/.*login/);

    // Take screenshot for verification
    await page.screenshot({ path: 'test-results/invalid-login-error.png' });
  });

  test('should require authentication to access users pages', async ({ page }) => {
    // Try to access users page without authentication
    await page.goto('/users');

    // Should be redirected to login page
    await page.waitForURL('**/login');
    await expect(page).toHaveURL(/.*login/);

    // Try to access add user page without authentication
    await page.goto('/users/new');

    // Should be redirected to login page
    await page.waitForURL('**/login');
    await expect(page).toHaveURL(/.*login/);

    // Take screenshot for verification
    await page.screenshot({ path: 'test-results/authentication-required.png' });
  });
});