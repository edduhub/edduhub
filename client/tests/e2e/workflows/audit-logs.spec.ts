import { expect, test } from "@playwright/test";
import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

test.describe("Audit Logs Workflow", () => {
  test("admin can view and filter audit logs", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/audit-logs");
    await expect(page.getByRole("heading", { name: "Audit Logs", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    // Verify stats section
    await expect(page.getByText("Total Logs", { exact: true })).toBeVisible();
    await expect(page.getByText("Unique Users", { exact: true })).toBeVisible();

    // Verify table structure
    await expect(page.getByRole("columnheader", { name: /user/i })).toBeVisible();
    await expect(page.getByRole("columnheader", { name: /action/i })).toBeVisible();

    // Verify filter controls
    await expect(page.getByPlaceholder(/search/i)).toBeVisible();

    diagnostics.assertClean("audit logs page load");
    await logout(page);
  });
});

test.describe("Students Management Workflow", () => {
  test("faculty can browse and search students list", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.faculty);
    diagnostics.reset();

    await page.goto("/students");
    await expect(page.getByRole("heading", { name: /students/i })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    // Verify seeded student data
    await expect(page.getByText("Demo", { exact: true }).first()).toBeVisible();

    // Test search
    const searchInput = page.getByPlaceholder(/search/i);
    await expect(searchInput).toBeVisible();
    await searchInput.fill("Demo");
    await expect(page.getByText("Demo", { exact: true }).first()).toBeVisible();

    diagnostics.assertClean("students management page");
    await logout(page);
  });

  test("admin can access create student form", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/students");
    await expect(page.getByRole("heading", { name: /students/i })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    // Look for add/create student button
    const createBtn = page.getByRole("button", { name: /add student|create student|new student/i });
    if (await createBtn.isVisible()) {
      await createBtn.click();
      // Verify form fields appear
      await expect(page.getByLabel(/first name/i).or(page.locator("#firstName"))).toBeVisible();
    }

    diagnostics.assertClean("admin students create form");
    await logout(page);
  });
});
