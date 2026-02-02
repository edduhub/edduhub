"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  BookOpen,
  CheckCircle,
  AlertCircle,
  FileText,
  Calendar,
  Award,
} from "lucide-react";
import { format } from "date-fns";
import { logger } from '@/lib/logger';

type StudentDashboardData = {
  student: {
    id: number;
    rollNo: string;
    firstName: string;
    lastName: string;
    email: string;
    semester: number;
    department: number;
  };
  academicOverview: {
    gpa: number;
    totalCredits: number;
    enrolledCourses: number;
    attendanceRate: number;
    totalPresentSessions: number;
    totalAttendanceSessions: number;
  };
  courses: Array<{
    id: number;
    code: string;
    name: string;
    credits: number;
    semester: string;
    averageGrade: number;
    attendanceRate: number;
    totalSessions: number;
    presentSessions: number;
    enrollmentStatus: string;
  }>;
  assignments: {
    upcoming: Array<{
      id: number;
      title: string;
      courseID: number;
      dueDate: string;
      maxScore: number;
      isSubmitted: boolean;
    }>;
    completed: Array<{
      id: number;
      title: string;
      courseID: number;
      dueDate: string;
      maxScore: number;
      isSubmitted: boolean;
      submittedAt: string;
      score: number;
      feedback: string;
    }>;
    overdue: Array<{
      id: number;
      title: string;
      courseID: number;
      dueDate: string;
      maxScore: number;
      isSubmitted: boolean;
    }>;
    summary: {
      upcomingCount: number;
      completedCount: number;
      overdueCount: number;
    };
  };
  recentGrades: Array<{
    id: number;
    courseName: string;
    courseCode: string;
    assessmentName: string;
    assessmentType: string;
    obtainedMarks: number;
    totalMarks: number;
    percentage: number;
    gradedDate: string;
  }>;
  upcomingEvents: Array<{
    id: number;
    title: string;
    description: string;
    date: string;
    type: string;
  }>;
  announcements: Array<{
    id: number;
    title: string;
    content: string;
    priority: string;
  }>;
};

