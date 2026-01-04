// Core type definitions for EdduHub

export type UserRole = 'student' | 'faculty' | 'admin' | 'super_admin' | 'parent';

export type Parent = {
  id: number;
  userId: string;
  firstName: string;
  lastName: string;
  email: string;
  phone?: string;
  relation: 'father' | 'mother' | 'guardian';
  studentIds: number[];
  students?: Student[];
  collegeId: string;
  verified: boolean;
  createdAt: string;
  updatedAt: string;
};

export type StudentParentRelationship = {
  id: number;
  studentId: number;
  studentName?: string;
  parentId: number;
  parentName?: string;
  relation: 'father' | 'mother' | 'guardian';
  primaryContact: boolean;
  receiveNotifications: boolean;
  createdAt: string;
};

export type User = {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: UserRole;
  collegeId: string;
  collegeName: string;
  avatar?: string;
  verified: boolean;
};

export type UserPreferences = {
  emailNotifications?: boolean;
  pushNotifications?: boolean;
  theme?: 'light' | 'dark' | 'system';
  language?: string;
  timezone?: string;
  dateFormat?: string;
  [key: string]: string | boolean | number | undefined;
};

export type SocialLinks = {
  linkedin?: string;
  twitter?: string;
  github?: string;
  website?: string;
  [key: string]: string | undefined;
};

export type Profile = {
  id: number;
  user_id: string;
  college_id: string;
  bio: string;
  profile_image: string;
  phone_number: string;
  address: string;
  date_of_birth: string;
  joined_at: string;
  last_active: string;
  preferences: UserPreferences;
  social_links: SocialLinks;
  created_at: string;
  updated_at: string;
};

export type AuthSession = {
  token: string;
  user: User;
  expiresAt: string;
};

export type Course = {
  id: number;
  code: string;
  name: string;
  description?: string;
  credits: number;
  semester: string;
  departmentId: number;
  collegeId: string;
  instructorId?: string;
  instructorName?: string;
  enrollmentCount?: number;
  maxEnrollment?: number;
  createdAt: string;
  updatedAt: string;
};

export type Student = {
  id: number;
  userId: string;
  rollNo: string;
  firstName: string;
  lastName: string;
  email: string;
  phone?: string;
  dateOfBirth?: string;
  departmentId: number;
  departmentName?: string;
  semester: number;
  collegeId: string;
  status: 'active' | 'inactive' | 'suspended' | 'graduated';
  enrolledCourses?: number;
  gpa?: number;
  createdAt: string;
  updatedAt: string;
};

export type Assignment = {
  id: number;
  courseId: number;
  courseName?: string;
  title: string;
  description: string;
  dueDate: string;
  maxScore: number;
  attachments?: string[];
  collegeId: string;
  createdAt: string;
  updatedAt: string;
};

export type AssignmentSubmission = {
  id: number;
  assignmentId: number;
  studentId: number;
  studentName?: string;
  content?: string;
  attachments?: string[];
  submittedAt: string;
  score?: number;
  feedback?: string;
  gradedAt?: string;
  status: 'submitted' | 'graded' | 'late';
};

export type Attendance = {
  id: number;
  studentId: number;
  courseId: number;
  lectureId?: number;
  date: string;
  status: 'present' | 'absent' | 'late' | 'excused';
  markedBy: string;
  remarks?: string;
  collegeId: string;
  createdAt: string;
};

export type Grade = {
  id: number;
  studentId: number;
  courseId: number;
  assessmentType: string;
  assessmentName: string;
  score: number;
  maxScore: number;
  percentage: number;
  weightage?: number;
  gradedDate: string;
  remarks?: string;
  collegeId: string;
};

export type Announcement = {
  id: number;
  title: string;
  content: string;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  targetAudience: string[];
  courseId?: number;
  courseName?: string;
  departmentId?: number;
  departmentName?: string;
  publishedAt: string;
  expiresAt?: string;
  attachments?: string[];
  authorId: string;
  authorName: string;
  collegeId: string;
  isPinned?: boolean;
};

