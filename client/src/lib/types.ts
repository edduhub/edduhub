// Core type definitions for EdduHub

export type UserRole = 'student' | 'faculty' | 'admin' | 'super_admin';

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
  preferences: Record<string, any>;
  social_links: Record<string, any>;
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

export type Notification = {
  id: number;
  userId: string;
  title: string;
  message: string;
  type: 'info' | 'success' | 'warning' | 'error';
  category: string;
  isRead: boolean;
  actionUrl?: string;
  metadata?: Record<string, any>;
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
  upcomingEvents: number;
  pendingSubmissions?: number;
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