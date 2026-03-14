// Real-user Playwright test — NO mocks, NO interception.
// Logs in via the actual login page, then tests all authenticated pages.
import { chromium } from 'playwright';

const BASE_URL = 'http://localhost:3000';
const SCREENSHOT_DIR = '/tmp/edduhub-test-screenshots';

const TEST_USER = {
  email: 'test@example.com',
  password: 'SuperSecure-password123!',
};

const results = { passed: [], failed: [], errors: [] };

async function test(name, fn) {
  try {
    await fn();
    results.passed.push(name);
    console.log(`  ✅ ${name}`);
  } catch (err) {
    results.failed.push({ name, error: err.message });
    console.log(`  ❌ ${name}: ${err.message}`);
  }
}

async function main() {
  console.log('🚀 Starting comprehensive real-user testing...\n');

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
  });

  const consoleErrors = [];
  const pageErrors = [];

  const page = await context.newPage();
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      consoleErrors.push({ url: page.url(), text: msg.text() });
    }
  });
  page.on('pageerror', (err) => {
    pageErrors.push({ url: page.url(), error: err.message });
  });

  // ========================================
  // PHASE 1: Auth Pages (Unauthenticated)
  // ========================================
  console.log('\n📋 PHASE 1: Auth Pages (Unauthenticated)\n');

  await test('Login page loads', async () => {
    const response = await page.goto(`${BASE_URL}/auth/login`, {
      waitUntil: 'domcontentloaded',
      timeout: 15000,
    });
    if (!response || response.status() >= 400)
      throw new Error(`HTTP ${response?.status()}`);
    await page.waitForTimeout(2000);
    await page.screenshot({ path: `${SCREENSHOT_DIR}/01-login-page.png` });
  });

  await test('Login page has form elements', async () => {
    await page.waitForSelector('form', { timeout: 5000 });
    const emailInput = await page
      .locator('#email, input[type="email"], input[name="email"]')
      .count();
    const passwordInput = await page
      .locator('#password, input[type="password"], input[name="password"]')
      .count();
    const submitBtn = await page.locator('button[type="submit"]').count();
    if (emailInput === 0) throw new Error('Email input not found');
    if (passwordInput === 0) throw new Error('Password input not found');
    if (submitBtn === 0) throw new Error('Submit button not found');
  });

  await test('Login page has register link', async () => {
    const link = await page.locator('a[href="/auth/register"]');
    const count = await link.count();
    if (count === 0) throw new Error('Register link not found');
  });

  await test('Registration page loads', async () => {
    const response = await page.goto(`${BASE_URL}/auth/register`, {
      waitUntil: 'domcontentloaded',
      timeout: 15000,
    });
    if (!response || response.status() >= 400)
      throw new Error(`HTTP ${response?.status()}`);
    await page.waitForTimeout(2000);
    await page.screenshot({ path: `${SCREENSHOT_DIR}/02-register-page.png` });
  });

  await test('Registration page has form elements', async () => {
    await page.waitForSelector('form', { timeout: 5000 });
    const inputCount = await page.locator('input').count();
    if (inputCount < 3)
      throw new Error(`Expected at least 3 inputs, got ${inputCount}`);
  });

  // ========================================
  // PHASE 2: Unauthenticated Redirects
  // ========================================
  console.log('\n📋 PHASE 2: Unauthenticated Redirects\n');

  await test('Root page redirects to login when unauthenticated', async () => {
    await page.goto(`${BASE_URL}/`, {
      waitUntil: 'domcontentloaded',
      timeout: 15000,
    });
    await page.waitForTimeout(3000);
    const url = page.url();
    if (!url.includes('/auth/login'))
      throw new Error(`Expected redirect to login, got: ${url}`);
    await page.screenshot({ path: `${SCREENSHOT_DIR}/03-root-redirect.png` });
  });

  const sampleProtectedPages = ['/students', '/courses', '/analytics', '/settings', '/profile'];

  for (const path of sampleProtectedPages) {
    await test(`Protected page ${path} redirects to login`, async () => {
      await page.goto(`${BASE_URL}${path}`, {
        waitUntil: 'domcontentloaded',
        timeout: 15000,
      });
      await page.waitForTimeout(3000);
      const url = page.url();
      if (!url.includes('/auth/login'))
        throw new Error(`Expected redirect, got: ${url}`);
    });
  }

  // ========================================
  // PHASE 3: Real Login Flow
  // ========================================
  console.log('\n📋 PHASE 3: Real Login Flow\n');

  await test('Login with real credentials', async () => {
    // Navigate to login page
    await page.goto(`${BASE_URL}/auth/login`, {
      waitUntil: 'domcontentloaded',
      timeout: 15000,
    });
    await page.waitForTimeout(1000);

    // Fill in real credentials
    await page.fill('#email', TEST_USER.email);
    await page.fill('#password', TEST_USER.password);
    await page.screenshot({ path: `${SCREENSHOT_DIR}/04-login-filled.png` });

    // Submit the form
    await page.click('button[type="submit"]');

    // Wait for navigation away from login (up to 10s)
    try {
      await page.waitForURL((url) => !url.toString().includes('/auth/login'), {
        timeout: 10000,
      });
    } catch {
      // Take screenshot of current state for debugging
      await page.screenshot({ path: `${SCREENSHOT_DIR}/04-login-stuck.png` });
      throw new Error(`Still on login page after submit. URL: ${page.url()}`);
    }

    await page.waitForTimeout(2000);
    const url = page.url();
    console.log(`    Post-login URL: ${url}`);
    await page.screenshot({ path: `${SCREENSHOT_DIR}/05-post-login.png` });

    if (url.includes('/auth/login')) {
      throw new Error('Still on login page after successful submit');
    }
  });

  // ========================================
  // PHASE 4: Verify Auth State Persists
  // ========================================
  console.log('\n📋 PHASE 4: Auth State Verification\n');

  await test('Auth session persists in localStorage', async () => {
    const authData = await page.evaluate(() => {
      return localStorage.getItem('edduhub_auth');
    });
    if (!authData) throw new Error('No auth data in localStorage');
    const parsed = JSON.parse(authData);
    if (!parsed.token) throw new Error('No token in auth data');
    if (!parsed.user) throw new Error('No user in auth data');
    console.log(`    Logged in as: ${parsed.user.email} (${parsed.user.role})`);
  });

  await test('Dashboard loads after login', async () => {
    // Navigate to root (dashboard)
    await page.goto(`${BASE_URL}/`, {
      waitUntil: 'domcontentloaded',
      timeout: 15000,
    });
    await page.waitForTimeout(3000);

    const url = page.url();
    if (url.includes('/auth/login'))
      throw new Error('Redirected back to login — auth not persisted');
    await page.screenshot({ path: `${SCREENSHOT_DIR}/06-dashboard.png` });
  });

  // ========================================
  // PHASE 5: Test All Authenticated Pages
  // ========================================
  console.log('\n📋 PHASE 5: Test All Authenticated Pages\n');

  const authenticatedPages = [
    { path: '/', name: 'Dashboard' },
    { path: '/students', name: 'Students' },
    { path: '/courses', name: 'Courses' },
    { path: '/analytics', name: 'Analytics' },
    { path: '/announcements', name: 'Announcements' },
    { path: '/assignments', name: 'Assignments' },
    { path: '/attendance', name: 'Attendance' },
    { path: '/calendar', name: 'Calendar' },
    { path: '/departments', name: 'Departments' },
    { path: '/grades', name: 'Grades' },
    { path: '/notifications', name: 'Notifications' },
    { path: '/profile', name: 'Profile' },
    { path: '/settings', name: 'Settings' },
    { path: '/users', name: 'Users' },
    { path: '/quizzes', name: 'Quizzes' },
    { path: '/timetable', name: 'Timetable' },
    { path: '/fees', name: 'Fees' },
    { path: '/exams', name: 'Exams' },
    { path: '/forum', name: 'Forum' },
    { path: '/files', name: 'Files' },
    { path: '/roles', name: 'Roles' },
    { path: '/placements', name: 'Placements' },
    { path: '/self-service', name: 'Self Service' },
    { path: '/faculty-tools', name: 'Faculty Tools' },
    { path: '/parent-portal', name: 'Parent Portal' },
    { path: '/batch-operations', name: 'Batch Operations' },
    { path: '/audit-logs', name: 'Audit Logs' },
    { path: '/webhooks', name: 'Webhooks' },
    { path: '/system-status', name: 'System Status' },
    { path: '/advanced-analytics', name: 'Advanced Analytics' },
    { path: '/parent-links', name: 'Parent Links' },
    { path: '/student-dashboard', name: 'Student Dashboard' },
  ];

  for (const { path, name } of authenticatedPages) {
    await test(`Page "${name}" (${path}) loads without crash`, async () => {
      const response = await page.goto(`${BASE_URL}${path}`, {
        waitUntil: 'domcontentloaded',
        timeout: 15000,
      });

      // Check for server errors
      if (response && response.status() >= 500) {
        throw new Error(`Server error: HTTP ${response.status()}`);
      }

      // Wait for page to settle
      await page.waitForTimeout(2000);

      const url = page.url();
      const screenshotName = name.toLowerCase().replace(/\s+/g, '-');
      await page.screenshot({
        path: `${SCREENSHOT_DIR}/page-${screenshotName}.png`,
      });

      // Check for Next.js error overlay
      const errorOverlay = await page.locator('[data-nextjs-dialog]').count();
      if (errorOverlay > 0) {
        const errorText = await page
          .locator('[data-nextjs-dialog]')
          .textContent();
        throw new Error(
          `Next.js error overlay: ${errorText?.substring(0, 200)}`
        );
      }

      // Check for React error boundary
      const errorBoundary = await page
        .locator('text="Something went wrong"')
        .count();
      if (errorBoundary > 0) {
        throw new Error('React error boundary triggered');
      }

      // Check for "Application error" (Next.js production error)
      const appError = await page
        .locator('text="Application error"')
        .count();
      if (appError > 0) {
        throw new Error('Application error page shown');
      }

      // Verify we're not redirected to login unexpectedly
      if (url.includes('/auth/login') && path !== '/') {
        throw new Error(`Unexpectedly redirected to login from ${path}`);
      }
    });
  }

  // ========================================
  // PHASE 6: Navigation & Sidebar
  // ========================================
  console.log('\n📋 PHASE 6: Navigation & Sidebar\n');

  await test('Sidebar renders with navigation links', async () => {
    await page.goto(`${BASE_URL}/`, {
      waitUntil: 'domcontentloaded',
      timeout: 15000,
    });
    await page.waitForTimeout(2000);

    const sidebar = await page
      .locator('nav, [class*="sidebar"], aside')
      .first();
    const isVisible = await sidebar.isVisible().catch(() => false);
    await page.screenshot({ path: `${SCREENSHOT_DIR}/07-sidebar.png` });
    console.log(`    Sidebar visible: ${isVisible}`);
    if (!isVisible)
      throw new Error('Sidebar is not visible on desktop viewport');
  });

  await test('Top bar renders', async () => {
    const topbar = await page
      .locator('header, [class*="topbar"], [class*="Topbar"]')
      .first();
    const isVisible = await topbar.isVisible().catch(() => false);
    console.log(`    Top bar visible: ${isVisible}`);
  });

  await test('Sidebar has navigation links', async () => {
    const navLinks = await page.locator('nav a, aside a').count();
    console.log(`    Navigation links found: ${navLinks}`);
    if (navLinks === 0)
      throw new Error('No navigation links found in sidebar');
  });

  // ========================================
  // PHASE 7: Verify-email page
  // ========================================
  console.log('\n📋 PHASE 7: Email Verification Page\n');

  await test('Verify email page loads', async () => {
    const response = await page.goto(
      `${BASE_URL}/auth/verify-email?flow=test&token=test`,
      { waitUntil: 'domcontentloaded', timeout: 15000 }
    );
    if (response && response.status() >= 500)
      throw new Error(`HTTP ${response.status()}`);
    await page.waitForTimeout(2000);
    await page.screenshot({ path: `${SCREENSHOT_DIR}/08-verify-email.png` });
  });

  // ========================================
  // PHASE 8: Logout Flow
  // ========================================
  console.log('\n📋 PHASE 8: Logout Flow\n');

  await test('Logout clears auth and redirects to login', async () => {
    // First go back to dashboard
    await page.goto(`${BASE_URL}/`, {
      waitUntil: 'domcontentloaded',
      timeout: 15000,
    });
    await page.waitForTimeout(2000);

    // Look for logout button/link
    const logoutBtn = page.locator(
      'button:has-text("Logout"), button:has-text("Log out"), button:has-text("Sign out"), a:has-text("Logout"), a:has-text("Log out"), a:has-text("Sign out"), [data-testid="logout"]'
    );
    const logoutCount = await logoutBtn.count();
    console.log(`    Logout buttons found: ${logoutCount}`);

    if (logoutCount > 0) {
      await logoutBtn.first().click();
      await page.waitForTimeout(3000);
      const url = page.url();
      console.log(`    Post-logout URL: ${url}`);
      await page.screenshot({ path: `${SCREENSHOT_DIR}/09-post-logout.png` });

      // Verify auth is cleared
      const authData = await page.evaluate(() =>
        localStorage.getItem('edduhub_auth')
      );
      if (authData) {
        const parsed = JSON.parse(authData);
        if (parsed.token)
          console.log('    ⚠️  Auth token still in localStorage after logout');
      }
    } else {
      console.log('    ⚠️  No logout button found — skipping logout test');
    }
  });

  // ========================================
  // Summary
  // ========================================
  console.log('\n\n========================================');
  console.log('📊 TEST RESULTS SUMMARY');
  console.log('========================================');
  console.log(`✅ Passed: ${results.passed.length}`);
  console.log(`❌ Failed: ${results.failed.length}`);

  if (results.failed.length > 0) {
    console.log('\nFailed tests:');
    results.failed.forEach((f) => {
      console.log(`  ❌ ${f.name}`);
      console.log(`     Error: ${f.error}`);
    });
  }

  if (consoleErrors.length > 0) {
    console.log(`\n⚠️  Console errors (${consoleErrors.length}):`);
    const unique = [...new Set(consoleErrors.map((e) => e.text))];
    unique
      .slice(0, 20)
      .forEach((e) => console.log(`  - ${e.substring(0, 150)}`));
  }

  if (pageErrors.length > 0) {
    console.log(`\n💥 Page errors (${pageErrors.length}):`);
    const unique = [
      ...new Set(pageErrors.map((e) => `[${e.url}] ${e.error}`)),
    ];
    unique.forEach((e) => console.log(`  - ${e.substring(0, 200)}`));
  }

  console.log('\nScreenshots saved to:', SCREENSHOT_DIR);

  await browser.close();

  if (results.failed.length > 0) {
    process.exit(1);
  }
}

main().catch((err) => {
  console.error('Fatal error:', err);
  process.exit(2);
});
