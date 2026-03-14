import { expect, test, type Page } from "@playwright/test";

import { attachDiagnostics, login, logout, type Diagnostics } from "./fixtures/auth";

import {
  getAccessibleRoutesForRole,
  getRoleHomePath,
  type AppRoutePath,
} from "../../src/lib/route-access";

type DemoRole = "admin" | "faculty" | "student" | "parent";

type DemoUser = {
  role: DemoRole;
  email: string;
  password: string;
  visibleNav: AppRoutePath[];
  hiddenNav: AppRoutePath[];
  forbiddenRoutes: AppRoutePath[];
};

const DEMO_PASSWORD = "EduHub#2026!LocalSeed$A7q2";

const DEMO_USERS: DemoUser[] = [
  {
    role: "admin",
    email: "admin.demo@eduhub.local",
    password: DEMO_PASSWORD,
    visibleNav: ["/", "/users", "/webhooks", "/notifications", "/self-service"],
    hiddenNav: ["/student-dashboard", "/parent-portal"],
    forbiddenRoutes: ["/student-dashboard", "/parent-portal"],
  },
  {
    role: "faculty",
    email: "faculty.demo@eduhub.local",
    password: DEMO_PASSWORD,
    visibleNav: ["/", "/students", "/analytics", "/faculty-tools", "/self-service"],
    hiddenNav: ["/users", "/roles", "/webhooks", "/parent-portal"],
    forbiddenRoutes: ["/users", "/roles", "/webhooks", "/parent-portal"],
  },
  {
    role: "student",
    email: "student.demo@eduhub.local",
    password: DEMO_PASSWORD,
    visibleNav: ["/courses", "/attendance", "/self-service"],
    hiddenNav: ["/users", "/roles", "/webhooks", "/advanced-analytics"],
    forbiddenRoutes: ["/users", "/roles", "/webhooks", "/advanced-analytics"],
  },
  {
    role: "parent",
    email: "parent.demo@eduhub.local",
    password: DEMO_PASSWORD,
    visibleNav: ["/parent-portal", "/profile", "/notifications", "/settings"],
    hiddenNav: ["/courses", "/students", "/users", "/advanced-analytics"],
    forbiddenRoutes: ["/courses", "/students", "/users", "/advanced-analytics"],
  },
];

async function verifySidebar(page: Page, user: DemoUser) {
  for (const href of user.visibleNav) {
    await expect(page.locator(`aside nav a[href="${href}"]`).first()).toBeVisible();
  }

  for (const href of user.hiddenNav) {
    await expect(page.locator(`aside nav a[href="${href}"]`)).toHaveCount(0);
  }
}

async function verifyAllowedRoutes(page: Page, diagnostics: Diagnostics, role: DemoRole) {
  for (const path of getAccessibleRoutesForRole(role)) {
    diagnostics.reset();
    await page.goto(path, { waitUntil: "domcontentloaded" });
    await page.waitForTimeout(750);

    await expect
      .poll(() => new URL(page.url()).pathname, {
        message: `expected ${role} to remain on allowed route ${path}`,
      })
      .toBe(path);

    await expect(page.getByText("Unhandled Runtime Error")).toHaveCount(0);
    diagnostics.assertClean(`allowed route ${path} for ${role}`);
  }
}

async function verifyForbiddenRoutes(page: Page, diagnostics: Diagnostics, user: DemoUser) {
  const homePath = getRoleHomePath(user.role);

  for (const path of user.forbiddenRoutes) {
    diagnostics.reset();
    await page.goto(path, { waitUntil: "domcontentloaded" });
    await page.waitForTimeout(750);

    await expect
      .poll(() => new URL(page.url()).pathname, {
        message: `expected ${user.role} to be redirected away from ${path}`,
      })
      .toBe(homePath);

    diagnostics.assertClean(`forbidden route ${path} for ${user.role}`);
  }
}

for (const user of DEMO_USERS) {
  test(`${user.role} demo flow is clean`, async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, user);
    diagnostics.reset();

    await verifySidebar(page, user);
    await verifyAllowedRoutes(page, diagnostics, user.role);
    await verifyForbiddenRoutes(page, diagnostics, user);
    await logout(page);
  });
}
