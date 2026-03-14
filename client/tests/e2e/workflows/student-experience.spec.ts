import { expect, test } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

const SEEDED_ASSIGNMENT_TITLE = "Architecture Case Study";
const SEEDED_CALENDAR_EVENT_TITLE = "Demo Stakeholder Review";
const SEEDED_NOTIFICATION_TITLE = "New Exam Result Available";
const SEEDED_NOTIFICATION_MESSAGE = "Your practical exam result has been published.";

test.describe("Student experience workflows", () => {
  test("student can review assignments, generate a grade report, and inspect the seeded calendar", async ({
    page,
  }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/assignments");
    await expect(page.getByRole("heading", { name: "Assignments", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_ASSIGNMENT_TITLE, { exact: true })).toBeVisible();
    await expect(page.getByText(/view and submit your assignments/i)).toBeVisible();
    await expect(page.getByText("Total Assignments", { exact: true })).toBeVisible();
    diagnostics.assertClean("student assignments page");

    await page.goto("/grades");
    await expect(page.getByRole("heading", { name: "Grades", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText("Course Grades", { exact: true })).toBeVisible();
    await expect(page.getByText("Recent Grades", { exact: true })).toBeVisible();

    const gradeReportResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/api/reports/students/me/gradecard") &&
        response.request().method() === "GET" &&
        response.status() === 200
    );

    await page.getByRole("button", { name: /generate report/i }).click();
    await gradeReportResponse;
    diagnostics.assertClean("student grades report generation");

    await page.goto("/calendar");
    await expect(page.getByRole("heading", { name: "Calendar", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_CALENDAR_EVENT_TITLE, { exact: true })).toBeVisible();
    await expect(page.getByText("Main Conference Room", { exact: true })).toBeVisible();
    await expect(page.getByText("Upcoming Events", { exact: true })).toBeVisible();
    diagnostics.assertClean("student calendar review");

    await logout(page);
  });

  test("student can inspect notifications and use settings password validation", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/notifications");
    await expect(page.getByRole("heading", { name: "Notifications", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_NOTIFICATION_TITLE, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_NOTIFICATION_MESSAGE, { exact: true })).toBeVisible();

    await page.getByRole("button", { name: /filter:/i }).click();
    await expect(page.getByText(/no info notifications found/i)).toBeVisible();
    await page.getByRole("button", { name: /filter:/i }).click();
    await expect(page.getByText(SEEDED_NOTIFICATION_TITLE, { exact: true })).toBeVisible();
    diagnostics.assertClean("student notifications filtering");

    await page.goto("/settings");
    await expect(page.getByRole("heading", { name: "Settings", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText("Notifications", { exact: true }).first()).toBeVisible();
    await expect(page.getByText("Security", { exact: true })).toBeVisible();
    await expect(page.getByText("Preferences", { exact: true })).toBeVisible();

    await page.getByRole("button", { name: /update password/i }).click();
    await expect(page.getByRole("heading", { name: "Update Password", exact: true })).toBeVisible();

    await page.getByLabel("Current Password").fill("CurrentPassword1");
    await page.getByLabel("New Password").fill("NewPassword123");
    await page.getByLabel("Confirm New Password").fill("MismatchPassword123");
    await page.getByRole("button", { name: /update password/i }).last().click();

    await expect(page.getByText("New password and confirmation do not match", { exact: true })).toBeVisible();
    diagnostics.assertClean("student settings password validation");

    await logout(page);
  });
});
