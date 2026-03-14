import { expect, type Page } from "@playwright/test";

export type DemoRole = "admin" | "faculty" | "student" | "parent";

export type DemoUser = {
  role: DemoRole;
  email: string;
  password: string;
};

export const DEMO_PASSWORD = "EduHub#2026!LocalSeed$A7q2";
const AUTH_RETRY_DELAYS_MS = [1000, 2000, 4000, 8000, 12000];
export const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080";
const SESSION_COOKIE_NAMES = ["edduhub_access_token", "edduhub_refresh_token", "edduhub_session_token"] as const;
type AuthCookie = {
  name: string;
  value: string;
};
const SESSION_CACHE = new Map<DemoRole, AuthCookie[]>();

export const DEMO_USERS: Record<DemoRole, DemoUser> = {
  admin: {
    role: "admin",
    email: "admin.demo@eduhub.local",
    password: DEMO_PASSWORD,
  },
  faculty: {
    role: "faculty",
    email: "faculty.demo@eduhub.local",
    password: DEMO_PASSWORD,
  },
  student: {
    role: "student",
    email: "student.demo@eduhub.local",
    password: DEMO_PASSWORD,
  },
  parent: {
    role: "parent",
    email: "parent.demo@eduhub.local",
    password: DEMO_PASSWORD,
  },
};

export function getRoleHomePath(role: DemoRole): string {
  switch (role) {
    case "student":
      return "/student-dashboard";
    case "parent":
      return "/parent-portal";
    case "faculty":
    case "admin":
    default:
      return "/";
  }
}

export async function login(page: Page, user: DemoUser): Promise<void> {
  const authCookies = SESSION_CACHE.get(user.role) ?? (await fetchSession(page, user));
  SESSION_CACHE.set(user.role, authCookies);
  await page.context().addCookies(
    authCookies.map((cookie) => ({
      name: cookie.name,
      value: cookie.value,
      url: API_BASE,
      path: "/",
      httpOnly: true,
      secure: false,
      sameSite: "Strict",
    }))
  );
  await page.goto("/auth/login", { waitUntil: "domcontentloaded" });

  const targetPath = getRoleHomePath(user.role);
  await page.goto(targetPath, { waitUntil: "domcontentloaded" });
  await expect
    .poll(
      () => new URL(page.url()).pathname,
      { message: `expected ${user.role} to land on ${targetPath}` }
    )
    .toBe(targetPath);
}

export async function logout(page: Page): Promise<void> {
  await page.request.post(`${API_BASE}/auth/logout`, {
    headers: { "Content-Type": "application/json" },
  });
  await page.context().clearCookies();

  await page.goto("/auth/login", { waitUntil: "domcontentloaded" });
  await expect(page).toHaveURL(/\/auth\/login$/);
}

async function fetchSession(page: Page, user: DemoUser): Promise<AuthCookie[]> {
	let response = await page.context().request.post(`${API_BASE}/auth/login`, {
		data: {
			email: user.email,
			password: user.password,
		},
	});

	for (const delayMs of AUTH_RETRY_DELAYS_MS) {
		if (response.ok() || response.status() !== 429) {
			break;
		}

		await new Promise((resolve) => setTimeout(resolve, delayMs));
		response = await page.context().request.post(`${API_BASE}/auth/login`, {
			data: {
				email: user.email,
				password: user.password,
			},
		});
	}

	expect(response.ok(), `expected login API to succeed for ${user.role}`).toBeTruthy();
  const cookieHeaders = response.headersArray().filter((header) => header.name.toLowerCase() === "set-cookie");

  const cookies = cookieHeaders
    .map((header) => {
      const [firstPair] = header.value.split(";");
      const separatorIndex = firstPair.indexOf("=");
      if (separatorIndex <= 0) return null;
      return {
        name: firstPair.substring(0, separatorIndex),
        value: firstPair.substring(separatorIndex + 1),
      };
    })
    .filter((entry): entry is AuthCookie => Boolean(entry))
    .filter((entry) => SESSION_COOKIE_NAMES.includes(entry.name as (typeof SESSION_COOKIE_NAMES)[number]));

  expect(cookies.length, `expected auth cookies for ${user.role}`).toBeGreaterThan(0);
  return cookies;
}

export type Diagnostics = {
  consoleErrors: string[];
  pageErrors: string[];
  networkFailures: string[];
  unexpectedResponses: string[];
  reset: () => void;
  assertClean: (label: string) => void;
};


export function attachDiagnostics(page: Page): Diagnostics {
  const consoleErrors: string[] = [];
  const pageErrors: string[] = [];
  const networkFailures: string[] = [];
  const unexpectedResponses: string[] = [];

  page.on("console", (message) => {
    if (message.type() === "error") {
      consoleErrors.push(message.text());
    }
  });

  page.on("pageerror", (error) => {
    pageErrors.push(error.message);
  });

  page.on("requestfailed", (request) => {
    if (request.url().startsWith(API_BASE)) {
      networkFailures.push(
        `${request.method()} ${request.url()} ${request.failure()?.errorText || "failed"}`
      );
    }
  });

  page.on("response", (response) => {
    if (response.url().startsWith(API_BASE) && response.status() >= 400) {
      unexpectedResponses.push(
        `${response.status()} ${response.request().method()} ${response.url()}`
      );
    }
  });

  return {
    consoleErrors,
    pageErrors,
    networkFailures,
    unexpectedResponses,
    reset: () => {
      consoleErrors.length = 0;
      pageErrors.length = 0;
      networkFailures.length = 0;
      unexpectedResponses.length = 0;
    },
    assertClean: (label: string) => {
      expect(
        {
          consoleErrors,
          pageErrors,
          networkFailures,
          unexpectedResponses,
        },
        label
      ).toEqual({
        consoleErrors: [],
        pageErrors: [],
        networkFailures: [],
        unexpectedResponses: [],
      });
    },
  };
}
