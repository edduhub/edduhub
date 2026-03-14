import { expect, test } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

test.describe("Self-Service Workflow", () => {
  test("student submits a request and admin approves it", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);
    const requestCode = `PW-${Date.now()}`;
    const requestTitle = `Enrollment Request: ${requestCode}`;
    const adminResponse = `Approved by Playwright ${requestCode}`;

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/self-service");
    await expect(page.getByRole("heading", { name: "Student Self-Service" })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.fill("#courseCode", requestCode);
    await page.fill("#reason", `Need this course for automated browser verification ${requestCode}`);
    await page.fill("#specialRequests", "Please process via Playwright workflow.");

    const createRequestPromise = page.waitForResponse(
      (response) =>
        response.url().includes("/api/self-service/requests") &&
        response.request().method() === "POST" &&
        response.status() === 201
    );

    await page.getByRole("button", { name: /^submit request$/i }).click();
    await createRequestPromise;

    await expect(page.getByText("Enrollment request submitted successfully.")).toBeVisible();

    await page.getByRole("tab", { name: /history/i }).click();
    await expect(page.getByText(requestTitle, { exact: true })).toBeVisible();
    await expect(page.getByText("pending", { exact: true }).first()).toBeVisible();
    diagnostics.assertClean("student self-service request creation");

    await logout(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/self-service");
    await expect(page.getByRole("heading", { name: "Student Self-Service" })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.getByRole("tab", { name: /manage/i }).click({ force: true });
    await expect(page.getByRole("heading", { name: "Manage Requests" })).toBeVisible();
    await expect(page.getByText(requestTitle, { exact: true })).toBeVisible();

    await page
      .locator(
        `xpath=//*[normalize-space(text())="${requestTitle}"]/ancestor::div[.//button[normalize-space()="Process"]][1]//button[normalize-space()="Process"]`
      )
      .click();

    await expect(page.getByRole("heading", { name: "Process Request" })).toBeVisible();
    await page.fill("#response", adminResponse);

    const updateRequestPromise = page.waitForResponse(
      (response) =>
        response.url().includes("/api/self-service/requests/") &&
        response.request().method() === "PUT" &&
        response.status() === 200
    );

    await page.getByRole("button", { name: /^approve$/i }).click();
    await updateRequestPromise;

    await expect(page.getByText(requestTitle, { exact: true })).toBeVisible();
    await expect(page.getByText(adminResponse, { exact: true })).toBeVisible();
    await expect(page.getByText("approved", { exact: true }).first()).toBeVisible();
    diagnostics.assertClean("admin self-service approval");

    await logout(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/self-service");
    await expect(page.getByRole("heading", { name: "Student Self-Service" })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();
    await page.getByRole("tab", { name: /history/i }).click();
    await expect(page.getByText(requestTitle, { exact: true })).toBeVisible();
    await expect(page.getByText(adminResponse, { exact: true })).toBeVisible();
    await expect(page.getByText("approved", { exact: true }).first()).toBeVisible();
    diagnostics.assertClean("student self-service history");

    await logout(page);
  });

  test("admin can access self-service management", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/self-service");
    await expect(page.getByRole("heading", { name: "Student Self-Service" })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();
    await expect(page.getByRole("tab", { name: /manage/i })).toBeVisible();
    diagnostics.assertClean("admin self-service access");

    await logout(page);
  });
});
