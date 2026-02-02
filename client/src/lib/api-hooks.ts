import { useQuery, useMutation, useQueryClient, UseQueryOptions } from '@tanstack/react-query';
import { api, endpoints } from './api-client';
import type {
  DashboardResponse,
  StudentDashboardResponse,
  Student,
  Course,
  Attendance,
  Grade,
  Assignment,
  Quiz,
  Notification,
  CalendarEvent,
  Department,
  Announcement,
  TimetableBlock,
  FeeStructure,
  StudentFee,
  FileRecord,
  Folder,
  Webhook,
  AuditLog,
  Placement,
  ForumThread,
} from './types';

// Query keys for cache management
export const queryKeys = {
  dashboard: ['dashboard'] as const,
  studentDashboard: ['studentDashboard'] as const,
  students: ['students'] as const,
  student: (id: number) => ['students', id] as const,
  courses: ['courses'] as const,
  course: (id: number) => ['courses', id] as const,
  attendance: ['attendance'] as const,
  attendanceByCourse: (courseId: number) => ['attendance', 'course', courseId] as const,
  attendanceByStudent: (studentId: number) => ['attendance', 'student', studentId] as const,
  grades: ['grades'] as const,
  gradesByCourse: (courseId: number) => ['grades', 'course', courseId] as const,
  gradesByStudent: (studentId: number) => ['grades', 'student', studentId] as const,
  assignments: ['assignments'] as const,
  assignmentsByCourse: (courseId: number) => ['assignments', 'course', courseId] as const,
  quizzes: ['quizzes'] as const,
  quizzesByCourse: (courseId: number) => ['quizzes', 'course', courseId] as const,
  notifications: ['notifications'] as const,
  unreadCount: ['notifications', 'unreadCount'] as const,
  calendar: ['calendar'] as const,
  calendarEvents: (start: string, end: string) => ['calendar', start, end] as const,
  departments: ['departments'] as const,
  announcements: ['announcements'] as const,
  timetable: ['timetable'] as const,
  timetableByCourse: (courseId: number) => ['timetable', 'course', courseId] as const,
  feeStructures: ['feeStructures'] as const,
  studentFees: ['studentFees'] as const,
  files: ['files'] as const,
  folders: ['folders'] as const,
  webhooks: ['webhooks'] as const,
  auditLogs: ['auditLogs'] as const,
  placements: ['placements'] as const,
  forumThreads: ['forumThreads'] as const,
  forumThreadsByCourse: (courseId: number) => ['forumThreads', 'course', courseId] as const,
};

// Dashboard hooks
export function useDashboard(options?: Omit<UseQueryOptions<DashboardResponse, Error>, 'queryKey' | 'queryFn'>) {
  return useQuery({
    queryKey: queryKeys.dashboard,
    queryFn: async () => {
      const data = await api.get<DashboardResponse>('/api/dashboard');
      return data;
    },
    ...options,
  });
}

export function useStudentDashboard(options?: Omit<UseQueryOptions<StudentDashboardResponse, Error>, 'queryKey' | 'queryFn'>) {
  return useQuery({
    queryKey: queryKeys.studentDashboard,
    queryFn: async () => {
      const data = await api.get<StudentDashboardResponse>('/api/student/dashboard');
      return data;
    },
    ...options,
  });
}

// Student hooks
export function useStudents(options?: UseQueryOptions<Student[], Error>) {
  return useQuery({
    queryKey: queryKeys.students,
    queryFn: async () => {
      const data = await api.get<Student[]>(endpoints.students.list);
      return data || [];
    },
    ...options,
  });
}

export function useStudent(studentId: number, options?: UseQueryOptions<Student, Error>) {
  return useQuery({
    queryKey: queryKeys.student(studentId),
    queryFn: async () => {
      const data = await api.get<Student>(endpoints.students.get(studentId));
      return data;
    },
    enabled: !!studentId,
    ...options,
  });
}

export function useCreateStudent() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (studentData: Partial<Student>) => {
      const data = await api.post<Student>(endpoints.students.create, studentData);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.students });
    },
  });
}

export function useUpdateStudent() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: Partial<Student> }) => {
      const result = await api.patch<Student>(endpoints.students.update(id), data);
      return result;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.student(variables.id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.students });
    },
  });
}

export function useDeleteStudent() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (studentId: number) => {
      await api.delete(endpoints.students.delete(studentId));
      return studentId;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.students });
    },
  });
}

