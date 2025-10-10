const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function safeFetch<T>(path: string, fallback: T): Promise<T> {
  try {
    const res = await fetch(`${API_BASE}${path}`, { cache: "no-store" });
    if (!res.ok) {
      return fallback;
    }
    return (await res.json()) as T;
  } catch (error) {
    console.warn(`API fallback for ${path}`, error);
    return fallback;
  }
}

export type DashboardResponse = {
  metrics: {
    totalStudents: number;
    totalCourses: number;
    attendanceRate: number;
    announcements: number;
  };
  upcomingEvents: { id: number; title: string; start: string; course?: string }[];
  recentActivity: { id: number; entity: string; message: string; timestamp: string }[];
};

export async function fetchDashboard(): Promise<DashboardResponse> {
  const fallback: DashboardResponse = {
    metrics: {
      totalStudents: 1280,
      totalCourses: 54,
      attendanceRate: 87,
      announcements: 5,
    },
    upcomingEvents: [
      { id: 1, title: "Semester kickoff", start: new Date().toISOString(), course: "All" },
      { id: 2, title: "Data Structures Quiz", start: new Date(Date.now() + 864e5).toISOString(), course: "CS201" },
    ],
    recentActivity: [
      { id: 1, entity: "Attendance", message: "Marked attendance for CS201", timestamp: new Date().toISOString() },
      { id: 2, entity: "Grades", message: "Final exam grades released", timestamp: new Date(Date.now() - 36e5).toISOString() },
    ],
  };

  return safeFetch("/api/dashboard", fallback);
}

export type StudentRecord = {
  id: number;
  name: string;
  rollNo: string;
  courseCount: number;
  status: "active" | "inactive";
};

export async function fetchStudents(): Promise<StudentRecord[]> {
  const fallback: StudentRecord[] = [
    { id: 101, name: "Aarav Sharma", rollNo: "CS-23-001", courseCount: 5, status: "active" },
    { id: 102, name: "Mira Patel", rollNo: "CS-23-002", courseCount: 4, status: "active" },
    { id: 103, name: "Rahul Singh", rollNo: "CS-23-010", courseCount: 3, status: "inactive" },
  ];
  return safeFetch("/api/students", fallback);
}

export type CourseSummary = {
  id: number;
  name: string;
  instructor: string;
  enrollment: number;
  nextLecture?: string;
};

export async function fetchCourses(): Promise<CourseSummary[]> {
  const fallback: CourseSummary[] = [
    { id: 401, name: "Machine Learning", instructor: "Dr. Kapoor", enrollment: 82, nextLecture: "Tomorrow" },
    { id: 305, name: "Database Systems", instructor: "Prof. Rao", enrollment: 76, nextLecture: "In 2 days" },
  ];
  return safeFetch("/api/courses", fallback);
}

export type AnnouncementRecord = {
  id: number;
  title: string;
  priority: "low" | "normal" | "high" | "urgent";
  course?: string;
  publishedAt: string;
};

export async function fetchAnnouncements(): Promise<AnnouncementRecord[]> {
  const fallback: AnnouncementRecord[] = [
    { id: 1, title: "Labs rescheduled", priority: "high", course: "CS301", publishedAt: new Date().toISOString() },
    { id: 2, title: "Hackathon registrations open", priority: "normal", publishedAt: new Date(Date.now() - 864e5).toISOString() },
  ];
  return safeFetch("/api/announcements", fallback);
}

export type CalendarEvent = {
  id: number;
  title: string;
  start: string;
  end: string;
  course?: string;
  location?: string;
};

export async function fetchCalendar(): Promise<CalendarEvent[]> {
  const base = Date.now();
  const fallback: CalendarEvent[] = [
    { id: 1, title: "Guest Lecture", start: new Date(base + 2 * 864e5).toISOString(), end: new Date(base + 2 * 864e5 + 7200000).toISOString(), course: "CS401" },
    { id: 2, title: "Midterm Exams", start: new Date(base + 7 * 864e5).toISOString(), end: new Date(base + 8 * 864e5).toISOString() },
  ];
  return safeFetch("/api/calendar", fallback);
}

export type AnalyticsSummary = {
  label: string;
  value: number;
  delta?: number;
};

export async function fetchAnalytics(): Promise<AnalyticsSummary[]> {
  const fallback: AnalyticsSummary[] = [
    { label: "Average GPA", value: 3.4, delta: 0.1 },
    { label: "Attendance", value: 88, delta: 2 },
    { label: "Assignments Submitted", value: 92 },
  ];
  return safeFetch("/api/analytics", fallback);
}
