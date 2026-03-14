import { expect, test, type APIRequestContext, type Page } from "@playwright/test";

import { attachDiagnostics, DEMO_USERS, login, logout } from "../fixtures/auth";
import { discoverSeededAttendanceTarget } from "../fixtures/attendance";

test.describe("Quiz Workflow", () => {
  test("faculty creates a quiz and student completes it", async ({ page, request }) => {
    const diagnostics = attachDiagnostics(page);
    const target = await discoverSeededAttendanceTarget(request, "faculty");

    expect(target).not.toBeNull();

    const courseId = target!.courseId;
    const quizTitle = `Playwright Quiz ${Date.now()}`;
    const questionText = `What keeps a sprint plan testable for ${quizTitle}?`;
    const correctOption = `Stable fixtures for ${quizTitle}`;
    const alternateOption = `Random IDs for ${quizTitle}`;

    let createdQuizId: number | null = null;

    try {
      await login(page, DEMO_USERS.faculty);
      diagnostics.reset();

      await page.goto("/quizzes");
      await expect(page.getByRole("heading", { name: "Quizzes", exact: true })).toBeVisible();
      await page.waitForLoadState("networkidle");
      diagnostics.reset();

      await page.getByRole("button", { name: /^create quiz$/i }).click();
      await expect(page.getByText("New Quiz", { exact: true })).toBeVisible();

      await page.selectOption("#quiz-course-id", String(courseId));
      await page.fill("#quiz-title", quizTitle);
      await page.fill("#quiz-description", "Browser-first quiz workflow coverage.");
      await page.fill("#quiz-duration", "15");

      const createQuizResponse = page.waitForResponse(
        (response) =>
          response.url().includes(`/api/courses/${courseId}/quizzes`) &&
          response.request().method() === "POST" &&
          response.status() === 201
      );

      await page.getByRole("button", { name: /^create$/i }).click();

      const createQuizPayload = (await (await createQuizResponse).json()) as {
        data?: { id?: number };
      };
      createdQuizId = createQuizPayload.data?.id ?? null;

      await expect(page.getByText(quizTitle, { exact: true })).toBeVisible();

      await quizCardAction(page, quizTitle, "Questions").click();
      await expect(page.getByRole("heading", { name: "Quiz Questions" })).toBeVisible();
      await page.getByRole("button", { name: /^add question$/i }).last().click();
      await expect(page.getByRole("heading", { name: "Add Question" })).toBeVisible();

      await page.fill("#quiz-question-text", questionText);
      await page.selectOption("#quiz-question-type", "multiple_choice");
      await page.fill("#quiz-question-points", "5");
      await page.fill("#quiz-question-option-0", correctOption);
      await page.fill("#quiz-question-option-1", alternateOption);
      await page.getByRole("dialog").getByRole("radio").first().click();

      const createQuestionResponse = page.waitForResponse(
        (response) =>
          createdQuizId !== null &&
          response.url().includes(`/api/quizzes/${createdQuizId}/questions`) &&
          response.request().method() === "POST" &&
          response.status() === 201
      );

      await page.getByRole("button", { name: /^add question$/i }).last().click();
      await createQuestionResponse;

      await expect(page.getByText(questionText, { exact: true })).toBeVisible();
      await page.keyboard.press("Escape");
      await expect(page.getByRole("heading", { name: "Quiz Questions" })).not.toBeVisible();
      diagnostics.assertClean("faculty quiz creation");

      await logout(page);

      await login(page, DEMO_USERS.student);
      diagnostics.reset();

      await page.goto("/quizzes");
      await expect(page.getByRole("heading", { name: "Quizzes", exact: true })).toBeVisible();
      await page.waitForLoadState("networkidle");
      diagnostics.reset();
      await expect(page.getByText(quizTitle, { exact: true })).toBeVisible();

      const startAttemptResponse = page.waitForResponse(
        (response) =>
          createdQuizId !== null &&
          response.url().includes(`/api/quizzes/${createdQuizId}/attempts/start`) &&
          response.request().method() === "POST" &&
          response.status() === 201
      );

      await quizCardAction(page, quizTitle, "Start Quiz").click();
      await startAttemptResponse;

      await expect(page).toHaveURL(/\/quizzes\/\d+\/attempt\/\d+$/);
      await expect(page.getByText(questionText)).toBeVisible();
      await page.getByText(correctOption, { exact: true }).click();

      const submitAttemptResponse = page.waitForResponse(
        (response) =>
          /\/api\/attempts\/\d+\/submit$/.test(response.url()) &&
          response.request().method() === "POST" &&
          response.status() === 200
      );

      await page.getByRole("button", { name: /^submit attempt$/i }).click();
      await submitAttemptResponse;

      const finalScore = page.getByText(/Final Score:/);
      await expect(finalScore).toBeVisible();
      await expect(finalScore).toContainText("5");
      diagnostics.assertClean("student quiz completion");

      await logout(page);
    } finally {
      if (createdQuizId !== null) {
        try {
          await deleteQuiz(request, courseId, createdQuizId);
        } catch {
          // Cleanup should not hide the real test failure.
        }
      }
    }
  });
});

async function deleteQuiz(request: APIRequestContext, courseId: number, quizId: number): Promise<void> {
  const token = await loginAndGetToken(request, DEMO_USERS.faculty.email, DEMO_USERS.faculty.password);

  await request.delete(`http://127.0.0.1:8180/api/courses/${courseId}/quizzes/${quizId}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

async function loginAndGetToken(
  request: APIRequestContext,
  email: string,
  password: string
): Promise<string> {
  const response = await request.post("http://127.0.0.1:8180/auth/login", {
    data: { email, password },
  });

  expect(response.ok()).toBeTruthy();

  const payload = (await response.json()) as {
    data?: { token?: string };
  };

  expect(payload.data?.token).toBeTruthy();
  return payload.data!.token!;
}

function quizCardAction(page: Page, quizTitle: string, action: string) {
  return page.locator(
    `xpath=//*[normalize-space(text())="${quizTitle}"]/ancestor::div[.//button[normalize-space()="${action}"]][1]//button[normalize-space()="${action}"]`
  );
}