// Course hooks
export function useCourses(options?: UseQueryOptions<Course[], Error>) {
  return useQuery({
    queryKey: queryKeys.courses,
    queryFn: async () => {
      const data = await api.get<Course[]>(endpoints.courses.list);
      return data || [];
    },
    ...options,
  });
}

export function useCourse(courseId: number, options?: UseQueryOptions<Course, Error>) {
  return useQuery({
    queryKey: queryKeys.course(courseId),
    queryFn: async () => {
      const data = await api.get<Course>(endpoints.courses.get(courseId));
      return data;
    },
    enabled: !!courseId,
    ...options,
  });
}

export function useCreateCourse() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (courseData: Partial<Course>) => {
      const data = await api.post<Course>(endpoints.courses.create, courseData);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.courses });
    },
  });
}

export function useUpdateCourse() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: Partial<Course> }) => {
      const result = await api.patch<Course>(endpoints.courses.update(id), data);
      return result;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.course(variables.id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.courses });
    },
  });
}

export function useDeleteCourse() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (courseId: number) => {
      await api.delete(endpoints.courses.delete(courseId));
      return courseId;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.courses });
    },
  });
}

// Attendance hooks
export function useAttendanceByCourse(courseId: number, options?: UseQueryOptions<Attendance[], Error>) {
  return useQuery({
    queryKey: queryKeys.attendanceByCourse(courseId),
    queryFn: async () => {
      const data = await api.get<Attendance[]>(`/api/attendance/course/${courseId}`);
      return data || [];
    },
    enabled: !!courseId,
    ...options,
  });
}

export function useAttendanceByStudent(studentId: number, options?: Omit<UseQueryOptions<Attendance[], Error>, 'queryKey' | 'queryFn'>) {
  return useQuery({
    queryKey: queryKeys.attendanceByStudent(studentId),
    queryFn: async () => {
      const data = await api.get<Attendance[]>(endpoints.attendance.byStudent(studentId));
      return data || [];
    },
    enabled: !!studentId,
    ...options,
  });
}

export function useMarkAttendance() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ 
      courseId, 
      lectureId, 
      status 
    }: { 
      courseId: number; 
      lectureId: number; 
      status: string 
    }) => {
      const data = await api.post<Attendance>(
        endpoints.attendance.mark(courseId, lectureId),
        { status }
      );
      return data;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.attendanceByCourse(variables.courseId) });
    },
  });
}

// Grade hooks
export function useGradesByStudent(studentId: number, options?: UseQueryOptions<Grade[], Error>) {
  return useQuery({
    queryKey: queryKeys.gradesByStudent(studentId),
    queryFn: async () => {
      const data = await api.get<Grade[]>(endpoints.grades.byStudent(studentId));
      return data || [];
    },
    enabled: !!studentId,
    ...options,
  });
}

export function useGradesByCourse(courseId: number, options?: UseQueryOptions<Grade[], Error>) {
  return useQuery({
    queryKey: queryKeys.gradesByCourse(courseId),
    queryFn: async () => {
      const data = await api.get<Grade[]>(endpoints.grades.byCourse(courseId));
      return data || [];
    },
    enabled: !!courseId,
    ...options,
  });
}

// Assignment hooks
export function useAssignmentsByCourse(courseId: number, options?: UseQueryOptions<Assignment[], Error>) {
  return useQuery({
    queryKey: queryKeys.assignmentsByCourse(courseId),
    queryFn: async () => {
      const data = await api.get<Assignment[]>(endpoints.assignments.byCourse(courseId));
      return data || [];
    },
    enabled: !!courseId,
    ...options,
  });
}

// Quiz hooks
export function useQuizzesByCourse(courseId: number, options?: UseQueryOptions<Quiz[], Error>) {
  return useQuery({
    queryKey: queryKeys.quizzesByCourse(courseId),
    queryFn: async () => {
      const data = await api.get<Quiz[]>(endpoints.quizzes.byCourse(courseId));
      return data || [];
    },
    enabled: !!courseId,
    ...options,
  });
}

// Notification hooks
export function useNotifications(options?: UseQueryOptions<Notification[], Error>) {
  return useQuery({
    queryKey: queryKeys.notifications,
    queryFn: async () => {
      const data = await api.get<Notification[]>(endpoints.notifications.list);
      return data || [];
    },
    ...options,
  });
}

export function useUnreadCount(options?: UseQueryOptions<{ unread_count: number }, Error>) {
  return useQuery({
    queryKey: queryKeys.unreadCount,
    queryFn: async () => {
      const data = await api.get<{ unread_count: number }>(endpoints.notifications.unreadCount);
      return data;
    },
    ...options,
  });
}

