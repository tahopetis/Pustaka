#!/bin/bash

# User Management Test Execution Script
# This script runs comprehensive tests to verify user management functionality

set -e

echo "🚀 Starting User Management Functionality Tests"
echo "=============================================="

# Create results directory
mkdir -p test-results

# Check if Playwright is installed
if ! command -v npx playwright &> /dev/null; then
    echo "❌ Playwright is not installed. Installing..."
    npm install --save-dev @playwright/test
    npx playwright install
fi

# Ensure we're in the web directory
cd "$(dirname "$0")"

# Check if the application is running
echo "🔍 Checking if application is running on http://localhost:3000..."
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo "✅ Application is running"
else
    echo "❌ Application is not running. Please start the application first:"
    echo "   cd /home/syam/dev/pustaka && make dev"
    exit 1
fi

echo ""
echo "📋 Running User Management Tests..."
echo "--------------------------------"

# Run the user management tests
echo "🧪 Executing comprehensive user management tests..."
npx playwright test tests/e2e/user-management.spec.ts --reporter=list

echo ""
echo "📊 Test Results Summary:"
echo "========================"

# Check if tests passed
if [ $? -eq 0 ]; then
    echo "✅ All tests passed successfully!"
    echo ""
    echo "📸 Screenshots captured in test-results/"
    echo "📈 Detailed HTML report available at test-results/html-report/index.html"
    echo "📋 JSON results available at test-results/results.json"
else
    echo "❌ Some tests failed. Check the detailed report for more information."
    echo ""
    echo "🔍 Debug information:"
    echo "   - Screenshots: test-results/"
    echo "   - Traces: test-results/traces/"
    echo "   - Videos: test-results/videos/"
    echo "   - HTML Report: test-results/html-report/index.html"
fi

echo ""
echo "🎯 Testing Complete!"
echo "==================="
echo ""
echo "What was tested:"
echo "  ✓ Login functionality with admin/Admin@123"
echo "  ✓ Users page displays admin user (Issue #1 fix verification)"
echo "  ✓ Add User button navigation to /users/new (Issue #2 fix verification)"
echo "  ✓ User creation form functionality"
echo "  ✓ User list updates after creating new user"
echo "  ✓ User details page access"
echo "  ✓ User edit page access"
echo "  ✓ Authentication requirements"
echo "  ✓ Error handling for invalid credentials"
echo ""
echo "Screenshots for verification:"
echo "  📸 test-results/users-page-with-admin.png"
echo "  📸 test-results/add-user-page-loaded.png"
echo "  📸 test-results/user-created-successfully.png"
echo "  📸 test-results/user-details-page.png"
echo "  📸 test-results/edit-user-page.png"
echo "  📸 test-results/invalid-login-error.png"
echo "  📸 test-results/authentication-required.png"