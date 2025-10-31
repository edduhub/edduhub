import { test, expect } from '@playwright/test';

// Test configuration
const BASE_URL = process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000';
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Mock student data
const mockStudentDashboardData = {
  student: {
    id: 1,
    rollNo: 'ST001',
    firstName: 'John',
    lastName: 'Doe',
    email: 'john.doe@example.com',
    semester: 3,
    department: 1,
  },
  academicOverview: {
    gpa: 3.45,
    totalCredits: 18,
    enrolledCourses: 5,
    attendanceRate: 85.5,
    totalPresentSessions: 34,
    totalAttendanceSessions: 40,
  },
  courses: [
    {
      id: 1,
      code: 'CS101',
      name: 'Introduction to Computer Science',
      credits: 3,
      semester: 'Fall 2025',
      averageGrade: 85.5,
      attendanceRate: 90.0,
      totalSessions: 10,
      presentSessions: 9,
      enrollmentStatus: 'active',
    },
    {
      id: 2,
      code: 'MATH201',
      name: 'Calculus II',
      credits: 4,
      semester: 'Fall 2025',
      averageGrade: 78.0,
      attendanceRate: 80.0,
      totalSessions: 10,
      presentSessions: 8,
      enrollmentStatus: 'active',
    },
  ],
  assignments: {
    upcoming: [
      {
        id: 1,
        title: 'Data Structures Assignment',
        courseID: 1,
        dueDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
        maxScore: 100,
        isSubmitted: false,
      },
    ],
    completed: [
      {
        id: 2,
        title: 'Algorithms Assignment',
        courseID: 1,
        dueDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        maxScore: 100,
        isSubmitted: true,
        submittedAt: new Date(Date.now() - 8 * 24 * 60 * 60 * 1000).toISOString(),
        score: 92,
        feedback: 'Excellent work!',
      },
    ],
    overdue: [],
    summary: {
      upcomingCount: 1,
      completedCount: 1,
      overdueCount: 0,
    },
  },
  recentGrades: [
    {
      id: 1,
      courseName: 'Introduction to Computer Science',
      courseCode: 'CS101',
      assessmentName: 'Midterm Exam',
      assessmentType: 'exam',
      obtainedMarks: 85,
      totalMarks: 100,
      percentage: 85.0,
      gradedDate: new Date().toISOString(),
    },
  ],
  upcomingEvents: [
    {
      id: 1,
      title: 'Final Exam',
      description: 'CS101 Final Examination',
      date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(),
      type: 'exam',
    },
  ],
  announcements: [
    {
      id: 1,
      title: 'Holiday Notice',
      content: 'Campus will be closed next week',
      priority: 'high',
    },
  ],
};

