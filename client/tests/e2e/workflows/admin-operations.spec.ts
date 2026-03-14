import { expect, test, type Page } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

const SEEDED_USER_EMAIL = "student.demo@eduhub.local";
const SEEDED_WEBHOOK_URL = "https://hooks.example.com/eduhub-demo";
const SEEDED_COURSE_NAME = "Foundations of Software Engineering";

test.describe("Admin operations workflows", () => {
  test("admin can inspect user management and roles pages", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/users");
    await expect(page.getByRole("heading", { name: "User Management", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.getByPlaceholder("Search users...").fill(SEEDED_USER_EMAIL);
    await expect(page.getByText(SEEDED_USER_EMAIL, { exact: true })).toBeVisible();

    await userRowAction(page, SEEDED_USER_EMAIL, "View").click();
    await expect(page.getByRole("heading", { name: /demo student/i })).toBeVisible();
    await expect(page.getByText(SEEDED_USER_EMAIL, { exact: true })).toBeVisible();
    await page.getByRole("button", { name: "Close", exact: true }).click();
    diagnostics.assertClean("admin users coverage");

    await page.goto("/roles");
    await expect(page.getByRole("heading", { name: "Roles & Permissions", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByRole("tab", { name: /roles matrix/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /user assignments/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /available permissions/i })).toBeVisible();

    await page.getByRole("tab", { name: /user assignments/i }).click();
    await expect(page.getByText("User Role Assignments", { exact: true })).toBeVisible();
    await page.getByPlaceholder("Filter by name or email...").fill(SEEDED_USER_EMAIL);
    await expect(page.getByText(SEEDED_USER_EMAIL, { exact: true })).toBeVisible();

    await page.getByRole("tab", { name: /available permissions/i }).click();
    await expect(page.getByText("System Capabilities", { exact: true })).toBeVisible();
    await expect(page.getByText(/reference list of all available system actions and resources/i)).toBeVisible();
    diagnostics.assertClean("admin roles coverage");

    await logout(page);
  });

  test("admin can validate system status, seeded webhooks, and batch tooling", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/system-status");
    await expect(page.getByRole("heading", { name: "System Status", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText("Overall System Status", { exact: true })).toBeVisible();
    await expect(page.getByText("/health", { exact: true })).toBeVisible();
    await expect(page.getByText("/ready", { exact: true })).toBeVisible();
    await expect(page.getByText("/alive", { exact: true })).toBeVisible();
    diagnostics.assertClean("admin system status coverage");

    await page.goto("/webhooks");
    await expect(page.getByRole("heading", { name: "Webhook Management", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_WEBHOOK_URL, { exact: true })).toBeVisible();

    const webhookTestResponse = page.waitForResponse(
      (response) =>
        /\/api\/webhooks\/\d+\/test$/.test(response.url()) &&
        response.request().method() === "POST" &&
        response.status() === 200
    );

    await page.locator('button[title="Send test event"]').first().click();
    await webhookTestResponse;
    await expect(page.getByText("Test event sent successfully", { exact: true })).toBeVisible();
    diagnostics.assertClean("admin webhook coverage");

    await page.goto("/batch-operations");
    await expect(page.getByRole("heading", { name: "Batch Operations", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page
      .getByRole("button", { name: /download template/i })
      .first()
      .click();
    await expect(page.getByText("Student CSV template downloaded", { exact: true })).toBeVisible();

    await page.getByRole("tab", { name: /^grades$/i }).click();
    await expect(page.getByText("Select Course", { exact: true })).toBeVisible();
    await page.locator("#course-select").click();
    await expect(page.getByText(SEEDED_COURSE_NAME, { exact: true })).toBeVisible();
    await page.keyboard.press("Escape");

    await page.getByRole("tab", { name: /^enrollment$/i }).click();
    await expect(page.getByText("Enroll Students", { exact: true })).toBeVisible();
    await page.locator("#enrollment-course-select").click();
    await expect(page.getByText(SEEDED_COURSE_NAME, { exact: true })).toBeVisible();
    await page.keyboard.press("Escape");

    diagnostics.assertClean("admin batch operations coverage");
    await logout(page);
  });
});

function userRowAction(page: Page, email: string, action: string) {
  return page.locator(
    `xpath=//*[normalize-space(text())="${email}"]/ancestor::tr[1]//button[normalize-space()="${action}"]`
  );
}
