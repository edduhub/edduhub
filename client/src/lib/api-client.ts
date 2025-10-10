// Enhanced API client with authentication support

import { AuthSession } from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const AUTH_STORAGE_KEY = 'edduhub_auth';

export class APIError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'APIError';
  }
}

function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null;
  
  try {
    const stored = localStorage.getItem(AUTH_STORAGE_KEY);
    if (stored) {
      const session: AuthSession = JSON.parse(stored);
      if (new Date(session.expiresAt) > new Date()) {
        return session.token;
      }
    }
  } catch (error) {
    console.error('Failed to get auth token:', error);
  }
  return null;
}

type RequestOptions = {
  method?: string;
  body?: any;
  headers?: Record<string, string>;
  requireAuth?: boolean;
};

export async function apiClient<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<T> {
  const { method = 'GET', body, headers = {}, requireAuth = true } = options;

  const token = getAuthToken();
  
  const requestHeaders: HeadersInit = {
    'Content-Type': 'application/json',
    ...headers,
  };

  if (requireAuth && token) {
    requestHeaders['Authorization'] = `Bearer ${token}`;
  }

  const config: RequestInit = {
    method,
    headers: requestHeaders,
    cache: 'no-store',
  };

  if (body) {
    config.body = JSON.stringify(body);
  }

  const response = await fetch(`${API_BASE}${endpoint}`, config);

  if (!response.ok) {
    let message = 'Request failed';
    try {
      const errorData = await response.json();
      message = errorData.message || errorData.error || message;
    } catch {
      message = response.statusText || message;
    }
    throw new APIError(response.status, message);
  }

  // Handle empty responses
  const contentType = response.headers.get('content-type');
  if (!contentType || !contentType.includes('application/json')) {
    return {} as T;
  }

  return response.json();
}

// Convenience methods
export const api = {
  get: <T>(endpoint: string, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'GET', requireAuth }),
  
  post: <T>(endpoint: string, data: any, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'POST', body: data, requireAuth }),
  
  put: <T>(endpoint: string, data: any, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'PUT', body: data, requireAuth }),
  
  patch: <T>(endpoint: string, data: any, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'PATCH', body: data, requireAuth }),
  
  delete: <T>(endpoint: string, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'DELETE', requireAuth }),
};

