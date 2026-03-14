import {
  canAccessPath,
  findRouteAccessRule,
  getAccessibleRoutesForRole,
  getRoleHomePath,
} from "@/lib/route-access";

describe("route access policy", () => {
  it("redirects students to the student dashboard by default", () => {
    expect(getRoleHomePath("student")).toBe("/student-dashboard");
  });

  it("redirects parents to the parent portal by default", () => {
    expect(getRoleHomePath("parent")).toBe("/parent-portal");
  });

  it("uses the most specific matching route prefix", () => {
    expect(findRouteAccessRule("/advanced-analytics/details")?.path).toBe("/advanced-analytics");
  });

  it("blocks students from admin routes", () => {
    expect(canAccessPath("/users", "student")).toBe(false);
    expect(canAccessPath("/roles", "student")).toBe(false);
    expect(canAccessPath("/webhooks", "student")).toBe(false);
  });

  it("blocks staff from student-only modules", () => {
    expect(canAccessPath("/attendance", "admin")).toBe(false);
    expect(canAccessPath("/grades", "faculty")).toBe(false);
    expect(canAccessPath("/fees", "admin")).toBe(false);
  });

  it("keeps parent access limited to parent-safe routes", () => {
    expect(canAccessPath("/parent-portal", "parent")).toBe(true);
    expect(canAccessPath("/notifications", "parent")).toBe(true);
    expect(canAccessPath("/courses", "parent")).toBe(false);
  });

  it("returns the full set of staff-only routes for admins", () => {
    const adminRoutes = getAccessibleRoutesForRole("admin");
    expect(adminRoutes).toContain("/users");
    expect(adminRoutes).toContain("/analytics");
    expect(adminRoutes).not.toContain("/parent-portal");
  });
});
