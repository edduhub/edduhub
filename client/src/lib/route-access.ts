import type { UserRole } from "./types";

export type AppRoutePath =
  | "/"
  | "/advanced-analytics"
  | "/analytics"
  | "/announcements"
  | "/assignments"
  | "/attendance"
  | "/audit-logs"
  | "/batch-operations"
  | "/calendar"
  | "/courses"
  | "/departments"
  | "/exams"
  | "/faculty-tools"
  | "/fees"
  | "/files"
  | "/forum"
  | "/grades"
  | "/notifications"
  | "/parent-links"
  | "/parent-portal"
  | "/placements"
  | "/profile"
  | "/quizzes"
  | "/roles"
  | "/self-service"
  | "/settings"
  | "/student-dashboard"
  | "/students"
  | "/system-status"
  | "/timetable"
  | "/users"
  | "/webhooks";

export type RouteAccessRule = {
  path: AppRoutePath;
  allowedRoles: readonly UserRole[];
};

const ALL_STAFF_ROLES = ["admin", "faculty", "super_admin"] as const satisfies readonly UserRole[];
const STUDENT_AND_STAFF_ROLES = ["student", ...ALL_STAFF_ROLES] as const satisfies readonly UserRole[];
const ALL_AUTHENTICATED_ROLES = [
  "student",
  "faculty",
  "admin",
  "super_admin",
  "parent",
] as const satisfies readonly UserRole[];

export const ROUTE_ACCESS_RULES: readonly RouteAccessRule[] = [
  { path: "/", allowedRoles: ALL_STAFF_ROLES },
  { path: "/student-dashboard", allowedRoles: ["student"] },
  { path: "/parent-portal", allowedRoles: ["parent"] },
  { path: "/profile", allowedRoles: ALL_AUTHENTICATED_ROLES },
  { path: "/notifications", allowedRoles: ALL_AUTHENTICATED_ROLES },
  { path: "/settings", allowedRoles: ALL_AUTHENTICATED_ROLES },
  { path: "/courses", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/assignments", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/quizzes", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/attendance", allowedRoles: ["student"] },
  { path: "/grades", allowedRoles: ["student"] },
  { path: "/announcements", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/calendar", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/timetable", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/files", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/fees", allowedRoles: ["student"] },
  { path: "/placements", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/exams", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/forum", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/self-service", allowedRoles: STUDENT_AND_STAFF_ROLES },
  { path: "/students", allowedRoles: ALL_STAFF_ROLES },
  { path: "/analytics", allowedRoles: ALL_STAFF_ROLES },
  { path: "/advanced-analytics", allowedRoles: ALL_STAFF_ROLES },
  { path: "/faculty-tools", allowedRoles: ALL_STAFF_ROLES },
  { path: "/departments", allowedRoles: ["admin", "super_admin"] },
  { path: "/batch-operations", allowedRoles: ["admin", "super_admin"] },
  { path: "/webhooks", allowedRoles: ["admin", "super_admin"] },
  { path: "/audit-logs", allowedRoles: ["admin", "super_admin"] },
  { path: "/system-status", allowedRoles: ["admin", "super_admin"] },
  { path: "/users", allowedRoles: ["admin", "super_admin"] },
  { path: "/parent-links", allowedRoles: ["admin", "super_admin"] },
  { path: "/roles", allowedRoles: ["admin", "super_admin"] },
];

export function getRoleHomePath(role: UserRole): AppRoutePath {
  switch (role) {
    case "student":
      return "/student-dashboard";
    case "parent":
      return "/parent-portal";
    case "faculty":
    case "admin":
    case "super_admin":
    default:
      return "/";
  }
}

export function canAccessPath(pathname: string, role?: UserRole | null): boolean {
  if (!role) {
    return false;
  }

  const rule = findRouteAccessRule(pathname);
  if (!rule) {
    return true;
  }

  return rule.allowedRoles.includes(role);
}

export function findRouteAccessRule(pathname: string): RouteAccessRule | undefined {
  const normalizedPath = normalizePathname(pathname);

  return [...ROUTE_ACCESS_RULES]
    .sort((left, right) => right.path.length - left.path.length)
    .find((rule) => matchesRoutePrefix(normalizedPath, rule.path));
}

export function getAccessibleRoutesForRole(role: UserRole): AppRoutePath[] {
  return ROUTE_ACCESS_RULES
    .filter((rule) => rule.allowedRoles.includes(role))
    .map((rule) => rule.path);
}

function normalizePathname(pathname: string): string {
  const normalized = pathname.trim() || "/";
  return normalized.endsWith("/") && normalized !== "/" ? normalized.slice(0, -1) : normalized;
}

function matchesRoutePrefix(pathname: string, routePath: AppRoutePath): boolean {
  if (routePath === "/") {
    return pathname === "/";
  }

  return pathname === routePath || pathname.startsWith(`${routePath}/`);
}
