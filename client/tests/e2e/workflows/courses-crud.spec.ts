import { expect, test } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

const SEEDED_COURSE_NAME = "Foundations of Software Engineering";
const SEEDED_ANNOUNCEMENT_TITLE = "Demo Week Schedule";
const SEEDED_CALENDAR_EVENT = "Demo Stakeholder Review";

test.describe("Courses CRUD Workflow", () => {
  test("faculty searches courses and views course details", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.faculty);
    diagnostics.reset();

    await page.goto("/courses");
    await expect(page.getByRole("heading", { name: "Courses", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_COURSE_NAME, { exact: true })).toBeVisible();

    // Search for course
    await page.getByPlaceholder("Search courses...").fill("Foundations");
    await expect(page.getByText(SEEDED_COURSE_NAME, { exact: true })).toBeVisible();

    // View details
    await page.getByRole("button", { name: "View Details" }).first().click();
    await expect(page.getByRole("heading", { name: SEEDED_COURSE_NAME, exact: true })).toBeVisible();
    await page.getByRole("button", { name: "Close" }).click();

    diagnostics.assertClean("courses search and details");
    await logout(page);
  });

  test("faculty creates a new course", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);
    const courseName = `Playwright Course ${Date.now()}`;

    await login(page, DEMO_USERS.faculty);
    diagnostics.reset();

    await page.goto("/courses");
    await expect(page.getByRole("heading", { name: "Courses", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.getByRole("button", { name: /create course|add course|new course/i }).click();

    await page.fill('input[id="name"], input[name="name"]', courseName);
    await page.fill(
      'textarea[id="description"], textarea[name="description"], input[id="description"]',
      "Created by Playwright E2E test."
    );

    const createCourseResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/api/courses") &&
        response.request().method() === "POST" &&
        (response.status() === 200 || response.status() === 201)
    );

    await page.getByRole("button", { name: /^create$|^save$|^submit$/i }).click();
    await createCourseResponse;

    await expect(page.getByText(courseName, { exact: true })).toBeVisible();
    diagnostics.assertClean("course creation");

    await logout(page);
  });
});

test.describe("Announcements CRUD Workflow", () => {
  test("admin views seeded announcements and creates a new one", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);
    const announcementTitle = `Playwright Announcement ${Date.now()}`;

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/announcements");
    await expect(
      page.getByRole("heading", { name: "Announcements", exact: true })
    ).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_ANNOUNCEMENT_TITLE, { exact: true })).toBeVisible();

    // Create new announcement
    await page.getByRole("button", { name: /create|new|add/i }).first().click();

    await page.fill('input[id="title"], input[name="title"]', announcementTitle);
    await page.fill(
      'textarea[id="content"], textarea[name="content"]',
      "Automated E2E test announcement content."
    );

    const createResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/api/announcements") &&
        response.request().method() === "POST" &&
        (response.status() === 200 || response.status() === 201)
    );

    await page.getByRole("button", { name: /^publish$|^create$|^submit$|^save$/i }).click();
    await createResponse;

    await expect(page.getByText(announcementTitle, { exact: true })).toBeVisible();
    diagnostics.assertClean("announcement creation");

    await logout(page);
  });
});

test.describe("Calendar Workflow", () => {
  test("admin views seeded calendar events", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/calendar");
    await expect(page.getByRole("heading", { name: "Calendar", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_CALENDAR_EVENT, { exact: true })).toBeVisible();
    await expect(page.getByText("Upcoming Events", { exact: true })).toBeVisible();

    diagnostics.assertClean("calendar page with seeded events");
    await logout(page);
  });

  test("admin creates a new calendar event", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);
    const eventTitle = `Playwright Event ${Date.now()}`;

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/calendar");
    await expect(page.getByRole("heading", { name: "Calendar", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.getByRole("button", { name: /create|add|new/i }).first().click();

    await page.fill('input[id="title"], input[name="title"]', eventTitle);

    const createResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/api/calendar") &&
        response.request().method() === "POST" &&
        (response.status() === 200 || response.status() === 201)
    );

    await page.getByRole("button", { name: /^create$|^save$|^submit$/i }).click();
    await createResponse;

    await expect(page.getByText(eventTitle, { exact: true })).toBeVisible();
    diagnostics.assertClean("calendar event creation");

    await logout(page);
  });
});
