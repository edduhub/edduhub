import { useQuery, useMutation, useQueryClient, UseQueryOptions } from '@tanstack/react-query';
import { api, endpoints } from './api-client';
import { normalizeForumThread, type ForumThreadApi } from './forum';
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
  Rubric,
  RubricCriterion,
  OfficeHourSlot,
  OfficeHourBooking,
  CreateRubricInput,
  CreateOfficeHourInput,
  CreateOfficeHourBookingInput,
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
  facultyRubrics: ['faculty-tools', 'rubrics'] as const,
  facultyOfficeHours: ['faculty-tools', 'office-hours'] as const,
  facultyBookings: ['faculty-tools', 'bookings'] as const,
};

type NotificationAPI = {
  id: number;
  userId?: number | string;
  user_id?: number | string;
  title?: string;
  message?: string;
  type?: Notification['type'];
  category?: string;
  isRead?: boolean;
  is_read?: boolean;
  actionUrl?: string;
  action_url?: string;
  metadata?: Notification['metadata'];
  createdAt?: string;
  created_at?: string;
};

type AnnouncementAPI = {
  id?: number;
  title?: string;
  content?: string;
  priority?: string;
  targetAudience?: string[];
  target_audience?: string[];
  courseId?: number;
  course_id?: number;
  courseName?: string;
  course_name?: string;
  departmentId?: number;
  department_id?: number;
  departmentName?: string;
  department_name?: string;
  publishedAt?: string;
  published_at?: string;
  expiresAt?: string;
  expires_at?: string;
  attachments?: string[];
  authorId?: string;
  author_id?: string;
  authorName?: string;
  author_name?: string;
  author?: string;
  collegeId?: string | number;
  college_id?: string | number;
  isPinned?: boolean;
  is_pinned?: boolean;
};

type GradeAPI = {
  id: number;
  studentId?: number;
  student_id?: number;
  courseId?: number;
  course_id?: number;
  assessmentType?: string;
  assessment_type?: string;
  assessmentName?: string;
  assessment_name?: string;
  score?: number;
  obtainedMarks?: number;
  obtained_marks?: number;
  maxScore?: number;
  totalMarks?: number;
  total_marks?: number;
  percentage?: number;
  weightage?: number;
  gradedDate?: string;
  graded_at?: string;
  remarks?: string;
  collegeId?: string | number;
  college_id?: string | number;
  createdAt?: string;
  created_at?: string;
};

type AttendanceAPI = {
  id?: number;
  ID?: number;
  studentId?: number;
  studentID?: number;
  student_id?: number;
  courseId?: number;
  courseID?: number;
  course_id?: number;
  lectureId?: number;
  lectureID?: number;
  lecture_id?: number;
  date?: string;
  status?: string;
  markedBy?: string;
  marked_by?: string;
  remarks?: string;
  collegeId?: string | number;
  collegeID?: string | number;
  college_id?: string | number;
  createdAt?: string;
  created_at?: string;
  scannedAt?: string;
  scanned_at?: string;
};

type AssignmentAPI = {
  id: number;
  courseId?: number;
  course_id?: number;
  courseName?: string;
  course_name?: string;
  title?: string;
  description?: string;
  dueDate?: string;
  due_date?: string;
  maxScore?: number;
  max_points?: number;
  attachments?: string[];
  collegeId?: string | number;
  college_id?: string | number;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
};

type ParentStudentAPI = {
  id?: number;
  student_id?: number;
  studentId?: number;
  userId?: string;
  user_id?: string | number;
  rollNo?: string;
  roll_no?: string;
  firstName?: string;
  first_name?: string;
  lastName?: string;
  last_name?: string;
  email?: string;
  departmentId?: number;
  department_id?: number;
  departmentName?: string;
  department_name?: string;
  semester?: number;
  collegeId?: string | number;
  college_id?: string | number;
  status?: Student['status'];
  isActive?: boolean;
  is_active?: boolean;
  enrolledCourses?: number;
  enrolled_courses?: number;
  gpa?: number;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
};

type SelfServiceRequestAPI = {
  id: number;
  type?: string;
  title?: string;
  description?: string;
  status?: string;
  submittedAt?: string;
  submitted_at?: string;
  respondedAt?: string;
  responded_at?: string;
  response?: string;
  document_type?: string;
  delivery_method?: string;
};