export type CalendarEvent = {
  id: number;
  title: string;
  description?: string;
  start: string;
  end: string;
  type: 'lecture' | 'exam' | 'event' | 'holiday' | 'deadline';
  courseId?: number;
  courseName?: string;
  location?: string;
  isRecurring?: boolean;
  recurrencePattern?: string;
  organizerId: string;
  organizerName?: string;
  collegeId: string;
};

export type NotificationMetadata = {
  courseId?: number;
  courseName?: string;
  studentId?: number;
  studentName?: string;
  assignmentId?: number;
  quizId?: number;
  eventId?: number;
  [key: string]: string | number | undefined;
};

export type Notification = {
  id: number;
  userId: string;
  title: string;
  message: string;
  type: 'info' | 'success' | 'warning' | 'error';
  category: string;
  isRead: boolean;
  actionUrl?: string;
  metadata?: NotificationMetadata;
  createdAt: string;
};

export type Department = {
  id: number;
  name: string;
  code: string;
  description?: string;
  hodId?: string;
  hodName?: string;
  collegeId: string;
  studentCount?: number;
  facultyCount?: number;
  createdAt: string;
  updatedAt: string;
};

export type Quiz = {
  id: number;
  courseId: number;
  courseName?: string;
  title: string;
  description?: string;
  duration: number; // in minutes
  totalMarks: number;
  passingMarks: number;
  startTime?: string;
  endTime?: string;
  allowedAttempts: number;
  shuffleQuestions: boolean;
  showAnswers: boolean;
  status: 'draft' | 'published' | 'archived';
  collegeId: string;
  createdAt: string;
};

export type Question = {
  id: number;
  quizId: number;
  type: 'multiple_choice' | 'true_false' | 'short_answer' | 'essay';
  questionText: string;
  options?: string[];
  correctAnswer?: string | string[];
  marks: number;
  order: number;
  explanation?: string;
};

export type QuizAttempt = {
  id: number;
  quizId: number;
  quizTitle?: string;
  studentId: number;
  studentName?: string;
  startedAt: string;
  submittedAt?: string;
  score?: number;
  totalMarks: number;
  passed?: boolean;
  attemptNumber: number;
  answers?: Record<number, any>;
  status: 'in_progress' | 'submitted' | 'graded';
};

export type DashboardMetrics = {
  totalStudents: number;
  totalCourses: number;
  totalFaculty?: number;
  attendanceRate: number;
  averageGrade?: number;
  announcements: number;
  upcomingEvents?: number;
  pendingSubmissions?: number;
};

export type DashboardEvent = {
  id: number;
  title: string;
  start: string;
  end?: string;
  course?: string;
  type?: string;
};

export type DashboardActivity = {
  id: number;
  message: string;
  entity: string;
  timestamp: string;
};

export type DashboardResponse = {
  metrics: DashboardMetrics;
  upcomingEvents: DashboardEvent[];
  recentActivity: DashboardActivity[];
};

export type AnalyticsData = {
  label: string;
  value: number;
  trend?: 'up' | 'down' | 'stable';
  delta?: number;
  chartData?: { date: string; value: number }[];
};

export type APIResponse<T> = {
  data?: T;
  message?: string;
  error?: string;
  success: boolean;
};

export type PaginatedResponse<T> = {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
};

export type FileUploadResponse = {
  url: string;
  key: string;
  name: string;
  size: number;
  contentType: string;
};

// Course Materials & Modules
export type CourseModule = {
  id: number;
  courseId: number;
  title: string;
  description?: string;
  displayOrder: number;
  isPublished: boolean;
  collegeId: string;
  createdAt: string;
  updatedAt: string;
  materialsCount?: number;
};

export type MaterialType = 'document' | 'video' | 'link' | 'presentation' | 'audio' | 'image' | 'assignment' | 'quiz';

export type CourseMaterial = {
  id: number;
  courseId: number;
  moduleId?: number;
  title: string;
  description?: string;
  type: MaterialType;
  fileId?: number;
  url?: string;
  isPublished: boolean;
  dueDate?: string;
  uploadedBy: string;
  uploadedByName?: string;
  collegeId: string;
  createdAt: string;
  updatedAt: string;
  size?: number;
  accessCount?: number;
};

