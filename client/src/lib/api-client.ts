// Enhanced API client with authentication support, retry logic, and better error handling

import { AuthSession, ValidationError, Quiz, Profile, User } from './types';
import { APIError as CustomAPIError } from './errors';
import { logger } from './logger';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const AUTH_STORAGE_KEY = 'edduhub_auth';
const MAX_RETRIES = 3;
const RETRY_DELAY = 1000; // 1 second

function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null;
  const auth = localStorage.getItem(AUTH_STORAGE_KEY);
  if (!auth) return null;
  try {
    const session = JSON.parse(auth) as AuthSession;
    return session.token;
  } catch (err) {
    logger.error('Failed to parse auth token', err as Error);
    return null;
  }
}

// Re-export APIError from errors module for backward compatibility
export { CustomAPIError as APIError };

export class NetworkError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'NetworkError';
  }
}

type RequestBody =
  | string
  | number
  | boolean
  | null
  | { [key: string]: RequestBody | RequestBody[] }
  | RequestBody[];

type RequestOptions = {
  method?: string;
  body?: RequestBody;
  headers?: Record<string, string>;
  requireAuth?: boolean;
  retries?: number;
  retryDelay?: number;
};

async function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  retries: number = MAX_RETRIES,
  delayMs: number = RETRY_DELAY
): Promise<T> {
  let lastError: Error;

  for (let attempt = 0; attempt <= retries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error instanceof Error ? error : new Error(String(error));

      // Don't retry on certain status codes
      if (error instanceof CustomAPIError && [400, 401, 403, 404].includes(error.status)) {
        throw error;
      }

      // Stop retrying when attempts are exhausted
      if (attempt === retries) {
        throw error;
      }

      // Exponential backoff
      const backoffDelay = delayMs * Math.pow(2, attempt);
      await delay(backoffDelay);
    }
  }

  throw lastError!;
}

export async function apiClient<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<T> {
  const {
    method = 'GET',
    body,
    headers = {},
    requireAuth = true,
    retries = MAX_RETRIES,
    retryDelay = RETRY_DELAY
  } = options;

  const token = getAuthToken();

  const requestHeaders: HeadersInit = {
    'Content-Type': 'application/json',
    'X-Client-Version': '1.0.0',
    ...headers,
  };

  if (requireAuth && token) {
    requestHeaders['Authorization'] = `Bearer ${token}`;
  }

  const config: RequestInit = {
    method,
    headers: requestHeaders,
    cache: 'no-store',
    credentials: 'include',
  };

  if (body !== undefined) {
    config.body = JSON.stringify(body);
  }

  const attemptRequest = async (): Promise<T> => {
    let response: Response;
    try {
      response = await fetch(`${API_BASE}${endpoint}`, config);
    } catch (error) {
      throw new NetworkError(error instanceof Error ? error.message : 'Network request failed');
    }

    if (!response.ok) {
      let message = 'Request failed';
      let code: string | undefined;
      let details: unknown;
      let validationErrors: ValidationError[] | undefined;

      try {
        const errorData = await response.json();
        if (errorData && typeof errorData === 'object') {
          message = errorData.message || errorData.error || message;
          code = errorData.code;
          details = errorData.details;
          validationErrors = errorData.validationErrors;
        }
      } catch (err) {
        logger.error('Failed to parse error response', err as Error);
        message = response.statusText || message;
      }

      // Handle authentication errors
      if (response.status === 401) {
        // Clear invalid token
        if (typeof window !== 'undefined') {
          localStorage.removeItem(AUTH_STORAGE_KEY);
        }
        // Redirect to login
        if (typeof window !== 'undefined') {
          window.location.href = '/auth/login';
        }
      }

      const error = new CustomAPIError(response.status, message, code, details as Record<string, unknown> | undefined, validationErrors);
      throw error;
    }

    // Handle empty responses
    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      return {} as T;
    }

    const data = await response.json();
    // Backend returns {data: ..., message: ...} format
    // Extract the actual data
    if (data && typeof data === 'object' && 'data' in data) {
      return data.data as T;
    }
    return data;
  };

  // Use retry logic for GET requests and non-auth endpoints
  if (method === 'GET' || !requireAuth) {
    return retryWithBackoff(attemptRequest, retries, retryDelay);
  }

  return attemptRequest();
}