type RubricCriterionAPI = {
  id?: number;
  rubricId?: number;
  rubric_id?: number;
  name?: string;
  description?: string;
  weight?: number;
  maxScore?: number;
  max_score?: number;
  sortOrder?: number;
  sort_order?: number;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
};

type RubricAPI = {
  id?: number;
  facultyId?: number;
  faculty_id?: number;
  collegeId?: number;
  college_id?: number;
  name?: string;
  description?: string;
  courseId?: number;
  course_id?: number;
  isTemplate?: boolean;
  is_template?: boolean;
  isActive?: boolean;
  is_active?: boolean;
  maxScore?: number;
  max_score?: number;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
  criteria?: RubricCriterionAPI[];
};

type OfficeHourSlotAPI = {
  id?: number;
  facultyId?: number;
  faculty_id?: number;
  collegeId?: number;
  college_id?: number;
  dayOfWeek?: number;
  day_of_week?: number;
  startTime?: string;
  start_time?: string;
  endTime?: string;
  end_time?: string;
  location?: string;
  isVirtual?: boolean;
  is_virtual?: boolean;
  virtualLink?: string;
  virtual_link?: string;
  maxStudents?: number;
  max_students?: number;
  isActive?: boolean;
  is_active?: boolean;
  facultyName?: string;
  faculty_name?: string;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
};

type OfficeHourBookingAPI = {
  id?: number;
  officeHourId?: number;
  office_hour_id?: number;
  studentId?: number;
  student_id?: number;
  collegeId?: number;
  college_id?: number;
  bookingDate?: string;
  booking_date?: string;
  startTime?: string;
  start_time?: string;
  endTime?: string;
  end_time?: string;
  purpose?: string;
  status?: string;
  notes?: string;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
  officeHour?: OfficeHourSlotAPI;
  office_hour?: OfficeHourSlotAPI;
};

type ParentStudentsResponse = {
  students?: ParentStudentAPI[];
};

type ParentAttendanceResponse = {
  attendance?: AttendanceAPI[];
};

type ParentGradesResponse = {
  grades?: GradeAPI[];
};

type ParentAssignmentsResponse = {
  assignments?: AssignmentAPI[];
};

type SelfServiceRequestsResponse =
  | {
      requests?: SelfServiceRequestAPI[];
    }
  | SelfServiceRequestAPI[];

type RubricListResponse = {
  rubrics?: RubricAPI[];
};

type OfficeHourListResponse = {
  office_hours?: OfficeHourSlotAPI[];
};

type BookingListResponse = {
  bookings?: OfficeHourBookingAPI[];
};

type CreateAnnouncementInput = {
  title: string;
  content: string;
  priority?: Announcement['priority'];
  course_id?: number;
  is_published?: boolean;
  published_at?: string;
  expires_at?: string;
};

const toISODateString = (value?: string): string =>
  value || new Date().toISOString();

const mapNotification = (item: NotificationAPI): Notification => ({
  id: item.id,
  userId: String(item.userId ?? item.user_id ?? ''),
  title: item.title ?? '',
  message: item.message ?? '',
  type: item.type ?? 'info',
  category: item.category ?? item.type ?? 'general',
  isRead: item.isRead ?? item.is_read ?? false,
  actionUrl: item.actionUrl ?? item.action_url,
  metadata: item.metadata,
  createdAt: toISODateString(item.createdAt ?? item.created_at),
});

const mapAnnouncement = (item: AnnouncementAPI): Announcement => {
  const priority = item.priority;
  const normalizedPriority: Announcement['priority'] =
    priority === 'low' || priority === 'normal' || priority === 'high' || priority === 'urgent'
      ? priority
      : 'normal';

  const courseId = item.courseId ?? item.course_id;
  const departmentId = item.departmentId ?? item.department_id;
  const collegeIdRaw = item.collegeId ?? item.college_id;

  return {
    id: item.id ?? 0,
    title: item.title ?? '',
    content: item.content ?? '',
    priority: normalizedPriority,
    targetAudience: item.targetAudience ?? item.target_audience ?? [],
    courseId,
    courseName: item.courseName ?? item.course_name,
    departmentId,
    departmentName: item.departmentName ?? item.department_name,
    publishedAt: toISODateString(item.publishedAt ?? item.published_at),
    expiresAt: item.expiresAt ?? item.expires_at,
    attachments: item.attachments,
    authorId: item.authorId ?? item.author_id ?? '',
    authorName: item.authorName ?? item.author_name ?? item.author ?? 'System',
    collegeId: collegeIdRaw !== undefined ? String(collegeIdRaw) : '',
    isPinned: item.isPinned ?? item.is_pinned ?? false,
  };
};

