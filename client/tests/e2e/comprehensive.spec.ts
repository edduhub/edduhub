import { expect, test } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "./fixtures/auth";

const SEEDED_COURSE_NAME = "Foundations of Software Engineering";
const SEEDED_ASSIGNMENT_TITLE = "Architecture Case Study";
const SEEDED_ANNOUNCEMENT_TITLE = "Demo Week Schedule";
const SEEDED_ANNOUNCEMENT_CONTENT =
  "All demo stakeholders can use these seeded records to validate academic workflows.";
const SEEDED_DEPARTMENT_NAME = "Computer Science";
const SEEDED_PLACEMENT_COMPANY = "Acme Systems";
const SEEDED_PLACEMENT_ROLE = "Software Engineer Intern";
const SEEDED_PARENT_EMAIL = "parent.demo@eduhub.local";
const SEEDED_STUDENT_NAME = "Demo Student";

test.describe("Expanded Seeded E2E Coverage", () => {
  test("student dashboard surfaces seeded academic data", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/student-dashboard");
    await expect(page.getByText(/welcome back/i)).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(/current gpa/i)).toBeVisible();
    await expect(page.getByText(SEEDED_COURSE_NAME, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_ASSIGNMENT_TITLE, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_ANNOUNCEMENT_TITLE, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_ANNOUNCEMENT_CONTENT, { exact: true })).toBeVisible();

    diagnostics.assertClean("student dashboard seeded coverage");
    await logout(page);
  });

  test("student can browse seeded courses, announcements, profile and timetable", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/courses");
    await expect(page.getByRole("heading", { name: "Courses", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_COURSE_NAME, { exact: true })).toBeVisible();
    await page.getByPlaceholder("Search courses...").fill("Foundations");
    await expect(page.getByText(SEEDED_COURSE_NAME, { exact: true })).toBeVisible();
    await page.getByRole("button", { name: "View Details" }).first().click();
    await expect(page.getByRole("heading", { name: SEEDED_COURSE_NAME, exact: true })).toBeVisible();
    await expect(page.getByText("Core software engineering concepts for the demo cohort", { exact: true })).toBeVisible();
    await page.getByRole("button", { name: "Close" }).click();
    await expect(page.getByRole("heading", { name: SEEDED_COURSE_NAME, exact: true })).toHaveCount(0);
    diagnostics.assertClean("courses seeded coverage");

    await page.goto("/announcements");
    await expect(page.getByRole("heading", { name: "Announcements", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_ANNOUNCEMENT_TITLE, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_ANNOUNCEMENT_CONTENT, { exact: true })).toBeVisible();
    diagnostics.assertClean("announcements seeded coverage");

    await page.goto("/profile");
    await expect(page.getByRole("heading", { name: "Profile", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText("STUDENT", { exact: true })).toBeVisible();
    await expect(page.getByText("Academic Information", { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_DEPARTMENT_NAME, { exact: true })).toBeVisible();
    await expect(page.getByText("1", { exact: true }).first()).toBeVisible();
    diagnostics.assertClean("profile seeded coverage");

    await page.goto("/timetable");
    await expect(page.getByRole("heading", { name: "Timetable", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_COURSE_NAME, { exact: true })).toBeVisible();
    await expect(page.getByText("A-101", { exact: true })).toBeVisible();
    await expect(page.getByText("A-102", { exact: true })).toBeVisible();
    diagnostics.assertClean("timetable seeded coverage");

    await logout(page);
  });

  test("parent portal shows linked student context and seeded announcements", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.parent);
    diagnostics.reset();

    await page.goto("/parent-portal");
    await expect(page.getByRole("heading", { name: "Parent Portal", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByRole("button", { name: /demo student/i })).toBeVisible();
    await expect(page.getByText(SEEDED_DEPARTMENT_NAME, { exact: true })).toBeVisible();

    await page.getByRole("tab", { name: /^attendance$/i }).click();
    await expect(page.getByText("Attendance Overview", { exact: true })).toBeVisible();
    await expect(page.getByText(/present/i).first()).toBeVisible();

    await page.getByRole("tab", { name: /^announcements$/i }).click();
    await expect(page.getByText("School Announcements", { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_ANNOUNCEMENT_TITLE, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_ANNOUNCEMENT_CONTENT, { exact: true })).toBeVisible();

    diagnostics.assertClean("parent portal seeded coverage");
    await logout(page);
  });

  test("admin can review seeded departments, placements and parent links", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/departments");
    await expect(page.getByRole("heading", { name: "Departments", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_DEPARTMENT_NAME, { exact: true })).toBeVisible();
    await page.getByRole("button", { name: "View Details" }).first().click();
    await expect(page.getByRole("heading", { name: SEEDED_DEPARTMENT_NAME, exact: true })).toBeVisible();
    await expect(page.getByText("Department of Computer Science and Engineering", { exact: true })).toBeVisible();
    diagnostics.assertClean("departments seeded coverage");

    await page.goto("/placements");
    await expect(page.getByRole("heading", { name: "Placements", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_PLACEMENT_COMPANY, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_PLACEMENT_ROLE, { exact: true })).toBeVisible();
    await page.getByRole("button", { name: /company statistics/i }).click();
    await expect(page.getByText(SEEDED_PLACEMENT_COMPANY, { exact: true })).toBeVisible();
    diagnostics.assertClean("placements seeded coverage");

    await page.goto("/parent-links");
    await expect(page.getByRole("heading", { name: "Parent-Student Links", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_PARENT_EMAIL, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_STUDENT_NAME, { exact: true })).toBeVisible();
    diagnostics.assertClean("parent links seeded coverage");

    await logout(page);
  });
});