// Convenience methods
export const api = {
  get: <T>(endpoint: string, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'GET', requireAuth }),

  post: <T>(endpoint: string, data: RequestBody, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'POST', body: data, requireAuth }),

  put: <T>(endpoint: string, data: RequestBody, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'PUT', body: data, requireAuth }),

  patch: <T>(endpoint: string, data: RequestBody, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'PATCH', body: data, requireAuth }),

  delete: <T>(endpoint: string, requireAuth = true) =>
    apiClient<T>(endpoint, { method: 'DELETE', requireAuth }),
};

// API endpoints
export const endpoints = {
  // Auth
  auth: {
    login: '/auth/login',
    register: '/auth/register',
    logout: '/auth/logout',
    refresh: '/auth/refresh',
    profile: '/api/profile',
    changePassword: '/auth/change-password',
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
    // Student convenience endpoint (list all assignments across enrolled courses)
    list: '/api/assignments',
    // Course-scoped endpoints
    listByCourse: (courseId: number) => `/api/courses/${courseId}/assignments`,
    get: (courseId: number, id: number) => `/api/courses/${courseId}/assignments/${id}`,
    create: (courseId: number) => `/api/courses/${courseId}/assignments`,
    update: (courseId: number, id: number) => `/api/courses/${courseId}/assignments/${id}`,
    delete: (courseId: number, id: number) => `/api/courses/${courseId}/assignments/${id}`,
    submit: (courseId: number, id: number) => `/api/courses/${courseId}/assignments/${id}/submit`,
    grade: (courseId: number, submissionId: number) =>
      `/api/courses/${courseId}/assignments/submissions/${submissionId}/grade`,
  },

  // Attendance
  attendance: {
    myAttendance: '/api/attendance/student/me',
    myCourseStats: '/api/attendance/stats/courses',
    mark: (courseId: number, lectureId: number) => `/api/attendance/mark/course/${courseId}/lecture/${lectureId}`,
    markBulk: (courseId: number, lectureId: number) => `/api/attendance/mark/bulk/course/${courseId}/lecture/${lectureId}`,
    generateQR: (courseId: number, lectureId: number) => `/api/attendance/course/${courseId}/lecture/${lectureId}/qrcode`,
    processQR: '/api/attendance/process-qr',
    byStudent: (studentId: number) => `/api/attendance/student/${studentId}`,
    byCourse: (courseId: number) => `/api/attendance/course/${courseId}`,
    update: (courseId: number, lectureId: number, studentId: number) => `/api/attendance/course/${courseId}/lecture/${lectureId}/student/${studentId}`,
  },

  // Grades
  grades: {
    myGrades: '/api/grades',
    myCourseGrades: '/api/grades/courses',
    byCourse: (courseId: number) => `/api/grades/course/${courseId}`,
    byStudent: (studentId: number) => `/api/grades/student/${studentId}`,
    createAssessment: (courseId: number) => `/api/grades/course/${courseId}`,
    updateAssessment: (courseId: number, assessmentId: number) => `/api/grades/course/${courseId}/assessment/${assessmentId}`,
    deleteAssessment: (courseId: number, assessmentId: number) => `/api/grades/course/${courseId}/assessment/${assessmentId}`,
    submitScores: (courseId: number, assessmentId: number) => `/api/grades/course/${courseId}/assessment/${assessmentId}/scores`,
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
    unreadCount: '/api/notifications/unread/count',
    markAsRead: (id: number) => `/api/notifications/${id}/read`,
    markAllAsRead: '/api/notifications/mark-all-read',
    delete: (id: number) => `/api/notifications/${id}`,
  },

  // Quizzes
  quizzes: {
    // Student convenience endpoint (list all quizzes across enrolled courses)
    myQuizzes: '/api/quizzes',
    // Course-scoped endpoints
    listByCourse: (courseId: number) => `/api/courses/${courseId}/quizzes`,
    get: (courseId: number, id: number) => `/api/courses/${courseId}/quizzes/${id}`,
    create: (courseId: number) => `/api/courses/${courseId}/quizzes`,
    update: (courseId: number, id: number) => `/api/courses/${courseId}/quizzes/${id}`,
    delete: (courseId: number, id: number) => `/api/courses/${courseId}/quizzes/${id}`,
    questions: (id: number) => `/api/quizzes/${id}/questions`,
  },

  // Quiz Attempts
  quizAttempts: {
    start: (quizId: number) => `/api/quizzes/${quizId}/attempts/start`,
    submit: (attemptId: number) => `/api/attempts/${attemptId}/submit`,
    get: (attemptId: number) => `/api/attempts/${attemptId}`,
    listByQuiz: (quizId: number) => `/api/quizzes/${quizId}/attempts`,
    listByStudent: (studentId: number) => `/api/attempts/student/${studentId}`,
  },

  // Analytics
  analytics: {
    collegeDashboard: '/api/analytics/dashboard',
    courseAnalytics: (courseId: number) => `/api/analytics/courses/${courseId}/analytics`,
    gradeDistribution: (courseId: number) =>
      `/api/analytics/courses/${courseId}/grades/distribution`,
    studentPerformance: (studentId: number) =>
      `/api/analytics/students/${studentId}/performance`,
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

  // Users
  users: {
    list: '/api/users',
    get: (id: number) => `/api/users/${id}`,
    create: '/api/users',
    update: (id: number) => `/api/users/${id}`,
    delete: (id: number) => `/api/users/${id}`,
    updateRole: (id: number) => `/api/users/${id}/role`,
    updateStatus: (id: number) => `/api/users/${id}/status`,
  },

  // File Upload
  files: {
    upload: '/api/files/upload',
    delete: (key: string) => `/api/files/${key}`,
    getUrl: (key: string) => `/api/files/${key}/url`,
  },

  // Reports
  reports: {
    // Student convenience endpoints
    myGradeCard: '/api/reports/students/me/gradecard',
    myTranscript: '/api/reports/students/me/transcript',
    // Admin/Faculty endpoints
    gradeCard: (studentId: number) => `/api/reports/students/${studentId}/gradecard`,
    transcript: (studentId: number) => `/api/reports/students/${studentId}/transcript`,
    courseAttendance: (courseId: number) => `/api/reports/courses/${courseId}/attendance`,
    courseReport: (courseId: number) => `/api/reports/courses/${courseId}/report`,
  },

  // Roles and Permissions
  roles: {
    list: '/api/roles',
    get: (id: number) => `/api/roles/${id}`,
    create: '/api/roles',
    update: (id: number) => `/api/roles/${id}`,
    delete: (id: number) => `/api/roles/${id}`,
    assignPermissions: (id: number) => `/api/roles/${id}/permissions`,
  },

  permissions: {
    list: '/api/permissions',
  },

  userRoles: {
    assign: '/api/user-roles',
    getUserRoles: (userId: number) => `/api/user-roles/users/${userId}`,
  },

  // Fees
  fees: {
    structures: {
      list: '/api/fees/structures',
      create: '/api/fees/structures',
      update: (feeId: number) => `/api/fees/structures/${feeId}`,
      delete: (feeId: number) => `/api/fees/structures/${feeId}`,
    },
    assign: '/api/fees/assign',
    bulkAssign: '/api/fees/bulk-assign',
    myFees: '/api/fees/my-fees',
    myFeesSummary: '/api/fees/my-fees/summary',
    payments: {
      create: '/api/fees/payments',
      initiateOnline: '/api/fees/payments/online',
      list: '/api/fees/my-payments',
    },
  },

  // Timetable
  timetable: {
    list: '/api/timetable',
    create: '/api/timetable',
    update: (blockId: number) => `/api/timetable/${blockId}`,
    delete: (blockId: number) => `/api/timetable/${blockId}`,
    myTimetable: '/api/timetable/my-timetable',
  },

  // Placements
  placements: {
    list: '/api/placements',
    get: (id: number) => `/api/placements/${id}`,
    create: '/api/placements',
    update: (id: number) => `/api/placements/${id}`,
    delete: (id: number) => `/api/placements/${id}`,
    stats: '/api/placements/stats',
    companyStats: '/api/placements/company-stats',
    company: (name: string) => `/api/placements/company/${name}`,
    studentPlacements: (studentId: number) => `/api/students/${studentId}/placements`,
  },

  // Exams
  exams: {
    list: '/api/exams',
    get: (examId: number) => `/api/exams/${examId}`,
    create: '/api/exams',
    update: (examId: number) => `/api/exams/${examId}`,
    delete: (examId: number) => `/api/exams/${examId}`,
    stats: (examId: number) => `/api/exams/${examId}/stats`,
    results: (examId: number) => `/api/exams/${examId}/results`,
    studentResult: (examId: number, studentId: number) => `/api/exams/${examId}/results/${studentId}`,
    studentResults: (studentId: number) => `/api/students/${studentId}/exam-results`,
  },

  // Exam Rooms
  examRooms: {
    list: '/api/exam-rooms',
    create: '/api/exam-rooms',
    get: (roomId: number) => `/api/exam-rooms/${roomId}`,
    update: (roomId: number) => `/api/exam-rooms/${roomId}`,
    delete: (roomId: number) => `/api/exam-rooms/${roomId}`,
    availability: (roomId: number) => `/api/exam-rooms/${roomId}/availability`,
  },

  // Revaluation
  revaluation: {
    list: '/api/revaluation-requests',
    create: '/api/revaluation-requests',
    approve: (requestId: number) => `/api/revaluation-requests/${requestId}/approve`,
    reject: (requestId: number) => `/api/revaluation-requests/${requestId}/reject`,
  },

  // Self-Service
  selfService: {
    requests: '/api/self-service/requests',
    request: (requestId: number) => `/api/self-service/requests/${requestId}`,
    types: '/api/self-service/types',
  },

  // Faculty Tools
  facultyTools: {
    rubrics: '/api/faculty-tools/rubrics',
    rubric: (rubricId: number) => `/api/faculty-tools/rubrics/${rubricId}`,
    officeHours: '/api/faculty-tools/office-hours',
    officeHour: (officeHourId: number) => `/api/faculty-tools/office-hours/${officeHourId}`,
    bookings: '/api/faculty-tools/bookings',
    bookingStatus: (bookingId: number) => `/api/faculty-tools/bookings/${bookingId}/status`,
    officeHourBookings: (officeHourId: number) => `/api/faculty-tools/office-hours/${officeHourId}/bookings`,
  },

  // Forum
  forum: {
    threads: '/api/forum/threads',
    thread: (threadId: number) => `/api/forum/threads/${threadId}`,
    replies: (threadId: number) => `/api/forum/threads/${threadId}/replies`,
    createThread: '/api/forum/threads',
    createReply: (threadId: number) => `/api/forum/threads/${threadId}/replies`,
  },
};

export async function fetchQuizzes() {
  return api.get<Quiz[]>(endpoints.quizzes.myQuizzes);
}

export async function fetchProfile() {
  return api.get<Profile>(endpoints.auth.profile);
}

export async function fetchUsers() {
  return api.get<User[]>(endpoints.users.list);
}