const mapGrade = (item: GradeAPI): Grade => ({
  id: item.id,
  studentId: item.studentId ?? item.student_id ?? 0,
  courseId: item.courseId ?? item.course_id ?? 0,
  assessmentType: item.assessmentType ?? item.assessment_type ?? '',
  assessmentName: item.assessmentName ?? item.assessment_name ?? '',
  score: item.score ?? item.obtainedMarks ?? item.obtained_marks ?? 0,
  maxScore: item.maxScore ?? item.totalMarks ?? item.total_marks ?? 0,
  percentage: item.percentage ?? 0,
  weightage: item.weightage,
  gradedDate: toISODateString(item.gradedDate ?? item.graded_at ?? item.createdAt ?? item.created_at),
  remarks: item.remarks,
  collegeId: String(item.collegeId ?? item.college_id ?? ''),
});

const mapAttendance = (item: AttendanceAPI): Attendance => ({
  id: item.id ?? item.ID ?? 0,
  studentId: item.studentId ?? item.studentID ?? item.student_id ?? 0,
  courseId: item.courseId ?? item.courseID ?? item.course_id ?? 0,
  lectureId: item.lectureId ?? item.lectureID ?? item.lecture_id,
  date: toISODateString(item.date),
  status: (item.status ?? 'absent').toLowerCase() as Attendance['status'],
  markedBy: item.markedBy ?? item.marked_by ?? '',
  remarks: item.remarks,
  collegeId: String(item.collegeId ?? item.collegeID ?? item.college_id ?? ''),
  createdAt: toISODateString(item.createdAt ?? item.created_at ?? item.scannedAt ?? item.scanned_at ?? item.date),
});

const mapAssignment = (item: AssignmentAPI): Assignment => ({
  id: item.id,
  courseId: item.courseId ?? item.course_id ?? 0,
  courseName: item.courseName ?? item.course_name,
  title: item.title ?? '',
  description: item.description ?? '',
  dueDate: toISODateString(item.dueDate ?? item.due_date),
  maxScore: item.maxScore ?? item.max_points ?? 0,
  attachments: item.attachments,
  collegeId: String(item.collegeId ?? item.college_id ?? ''),
  createdAt: toISODateString(item.createdAt ?? item.created_at),
  updatedAt: toISODateString(item.updatedAt ?? item.updated_at),
});

const mapParentStudent = (item: ParentStudentAPI): Student => {
  const id = item.id ?? item.studentId ?? item.student_id ?? 0;
  const isActive = item.isActive ?? item.is_active ?? true;
  return {
    id,
    userId: String(item.userId ?? item.user_id ?? ''),
    rollNo: item.rollNo ?? item.roll_no ?? '',
    firstName: item.firstName ?? item.first_name ?? 'Student',
    lastName: item.lastName ?? item.last_name ?? String(id),
    email: item.email ?? '',
    departmentId: item.departmentId ?? item.department_id ?? 0,
    departmentName: item.departmentName ?? item.department_name ?? 'N/A',
    semester: item.semester ?? 0,
    collegeId: String(item.collegeId ?? item.college_id ?? ''),
    status: item.status ?? (isActive ? 'active' : 'inactive'),
    enrolledCourses: item.enrolledCourses ?? item.enrolled_courses ?? 0,
    gpa: item.gpa,
    createdAt: toISODateString(item.createdAt ?? item.created_at),
    updatedAt: toISODateString(item.updatedAt ?? item.updated_at),
  };
};

const mapSelfServiceRequest = (item: SelfServiceRequestAPI): SelfServiceRequest => {
  const type = item.type;
  const status = item.status;
  const normalizedType: SelfServiceRequest['type'] =
    type === 'enrollment' || type === 'schedule' || type === 'transcript' || type === 'document'
      ? type
      : 'document';
  const normalizedStatus: SelfServiceRequest['status'] =
    status === 'pending' || status === 'approved' || status === 'rejected' || status === 'processing'
      ? status
      : 'pending';

  return {
    id: item.id ?? 0,
    type: normalizedType,
    title: item.title ?? '',
    description: item.description ?? '',
    status: normalizedStatus,
    submittedAt: toISODateString(item.submittedAt ?? item.submitted_at),
    respondedAt: item.respondedAt ?? item.responded_at,
    response: item.response ?? (item as SelfServiceRequestAPI & { admin_response?: string }).admin_response,
  };
};

