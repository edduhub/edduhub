import { expect, test, type Page } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

const SEEDED_EXAM_TITLE = "Midterm Practical";

test.describe("Exams Workflow", () => {
  test("student reviews seeded exam details and result", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/exams");
    await expect(page.getByRole("heading", { name: "Exams Portal", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_EXAM_TITLE, { exact: true })).toBeVisible();
    await examRowAction(page, SEEDED_EXAM_TITLE, "View").click();

    await expect(page.getByRole("heading", { name: SEEDED_EXAM_TITLE, exact: true })).toBeVisible();
    await expect(page.getByText(/practical/i).first()).toBeVisible();
    await expect(page.getByText(/scheduled/i).last()).toBeVisible();
    await expect(page.getByText("TBD", { exact: true }).last()).toBeVisible();
    await expect(page.getByText("100", { exact: true }).last()).toBeVisible();
    await expect(page.getByText("40", { exact: true }).last()).toBeVisible();
    await expect(page.getByText("Bring your laptop", { exact: true })).toBeVisible();
    await expect(page.getByText("Notebook", { exact: true })).toBeVisible();
    await expect(page.getByText("Grade: A", { exact: true })).toBeVisible();
    await expect(page.getByText("Marks: 88", { exact: true })).toBeVisible();
    await expect(page.getByText("Percentage: 88.0%", { exact: true })).toBeVisible();
    await expect(page.getByText("Result: pass", { exact: true })).toBeVisible();

    diagnostics.assertClean("exams seeded exam detail flow");
    await logout(page);
  });
});

function examRowAction(page: Page, title: string, action: string) {
  return page.locator(
    `xpath=//*[normalize-space(text())="${title}"]/ancestor::tr[1]//button[normalize-space()="${action}"]`
  );
}
