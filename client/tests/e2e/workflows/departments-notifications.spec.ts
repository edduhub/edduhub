import { expect, test } from "@playwright/test";
import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

const SEEDED_DEPARTMENT = "Computer Science";
const SEEDED_DEPARTMENT_DESC = "Department of Computer Science and Engineering";
const SEEDED_NOTIFICATION_TITLE = "New Exam Result Available";
const SEEDED_NOTIFICATION_MSG = "Your practical exam result has been published.";

test.describe("Departments CRUD Workflow", () => {
  test("admin views and inspects seeded department", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/departments");
    await expect(page.getByRole("heading", { name: "Departments", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_DEPARTMENT, { exact: true })).toBeVisible();

    await page.getByRole("button", { name: "View Details" }).first().click();
    await expect(page.getByRole("heading", { name: SEEDED_DEPARTMENT, exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_DEPARTMENT_DESC, { exact: true })).toBeVisible();

    diagnostics.assertClean("departments view and inspect");
    await logout(page);
  });

  test("admin creates a new department", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);
    const deptName = `Playwright Dept ${Date.now()}`;

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/departments");
    await expect(page.getByRole("heading", { name: "Departments", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.getByRole("button", { name: /create|add|new/i }).first().click();

    await page.fill('input[id="name"], input[name="name"]', deptName);
    await page.fill('textarea[id="description"], textarea[name="description"], input[id="description"]', "Created by Playwright E2E test.");

    const createResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/api/departments") &&
        response.request().method() === "POST" &&
        (response.status() === 200 || response.status() === 201)
    );

    await page.getByRole("button", { name: /^create$|^save$|^submit$/i }).click();
    await createResponse;

    await expect(page.getByText(deptName, { exact: true })).toBeVisible();
    diagnostics.assertClean("department creation");

    await logout(page);
  });
});

test.describe("Notifications Deep Workflow", () => {
  test("student views seeded notifications and uses mark all read", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/notifications");
    await expect(page.getByRole("heading", { name: "Notifications", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_NOTIFICATION_TITLE, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_NOTIFICATION_MSG, { exact: true })).toBeVisible();

    // Mark all as read
    const markAllBtn = page.getByRole("button", { name: /mark all.*(read|as read)/i });
    if (await markAllBtn.isVisible()) {
      const markAllResponse = page.waitForResponse(
        (response) =>
          response.url().includes("/api/notifications/mark-all-read") &&
          response.request().method() === "POST" &&
          response.status() === 200
      );

      await markAllBtn.click();
      await markAllResponse;
    }

    // Test filtering
    await page.getByRole("button", { name: /filter:/i }).click();
    await page.getByRole("button", { name: /filter:/i }).click();
    await expect(page.getByText(SEEDED_NOTIFICATION_TITLE, { exact: true })).toBeVisible();

    diagnostics.assertClean("notifications deep workflow");
    await logout(page);
  });
});

test.describe("Parent Portal Deep Workflow", () => {
  test("parent navigates through all child dashboard tabs", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.parent);
    diagnostics.reset();

    await page.goto("/parent-portal");
    await expect(page.getByRole("heading", { name: "Parent Portal", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    // Verify linked child
    await expect(page.getByRole("button", { name: /demo student/i })).toBeVisible();
    await expect(page.getByText(SEEDED_DEPARTMENT, { exact: true })).toBeVisible();

    // Attendance tab
    await page.getByRole("tab", { name: /^attendance$/i }).click();
    await expect(page.getByText("Attendance Overview", { exact: true })).toBeVisible();

    // Grades tab
    await page.getByRole("tab", { name: /^grades$/i }).click();
    await expect(page.getByText(/grades/i).first()).toBeVisible();

    // Assignments tab
    await page.getByRole("tab", { name: /^assignments$/i }).click();
    await expect(page.getByText(/assignment/i).first()).toBeVisible();

    // Announcements tab
    await page.getByRole("tab", { name: /^announcements$/i }).click();
    await expect(page.getByText("School Announcements", { exact: true })).toBeVisible();

    diagnostics.assertClean("parent portal all tabs");
    await logout(page);
  });
});
