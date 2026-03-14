import { expect, test } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

const SEEDED_ANNOUNCEMENT_TITLE = "Demo Week Schedule";
const SEEDED_RUBRIC_NAME = "Presentation Rubric";
const SEEDED_RUBRIC_CRITERION = "Clarity";
const SEEDED_BOOKING_PURPOSE = "Discuss the case study feedback";
const SEEDED_BOOKING_NOTE = "Seeded booking for demo";

test.describe("Faculty operations workflows", () => {
  test("faculty can review seeded productivity tools data", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.faculty);
    diagnostics.reset();

    await page.goto("/faculty-tools");
    await expect(
      page.getByRole("heading", { name: /faculty productivity tools/i })
    ).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByRole("tab", { name: /bulk announcements/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /grading rubrics/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /office hours/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /bookings/i })).toBeVisible();

    await expect(page.getByText(SEEDED_ANNOUNCEMENT_TITLE, { exact: true })).toBeVisible();

    await page.getByRole("tab", { name: /grading rubrics/i }).click();
    await expect(page.getByText(SEEDED_RUBRIC_NAME, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_RUBRIC_CRITERION, { exact: true })).toBeVisible();

    await page.getByRole("tab", { name: /office hours/i }).click();
    await expect(page.getByText("Tuesday", { exact: true })).toBeVisible();
    await expect(page.getByText(/14:00/).first()).toBeVisible();
    await expect(page.getByText(/15:00/).first()).toBeVisible();
    await expect(page.getByRole("link", { name: /join meeting/i })).toBeVisible();

    await page.getByRole("tab", { name: /bookings/i }).click();
    await expect(page.getByText(SEEDED_BOOKING_PURPOSE, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_BOOKING_NOTE, { exact: true })).toBeVisible();
    await expect(page.getByText("confirmed", { exact: true }).first()).toBeVisible();

    diagnostics.assertClean("faculty productivity tools seeded coverage");
    await logout(page);
  });
});