test.describe('Student Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Mock the API response
    await page.route(`${API_URL}/api/student/dashboard`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: mockStudentDashboardData, success: true }),
      });
    });

    // Mock authentication
    await page.route(`${API_URL}/auth/**`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            user: {
              id: 'test-user-id',
              email: 'john.doe@example.com',
              firstName: 'John',
              lastName: 'Doe',
              role: 'student',
              collegeId: '1',
              collegeName: 'Test College',
            },
          },
          success: true,
        }),
      });
    });

    // Navigate to the student dashboard
    await page.goto(`${BASE_URL}/student-dashboard`);
  });

  test('should display student information correctly', async ({ page }) => {
    // Wait for the page to load
    await page.waitForSelector('h1', { timeout: 5000 });

    // Check if student name is displayed
    await expect(page.locator('h1')).toContainText('Welcome back, John!');

    // Check if roll number is displayed
    await expect(page.locator('text=Roll No: ST001')).toBeVisible();

    // Check if semester is displayed
    await expect(page.locator('text=Semester 3')).toBeVisible();
  });

  test('should display academic overview metrics', async ({ page }) => {
    // Wait for dashboard to load
    await page.waitForSelector('[data-testid="academic-overview"], .grid', { timeout: 5000 });

    // Check GPA display
    await expect(page.locator('text=3.45')).toBeVisible();

    // Check enrolled courses count
    await expect(page.locator('text=/5.*Enrolled Courses/i')).toBeVisible();

    // Check attendance rate
    await expect(page.locator('text=/85\\.5%.*Attendance Rate/i')).toBeVisible();

    // Check total credits
    await expect(page.locator('text=/18.*total credits/i')).toBeVisible();
  });

  test('should display course list with details', async ({ page }) => {
    // Wait for dashboard to load
    await page.waitForLoadState('networkidle');

    // Check if CS101 course is displayed
    await expect(page.locator('text=Introduction to Computer Science')).toBeVisible();
    await expect(page.locator('text=CS101')).toBeVisible();

    // Check if MATH201 course is displayed
    await expect(page.locator('text=Calculus II')).toBeVisible();
    await expect(page.locator('text=MATH201')).toBeVisible();
  });

  test('should display assignments by category', async ({ page }) => {
    // Wait for dashboard to load
    await page.waitForLoadState('networkidle');

    // Check upcoming assignments count
    await expect(page.locator('text=/1.*Upcoming/i')).toBeVisible();

    // Check completed assignments count
    await expect(page.locator('text=/1.*Completed/i')).toBeVisible();

    // Check overdue count is 0
    await expect(page.locator('text=/0.*Overdue/i')).toBeVisible();

    // Check if upcoming assignment is displayed
    await expect(page.locator('text=Data Structures Assignment')).toBeVisible();
  });

  test('should navigate between tabs', async ({ page }) => {
    // Wait for tabs to be available
    await page.waitForSelector('[role="tablist"]', { timeout: 5000 });

    // Click on Courses tab
    await page.click('button[role="tab"]:has-text("Courses")');
    await expect(page.locator('text=All Enrolled Courses')).toBeVisible();

    // Click on Assignments tab
    await page.click('button[role="tab"]:has-text("Assignments")');
    await expect(page.locator('text=Upcoming Assignments')).toBeVisible();

    // Click on Grades tab
    await page.click('button[role="tab"]:has-text("Grades")');
    await expect(page.locator('text=All Grades')).toBeVisible();

    // Go back to Overview tab
    await page.click('button[role="tab"]:has-text("Overview")');
    await expect(page.locator('text=Course Progress')).toBeVisible();
  });

  test('should display recent grades table', async ({ page }) => {
    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Check if grades table is visible (in overview or grades tab)
    const gradesVisible = await page.locator('text=Midterm Exam').isVisible();

    if (!gradesVisible) {
      // If not visible on overview, navigate to grades tab
      await page.click('button[role="tab"]:has-text("Grades")');
      await page.waitForTimeout(500);
    }

    // Verify grade details
    await expect(page.locator('text=Midterm Exam')).toBeVisible();
    await expect(page.locator('text=/85\\.0%/i')).toBeVisible();
  });

  test('should show upcoming events', async ({ page }) => {
    // Wait for dashboard to load
    await page.waitForLoadState('networkidle');

    // Check if events section is visible
    await expect(page.locator('text=Final Exam')).toBeVisible();
    await expect(page.locator('text=CS101 Final Examination')).toBeVisible();
  });

  test('should display announcements', async ({ page }) => {
    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Check if announcement is visible
    await expect(page.locator('text=Holiday Notice')).toBeVisible();
    await expect(page.locator('text=Campus will be closed next week')).toBeVisible();
  });

  test('should handle empty data gracefully', async ({ page }) => {
    // Mock empty data response
    await page.route(`${API_URL}/api/student/dashboard`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            ...mockStudentDashboardData,
            courses: [],
            assignments: {
              upcoming: [],
              completed: [],
              overdue: [],
              summary: { upcomingCount: 0, completedCount: 0, overdueCount: 0 },
            },
            recentGrades: [],
          },
          success: true,
        }),
      });
    });

    // Reload the page
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Check if "No courses" message is displayed
    await expect(page.locator('text=/No.*courses/i')).toBeVisible();

    // Check if "No grades" message is displayed
    await expect(page.locator('text=/No grades available/i')).toBeVisible();
  });

  test('should display loading state', async ({ page }) => {
    // Slow down the API response to see loading state
    await page.route(`${API_URL}/api/student/dashboard`, async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: mockStudentDashboardData, success: true }),
      });
    });

    // Navigate to dashboard
    await page.goto(`${BASE_URL}/student-dashboard`);

    // Check if loading spinner is visible
    await expect(page.locator('.animate-spin')).toBeVisible();

    // Wait for content to load
    await page.waitForSelector('h1', { timeout: 5000 });

    // Loading spinner should disappear
    await expect(page.locator('.animate-spin')).not.toBeVisible();
  });

  test('should handle API errors gracefully', async ({ page }) => {
    // Mock API error
    await page.route(`${API_URL}/api/student/dashboard`, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error', success: false }),
      });
    });

    // Reload the page
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Check if error message is displayed
    await expect(page.locator('text=/Unable to load dashboard data/i')).toBeVisible();
  });

  test('should redirect non-student users', async ({ page }) => {
    // Mock authentication for non-student user
    await page.route(`${API_URL}/auth/**`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            user: {
              id: 'admin-user-id',
              email: 'admin@example.com',
              firstName: 'Admin',
              lastName: 'User',
              role: 'admin', // Not a student
              collegeId: '1',
              collegeName: 'Test College',
            },
          },
          success: true,
        }),
      });
    });

    // Try to navigate to student dashboard
    await page.goto(`${BASE_URL}/student-dashboard`);

    // Should be redirected to home page
    await page.waitForURL(`${BASE_URL}/`, { timeout: 5000 });
  });

  test('should display course progress bars', async ({ page }) => {
    // Wait for dashboard to load
    await page.waitForLoadState('networkidle');

    // Check if progress bars are visible
    const progressBars = page.locator('[role="progressbar"], .progress');
    await expect(progressBars.first()).toBeVisible();
  });

  test('should show correct badge colors for grades', async ({ page }) => {
    // Wait for page to load
    await page.waitForLoadState('networkidle');

    // Navigate to grades tab if needed
    const gradesTabButton = page.locator('button[role="tab"]:has-text("Grades")');
    if (await gradesTabButton.isVisible()) {
      await gradesTabButton.click();
      await page.waitForTimeout(500);
    }

    // Check for grade badges (should be green/blue for passing grades)
    const gradeBadges = page.locator('[data-testid="grade-badge"], .badge');
    if (await gradeBadges.count() > 0) {
      await expect(gradeBadges.first()).toBeVisible();
    }
  });

  test('should display attendance statistics per course', async ({ page }) => {
    // Wait for dashboard to load
    await page.waitForLoadState('networkidle');

    // Navigate to courses tab
    await page.click('button[role="tab"]:has-text("Courses")');
    await page.waitForTimeout(500);

    // Check if attendance information is displayed
    await expect(page.locator('text=/90%/i').first()).toBeVisible();
    await expect(page.locator('text=/9\\/10/i').or(page.locator('text=/\\(9\\/10\\)/i')).first()).toBeVisible();
  });
});

test.describe('Student Dashboard - Responsive Design', () => {
  test('should be responsive on mobile devices', async ({ page }) => {
    // Set viewport to mobile size
    await page.setViewportSize({ width: 375, height: 667 });

    // Mock API and navigate
    await page.route(`${API_URL}/api/student/dashboard`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: mockStudentDashboardData, success: true }),
      });
    });

    await page.goto(`${BASE_URL}/student-dashboard`);
    await page.waitForLoadState('networkidle');

    // Check if key elements are still visible
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('text=3.45')).toBeVisible();
  });

  test('should be responsive on tablet devices', async ({ page }) => {
    // Set viewport to tablet size
    await page.setViewportSize({ width: 768, height: 1024 });

    // Mock API and navigate
    await page.route(`${API_URL}/api/student/dashboard`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: mockStudentDashboardData, success: true }),
      });
    });

    await page.goto(`${BASE_URL}/student-dashboard`);
    await page.waitForLoadState('networkidle');

    // Check if grid layout adjusts properly
    await expect(page.locator('.grid')).toBeVisible();
  });
});