export type MaterialAccessLog = {
  id: number;
  materialId: number;
  studentId: number;
  accessedAt: string;
  duration?: number;
  collegeId: string;
};

export type StudentProgress = {
  studentId: number;
  courseId: number;
  modulesCompleted: number;
  totalModules: number;
  materialsAccessed: number;
  totalMaterials: number;
  completionPercentage: number;
  lastAccessedAt?: string;
};

// File Management
export type FileMetadata = {
  alt?: string;
  description?: string;
  author?: string;
  category?: string;
  tags?: string[];
  [key: string]: string | string[] | undefined;
};

export type FileRecord = {
  id: number;
  collegeId: string;
  name: string;
  type: string;
  size: number;
  path: string;
  uploaderId: string;
  uploaderName?: string;
  folderId?: number;
  folderName?: string;
  storageType: 'local' | 's3' | 'minio';
  tags?: string[];
  metadata?: FileMetadata;
  currentVersion?: number;
  createdAt: string;
  updatedAt: string;
};

export type FileVersion = {
  id: number;
  fileId: number;
  versionNumber: number;
  changeDescription?: string;
  size: number;
  path: string;
  createdBy: string;
  createdByName?: string;
  createdAt: string;
};

export type Folder = {
  id: number;
  collegeId: string;
  name: string;
  parentId?: number;
  parentName?: string;
  path: string;
  createdBy: string;
  createdByName?: string;
  fileCount?: number;
  createdAt: string;
  updatedAt: string;
};

// Fee Management
export type FeeStructure = {
  id: number;
  collegeId: string;
  name: string;
  description?: string;
  amount: number;
  currency: string;
  category?: string;
  academicYear?: string;
  semester?: number;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
};

export type StudentFee = {
  id: number;
  studentId: number;
  studentName?: string;
  feeStructureId: number;
  feeStructureName?: string;
  semester?: number;
  dueDate: string;
  amount: number;
  paidAmount: number;
  remainingAmount: number;
  status: 'pending' | 'partial' | 'paid' | 'overdue';
  createdAt: string;
  updatedAt: string;
};

export type FeePayment = {
  id: number;
  studentFeeId: number;
  amount: number;
  paymentDate: string;
  paymentMethod: 'cash' | 'card' | 'online' | 'cheque' | 'upi';
  transactionId?: string;
  status: 'pending' | 'completed' | 'failed' | 'refunded';
  remarks?: string;
  createdAt: string;
};

export type FeeSummary = {
  totalAmount: number;
  paidAmount: number;
  remainingAmount: number;
  overdueAmount: number;
  feeCount: number;
  paidCount: number;
  pendingCount: number;
  overdueCount: number;
};

// Timetable
export type TimetableBlock = {
  id: number;
  collegeId: string;
  courseId: number;
  courseName?: string;
  courseCode?: string;
  dayOfWeek: number; // 0-6 (Sunday-Saturday)
  dayName?: string;
  startTime: string;
  endTime: string;
  room?: string;
  instructorId?: string;
  instructorName?: string;
  type?: 'lecture' | 'lab' | 'tutorial';
  createdAt: string;
  updatedAt: string;
};

// Lectures
export type Lecture = {
  id: number;
  courseId: number;
  courseName?: string;
  title: string;
  date: string;
  startTime: string;
  endTime: string;
  duration: number;
  room?: string;
  instructorId: string;
  instructorName?: string;
  description?: string;
  attachments?: string[];
  status: 'scheduled' | 'ongoing' | 'completed' | 'cancelled';
  attendanceMarked: boolean;
  collegeId: string;
  createdAt: string;
  updatedAt: string;
};

// Audit Logs
export type AuditLog = {
  id: number;
  collegeId: string;
  userId: string;
  userName?: string;
  action: string;
  entityType: string;
  entityId?: string;
  changes?: Record<string, unknown>;
  ipAddress?: string;
  userAgent?: string;
  timestamp: string;
};

