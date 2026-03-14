import { expect, test } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

const SEEDED_THREAD_TITLE = "How should we structure the service layer?";
const SEEDED_THREAD_CONTENT =
  "Looking for advice on separating handlers, services, and repositories.";
const SEEDED_ACCEPTED_REPLY =
  "Start with handlers for HTTP concerns, services for workflow logic, and repositories for persistence.";

test.describe("Forum Workflow", () => {
  test("student opens seeded discussion and posts a reply", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);
    const uniqueReply = `Playwright forum reply ${Date.now()}`;

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/forum");
    await expect(page.getByRole("heading", { name: "Discussion Forum", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_THREAD_TITLE, { exact: true })).toBeVisible();
    await page.getByRole("link", { name: SEEDED_THREAD_TITLE }).click();

    await expect(page).toHaveURL(/\/forum\/\d+$/);
    await expect(page.getByRole("heading", { name: SEEDED_THREAD_TITLE, exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_THREAD_CONTENT, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_ACCEPTED_REPLY, { exact: true })).toBeVisible();
    await expect(page.getByText("Accepted", { exact: true })).toBeVisible();

    const createReplyResponse = page.waitForResponse(
      (response) =>
        /\/api\/forum\/threads\/\d+\/replies$/.test(response.url()) &&
        response.request().method() === "POST" &&
        response.status() === 201
    );

    await page.getByPlaceholder("Write your reply...").fill(uniqueReply);
    await page.getByRole("button", { name: /^post reply$/i }).click();
    await createReplyResponse;

    await expect(page.getByText(uniqueReply, { exact: true })).toBeVisible();

    diagnostics.assertClean("forum seeded thread flow");
    await logout(page);
  });
});