export function useMarkNotificationAsRead() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (notificationId: number) => {
      await api.patch(endpoints.notifications.markAsRead(notificationId), {});
      return notificationId;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      queryClient.invalidateQueries({ queryKey: queryKeys.unreadCount });
    },
  });
}

export function useMarkAllNotificationsAsRead() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async () => {
      await api.post(endpoints.notifications.markAllAsRead, {});
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      queryClient.invalidateQueries({ queryKey: queryKeys.unreadCount });
    },
  });
}

// Calendar hooks
export function useCalendarEvents(start: string, end: string, options?: UseQueryOptions<CalendarEvent[], Error>) {
  return useQuery({
    queryKey: queryKeys.calendarEvents(start, end),
    queryFn: async () => {
      const data = await api.get<CalendarEvent[]>(`${endpoints.calendar.list}?start=${start}&end=${end}`);
      return data || [];
    },
    enabled: !!start && !!end,
    ...options,
  });
}

// Department hooks
export function useDepartments(options?: UseQueryOptions<Department[], Error>) {
  return useQuery({
    queryKey: queryKeys.departments,
    queryFn: async () => {
      const data = await api.get<Department[]>(endpoints.departments.list);
      return data || [];
    },
    ...options,
  });
}

// Announcement hooks
export function useAnnouncements(options?: UseQueryOptions<Announcement[], Error>) {
  return useQuery({
    queryKey: queryKeys.announcements,
    queryFn: async () => {
      const data = await api.get<Announcement[]>(endpoints.announcements.list);
      return data || [];
    },
    ...options,
  });
}

// Timetable hooks
export function useTimetable(options?: UseQueryOptions<TimetableBlock[], Error>) {
  return useQuery({
    queryKey: queryKeys.timetable,
    queryFn: async () => {
      const data = await api.get<TimetableBlock[]>(endpoints.timetable.list);
      return data || [];
    },
    ...options,
  });
}

export function useTimetableByCourse(courseId: number, options?: UseQueryOptions<TimetableBlock[], Error>) {
  return useQuery({
    queryKey: queryKeys.timetableByCourse(courseId),
    queryFn: async () => {
      const data = await api.get<TimetableBlock[]>(`${endpoints.timetable.list}?courseId=${courseId}`);
      return data || [];
    },
    enabled: !!courseId,
    ...options,
  });
}

// Fee hooks
export function useFeeStructures(options?: UseQueryOptions<FeeStructure[], Error>) {
  return useQuery({
    queryKey: queryKeys.feeStructures,
    queryFn: async () => {
      const data = await api.get<FeeStructure[]>(endpoints.fees.structures.list);
      return data || [];
    },
    ...options,
  });
}

export function useStudentFees(options?: UseQueryOptions<StudentFee[], Error>) {
  return useQuery({
    queryKey: queryKeys.studentFees,
    queryFn: async () => {
      const data = await api.get<StudentFee[]>(endpoints.fees.myFees);
      return data || [];
    },
    ...options,
  });
}

// File hooks
export function useFiles(folderId?: number, options?: UseQueryOptions<FileRecord[], Error>) {
  return useQuery({
    queryKey: [...queryKeys.files, folderId],
    queryFn: async () => {
      const url = folderId ? `/api/file-management?folderId=${folderId}` : '/api/file-management';
      const data = await api.get<FileRecord[]>(url);
      return data || [];
    },
    ...options,
  });
}

export function useFolders(options?: UseQueryOptions<Folder[], Error>) {
  return useQuery({
    queryKey: queryKeys.folders,
    queryFn: async () => {
      const data = await api.get<Folder[]>(endpoints.folders.list);
      return data || [];
    },
    ...options,
  });
}

// Webhook hooks
export function useWebhooks(options?: UseQueryOptions<Webhook[], Error>) {
  return useQuery({
    queryKey: queryKeys.webhooks,
    queryFn: async () => {
      const data = await api.get<Webhook[]>(endpoints.webhooks.list);
      return data || [];
    },
    ...options,
  });
}

// Audit log hooks
export function useAuditLogs(options?: UseQueryOptions<AuditLog[], Error>) {
  return useQuery({
    queryKey: queryKeys.auditLogs,
    queryFn: async () => {
      const data = await api.get<AuditLog[]>(endpoints.audit.logs);
      return data || [];
    },
    ...options,
  });
}

// Placement hooks
export function usePlacements(options?: UseQueryOptions<Placement[], Error>) {
  return useQuery({
    queryKey: queryKeys.placements,
    queryFn: async () => {
      const data = await api.get<Placement[]>(endpoints.placements.list);
      return data || [];
    },
    ...options,
  });
}