export type AuditStatistics = {
  totalActions: number;
  userActions: number;
  entityChanges: number;
  periodStart: string;
  periodEnd: string;
  topActions: { action: string; count: number }[];
  topUsers: { userId: string; userName: string; count: number }[];
};

// Webhooks
export type Webhook = {
  id: number;
  collegeId: string;
  url: string;
  event: string;
  secret?: string;
  isActive: boolean;
  lastTriggeredAt?: string;
  failureCount: number;
  createdAt: string;
  updatedAt: string;
};

export type WebhookEventData = {
  entityType: string;
  entityId: string | number;
  action: string;
  changes?: Record<string, unknown>;
  userId?: string;
  timestamp?: string;
  [key: string]: unknown;
};

export type WebhookEvent = {
  event: string;
  timestamp: string;
  data: WebhookEventData;
};

// Analytics Types
export type PerformanceMetrics = {
  studentId?: number;
  courseId?: number;
  averageScore: number;
  highestScore: number;
  lowestScore: number;
  totalAssessments: number;
  trend: 'improving' | 'declining' | 'stable';
  percentile?: number;
  gradeDistribution?: { grade: string; count: number }[];
};

export type AttendanceTrend = {
  period: string;
  presentCount: number;
  absentCount: number;
  lateCount: number;
  excusedCount: number;
  attendanceRate: number;
};

export type CourseEngagement = {
  courseId: number;
  courseName: string;
  enrolledStudents: number;
  activeStudents: number;
  materialAccessCount: number;
  averageAccessTime: number;
  assignmentSubmissionRate: number;
  quizCompletionRate: number;
  attendanceRate: number;
};

export type PredictiveInsight = {
  studentId: number;
  studentName: string;
  riskLevel: 'low' | 'medium' | 'high';
  factors: string[];
  recommendations: string[];
  confidenceScore: number;
};

export type LearningAnalytics = {
  period: string;
  engagementRate: number;
  completionRate: number;
  averageTimeSpent: number;
  mostAccessedMaterials: { materialId: number; title: string; accessCount: number }[];
  leastAccessedMaterials: { materialId: number; title: string; accessCount: number }[];
  peakActivityHours: { hour: number; activityCount: number }[];
};

export type PerformanceTrend = {
  date: string;
  score: number;
  rank?: number;
  percentile?: number;
};

export type ComparativeCourseAnalysis = {
  courseId: number;
  courseName: string;
  metrics: {
    averageScore: number;
    passRate: number;
    attendanceRate: number;
    engagementScore: number;
    completionRate: number;
  };
};

// Advanced Dashboard Types
export type StudentDashboardData = {
  profile: {
    name: string;
    rollNo: string;
    semester: number;
    department: string;
    email: string;
    phone?: string;
  };
  academicOverview: {
    enrolledCourses: number;
    completedCredits: number;
    cgpa: number;
    currentSemesterGPA: number;
    attendancePercentage: number;
  };
  recentGrades: Grade[];
  upcomingAssignments: Assignment[];
  recentAttendance: Attendance[];
  enrolledCourses: Course[];
  announcements: Announcement[];
  upcomingEvents: CalendarEvent[];
  pendingQuizzes: Quiz[];
  courseProgress: StudentProgress[];
};

// Discussion Forums
export type ForumCategory = 'general' | 'academic' | 'assignment' | 'question' | 'announcement';

export type ForumThread = {
  id: number;
  courseId: number;
  courseName?: string;
  category: ForumCategory;
  title: string;
  content: string;
  authorId: number;
  authorName: string;
  authorAvatar?: string;
  isPinned: boolean;
  isLocked: boolean;
  viewCount: number;
  replyCount: number;
  lastReplyAt?: string;
  lastReplyBy?: number;
  tags: string[];
  createdAt: string;
  updatedAt: string;
  collegeId: number;
};

export type ForumReply = {
  id: number;
  threadId: number;
  parentId?: number;
  content: string;
  authorId: number;
  authorName: string;
  authorAvatar?: string;
  isAcceptedAnswer: boolean;
  likeCount: number;
  hasLiked: boolean;
  createdAt: string;
  updatedAt: string;
  collegeId: number;
};

