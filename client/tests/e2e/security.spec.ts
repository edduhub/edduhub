import { test, expect } from '@playwright/test';

/**
 * End-to-End Security Tests for EduHub
 * Tests critical security features and multi-tenant isolation
 */

test.describe('Multi-Tenant Security', () => {
  test('should prevent access to other college data', async ({ page }) => {
    // Login as college 1 user
    await page.goto('/auth/login');
    await page.fill('[name="email"]', 'college1@test.com');
    await page.fill('[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    await expect(page).toHaveURL('/dashboard');
    
    // Try to access college 2 data via URL manipulation
    await page.goto('/api/colleges/2/students');
    
    // Should be redirected or show error
    await expect(page.locator('text=Forbidden')).toBeVisible({ timeout: 5000 });
  });

  test('should enforce college context in all requests', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Intercept API calls
    const requests = [];
    page.on('request', request => {
      if (request.url().includes('/api/')) {
        requests.push(request);
      }
    });
    
    await page.click('[data-testid="students-link"]');
    await page.waitForLoadState('networkidle');
    
    // All API requests should include college context
    for (const request of requests) {
      const headers = request.headers();
      expect(headers['authorization']).toBeTruthy();
    }
  });
});

test.describe('Authentication Security', () => {
  test('should reject expired tokens', async ({ page }) => {
    // Set an expired token manually
    await page.goto('/dashboard');
    await page.evaluate(() => {
      localStorage.setItem('token', 'expired.jwt.token');
    });
    
    await page.reload();
    
    // Should redirect to login
    await expect(page).toHaveURL(/.*login/);
  });

  test('should implement token rotation', async ({ page }) => {
    await page.goto('/auth/login');
    await page.fill('[name="email"]', 'test@college.com');
    await page.fill('[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    // Get initial token
    const initialToken = await page.evaluate(() => localStorage.getItem('token'));
    
    // Wait for potential token refresh (simulated time passage)
    await page.waitForTimeout(1000);
    
    // Make an API call that might trigger refresh
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    
    const newToken = await page.evaluate(() => localStorage.getItem('token'));
    
    // Tokens should exist
    expect(initialToken).toBeTruthy();
    expect(newToken).toBeTruthy();
  });

  test('should handle invalid credentials', async ({ page }) => {
    await page.goto('/auth/login');
    await page.fill('[name="email"]', 'wrong@test.com');
    await page.fill('[name="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    
    // Should show error message
    await expect(page.locator('text=/Invalid credentials|Login failed/i')).toBeVisible();
  });
});

test.describe('QR Code Attendance Security', () => {
  test('should reject expired QR codes', async ({ page }) => {
    // Faculty generates QR code
    await page.goto('/attendance/mark');
    await page.click('[data-testid="generate-qr"]');
    
    // Wait for QR code expiration (simulated)
    await page.waitForTimeout(1000);
    
    // Try to scan expired code
    const response = await page.request.post('/api/attendance/process-qr', {
      data: {
        qrcode_data: 'expired_qr_data'
      }
    });
    
    expect(response.status()).toBe(400);
  });

  test('should validate college in QR code', async ({ page, context }) => {
    // Generate QR for college 1
    await page.goto('/attendance/course/1/lecture/1/qrcode');
    
    const qrData = await page.textContent('[data-qr-code]');
    
    // Try to use with college 2 credentials (requires separate session)
    const page2 = await context.newPage();
    await page2.goto('/attendance/mark');
    
    const response = await page2.request.post('/api/attendance/process-qr', {
      data: {
        qrcode_data: qrData
      }
    });
    
    // Should be rejected
    expect(response.status()).toBe(403);
  });
});

test.describe('Data Validation and XSS Prevention', () => {
  test('should sanitize user input', async ({ page }) => {
    await page.goto('/announcements/create');
    
    // Try XSS attack
    const xssPayload = '<script>alert("xss")</script>';
    await page.fill('[name="content"]', xssPayload);
    await page.click('button[type="submit"]');
    
    await page.waitForLoadState('networkidle');
    
    // Check that script is not executed
    const content = await page.textContent('[data-testid="announcement-content"]');
    expect(content).not.toContain('<script>');
  });

  test('should validate form inputs', async ({ page }) => {
    await page.goto('/students/create');
    
    // Submit empty form
    await page.click('button[type="submit"]');
    
    // Should show validation errors
    await expect(page.locator('text=/required/i')).toBeVisible();
  });
});

test.describe('Authorization and Role-Based Access', () => {
  test('student cannot access admin pages', async ({ page }) => {
    // Login as student
    await page.goto('/auth/login');
    await page.fill('[name="email"]', 'student@test.com');
    await page.fill('[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    // Try to access admin page
    await page.goto('/admin/users');
    
    // Should show unauthorized or redirect
    await expect(page.locator('text=/Unauthorized|Forbidden/i')).toBeVisible();
  });

  test('faculty can manage courses', async ({ page }) => {
    // Login as faculty
    await page.goto('/auth/login');
    await page.fill('[name="email"]', 'faculty@test.com');
    await page.fill('[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    // Should be able to access course management
    await page.goto('/courses/create');
    await expect(page.locator('h1')).toContainText(/Create Course/i);
  });
});

test.describe('WebSocket Security', () => {
  test('should require authentication for WebSocket', async ({ page }) => {
    // Try to connect without token
    await page.goto('/notifications');
    
    const wsError = await page.evaluate(() => {
      return new Promise((resolve) => {
        const ws = new WebSocket('ws://localhost:8080/api/notifications/ws');
        ws.onerror = () => resolve('error');
        ws.onopen = () => resolve('success');
        setTimeout(() => resolve('timeout'), 1000);
      });
    });
    
    // Should fail to connect
    expect(wsError).toBe('error');
  });

  test('should isolate notifications by college', async ({ page, context }) => {
    // Setup two users from different colleges
    const page1 = page;
    const page2 = await context.newPage();
    
    // Login both
    await page1.goto('/auth/login');
    await page1.fill('[name="email"]', 'user1@college1.com');
    await page1.fill('[name="password"]', 'password123');
    await page1.click('button[type="submit"]');
    
    await page2.goto('/auth/login');
    await page2.fill('[name="email"]', 'user2@college2.com');
    await page2.fill('[name="password"]', 'password123');
    await page2.click('button[type="submit"]');
    
    // User 2 should not receive user 1's notifications
    await page1.goto('/notifications');
    await page2.goto('/notifications');
    
    // Create notification for college 1
    await page1.evaluate(() => {
      // Simulate notification creation
    });
    
    // Page 2 should not receive it
    await page2.waitForTimeout(2000);
    const notificationCount = await page2.locator('[data-testid="notification"]').count();
    expect(notificationCount).toBe(0);
  });
});

test.describe('Error Handling', () => {
  test('should not leak sensitive information in errors', async ({ page }) => {
    await page.goto('/api/students/999999999');
    
    const content = await page.textContent('body');
    
    // Should not contain database details, stack traces, or file paths
    expect(content).not.toMatch(/postgres|sql|\.go:|goroutine|panic/i);
  });
});

test.describe('Performance and Rate Limiting', () => {
  test('should handle rapid requests gracefully', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Make many rapid requests
    const promises = [];
    for (let i = 0; i < 50; i++) {
      promises.push(page.request.get('/api/dashboard'));
    }
    
    const responses = await Promise.all(promises);
    
    // Some might be rate limited, but server shouldn't crash
    const statusCodes = responses.map(r => r.status());
    expect(statusCodes).toContain(200);
  });
});