// API endpoints
export const endpoints = {
  // Auth
  auth: {
    login: '/api/auth/login',
    register: '/api/auth/register',
    logout: '/api/auth/logout',
    refresh: '/api/auth/refresh',
    profile: '/api/profile',
  },
  
  // Students
  students: {
    list: '/api/students',
    get: (id: number) => `/api/students/${id}`,
    create: '/api/students',
    update: (id: number) => `/api/students/${id}`,
    delete: (id: number) => `/api/students/${id}`,
    freeze: (id: number) => `/api/students/${id}/freeze`,
  },
  
  // Courses
  courses: {
    list: '/api/courses',
    get: (id: number) => `/api/courses/${id}`,
    create: '/api/courses',
    update: (id: number) => `/api/courses/${id}`,
    delete: (id: number) => `/api/courses/${id}`,
    enroll: (id: number) => `/api/courses/${id}/enroll`,
    students: (id: number) => `/api/courses/${id}/students`,
  },
  
  // Assignments
  assignments: {
    list: '/api/assignments',
    get: (id: number) => `/api/assignments/${id}`,
    create: '/api/assignments',
    update: (id: number) => `/api/assignments/${id}`,
    delete: (id: number) => `/api/assignments/${id}`,
    submit: (id: number) => `/api/assignments/${id}/submit`,
    submissions: (id: number) => `/api/assignments/${id}/submissions`,
    grade: (assignmentId: number, submissionId: number) =>
      `/api/assignments/${assignmentId}/submissions/${submissionId}/grade`,
  },
  
  // Attendance
  attendance: {
    mark: '/api/attendance/mark',
    markBulk: '/api/attendance/mark-bulk',
    generateQR: '/api/attendance/qr/generate',
    processQR: '/api/attendance/qr/process',
    byStudent: (studentId: number) => `/api/attendance/student/${studentId}`,
    byCourse: (courseId: number) => `/api/attendance/course/${courseId}`,
    update: (id: number) => `/api/attendance/${id}`,
  },
  
  // Grades
  grades: {
    byCourse: (courseId: number) => `/api/grades/course/${courseId}`,
    byStudent: (studentId: number) => `/api/grades/student/${studentId}`,
    createAssessment: '/api/grades/assessment',
    updateAssessment: (id: number) => `/api/grades/assessment/${id}`,
    deleteAssessment: (id: number) => `/api/grades/assessment/${id}`,
    submitScores: '/api/grades/scores',
  },
  
  // Announcements
  announcements: {
    list: '/api/announcements',
    get: (id: number) => `/api/announcements/${id}`,
    create: '/api/announcements',
    update: (id: number) => `/api/announcements/${id}`,
    delete: (id: number) => `/api/announcements/${id}`,
  },
  
  // Calendar
  calendar: {
    list: '/api/calendar',
    get: (id: number) => `/api/calendar/${id}`,
    create: '/api/calendar',
    update: (id: number) => `/api/calendar/${id}`,
    delete: (id: number) => `/api/calendar/${id}`,
  },
  
  // Notifications
  notifications: {
    list: '/api/notifications',
    unreadCount: '/api/notifications/unread-count',
    markAsRead: (id: number) => `/api/notifications/${id}/read`,
    markAllAsRead: '/api/notifications/read-all',
    delete: (id: number) => `/api/notifications/${id}`,
  },
  
  // Quizzes
  quizzes: {
    list: '/api/quizzes',
    get: (id: number) => `/api/quizzes/${id}`,
    create: '/api/quizzes',
    update: (id: number) => `/api/quizzes/${id}`,
    delete: (id: number) => `/api/quizzes/${id}`,
    questions: (id: number) => `/api/quizzes/${id}/questions`,
  },
  
  // Quiz Attempts
  quizAttempts: {
    start: (quizId: number) => `/api/quizzes/${quizId}/attempts/start`,
    submit: (quizId: number, attemptId: number) =>
      `/api/quizzes/${quizId}/attempts/${attemptId}/submit`,
    get: (quizId: number, attemptId: number) =>
      `/api/quizzes/${quizId}/attempts/${attemptId}`,
    listByQuiz: (quizId: number) => `/api/quizzes/${quizId}/attempts`,
    listByStudent: (studentId: number) => `/api/students/${studentId}/quiz-attempts`,
  },
  
  // Analytics
  analytics: {
    collegeDashboard: '/api/analytics/college/dashboard',
    courseAnalytics: (courseId: number) => `/api/analytics/course/${courseId}`,
    gradeDistribution: (courseId: number) =>
      `/api/analytics/course/${courseId}/grade-distribution`,
    studentPerformance: (studentId: number) =>
      `/api/analytics/student/${studentId}/performance`,
    attendanceTrends: '/api/analytics/attendance/trends',
  },
  
  // Departments
  departments: {
    list: '/api/departments',
    get: (id: number) => `/api/departments/${id}`,
    create: '/api/departments',
    update: (id: number) => `/api/departments/${id}`,
    delete: (id: number) => `/api/departments/${id}`,
  },
  
  // File Upload
  files: {
    upload: '/api/files/upload',
    delete: (key: string) => `/api/files/${key}`,
    getUrl: (key: string) => `/api/files/${key}/url`,
  },
  
  // Reports
  reports: {
    gradeCard: (studentId: number) => `/api/reports/grade-card/${studentId}`,
    transcript: (studentId: number) => `/api/reports/transcript/${studentId}`,
    attendanceReport: (studentId: number) =>
      `/api/reports/attendance/${studentId}`,
    courseReport: (courseId: number) => `/api/reports/course/${courseId}`,
  },
};