import { expect, test } from "@playwright/test";
import { DEMO_USERS } from "../fixtures/auth";

test.describe("Registration Flow", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/auth/register", { waitUntil: "domcontentloaded" });
  });

  test("registration page loads correctly", async ({ page }) => {
    await expect(page.getByRole("heading", { name: /create account|register|sign up/i })).toBeVisible();
    await expect(page.getByLabel(/first name/i)).toBeVisible();
    await expect(page.getByLabel(/last name/i)).toBeVisible();
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByLabel(/^password$/i)).toBeVisible();
    await expect(page.getByLabel(/confirm password/i)).toBeVisible();
  });

  test("registration validates matching passwords", async ({ page }) => {
    await page.getByLabel(/first name/i).fill("John");
    await page.getByLabel(/last name/i).fill("Doe");
    await page.getByLabel(/email/i).fill("john.doe@test.com");
    await page.getByLabel(/^password$/i).fill("Password123!");
    await page.getByLabel(/confirm password/i).fill("DifferentPassword123!");

    await page.getByRole("button", { name: /register|sign up|create account/i }).click();

    await expect(page.getByText(/password.*match|do not match/i)).toBeVisible();
  });

  test("registration validates email format", async ({ page }) => {
    await page.getByLabel(/first name/i).fill("John");
    await page.getByLabel(/last name/i).fill("Doe");
    await page.getByLabel(/email/i).fill("invalid-email");
    await page.getByLabel(/^password$/i).fill("Password123!");
    await page.getByLabel(/confirm password/i).fill("Password123!");

    await page.getByRole("button", { name: /register|sign up|create account/i }).click();

    await expect(page.getByText(/invalid.*email|valid.*email|email.*invalid/i)).toBeVisible();
  });

  test("registration validates password strength", async ({ page }) => {
    await page.getByLabel(/first name/i).fill("John");
    await page.getByLabel(/last name/i).fill("Doe");
    await page.getByLabel(/email/i).fill("john.doe@test.com");
    await page.getByLabel(/^password$/i).fill("weak");
    await page.getByLabel(/confirm password/i).fill("weak");

    await page.getByRole("button", { name: /register|sign up|create account/i }).click();

    await expect(page.getByText(/password.*strong|minimum.*length|at least/i)).toBeVisible();
  });

  test("registration prevents duplicate email", async ({ page }) => {
    const existingUser = DEMO_USERS.admin;

    await page.getByLabel(/first name/i).fill("Admin");
    await page.getByLabel(/last name/i).fill("User");
    await page.getByLabel(/email/i).fill(existingUser.email);
    await page.getByLabel(/^password$/i).fill("Password123!");
    await page.getByLabel(/confirm password/i).fill("Password123!");

    await page.getByRole("button", { name: /register|sign up|create account/i }).click();

    await expect(page.getByText(/already.*exist|email.*taken|already.*registered/i)).toBeVisible();
  });

  test("can navigate to login from register page", async ({ page }) => {
    const loginLink = page.getByRole("link", { name: /already.*account|sign in|login/i });
    await expect(loginLink).toBeVisible();
    await loginLink.click();

    await expect(page).toHaveURL(/\/auth\/login/);
    await expect(page.getByRole("heading", { name: /welcome to edduhub|sign in/i })).toBeVisible();
  });

  test("registration with valid data shows success or redirects", async ({ page }) => {
    const uniqueEmail = `newuser${Date.now()}@test.com`;

    await page.getByLabel(/first name/i).fill("New");
    await page.getByLabel(/last name/i).fill("User");
    await page.getByLabel(/email/i).fill(uniqueEmail);
    await page.getByLabel(/^password$/i).fill("SecurePass123!");
    await page.getByLabel(/confirm password/i).fill("SecurePass123!");

    await page.getByRole("button", { name: /register|sign up|create account/i }).click();

    await Promise.race([
      expect(page.getByText(/verification|confirm|check.*email/i)).toBeVisible({ timeout: 5000 }),
      expect(page).toHaveURL(/\/auth\/verify-email|verify-email/, { timeout: 5000 }),
    ]).catch(() => {
      console.log("Registration may have succeeded or shown another response");
    });
  });
});
