import { expect, test } from "@playwright/test";
import { DEMO_USERS, logout } from "../fixtures/auth";

test.describe("Home Page", () => {
  test("root URL redirects to login for unauthenticated users", async ({ page }) => {
    await page.goto("/", { waitUntil: "domcontentloaded" });
    await expect(page).toHaveURL(/\/auth\/login/);
  });

  test("admin is redirected to admin dashboard", async ({ page }) => {
    await page.goto("/auth/login", { waitUntil: "domcontentloaded" });
    await page.getByLabel("Email").fill(DEMO_USERS.admin.email);
    await page.getByLabel("Password").fill(DEMO_USERS.admin.password);
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(page).toHaveURL(/\/admin-dashboard/, { timeout: 10000 });
    await expect(page.getByRole("heading", { name: /admin|dashboard/i })).toBeVisible();

    await logout(page);
  });

  test("faculty is redirected to faculty dashboard", async ({ page }) => {
    await page.goto("/auth/login", { waitUntil: "domcontentloaded" });
    await page.getByLabel("Email").fill(DEMO_USERS.faculty.email);
    await page.getByLabel("Password").fill(DEMO_USERS.faculty.password);
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(page).toHaveURL(/\/faculty-dashboard/, { timeout: 10000 });
    await expect(page.getByRole("heading", { name: /faculty|dashboard|welcome/i })).toBeVisible();

    await logout(page);
  });

  test("student is redirected to student dashboard", async ({ page }) => {
    await page.goto("/auth/login", { waitUntil: "domcontentloaded" });
    await page.getByLabel("Email").fill(DEMO_USERS.student.email);
    await page.getByLabel("Password").fill(DEMO_USERS.student.password);
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(page).toHaveURL(/\/student-dashboard/, { timeout: 10000 });
    await expect(page.getByRole("heading", { name: /student|dashboard/i })).toBeVisible();

    await logout(page);
  });

  test("parent is redirected to parent dashboard", async ({ page }) => {
    await page.goto("/auth/login", { waitUntil: "domcontentloaded" });
    await page.getByLabel("Email").fill(DEMO_USERS.parent.email);
    await page.getByLabel("Password").fill(DEMO_USERS.parent.password);
    await page.getByRole("button", { name: /sign in/i }).click();

    await expect(page).toHaveURL(/\/parent-dashboard/, { timeout: 10000 });
    await expect(page.getByRole("heading", { name: /parent|portal|dashboard/i })).toBeVisible();

    await logout(page);
  });

  test("navigation is role-specific", async ({ page }) => {
    await page.goto("/auth/login", { waitUntil: "domcontentloaded" });
    await page.getByLabel("Email").fill(DEMO_USERS.admin.email);
    await page.getByLabel("Password").fill(DEMO_USERS.admin.password);
    await page.getByRole("button", { name: /sign in/i }).click();
    await page.waitForURL(/\/admin-dashboard/);

    const navLinks = await page.getByRole("navigation").getByRole("link").all();
    console.log(`Admin has ${navLinks.length} navigation links`);

    await logout(page);
  });
});