// Forum hooks
export function useForumThreads(courseId?: number, options?: UseQueryOptions<ForumThread[], Error>) {
  return useQuery({
    queryKey: courseId ? queryKeys.forumThreadsByCourse(courseId) : queryKeys.forumThreads,
    queryFn: async () => {
      const url = courseId ? `${endpoints.forum.threads}?courseId=${courseId}` : endpoints.forum.threads;
      const data = await api.get<ForumThread[]>(url);
      return data || [];
    },
    ...options,
  });
}

// Analytics hooks
export function useAnalyticsDashboard(options?: UseQueryOptions<DashboardResponse, Error>) {
  return useQuery({
    queryKey: ['analytics', 'dashboard'],
    queryFn: async () => {
      const data = await api.get<DashboardResponse>(endpoints.analytics.collegeDashboard);
      return data;
    },
    ...options,
  });
}

export function useStudentPerformance(studentId: number, options?: UseQueryOptions<unknown, Error>) {
  return useQuery({
    queryKey: ['analytics', 'student', studentId, 'performance'],
    queryFn: async () => {
      const data = await api.get<unknown>(endpoints.analytics.studentPerformance(studentId));
      return data;
    },
    enabled: !!studentId,
    ...options,
  });
}

export function useAttendanceTrends(options?: UseQueryOptions<unknown, Error>) {
  return useQuery({
    queryKey: ['analytics', 'attendance', 'trends'],
    queryFn: async () => {
      const data = await api.get<unknown>(endpoints.analytics.attendanceTrends);
      return data;
    },
    ...options,
  });
}

// Parent Portal hooks
export function useParentChildren(options?: UseQueryOptions<Student[], Error>) {
  return useQuery({
    queryKey: ['parent', 'children'],
    queryFn: async () => {
      const data = await api.get<Student[]>('/api/parent/children');
      return data || [];
    },
    ...options,
  });
}

export function useParentChildDashboard(studentId: number, options?: UseQueryOptions<DashboardResponse, Error>) {
  return useQuery({
    queryKey: ['parent', 'child', studentId, 'dashboard'],
    queryFn: async () => {
      const data = await api.get<DashboardResponse>(`/api/parent/children/${studentId}/dashboard`);
      return data;
    },
    enabled: !!studentId,
    ...options,
  });
}

export function useParentChildAttendance(studentId: number, options?: UseQueryOptions<Attendance[], Error>) {
  return useQuery({
    queryKey: ['parent', 'child', studentId, 'attendance'],
    queryFn: async () => {
      const data = await api.get<Attendance[]>(`/api/parent/children/${studentId}/attendance`);
      return data || [];
    },
    enabled: !!studentId,
    ...options,
  });
}

export function useParentChildGrades(studentId: number, options?: UseQueryOptions<Grade[], Error>) {
  return useQuery({
    queryKey: ['parent', 'child', studentId, 'grades'],
    queryFn: async () => {
      const data = await api.get<Grade[]>(`/api/parent/children/${studentId}/grades`);
      return data || [];
    },
    enabled: !!studentId,
    ...options,
  });
}

export function useParentChildAssignments(studentId: number, options?: UseQueryOptions<Assignment[], Error>) {
  return useQuery({
    queryKey: ['parent', 'child', studentId, 'assignments'],
    queryFn: async () => {
      const data = await api.get<Assignment[]>(`/api/parent/children/${studentId}/assignments`);
      return data || [];
    },
    enabled: !!studentId,
    ...options,
  });
}

// Self-Service hooks
type SelfServiceRequest = {
  id: number;
  type: 'enrollment' | 'schedule' | 'transcript' | 'document';
  title: string;
  description: string;
  status: 'pending' | 'approved' | 'rejected' | 'processing';
  submittedAt: string;
  respondedAt?: string;
  response?: string;
};

type CreateSelfServiceRequestInput = {
  type: 'enrollment' | 'schedule' | 'transcript' | 'document';
  title: string;
  description: string;
};

export function useSelfServiceRequests(options?: UseQueryOptions<SelfServiceRequest[], Error>) {
  return useQuery({
    queryKey: ['self-service', 'requests'],
    queryFn: async () => {
      const data = await api.get<SelfServiceRequest[]>('/api/self-service/requests');
      return data || [];
    },
    ...options,
  });
}

export function useCreateSelfServiceRequest() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (input: CreateSelfServiceRequestInput) => {
      const data = await api.post<SelfServiceRequest>('/api/self-service/requests', input);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['self-service', 'requests'] });
    },
  });
}
