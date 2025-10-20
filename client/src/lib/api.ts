// This file is deprecated. Use api-client.ts instead.
// Kept for backwards compatibility but should be removed.

export type DashboardResponse = {
  metrics: {
    totalStudents: number;
    totalCourses: number;
    totalFaculty?: number;
    attendanceRate: number;
    announcements: number;
    pendingSubmissions?: number;
  };
  upcomingEvents: { id: number; title: string; start: string; course?: string }[];
  recentActivity: { id: number; entity: string; message: string; timestamp: string }[];
};

export type StudentRecord = {
  id: number;
  name: string;
  rollNo: string;
  courseCount: number;
  status: "active" | "inactive";
};

export type CourseSummary = {
  id: number;
  name: string;
  instructor: string;
  enrollment: number;
  nextLecture?: string;
};

export type AnnouncementRecord = {
  id: number;
  title: string;
  priority: "low" | "normal" | "high" | "urgent";
  course?: string;
  publishedAt: string;
};

export type CalendarEvent = {
  id: number;
  title: string;
  start: string;
  end: string;
  course?: string;
  location?: string;
};

export type AnalyticsSummary = {
  label: string;
  value: number;
  delta?: number;
};
