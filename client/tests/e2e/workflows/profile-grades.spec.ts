import { expect, test } from "@playwright/test";
import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

const SEEDED_COURSE_NAME = "Foundations of Software Engineering";
const SEEDED_DEPARTMENT = "Computer Science";
const SEEDED_PLACEMENT_COMPANY = "Acme Systems";
const SEEDED_PLACEMENT_ROLE = "Software Engineer Intern";

test.describe("Profile Workflow", () => {
  test("student views profile with academic information", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/profile");
    await expect(page.getByRole("heading", { name: "Profile", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText("STUDENT", { exact: true })).toBeVisible();
    await expect(page.getByText("Academic Information", { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_DEPARTMENT, { exact: true })).toBeVisible();

    diagnostics.assertClean("student profile page");
    await logout(page);
  });

  test("faculty views own profile", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.faculty);
    diagnostics.reset();

    await page.goto("/profile");
    await expect(page.getByRole("heading", { name: "Profile", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText("FACULTY", { exact: true })).toBeVisible();

    diagnostics.assertClean("faculty profile page");
    await logout(page);
  });
});

test.describe("Grades Workflow", () => {
  test("student views grades and generates grade report", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/grades");
    await expect(page.getByRole("heading", { name: "Grades", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText("Course Grades", { exact: true })).toBeVisible();
    await expect(page.getByText("Recent Grades", { exact: true })).toBeVisible();

    // Generate grade report
    const gradeReportResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/api/reports/students/me/gradecard") &&
        response.request().method() === "GET" &&
        response.status() === 200
    );

    await page.getByRole("button", { name: /generate report/i }).click();
    await gradeReportResponse;

    diagnostics.assertClean("student grades and report generation");
    await logout(page);
  });
});

test.describe("Placements Workflow", () => {
  test("admin views seeded placements and company statistics", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.admin);
    diagnostics.reset();

    await page.goto("/placements");
    await expect(page.getByRole("heading", { name: "Placements", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_PLACEMENT_COMPANY, { exact: true })).toBeVisible();
    await expect(page.getByText(SEEDED_PLACEMENT_ROLE, { exact: true })).toBeVisible();

    // View company statistics
    await page.getByRole("button", { name: /company statistics/i }).click();
    await expect(page.getByText(SEEDED_PLACEMENT_COMPANY, { exact: true })).toBeVisible();

    diagnostics.assertClean("placements page with seeded data");
    await logout(page);
  });
});

test.describe("Timetable Workflow", () => {
  test("student views seeded timetable", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/timetable");
    await expect(page.getByRole("heading", { name: "Timetable", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(SEEDED_COURSE_NAME, { exact: true })).toBeVisible();
    await expect(page.getByText("A-101", { exact: true })).toBeVisible();
    await expect(page.getByText("A-102", { exact: true })).toBeVisible();

    diagnostics.assertClean("student timetable page");
    await logout(page);
  });
});
