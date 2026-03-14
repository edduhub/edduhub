import { expect, test, type Page } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

test.describe("Fees Workflow", () => {
  test("student views seeded fees and downloads statement and invoice", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/fees");
    await expect(page.getByRole("heading", { name: "Fees & Payments", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText("Course Fee", { exact: true })).toBeVisible();
    await expect(page.getByText("Outstanding Balance", { exact: true })).toBeVisible();
    await expect(page.getByText("₹1,000", { exact: true }).first()).toBeVisible();
    await expect(page.getByText("₹500", { exact: true }).first()).toBeVisible();
    await expect(page.getByText("₹1,500", { exact: true }).first()).toBeVisible();

    await page.getByRole("button", { name: /download statement/i }).click();
    await expect(page.getByText("Statement downloaded successfully", { exact: true })).toBeVisible();

    await page.getByRole("tab", { name: /payment history/i }).click();
    await expect(page.getByText("demo-fee-payment-001", { exact: true })).toBeVisible();

    await page.getByRole("button", { name: "Download invoice for demo-fee-payment-001" }).click();
    await expect(page.getByText("Invoice downloaded", { exact: true })).toBeVisible();

    diagnostics.assertClean("fees seeded history and downloads");
    await logout(page);
  });

  test("student sees payment configuration error on secure pay", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/fees");
    await expect(page.getByRole("heading", { name: "Fees & Payments", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByRole("button", { name: /quick pay/i })).toBeDisabled();
    await feeCardAction(page, "Course Fee", "Pay Securely").click();
    await expect(page.getByText("Payment system is not configured. Please contact support.", { exact: true })).toBeVisible();

    diagnostics.assertClean("fees payment config guard");
    await logout(page);
  });
});

function feeCardAction(page: Page, title: string, action: string) {
  return page.locator(
    `xpath=//*[normalize-space(text())="${title}"]/ancestor::div[.//button[normalize-space()="${action}"]][1]//button[normalize-space()="${action}"]`
  );
}
