import { test as base, expect } from '@playwright/test';

// Extend base test with custom fixtures
export const test = base.extend({
  // Add any custom fixtures here if needed
});

// Export expect from playwright
export { expect };

// Global test setup
test.beforeAll(async () => {
  // Any global setup before all tests
  console.log('Starting user management E2E tests...');
});

test.afterAll(async () => {
  // Any global cleanup after all tests
  console.log('User management E2E tests completed.');
});

// Custom timeout for network operations
test.setTimeout(60000);