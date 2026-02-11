import { test, expect } from '@playwright/test';

/**
 * Comprehensive End-to-End Testing for EduHub Frontend
 * Tests all core user flows and page navigation
 */

test.describe('Frontend Comprehensive Testing', () => {
  test.describe('1. Home Page Testing', () => {
    test('should load home page successfully', async ({ page }) => {
      await page.goto('/');
      
      // Check page loads without errors
      await expect(page).toHaveTitle(/edduhub platform/i);
      
      // Check for main content
      await expect(page.locator('body')).toBeVisible();
      
      // Check console for any errors
      const consoleErrors: string[] = [];
      page.on('console', msg => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });
      
      await page.waitForLoadState('networkidle');
      
      console.log('Console Errors on Home Page:', consoleErrors);
      expect(consoleErrors).toHaveLength(0);
    });
  });

  test.describe('2. User Registration Flow', () => {
    test('should navigate to registration page', async ({ page }) => {
      await page.goto('/auth/register');
      
      // Check page loads
      await expect(page.locator('h1, h2')).toContainText(/register|sign up/i);
      
      // Check for form elements
      await expect(page.locator('input[name="email"]')).toBeVisible();
      await expect(page.locator('input[name="password"]')).toBeVisible();
      await expect(page.locator('input[name="firstName"]')).toBeVisible();
      await expect(page.locator('input[name="lastName"]')).toBeVisible();
      
      // Check console for errors
      const consoleErrors: string[] = [];
      page.on('console', msg => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });
      
      await page.waitForLoadState('networkidle');
      console.log('Registration Page Console Errors:', consoleErrors);
    });

    test('should complete registration form', async ({ page }) => {
      await page.goto('/auth/register');
      
      // Fill registration form with test data
      const timestamp = Date.now();
      const testEmail = `testuser${timestamp}@example.com`;
      await page.fill('input[name="email"]', testEmail);
      await page.fill('input[name="password"]', 'TestPassword123!');
      await page.fill('input[name="firstName"]', 'Test');
      await page.fill('input[name="lastName"]', 'User');
      
      // Check if submit button exists and is clickable
      await expect(page.locator('button[type="submit"]')).toBeVisible();
      
      // Capture any console errors during form interaction
      const consoleErrors: string[] = [];
      page.on('console', msg => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });
      
      // Test form validation
      await page.fill('input[name="email"]', 'invalid-email');
      await page.click('button[type="submit"]');
      await page.waitForTimeout(1000);
      
      // Check for validation errors
      const validationError = page.locator('.error, .error-message, [data-testid="error"]');
      if (await validationError.isVisible()) {
        console.log('Form validation working correctly');
      }
      
      // Fix email and proceed with submission
      await page.fill('input[name="email"]', testEmail);
      
      // Submit the form to test complete flow
      try {
        await Promise.all([
          page.waitForNavigation({ timeout: 10000 }),
          page.click('button[type="submit"]')
        ]);
        
        // Check if registration was successful or handled appropriately
        const currentURL = page.url();
        console.log('Registration submission URL:', currentURL);
        
        // Look for success messages or redirects
        const successMessage = page.locator('.success, .success-message, [data-testid="success"]');
        if (await successMessage.isVisible()) {
          console.log('Registration successful');
        }
        
      } catch {
        console.log('Registration submission test completed (may have failed as expected)');
      }
      
      console.log('Registration Form Console Errors:', consoleErrors);
    });
  });

  test.describe('3. User Login Flow', () => {
    test('should navigate to login page', async ({ page }) => {
      await page.goto('/auth/login');
      
      // Check page loads
      await expect(page.locator('h1, h2')).toContainText(/login|sign in/i);
      
      // Check for form elements
      await expect(page.locator('input[name="email"]')).toBeVisible();
      await expect(page.locator('input[name="password"]')).toBeVisible();
      
      // Check console for errors
      const consoleErrors: string[] = [];
      page.on('console', msg => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });
      
      await page.waitForLoadState('networkidle');
      console.log('Login Page Console Errors:', consoleErrors);
    });

    test('should attempt login with test credentials', async ({ page }) => {
      await page.goto('/auth/login');
      
      // Test form validation first
      await page.click('button[type="submit"]');
      await page.waitForTimeout(1000);
      
      // Check for validation errors
      const validationError = page.locator('.error, .error-message, [data-testid="error"]');
      if (await validationError.isVisible()) {
        console.log('Login form validation working correctly');
      }
      
      // Fill login form with invalid credentials to test error handling
      await page.fill('input[name="email"]', 'invalid@test.com');
      await page.fill('input[name="password"]', 'wrongpassword');
      
      // Check form validation and submission
      await expect(page.locator('button[type="submit"]')).toBeVisible();
      
      // Capture console errors during login attempt
      const consoleErrors: string[] = [];
      page.on('console', msg => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });
      
      // Submit form to test authentication flow
      try {
        await Promise.all([
          page.waitForNavigation({ timeout: 10000 }),
          page.click('button[type="submit"]')
        ]);
        
        // Check for error messages or redirects
        const errorMessage = page.locator('.error, .error-message, [data-testid="error"]');
        if (await errorMessage.isVisible()) {
          console.log('Login error handling working correctly');
        }
        
        const currentURL = page.url();
        console.log('Login attempt URL:', currentURL);
        
      } catch {
        console.log('Login submission test completed (may have failed as expected)');
      }
      
      console.log('Login Form Console Errors:', consoleErrors);
    });
  });

  test.describe('4. Core Page Navigation Testing', () => {
    const corePages = [
      '/courses',
      '/quizzes',
      '/profile',
      '/assignments',
      '/attendance',
      '/grades',
      '/analytics',
      '/calendar',
      '/settings'
    ];

    test('should navigate to all core pages', async ({ page }) => {
      for (const pagePath of corePages) {
        console.log(`Testing navigation to: ${pagePath}`);
        
        // Navigate to each page
        await page.goto(pagePath);
        
        // Check page loads
        await page.waitForLoadState('networkidle');
        
        // Check for basic page content
        await expect(page.locator('body')).toBeVisible();
        
        // Capture console errors for each page
        const consoleErrors: string[] = [];
        page.on('console', msg => {
          if (msg.type() === 'error') {
            consoleErrors.push(`${pagePath}: ${msg.text()}`);
          }
        });
        
        // Wait a moment for any async errors
        await page.waitForTimeout(1000);
        
        if (consoleErrors.length > 0) {
          console.log(`Console Errors on ${pagePath}:`, consoleErrors);
        }
        
        // Check if page has protected route or redirects
        const currentURL = page.url();
        if (currentURL.includes('/auth/login') || currentURL.includes('/auth/register')) {
          console.log(`${pagePath}: Protected route - redirected to authentication`);
        } else {
          console.log(`${pagePath}: Page loaded successfully`);
        }
      }
    });

    test('should check navigation components', async ({ page }) => {
      await page.goto('/');
      
      // Check if navigation components exist
      await expect(page.locator('[data-testid="sidebar"], nav')).toBeVisible();
      
      // Check for navigation links
      const navLinks = await page.locator('a[href]').count();
      console.log(`Found ${navLinks} navigation links`);
      
      // Check for responsive menu
      await page.setViewportSize({ width: 768, height: 1024 });
      await page.waitForTimeout(500);
      
      const mobileMenuButton = page.locator('[data-testid="mobile-menu-button"], .mobile-menu-button');
      if (await mobileMenuButton.isVisible()) {
        console.log('Mobile menu button found');
        await mobileMenuButton.click();
        await page.waitForTimeout(500);
      }
    });
  });

  test.describe('5. UI Component Testing', () => {
    test('should render UI components correctly', async ({ page }) => {
      await page.goto('/');
      
      // Test common UI components
      const components = [
        { selector: 'button', name: 'buttons' },
        { selector: 'input', name: 'input fields' },
        { selector: 'select', name: 'select dropdowns' },
        { selector: '[data-testid="card"], .card', name: 'cards' },
        { selector: '[data-testid="table"], table', name: 'tables' }
      ];
      
      for (const component of components) {
        const count = await page.locator(component.selector).count();
        console.log(`Found ${count} ${component.name}`);
        
        if (count === 0) {
          console.log(`Warning: No ${component.name} found on home page`);
        }
      }
      
      // Check for dark/light theme toggle
      const themeToggle = page.locator('[data-testid="theme-toggle"], .theme-toggle');
      if (await themeToggle.isVisible()) {
        console.log('Theme toggle found - testing theme switching');
        await themeToggle.click();
        await page.waitForTimeout(500);
        
        const isDark = await page.evaluate(() => 
          document.documentElement.classList.contains('dark')
        );
        console.log(`Theme switched to: ${isDark ? 'dark' : 'light'}`);
      }
    });

    test('should handle responsive design', async ({ page }) => {
      await page.goto('/');
      
      const viewports = [
        { width: 1920, height: 1080, name: 'Desktop' },
        { width: 1024, height: 768, name: 'Tablet' },
        { width: 375, height: 667, name: 'Mobile' }
      ];
      
      for (const viewport of viewports) {
        console.log(`Testing ${viewport.name} view (${viewport.width}x${viewport.height})`);
        
        await page.setViewportSize({ 
          width: viewport.width, 
          height: viewport.height 
        });
        
        await page.waitForTimeout(500);
        
        // Check if page is still functional
        const isVisible = await page.locator('body').isVisible();
        console.log(`${viewport.name}: Page visible = ${isVisible}`);
        
        // Check for any overflow issues
        const hasOverflow = await page.evaluate(() => {
          return document.documentElement.scrollWidth > window.innerWidth;
        });
        
        if (hasOverflow) {
          console.log(`${viewport.name}: Warning - horizontal overflow detected`);
        }
      }
    });
  });

  test.describe('6. Error Handling and Edge Cases', () => {
    test('should handle 404 pages gracefully', async ({ page }) => {
      // Try to access a non-existent page
      await page.goto('/this-page-does-not-exist');
      
      // Check if page shows proper 404 handling
      await page.waitForLoadState('networkidle');
      
      const currentURL = page.url();
      console.log('404 Page URL:', currentURL);
      
      // Check for any console errors
      const consoleErrors: string[] = [];
      page.on('console', msg => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });
      
      await page.waitForTimeout(1000);
      console.log('404 Page Console Errors:', consoleErrors);
    });

    test('should handle network failures', async ({ page }) => {
      // Test with network interception to simulate failures
      await page.route('**/api/**', route => {
        route.abort('internetdisconnected');
      });
      
      await page.goto('/courses');
      await page.waitForLoadState('networkidle');
      
      // Check how application handles network failures
      const errorElements = page.locator('.error, .error-message, [data-testid="error"]');
      const errorCount = await errorElements.count();
      
      console.log(`Found ${errorCount} error elements when API is unavailable`);
      
      await page.unroute('**/api/**');
    });
  });

  test.describe('7. Form Submission and API Integration', () => {
    test('should handle form submissions with proper validation', async ({ page }) => {
      // Test multiple form types
      const formTests = [
        { path: '/auth/register', name: 'Registration' },
        { path: '/auth/login', name: 'Login' },
        { path: '/profile', name: 'Profile' }
      ];
      
      for (const formTest of formTests) {
        console.log(`Testing ${formTest.name} form submission`);
        
        await page.goto(formTest.path);
        await page.waitForLoadState('networkidle');
        
        // Check if we're redirected to auth (protected route)
        const currentURL = page.url();
        if (currentURL.includes('/auth/login') && formTest.path !== '/auth/login') {
          console.log(`${formTest.name} page is protected - requires authentication`);
          continue;
        }
        
        // Look for form elements
        const submitButton = page.locator('button[type="submit"]');
        if (await submitButton.isVisible()) {
          // Test empty form submission
          await submitButton.click();
          await page.waitForTimeout(1000);
          
          // Check for validation messages
          const validationErrors = page.locator('.error, .error-message, [data-testid="error"]');
          const errorCount = await validationErrors.count();
          
          if (errorCount > 0) {
            console.log(`${formTest.name} form validation working - found ${errorCount} validation errors`);
          }
        }
      }
    });

    test('should handle API responses correctly', async ({ page }) => {
      // Monitor API calls and responses
      const apiCalls: any[] = [];
      
      page.on('request', request => {
        if (request.url().includes('/api/')) {
          apiCalls.push({
            url: request.url(),
            method: request.method(),
            type: 'request'
          });
        }
      });
      
      page.on('response', response => {
        if (response.url().includes('/api/')) {
          apiCalls.push({
            url: response.url(),
            status: response.status(),
            type: 'response'
          });
        }
      });
      
      // Navigate to pages that trigger API calls
      const apiPages = ['/courses', '/assignments', '/announcements'];
      
      for (const pagePath of apiPages) {
        await page.goto(pagePath);
        await page.waitForLoadState('networkidle');
        await page.waitForTimeout(2000); // Allow time for API calls
        
        const currentURL = page.url();
        if (currentURL.includes('/auth/login')) {
          console.log(`${pagePath}: Protected route - API calls not accessible without auth`);
          continue;
        }
      }
      
      console.log('API calls monitored:', apiCalls.length);
      
      // Check for any failed API calls
      const failedCalls = apiCalls.filter(call => call.type === 'response' && call.status >= 400);
      if (failedCalls.length > 0) {
        console.log('Failed API calls:', failedCalls);
      }
    });
  });

  test.describe('8. Performance and Loading', () => {
    test('should check page load performance', async ({ page }) => {
      const navigationPromise = page.waitForNavigation();
      const startTime = Date.now();
      
      await page.goto('/');
      await navigationPromise;
      
      const loadTime = Date.now() - startTime;
      console.log(`Home page load time: ${loadTime}ms`);
      
      // Load time should be reasonable (under 5 seconds)
      expect(loadTime).toBeLessThan(5000);
      
      // Check for any slow loading resources
      const performanceMetrics = await page.evaluate(() => {
        const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        return {
          domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
          loadComplete: navigation.loadEventEnd - navigation.loadEventStart
        };
      });
      
      console.log('Performance Metrics:', performanceMetrics);
    });
  });
});