const mapRubricCriterion = (item: RubricCriterionAPI): RubricCriterion => ({
  id: item.id ?? 0,
  rubricId: item.rubricId ?? item.rubric_id ?? 0,
  name: item.name ?? '',
  description: item.description,
  weight: item.weight ?? 0,
  maxScore: item.maxScore ?? item.max_score ?? 0,
  sortOrder: item.sortOrder ?? item.sort_order ?? 0,
  createdAt: item.createdAt ?? item.created_at,
  updatedAt: item.updatedAt ?? item.updated_at,
});

const mapRubric = (item: RubricAPI): Rubric => ({
  id: item.id ?? 0,
  facultyId: item.facultyId ?? item.faculty_id ?? 0,
  collegeId: item.collegeId ?? item.college_id ?? 0,
  name: item.name ?? '',
  description: item.description,
  courseId: item.courseId ?? item.course_id,
  isTemplate: item.isTemplate ?? item.is_template ?? false,
  isActive: item.isActive ?? item.is_active ?? true,
  maxScore: item.maxScore ?? item.max_score ?? 0,
  createdAt: toISODateString(item.createdAt ?? item.created_at),
  updatedAt: toISODateString(item.updatedAt ?? item.updated_at),
  criteria: (item.criteria ?? []).map(mapRubricCriterion),
});

const mapOfficeHourSlot = (item: OfficeHourSlotAPI): OfficeHourSlot => ({
  id: item.id ?? 0,
  facultyId: item.facultyId ?? item.faculty_id ?? 0,
  collegeId: item.collegeId ?? item.college_id ?? 0,
  dayOfWeek: item.dayOfWeek ?? item.day_of_week ?? 0,
  startTime: item.startTime ?? item.start_time ?? '',
  endTime: item.endTime ?? item.end_time ?? '',
  location: item.location,
  isVirtual: item.isVirtual ?? item.is_virtual ?? false,
  virtualLink: item.virtualLink ?? item.virtual_link,
  maxStudents: item.maxStudents ?? item.max_students ?? 1,
  isActive: item.isActive ?? item.is_active ?? true,
  facultyName: item.facultyName ?? item.faculty_name,
  createdAt: toISODateString(item.createdAt ?? item.created_at),
  updatedAt: toISODateString(item.updatedAt ?? item.updated_at),
});

const mapOfficeHourBooking = (item: OfficeHourBookingAPI): OfficeHourBooking => {
  const status = item.status;
  const normalizedStatus: OfficeHourBooking['status'] =
    status === 'confirmed' || status === 'cancelled' || status === 'completed' || status === 'no_show'
      ? status
      : 'confirmed';

  return {
    id: item.id ?? 0,
    officeHourId: item.officeHourId ?? item.office_hour_id ?? 0,
    studentId: item.studentId ?? item.student_id ?? 0,
    collegeId: item.collegeId ?? item.college_id ?? 0,
    bookingDate: toISODateString(item.bookingDate ?? item.booking_date),
    startTime: item.startTime ?? item.start_time ?? '',
    endTime: item.endTime ?? item.end_time ?? '',
    purpose: item.purpose,
    status: normalizedStatus,
    notes: item.notes,
    createdAt: toISODateString(item.createdAt ?? item.created_at),
    updatedAt: toISODateString(item.updatedAt ?? item.updated_at),
    officeHour: item.officeHour || item.office_hour ? mapOfficeHourSlot((item.officeHour ?? item.office_hour) as OfficeHourSlotAPI) : undefined,
  };
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
      const data = await api.get<Assignment[]>(endpoints.assignments.listByCourse(courseId));
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
      const data = await api.get<Quiz[]>(endpoints.quizzes.listByCourse(courseId));
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
      const data = await api.get<NotificationAPI[]>(endpoints.notifications.list);
      return (data || []).map(mapNotification);
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
      const data = await api.get<AnnouncementAPI[]>(endpoints.announcements.list);
      return (data || []).map(mapAnnouncement);
    },
    ...options,
  });
}

