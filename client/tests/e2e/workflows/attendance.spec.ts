import { expect, test } from "@playwright/test";
import { DEMO_USERS, login, logout, attachDiagnostics } from "../fixtures/auth";
import { buildQRCodePayload, discoverSeededAttendanceTarget } from "../fixtures/attendance";

test.describe("Attendance Workflow", () => {
  test("faculty generates QR and student marks attendance", async ({ page, request }) => {
    const diagnostics = attachDiagnostics(page);
    const facultyUser = DEMO_USERS.faculty;
    const studentUser = DEMO_USERS.student;

    const target = await discoverSeededAttendanceTarget(request, "faculty");
    expect(target).not.toBeNull();
    const { courseId, lectureId } = target!;

    await login(page, facultyUser);
    diagnostics.reset();

    await page.goto("/attendance");
    await expect(page.getByRole("heading", { name: "Attendance", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.getByRole("button", { name: /generate qr code/i }).click();
    await expect(page.getByRole("dialog")).toBeVisible();

    await page.fill('input[id="courseId"]', String(courseId));
    await page.fill('input[id="lectureId"]', String(lectureId));

    const qrGenerationPromise = page.waitForResponse(
      (response) =>
        response.url().includes(`/api/attendance/course/${courseId}/lecture/${lectureId}/qrcode`) &&
        response.status() === 200
    );
    await page.getByRole("button", { name: /generate qr/i }).click();

    const qrResponse = await qrGenerationPromise;
    expect(qrResponse.status()).toBe(200);

    const qrImage = page.locator('img[alt="Attendance QR Code"]');
    await expect(qrImage).toBeVisible({ timeout: 10000 });
    diagnostics.assertClean("faculty QR generation");

    const qrData = buildQRCodePayload(target!);
    await page.keyboard.press("Escape");
    await expect(page.getByRole("dialog")).not.toBeVisible();
    await logout(page);

    await login(page, studentUser);
    diagnostics.reset();

    await page.goto("/attendance");
    await expect(page.getByRole("heading", { name: "Attendance", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.getByRole("tab", { name: /scan qr/i }).click();
    await expect(page.getByRole("heading", { name: "Scan QR Code" })).toBeVisible();

    page.once("dialog", async (dialog) => {
      await dialog.accept(qrData);
    });

    const markAttendancePromise = page.waitForResponse(
      (response) =>
        response.url().includes("/api/attendance/process-qr") &&
        response.status() === 200
    );

    await page.getByRole("button", { name: /scan & mark attendance/i }).click();

    const markResponse = await markAttendancePromise;
    expect(markResponse.status()).toBe(200);

    await expect(page.getByText("Attendance marked successfully!")).toBeVisible({
      timeout: 10000,
    });

    await page.waitForLoadState("networkidle");
    diagnostics.assertClean("student attendance marked");

    await page.reload();
    await expect(page.getByRole("heading", { name: "Attendance", exact: true })).toBeVisible();
    await expect(page.getByText(target!.courseName, { exact: true })).toBeVisible();

    await page.getByRole("tab", { name: /records/i }).click();
    await expect(page.getByRole("cell", { name: target!.courseName }).first()).toBeVisible();

    await logout(page);
  });

  test("student views attendance records", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);
    const studentUser = DEMO_USERS.student;

    await login(page, studentUser);
    diagnostics.reset();

    await page.goto("/attendance");
    await expect(page.getByRole("heading", { name: "Attendance", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();

    await page.getByRole("tab", { name: /records/i }).click();

    const hasRecords =
      (await page.locator("table").count()) > 0 ||
      (await page.getByText("No attendance records found").count()) > 0;
    expect(hasRecords).toBe(true);

    await page.getByRole("tab", { name: /overview/i }).click();
    await expect(page.getByText("Overall Attendance")).toBeVisible();

    diagnostics.assertClean("student attendance records view");

    await logout(page);
  });

  test("faculty can access attendance page", async ({ page }) => {
    const diagnostics = attachDiagnostics(page);
    const facultyUser = DEMO_USERS.faculty;

    await login(page, facultyUser);
    diagnostics.reset();

    await page.goto("/attendance");
    await expect(page.getByRole("heading", { name: "Attendance", exact: true })).toBeVisible();
    await page.waitForLoadState("networkidle");
    diagnostics.reset();
    await expect(page.getByRole("button", { name: /generate qr code/i })).toBeVisible();

    diagnostics.assertClean("faculty attendance page");

    await logout(page);
  });
});
