import { expect, test } from "@playwright/test";
import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

test.describe("Analytics Workflows", () => {
  test("student analytics page displays metrics and tabs", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.faculty);
    diagnostics.reset();

    await page.goto("/analytics");
    await expect(page.getByRole("heading", { name: "Student Analytics", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    // Verify metric cards
    await expect(page.getByText("Average Score", { exact: true }).first()).toBeVisible();
    await expect(page.getByText("Attendance Rate", { exact: true }).first()).toBeVisible();
    await expect(page.getByText("Completion Rate", { exact: true }).first()).toBeVisible();
    await expect(page.getByText("Assessments", { exact: true }).first()).toBeVisible();

    // Verify predictive insights
    await expect(page.getByText("Insights & Recommendations", { exact: true })).toBeVisible();

    // Performance tab (default)
    await expect(page.getByText("Performance Overview", { exact: true })).toBeVisible();
    await expect(page.getByText("Performance Trends", { exact: true })).toBeVisible();

    // Attendance tab
    await page.getByRole("tab", { name: "Attendance" }).click();
    await expect(page.getByText("Attendance Trends", { exact: true })).toBeVisible();

    // Learning tab
    await page.getByRole("tab", { name: "Learning" }).click();
    await expect(page.getByText("Learning Engagement", { exact: true })).toBeVisible();
    await expect(page.getByText("Most Accessed Materials", { exact: true })).toBeVisible();
    await expect(page.getByText("Peak Activity Hours", { exact: true })).toBeVisible();

    // Progress tab
    await page.getByRole("tab", { name: "Progress" }).click();
    await expect(page.getByText("Course Progress", { exact: true })).toBeVisible();

    diagnostics.assertClean("analytics page tabs and metrics");

    // Export report
    await page.getByRole("button", { name: /export report/i }).click();
    // The export creates a blob download, no network request to wait for
    // Just verify no errors occur
    diagnostics.assertClean("analytics export report");

    await logout(page);
  });

  test("advanced analytics page loads for admin", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/advanced-analytics");
    await expect(page.getByRole("heading", { name: "Advanced Analytics", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    // Verify tab structure exists
    const tabLabels = [
      /student progression/i,
      /course engagement/i,
      /predictive/i,
      /learning/i,
      /performance/i,
      /comparative/i,
    ];

    for (const label of tabLabels) {
      await expect(page.getByRole("tab", { name: label })).toBeVisible();
    }

    diagnostics.assertClean("advanced analytics page load");
    await logout(page);
  });
});