export function useCreateAnnouncement() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (announcementData: CreateAnnouncementInput) => {
      const data = await api.post<AnnouncementAPI>(endpoints.announcements.create, announcementData);
      return mapAnnouncement(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.announcements });
    },
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
      const data = await api.get<Folder[]>('/api/folders');
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
      const data = await api.get<Webhook[]>('/api/webhooks');
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
      const data = await api.get<AuditLog[]>('/api/audit/logs');
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
      const url = courseId ? `${endpoints.forum.threads}?course_id=${courseId}` : endpoints.forum.threads;
      const data = await api.get<ForumThreadApi[]>(url);
      return Array.isArray(data) ? data.map(normalizeForumThread) : [];
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
      const data = await api.get<ParentStudentsResponse>('/api/parent/children');
      return (data?.students || []).map(mapParentStudent);
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
      const data = await api.get<ParentAttendanceResponse>(`/api/parent/children/${studentId}/attendance`);
      return (data?.attendance || []).map(mapAttendance);
    },
    enabled: !!studentId,
    ...options,
  });
}

export function useParentChildGrades(studentId: number, options?: UseQueryOptions<Grade[], Error>) {
  return useQuery({
    queryKey: ['parent', 'child', studentId, 'grades'],
    queryFn: async () => {
      const data = await api.get<ParentGradesResponse>(`/api/parent/children/${studentId}/grades`);
      return (data?.grades || []).map(mapGrade);
    },
    enabled: !!studentId,
    ...options,
  });
}

export function useParentChildAssignments(studentId: number, options?: UseQueryOptions<Assignment[], Error>) {
  return useQuery({
    queryKey: ['parent', 'child', studentId, 'assignments'],
    queryFn: async () => {
      const data = await api.get<ParentAssignmentsResponse>(`/api/parent/children/${studentId}/assignments`);
      return (data?.assignments || []).map(mapAssignment);
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
  document_type?: string;
  delivery_method?: string;
};

export function useSelfServiceRequests(options?: UseQueryOptions<SelfServiceRequest[], Error>) {
  return useQuery({
    queryKey: ['self-service', 'requests'],
    queryFn: async () => {
      const data = await api.get<SelfServiceRequestsResponse>(endpoints.selfService.requests);
      const requests = Array.isArray(data) ? data : (data?.requests || []);
      return requests.map(mapSelfServiceRequest);
    },
    ...options,
  });
}

export function useCreateSelfServiceRequest() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (input: CreateSelfServiceRequestInput) => {
      const data = await api.post<SelfServiceRequestAPI>(endpoints.selfService.requests, input);
      return mapSelfServiceRequest(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['self-service', 'requests'] });
    },
  });
}

export type UpdateSelfServiceRequestInput = {
  status?: 'pending' | 'approved' | 'rejected' | 'processing';
  response?: string;
};

export function useUpdateSelfServiceRequest() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ requestId, input }: { requestId: number; input: UpdateSelfServiceRequestInput }) => {
      const data = await api.put<SelfServiceRequestAPI>(endpoints.selfService.request(requestId), input);
      return mapSelfServiceRequest(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['self-service', 'requests'] });
      queryClient.invalidateQueries({ queryKey: ['self-service', 'all-requests'] });
    },
  });
}

// Admin: Fetch all requests (not just own)
export function useAllSelfServiceRequests(options?: UseQueryOptions<SelfServiceRequest[], Error>) {
  return useQuery({
    queryKey: ['self-service', 'all-requests'],
    queryFn: async () => {
      const data = await api.get<SelfServiceRequestsResponse>(`${endpoints.selfService.requests}?all=true`);
      const requests = Array.isArray(data) ? data : (data?.requests || []);
      return requests.map(mapSelfServiceRequest);
    },
    ...options,
  });
}

export function useFacultyRubrics(facultyId?: number, options?: UseQueryOptions<Rubric[], Error>) {
  return useQuery({
    queryKey: [...queryKeys.facultyRubrics, facultyId ?? 'all'],
    queryFn: async () => {
      const endpoint = facultyId
        ? `${endpoints.facultyTools.rubrics}?faculty_id=${facultyId}`
        : endpoints.facultyTools.rubrics;
      const data = await api.get<RubricListResponse>(endpoint);
      return (data?.rubrics || []).map(mapRubric);
    },
    ...options,
  });
}

export function useCreateRubric() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateRubricInput) => {
      const data = await api.post<RubricAPI>(endpoints.facultyTools.rubrics, input);
      return mapRubric(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyRubrics });
    },
  });
}

