import { expect, test } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";

test.describe("Files Workflow", () => {
  test("student views seeded file versions and download link", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);
    const fileName = "demo-handbook";

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/files");
    await expect(page.getByRole("heading", { name: "File Management", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await expect(page.getByText(fileName, { exact: true })).toBeVisible();

    const versionsResponsePromise = page.waitForResponse(
      (response) =>
        /\/api\/file-management\/\d+\/versions$/.test(response.url()) &&
        response.request().method() === "GET" &&
        response.status() === 200
    );

    await page.getByRole("button", { name: `View versions for ${fileName}` }).click();
    await versionsResponsePromise;

    await expect(page.getByRole("heading", { name: `Version History - ${fileName}` })).toBeVisible();
    await expect(page.getByRole("dialog").getByText("v", { exact: true })).toBeVisible();
    await expect(page.getByText("Current", { exact: true })).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(page.getByRole("heading", { name: `Version History - ${fileName}` })).not.toBeVisible();

    await page.evaluate(() => {
      const openedUrls: string[] = [];
      window.open = ((url?: string | URL | undefined) => {
        if (url) {
          openedUrls.push(String(url));
        }
        return null;
      }) as typeof window.open;
      (window as typeof window & { __playwrightOpenedUrls?: string[] }).__playwrightOpenedUrls = openedUrls;
    });

    const downloadResponsePromise = page.waitForResponse(
      (response) =>
        /\/api\/file-management\/\d+\/download$/.test(response.url()) &&
        response.request().method() === "GET" &&
        response.status() === 200
    );

    await page.getByRole("button", { name: `Download ${fileName}` }).click();
    const downloadPayload = (await (await downloadResponsePromise).json()) as {
      data?: { url?: string };
    };

    const openedUrls = await page.evaluate(() => {
      return ((window as typeof window & { __playwrightOpenedUrls?: string[] }).__playwrightOpenedUrls ?? []).slice();
    });

    expect(downloadPayload.data?.url).toMatch(/^https?:\/\//);
    expect(openedUrls[0]).toBe(downloadPayload.data?.url);

    diagnostics.assertClean("files seeded file view");
    await logout(page);
  });

  test("student sees seeded demo folder", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);

    await login(page, DEMO_USERS.student);
    diagnostics.reset();

    await page.goto("/files");
    await expect(page.getByRole("heading", { name: "File Management", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.getByRole("tab", { name: /^folders$/i }).click();
    await expect(page.getByRole("heading", { name: "Folders" })).toBeVisible();
    await expect(page.getByText("demo", { exact: true })).toBeVisible();
    await expect(page.getByText(/files$/).first()).toBeVisible();

    diagnostics.assertClean("files seeded folder view");
    await logout(page);
  });
});
