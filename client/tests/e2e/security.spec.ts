import { expect, test, type Page } from "@playwright/test";

import { API_BASE, attachDiagnostics, DEMO_USERS, getRoleHomePath, logout } from "./fixtures/auth";

type UiLoginExpectation = {
  label: string;
  email: string;
  password: string;
  expectedPath: string;
  landingHeading: string | RegExp;
};

const UI_LOGIN_USERS: UiLoginExpectation[] = [
  {
    label: "admin",
    email: DEMO_USERS.admin.email,
    password: DEMO_USERS.admin.password,
    expectedPath: getRoleHomePath("admin"),
    landingHeading: "Admin Dashboard",
  },
  {
    label: "faculty",
    email: DEMO_USERS.faculty.email,
    password: DEMO_USERS.faculty.password,
    expectedPath: getRoleHomePath("faculty"),
    landingHeading: /welcome back/i,
  },
  {
    label: "student",
    email: DEMO_USERS.student.email,
    password: DEMO_USERS.student.password,
    expectedPath: getRoleHomePath("student"),
    landingHeading: "Student Dashboard",
  },
  {
    label: "parent",
    email: DEMO_USERS.parent.email,
    password: DEMO_USERS.parent.password,
    expectedPath: getRoleHomePath("parent"),
    landingHeading: "Parent Portal",
  },
];

async function loginViaUI(page: Page, email: string, password: string) {
  await page.goto("/auth/login", { waitUntil: "domcontentloaded" });
  await expect(page.getByRole("heading", { name: /welcome to edduhub/i })).toBeVisible();

  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(password);

  const loginResponse = page.waitForResponse(
    (response) =>
      response.url().includes("/auth/login") &&
      response.request().method() === "POST" &&
      response.status() === 200
  );

  await page.getByRole("button", { name: /sign in/i }).click();
  await loginResponse;
}

test.describe("Authentication and access control", () => {
  test("unauthenticated users are redirected to login from protected routes", async ({ page }) => {
    await page.goto("/settings", { waitUntil: "domcontentloaded" });

    await expect
      .poll(() => new URL(page.url()).pathname, {
        message: "expected protected route access to redirect to login",
      })
      .toBe("/auth/login");

    await expect(page.getByRole("heading", { name: /welcome to edduhub/i })).toBeVisible();
  });

  test("login form rejects invalid credentials", async ({ page }) => {
    await page.goto("/auth/login", { waitUntil: "domcontentloaded" });

    await page.getByLabel("Email").fill("wrong.user@edduhub.local");
    await page.getByLabel("Password").fill("wrong-password");

    const loginResponse = page.waitForResponse(
      (response) =>
        response.url().includes("/auth/login") &&
        response.request().method() === "POST" &&
        response.status() >= 400
    );

    await page.getByRole("button", { name: /sign in/i }).click();
    await loginResponse;

    await expect(
      page.getByText(/login failed|invalid credentials|failed to login/i)
    ).toBeVisible();
    await expect(page).toHaveURL(/\/auth\/login$/);
  });

  test("invalid stored auth is cleared and redirected back to login", async ({ page }) => {
    await page.context().addCookies([
      {
        name: "edduhub_access_token",
        value: "invalid-demo-token",
        url: API_BASE,
        path: "/",
        httpOnly: true,
        secure: false,
        sameSite: "Strict",
      },
    ]);

    await page.goto("/student-dashboard", { waitUntil: "domcontentloaded" });

    await expect
      .poll(() => new URL(page.url()).pathname, {
        message: "expected invalid stored auth to be rejected",
      })
      .toBe("/auth/login");

    await expect(page.getByRole("heading", { name: /welcome to edduhub/i })).toBeVisible();
  });

  test("verify-email without required parameters shows an error state", async ({ page }) => {
    await page.goto("/auth/verify-email", { waitUntil: "domcontentloaded" });

    await expect(page.getByRole("heading", { name: "Email Verification" })).toBeVisible();
    await expect(page.getByText(/invalid verification link/i)).toBeVisible();
    await expect(page.getByRole("link", { name: /back to sign in/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /create account/i })).toBeVisible();
  });

  for (const user of UI_LOGIN_USERS) {
    test(`${user.label} can sign in through the login UI`, async ({ page }) => {
      const diagnostics = attachDiagnostics(page);

      await loginViaUI(page, user.email, user.password);
      await page.waitForLoadState("networkidle");

      await expect
        .poll(() => new URL(page.url()).pathname, {
          message: `expected ${user.label} to land on ${user.expectedPath}`,
        })
        .toBe(user.expectedPath);

      await expect(page.getByRole("heading", { name: user.landingHeading })).toBeVisible();
      diagnostics.assertClean(`${user.label} ui login flow`);

      await logout(page);
    });
  }
});