export default function StudentDashboardPage() {
  const { user, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const [dashboardData, setDashboardData] = useState<StudentDashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const abortController = new AbortController();

    const fetchDashboardData = async () => {
      if (user) {
        try {
          setLoading(true);
          setError(null);
          const data = await api.get<StudentDashboardData>("/api/student/dashboard");
          // Check if component is still mounted
          if (!abortController.signal.aborted) {
            setDashboardData(data);
          }
        } catch (err: unknown) {
          // Ignore abort errors - they're expected on unmount
          if (err instanceof Error && err.name === 'AbortError') {
            return;
          }
          const errorMessage = err instanceof Error ? err.message : 'Failed to fetch dashboard data';
          logger.error('Dashboard fetch error:', err instanceof Error ? err : new Error(errorMessage));
          if (!abortController.signal.aborted) {
            setError(errorMessage);
          }
        } finally {
          if (!abortController.signal.aborted) {
            setLoading(false);
          }
        }
      }
    };

    fetchDashboardData();

    // Cleanup: abort request on unmount
    return () => {
      abortController.abort();
    };
  }, [user]);

  useEffect(() => {
    if (!authLoading && !user) {
      router.push("/auth/login");
    } else if (!authLoading && user && user.role !== 'student') {
      router.push("/");
    }
  }, [user, authLoading, router]);

  if (authLoading || loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-12 w-12 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
      </div>
    );
  }

  if (!user || user.role !== 'student') {
    return null;
  }

  if (error) {
    return (
      <div className="p-6">
        <Card>
          <CardHeader>
            <CardTitle>Student Dashboard</CardTitle>
            <CardDescription>Unable to load dashboard data</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-destructive">{error}</div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!dashboardData) {
    return null;
  }

  const { student, academicOverview, courses, assignments, recentGrades, upcomingEvents, announcements } = dashboardData;

  return (
    <div className="space-y-6">
      {/* Header Section */}
      <div>
        <h1 className="text-3xl font-bold">
          Welcome back, {student.firstName}!
        </h1>
        <p className="text-muted-foreground mt-1">
          Roll No: {student.rollNo} | Semester {student.semester} | {student.email}
        </p>
      </div>

      {/* Academic Overview Cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <Award className="h-4 w-4" />
              Current GPA
            </CardDescription>
            <CardTitle className="text-3xl font-bold text-primary">
              {academicOverview.gpa.toFixed(2)}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xs text-muted-foreground">
              {academicOverview.totalCredits} total credits
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <BookOpen className="h-4 w-4" />
              Enrolled Courses
            </CardDescription>
            <CardTitle className="text-3xl font-bold">
              {academicOverview.enrolledCourses}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xs text-muted-foreground">
              Active enrollment
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4" />
              Attendance Rate
            </CardDescription>
            <CardTitle className="text-3xl font-bold">
              {academicOverview.attendanceRate.toFixed(1)}%
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xs text-muted-foreground">
              {academicOverview.totalPresentSessions}/{academicOverview.totalAttendanceSessions} sessions
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <FileText className="h-4 w-4" />
              Pending Tasks
            </CardDescription>
            <CardTitle className="text-3xl font-bold">
              {assignments.summary.upcomingCount + assignments.summary.overdueCount}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xs text-destructive">
              {assignments.summary.overdueCount} overdue
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content Tabs */}
      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="courses">Courses</TabsTrigger>
          <TabsTrigger value="assignments">Assignments</TabsTrigger>
          <TabsTrigger value="grades">Grades</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-6 lg:grid-cols-2">
            {/* Course Progress */}
            <Card>
              <CardHeader>
                <CardTitle>Course Progress</CardTitle>
                <CardDescription>Your performance in enrolled courses</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {courses.slice(0, 5).map((course) => (
                  <div key={course.id} className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <div>
                        <span className="font-medium">{course.name}</span>
                        <span className="ml-2 text-muted-foreground">{course.code}</span>
                      </div>
                      <Badge variant={course.averageGrade >= 60 ? "default" : "destructive"}>
                        {course.averageGrade.toFixed(1)}%
                      </Badge>
                    </div>
                    <Progress value={course.averageGrade} />
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>Attendance: {course.attendanceRate.toFixed(0)}%</span>
                      <span>{course.credits} credits</span>
                    </div>
                  </div>
                ))}
                {courses.length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-4">
                    No enrolled courses
                  </p>
                )}
              </CardContent>
            </Card>

            {/* Assignments Overview */}
            <Card>
              <CardHeader>
                <CardTitle>Assignments Overview</CardTitle>
                <CardDescription>Track your assignment status</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-3 gap-4">
                  <div className="text-center">
                    <div className="text-2xl font-bold text-blue-600">
                      {assignments.summary.upcomingCount}
                    </div>
                    <div className="text-xs text-muted-foreground">Upcoming</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-green-600">
                      {assignments.summary.completedCount}
                    </div>
                    <div className="text-xs text-muted-foreground">Completed</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-red-600">
                      {assignments.summary.overdueCount}
                    </div>
                    <div className="text-xs text-muted-foreground">Overdue</div>
                  </div>
                </div>

                <div className="space-y-3 mt-4">
                  <h4 className="text-sm font-semibold">Upcoming Deadlines</h4>
                  {assignments.upcoming.slice(0, 5).map((assignment) => (
                    <div key={assignment.id} className="rounded-lg border p-3 space-y-1">
                      <div className="flex items-center justify-between">
                        <span className="font-medium text-sm">{assignment.title}</span>
                        <Badge variant="outline">
                          {assignment.maxScore} pts
                        </Badge>
                      </div>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <Calendar className="h-3 w-3" />
                        Due: {format(new Date(assignment.dueDate), 'MMM dd, yyyy')}
                      </div>
                    </div>
                  ))}
                  {assignments.upcoming.length === 0 && (
                    <p className="text-sm text-muted-foreground text-center py-2">
                      No upcoming assignments
                    </p>
                  )}
                </div>

                {assignments.overdue.length > 0 && (
                  <div className="space-y-3">
                    <h4 className="text-sm font-semibold text-destructive">
                      Overdue ({assignments.overdue.length})
                    </h4>
                    {assignments.overdue.slice(0, 3).map((assignment) => (
                      <div key={assignment.id} className="rounded-lg border border-destructive/50 p-3 space-y-1">
                        <div className="flex items-center justify-between">
                          <span className="font-medium text-sm">{assignment.title}</span>
                          <AlertCircle className="h-4 w-4 text-destructive" />
                        </div>
                        <div className="text-xs text-muted-foreground">
                          Was due: {format(new Date(assignment.dueDate), 'MMM dd')}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Recent Grades & Events */}
          <div className="grid gap-6 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Recent Grades</CardTitle>
                <CardDescription>Latest assessment results</CardDescription>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Assessment</TableHead>
                      <TableHead>Course</TableHead>
                      <TableHead className="text-right">Score</TableHead>
                      <TableHead className="text-right">%</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {recentGrades.slice(0, 5).map((grade) => (
                      <TableRow key={grade.id}>
                        <TableCell className="font-medium">{grade.assessmentName}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {grade.courseCode}
                        </TableCell>
                        <TableCell className="text-right text-sm">
                          {grade.obtainedMarks}/{grade.totalMarks}
                        </TableCell>
                        <TableCell className="text-right">
                          <Badge variant={grade.percentage >= 60 ? "default" : "destructive"}>
                            {grade.percentage.toFixed(1)}%
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
                {recentGrades.length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-4">
                    No grades available yet
                  </p>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Upcoming Events</CardTitle>
                <CardDescription>Important dates and events</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                {upcomingEvents.slice(0, 5).map((event) => (
                  <div key={event.id} className="rounded-lg border p-3 space-y-1">
                    <div className="flex items-center justify-between">
                      <span className="font-medium text-sm">{event.title}</span>
                      <Badge variant="outline">{event.type}</Badge>
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <Calendar className="h-3 w-3" />
                      {format(new Date(event.date), 'MMM dd, yyyy')}
                    </div>
                    {event.description && (
                      <div className="text-xs text-muted-foreground">{event.description}</div>
                    )}
                  </div>
                ))}
                {upcomingEvents.length === 0 && (
                  <p className="text-sm text-muted-foreground text-center py-4">
                    No upcoming events
                  </p>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Courses Tab */}
        <TabsContent value="courses">
          <Card>
            <CardHeader>
              <CardTitle>All Enrolled Courses</CardTitle>
              <CardDescription>Complete list of your courses</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Course</TableHead>
                    <TableHead>Code</TableHead>
                    <TableHead className="text-right">Credits</TableHead>
                    <TableHead className="text-right">Grade</TableHead>
                    <TableHead className="text-right">Attendance</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {courses.map((course) => (
                    <TableRow key={course.id}>
                      <TableCell className="font-medium">{course.name}</TableCell>
                      <TableCell>{course.code}</TableCell>
                      <TableCell className="text-right">{course.credits}</TableCell>
                      <TableCell className="text-right">
                        {course.averageGrade > 0 ? (
                          <Badge variant={course.averageGrade >= 60 ? "default" : "destructive"}>
                            {course.averageGrade.toFixed(1)}%
                          </Badge>
                        ) : (
                          <span className="text-muted-foreground text-sm">N/A</span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        {course.totalSessions > 0 ? (
                          <span className="text-sm">
                            {course.attendanceRate.toFixed(0)}% ({course.presentSessions}/{course.totalSessions})
                          </span>
                        ) : (
                          <span className="text-muted-foreground text-sm">N/A</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <Badge variant={course.enrollmentStatus === 'active' ? "default" : "secondary"}>
                          {course.enrollmentStatus}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              {courses.length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-8">
                  No courses enrolled
                </p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Assignments Tab */}
        <TabsContent value="assignments" className="space-y-4">
          {assignments.overdue.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-destructive">Overdue Assignments</CardTitle>
                <CardDescription>These assignments need immediate attention</CardDescription>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Title</TableHead>
                      <TableHead>Due Date</TableHead>
                      <TableHead className="text-right">Max Score</TableHead>
                      <TableHead className="text-right">Action</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {assignments.overdue.map((assignment) => (
                      <TableRow key={assignment.id}>
                        <TableCell className="font-medium">{assignment.title}</TableCell>
                        <TableCell className="text-destructive">
                          {format(new Date(assignment.dueDate), 'MMM dd, yyyy')}
                        </TableCell>
                        <TableCell className="text-right">{assignment.maxScore}</TableCell>
                        <TableCell className="text-right">
                          <Button size="sm" variant="default">Submit Now</Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          <Card>
            <CardHeader>
              <CardTitle>Upcoming Assignments</CardTitle>
              <CardDescription>Assignments due soon</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead>Due Date</TableHead>
                    <TableHead className="text-right">Max Score</TableHead>
                    <TableHead className="text-right">Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {assignments.upcoming.map((assignment) => (
                    <TableRow key={assignment.id}>
                      <TableCell className="font-medium">{assignment.title}</TableCell>
                      <TableCell>{format(new Date(assignment.dueDate), 'MMM dd, yyyy')}</TableCell>
                      <TableCell className="text-right">{assignment.maxScore}</TableCell>
                      <TableCell className="text-right">
                        <Button size="sm" variant="default">Submit</Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              {assignments.upcoming.length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-8">
                  No upcoming assignments
                </p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Completed Assignments</CardTitle>
              <CardDescription>Your submitted work</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead>Submitted</TableHead>
                    <TableHead className="text-right">Score</TableHead>
                    <TableHead className="text-right">Grade</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {assignments.completed.map((assignment) => (
                    <TableRow key={assignment.id}>
                      <TableCell className="font-medium">{assignment.title}</TableCell>
                      <TableCell>
                        {format(new Date(assignment.submittedAt), 'MMM dd, yyyy')}
                      </TableCell>
                      <TableCell className="text-right">
                        {assignment.score !== null && assignment.score !== undefined
                          ? `${assignment.score}/${assignment.maxScore}`
                          : 'Pending'}
                      </TableCell>
                      <TableCell className="text-right">
                        {assignment.score !== null && assignment.score !== undefined ? (
                          <Badge variant={(assignment.score / assignment.maxScore) >= 0.6 ? "default" : "destructive"}>
                            {((assignment.score / assignment.maxScore) * 100).toFixed(1)}%
                          </Badge>
                        ) : (
                          <Badge variant="secondary">Grading</Badge>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              {assignments.completed.length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-8">
                  No completed assignments yet
                </p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Grades Tab */}
        <TabsContent value="grades">
          <Card>
            <CardHeader>
              <CardTitle>All Grades</CardTitle>
              <CardDescription>Complete record of your academic performance</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Course</TableHead>
                    <TableHead>Assessment</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead className="text-right">Score</TableHead>
                    <TableHead className="text-right">Percentage</TableHead>
                    <TableHead>Date</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {recentGrades.map((grade) => (
                    <TableRow key={grade.id}>
                      <TableCell>
                        <div className="font-medium">{grade.courseName}</div>
                        <div className="text-xs text-muted-foreground">{grade.courseCode}</div>
                      </TableCell>
                      <TableCell className="font-medium">{grade.assessmentName}</TableCell>
                      <TableCell>
                        <Badge variant="outline">{grade.assessmentType}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        {grade.obtainedMarks}/{grade.totalMarks}
                      </TableCell>
                      <TableCell className="text-right">
                        <Badge variant={grade.percentage >= 60 ? "default" : "destructive"}>
                          {grade.percentage.toFixed(1)}%
                        </Badge>
                      </TableCell>
                      <TableCell>{format(new Date(grade.gradedDate), 'MMM dd, yyyy')}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              {recentGrades.length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-8">
                  No grades available yet
                </p>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Announcements Section */}
      {announcements.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Recent Announcements</CardTitle>
            <CardDescription>Important updates from your college</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {announcements.map((announcement) => (
              <div key={announcement.id} className="rounded-lg border p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <h4 className="font-semibold">{announcement.title}</h4>
                  <Badge variant={
                    announcement.priority === 'urgent' ? 'destructive' :
                      announcement.priority === 'high' ? 'default' : 'secondary'
                  }>
                    {announcement.priority}
                  </Badge>
                </div>
                <p className="text-sm text-muted-foreground">{announcement.content}</p>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