export type ForumSearchFilters = {
  courseId?: number;
  category?: ForumCategory;
  tag?: string;
  authorId?: string;
  searchQuery?: string;
  sortBy?: 'latest' | 'popular' | 'unanswered';
  pinnedOnly?: boolean;
};

// Role & Permission Types
export type Permission = {
  id: number;
  resource: string;
  action: string;
  description?: string;
  createdAt: string;
};

export type Role = {
  id: number;
  collegeId: string;
  name: string;
  description?: string;
  permissions: Permission[];
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
};

export type UserRoleAssignment = {
  id: number;
  userId: string;
  roleId: number;
  roleName: string;
  assignedBy: string;
  assignedAt: string;
  expiresAt?: string;
};

// Error Types
export type AppError = {
  message: string;
  code?: string;
  statusCode?: number;
  details?: Record<string, any>;
};

export type ValidationError = {
  field: string;
  message: string;
  value?: unknown;
};

export type ApiError = {
  error: string;
  message?: string;
  code?: string;
  statusCode: number;
  validationErrors?: ValidationError[];
  details?: Record<string, unknown>;
};

// Logger Types
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export type LogEntry = {
  timestamp: string;
  level: LogLevel;
  message: string;
  context?: Record<string, unknown>;
  error?: Error;
};

// Batch Operation Types
export type BatchImportResult = {
  total: number;
  successful: number;
  failed: number;
  errors: { row: number; error: string }[];
};

export type BatchExportOptions = {
  format: 'csv' | 'xlsx' | 'json';
  filters?: Record<string, string | number | boolean | (string | number | boolean)[]>;
  fields?: string[];
};

// Report Types
export type GradeCard = {
  studentId: number;
  studentName: string;
  rollNo: string;
  semester: number;
  academicYear: string;
  courses: {
    courseId: number;
    courseName: string;
    courseCode: string;
    credits: number;
    grade: string;
    gradePoint: number;
    assessments: Grade[];
  }[];
  semesterGPA: number;
  cgpa: number;
  generatedAt: string;
};

export type Transcript = {
  studentId: number;
  studentName: string;
  rollNo: string;
  department: string;
  enrollmentYear: number;
  semesters: {
    semester: number;
    academicYear: string;
    courses: {
      courseCode: string;
      courseName: string;
      credits: number;
      grade: string;
      gradePoint: number;
    }[];
    semesterGPA: number;
    creditsEarned: number;
  }[];
  cgpa: number;
  totalCreditsEarned: number;
  generatedAt: string;
};

export type AttendanceReport = {
  courseId: number;
  courseName: string;
  courseCode: string;
  period: string;
  totalLectures: number;
  students: {
    studentId: number;
    studentName: string;
    rollNo: string;
    presentCount: number;
    absentCount: number;
    lateCount: number;
    excusedCount: number;
    attendancePercentage: number;
  }[];
  generatedAt: string;
};

// College Types
export type College = {
  id: string;
  name: string;
  address?: string;
  city?: string;
  state?: string;
  country?: string;
  pincode?: string;
  phone?: string;
  email?: string;
  website?: string;
  logo?: string;
  establishedYear?: number;
  affiliatedTo?: string;
  accreditation?: string[];
  createdAt: string;
  updatedAt: string;
};

export type CollegeStats = {
  totalStudents: number;
  totalFaculty: number;
  totalStaff: number;
  totalCourses: number;
  totalDepartments: number;
  activeEnrollments: number;
  overallAttendanceRate: number;
  averageCGPA: number;
};

/**
 * Get the appropriate dashboard path based on user role
 * @param role - The user's role
 * @returns The path to redirect to after login
 */
export function getDashboardPathForRole(role: UserRole): string {
  switch (role) {
    case 'student':
      return '/student-dashboard';
    case 'faculty':
      return '/faculty-dashboard';
    case 'admin':
      return '/admin-dashboard';
    case 'super_admin':
      return '/super-admin';
    case 'parent':
      return '/parent-portal';
    default:
      return '/';
  }
}