export function useUpdateRubric() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ rubricId, input }: { rubricId: number; input: CreateRubricInput }) => {
      const data = await api.put<RubricAPI>(endpoints.facultyTools.rubric(rubricId), input);
      return mapRubric(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyRubrics });
    },
  });
}

export function useDeleteRubric() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (rubricId: number) => api.delete<{ deleted: boolean }>(endpoints.facultyTools.rubric(rubricId)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyRubrics });
    },
  });
}

export function useOfficeHours(params?: { activeOnly?: boolean; facultyId?: number }, options?: UseQueryOptions<OfficeHourSlot[], Error>) {
  return useQuery({
    queryKey: [...queryKeys.facultyOfficeHours, params?.activeOnly ?? false, params?.facultyId ?? 'all'],
    queryFn: async () => {
      const search = new URLSearchParams();
      if (params?.activeOnly) search.set('active_only', 'true');
      if (params?.facultyId) search.set('faculty_id', String(params.facultyId));
      const endpoint = search.size
        ? `${endpoints.facultyTools.officeHours}?${search.toString()}`
        : endpoints.facultyTools.officeHours;
      const data = await api.get<OfficeHourListResponse>(endpoint);
      return (data?.office_hours || []).map(mapOfficeHourSlot);
    },
    ...options,
  });
}

export function useCreateOfficeHour() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateOfficeHourInput) => {
      const data = await api.post<OfficeHourSlotAPI>(endpoints.facultyTools.officeHours, input);
      return mapOfficeHourSlot(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyOfficeHours });
    },
  });
}

export function useUpdateOfficeHour() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ officeHourId, input }: { officeHourId: number; input: CreateOfficeHourInput }) => {
      const data = await api.put<OfficeHourSlotAPI>(endpoints.facultyTools.officeHour(officeHourId), input);
      return mapOfficeHourSlot(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyOfficeHours });
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyBookings });
    },
  });
}

export function useDeleteOfficeHour() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (officeHourId: number) => api.delete<{ deleted: boolean }>(endpoints.facultyTools.officeHour(officeHourId)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyOfficeHours });
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyBookings });
    },
  });
}

export function useBookings(params?: { officeHourId?: number; studentId?: number; facultyId?: number }, options?: UseQueryOptions<OfficeHourBooking[], Error>) {
  return useQuery({
    queryKey: [...queryKeys.facultyBookings, params?.officeHourId ?? 'all', params?.studentId ?? 'all', params?.facultyId ?? 'all'],
    queryFn: async () => {
      const search = new URLSearchParams();
      if (params?.officeHourId) search.set('office_hour_id', String(params.officeHourId));
      if (params?.studentId) search.set('student_id', String(params.studentId));
      if (params?.facultyId) search.set('faculty_id', String(params.facultyId));
      const endpoint = search.size
        ? `${endpoints.facultyTools.bookings}?${search.toString()}`
        : endpoints.facultyTools.bookings;
      const data = await api.get<BookingListResponse>(endpoint);
      return (data?.bookings || []).map(mapOfficeHourBooking);
    },
    ...options,
  });
}

export function useOfficeHourBookings(officeHourId: number, options?: UseQueryOptions<OfficeHourBooking[], Error>) {
  return useQuery({
    queryKey: [...queryKeys.facultyBookings, 'office-hour', officeHourId],
    queryFn: async () => {
      const data = await api.get<BookingListResponse>(endpoints.facultyTools.officeHourBookings(officeHourId));
      return (data?.bookings || []).map(mapOfficeHourBooking);
    },
    enabled: !!officeHourId,
    ...options,
  });
}

export function useCreateBooking() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateOfficeHourBookingInput) => {
      const data = await api.post<OfficeHourBookingAPI>(endpoints.facultyTools.bookings, input);
      return mapOfficeHourBooking(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyBookings });
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyOfficeHours });
    },
  });
}

export function useUpdateBookingStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ bookingId, status, notes }: { bookingId: number; status: OfficeHourBooking['status']; notes?: string }) => {
      const body: { status: OfficeHourBooking['status']; notes?: string } = { status };
      if (notes !== undefined) {
        body.notes = notes;
      }
      const data = await api.patch<OfficeHourBookingAPI>(endpoints.facultyTools.bookingStatus(bookingId), body);
      return mapOfficeHourBooking(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.facultyBookings });
    },
  });
}
