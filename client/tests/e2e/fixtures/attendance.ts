import type { APIRequestContext } from "@playwright/test";

import { DEMO_USERS, type DemoRole } from "./auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080";
const SEEDED_COURSE_NAME = "Foundations of Software Engineering";
const SEEDED_LECTURE_TITLE = "Sprint Planning and Estimation";
const AUTH_RETRY_DELAYS_MS = [1000, 2000, 4000, 8000, 12000];

type LoginEnvelope = {
  data?: {
    token?: string;
    user?: {
      collegeId?: string | number;
      college_id?: string | number;
    };
  };
};

type CourseSummary = {
  id: number;
  name?: string;
  course_name?: string;
  collegeId?: string | number;
  college_id?: string | number;
};

type LectureSummary = {
  id: number;
  title?: string;
  lecture_title?: string;
  collegeId?: string | number;
  college_id?: string | number;
};

type DataEnvelope<T> = {
  data?: T;
};

export type SeededAttendanceTarget = {
  courseId: number;
  courseName: string;
  lectureId: number;
  lectureTitle: string;
  collegeId: number;
};

export type AttendanceRecord = {
  id?: number;
  courseId?: number;
  courseName?: string;
  date?: string;
  status?: string;
};

export async function discoverSeededAttendanceTarget(
  request: APIRequestContext,
  role: DemoRole = "faculty"
): Promise<SeededAttendanceTarget | null> {
  const session = await loginAsRole(request, role);
  if (!session) {
    return null;
  }

  const courses = await fetchData<CourseSummary[]>(request, "/api/courses", session.token);
  if (!courses || courses.length === 0) {
    return null;
  }

  const course =
    courses.find((entry) => normalizeText(entry.name ?? entry.course_name) === SEEDED_COURSE_NAME.toLowerCase()) ||
    courses[0];
  const courseId = normalizeNumber(course.id);

  if (!courseId) {
    return null;
  }

  const lectures = await fetchData<LectureSummary[]>(
    request,
    `/api/courses/${courseId}/lectures`,
    session.token
  );
  if (!lectures || lectures.length === 0) {
    return null;
  }

  const lecture =
    lectures.find(
      (entry) => normalizeText(entry.title ?? entry.lecture_title) === SEEDED_LECTURE_TITLE.toLowerCase()
    ) || lectures[0];
  const lectureId = normalizeNumber(lecture.id);
  const collegeId = normalizeNumber(lecture.collegeId ?? lecture.college_id ?? course.collegeId ?? course.college_id);

  if (!lectureId || !collegeId) {
    return null;
  }

  return {
    courseId,
    courseName: course.name ?? course.course_name ?? `Course ${courseId}`,
    lectureId,
    lectureTitle: lecture.title ?? lecture.lecture_title ?? `Lecture ${lectureId}`,
    collegeId,
  };
}

export async function fetchStudentAttendance(
  request: APIRequestContext,
  role: DemoRole = "student"
): Promise<AttendanceRecord[]> {
  const session = await loginAsRole(request, role);
  if (!session) {
    return [];
  }

  return (await fetchData<AttendanceRecord[]>(request, "/api/attendance/student/me", session.token)) ?? [];
}

export function buildQRCodePayload(target: SeededAttendanceTarget, now = new Date()): string {
  const expiresAt = new Date(now.getTime() + 10 * 60 * 1000);

  return JSON.stringify({
    course_id: target.courseId,
    lecture_id: target.lectureId,
    college_id: target.collegeId,
    time_stamp: now.toISOString(),
    expires_at: expiresAt.toISOString(),
  });
}

function normalizeText(value?: string): string {
  return (value ?? "").trim().toLowerCase();
}

function normalizeNumber(value: string | number | undefined): number | null {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  }

  return null;
}

async function loginAsRole(
  request: APIRequestContext,
  role: DemoRole
): Promise<{ token: string; collegeId: number | null } | null> {
  const user = DEMO_USERS[role];
  let response = await request.post(`${API_BASE}/auth/login`, {
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
    response = await request.post(`${API_BASE}/auth/login`, {
      data: {
        email: user.email,
        password: user.password,
      },
    });
  }

  if (!response.ok()) {
    return null;
  }

  const payload = (await response.json()) as LoginEnvelope;
  const token = payload.data?.token;

  if (!token) {
    return null;
  }

  return {
    token,
    collegeId: normalizeNumber(payload.data?.user?.collegeId ?? payload.data?.user?.college_id),
  };
}

async function fetchData<T>(request: APIRequestContext, path: string, token: string): Promise<T | null> {
  const response = await request.get(`${API_BASE}${path}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!response.ok()) {
    return null;
  }

  const payload = (await response.json()) as DataEnvelope<T> | T;

  if (payload && typeof payload === "object" && "data" in payload) {
    return (payload as DataEnvelope<T>).data ?? null;
  }

  return payload as T;
}
