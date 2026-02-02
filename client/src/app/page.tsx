"use client";

import { useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { useDashboard, useStudentDashboard } from "@/lib/api-hooks";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  Users,
  BookOpen,
  GraduationCap,
  TrendingUp,
  Calendar,
  Clock,
  Award,
  CheckCircle,
  AlertCircle,
  FileText,
} from "lucide-react";
import { format } from "date-fns";


export default function DashboardPage() {
  const { user, isLoading: authLoading } = useAuth();
  const router = useRouter();

  // Fetch dashboard data using React Query
  const {
    data: dashboardData,
    isLoading: dashboardLoading,
    error: _dashboardError,
  } = useDashboard({
    enabled: !!user && user.role !== 'student',
  });

  // Fetch student dashboard data using React Query
  const {
    data: studentDashboardData,
    isLoading: studentLoading,
    error: studentError,
  } = useStudentDashboard({
    enabled: !!user && user.role === 'student',
  });

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!authLoading && !user) {
      router.push("/auth/login");
    }
  }, [user, authLoading, router]);

  // Calculate student metrics using useMemo
  const studentMetrics = useMemo(() => {
    if (!studentDashboardData) {
      return {
        enrolledCourses: 0,
        gpa: "0.00",
        attendanceRate: 0,
        pendingTasks: 0,
        courseGrades: [],
        recentGrades: [],
      };
    }

    const courseGrades = studentDashboardData.metrics?.courseGrades || [];
    const attendance = studentDashboardData.metrics?.attendance || [];
    const assignments = studentDashboardData.metrics?.assignments || [];
    const grades = studentDashboardData.metrics?.grades || [];

    const enrolledCourses = courseGrades.length;

    // Calculate GPA
    const gradePoints: Record<string, number> = {
      'A+': 4.0, 'A': 3.7, 'A-': 3.3,
      'B+': 3.0, 'B': 2.7, 'B-': 2.3,
      'C+': 2.0, 'C': 1.7, 'C-': 1.3,
      'D': 1.0, 'F': 0.0
    };

    let gpa = 0;
    if (courseGrades.length > 0) {
      const totalCredits = courseGrades.reduce((acc, c) => acc + (c.credits || 3), 0);
      if (totalCredits > 0) {
        const totalPoints = courseGrades.reduce((acc, c) => {
          const grade = c.letterGrade || c.grade || 'C';
          return acc + (gradePoints[grade] || 0) * (c.credits || 3);
        }, 0);
        gpa = totalPoints / totalCredits;
      }
    }

    // Calculate attendance rate
    let attendanceRate = 0;
    if (attendance.length > 0) {
      const totalSessions = attendance.reduce((acc, c) => acc + (c.totalSessions || c.total || 0), 0);
      const presentSessions = attendance.reduce((acc, c) => acc + (c.presentCount || c.present || 0), 0);
      if (totalSessions > 0) {
        attendanceRate = Math.round((presentSessions / totalSessions) * 100);
      }
    }

    // Count pending tasks
    const pendingTasks = assignments.filter((a) => a.status === 'pending').length;

    return {
      enrolledCourses,
      gpa: gpa.toFixed(2),
      attendanceRate,
      pendingTasks,
      courseGrades: courseGrades.slice(0, 4),
      recentGrades: grades.slice(0, 3),
    };
  }, [studentDashboardData]);

  // Show loading state
  if (authLoading || (user?.role === 'student' ? studentLoading : dashboardLoading)) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
      </div>
    );
  }

  // Show error state for students
  if (user?.role === 'student' && studentError) {
    return (
      <div className="p-6">
        <Card>
          <CardHeader>
            <CardTitle>Student Dashboard</CardTitle>
            <CardDescription>Unable to load student data</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-destructive mb-4">{studentError.message}</div>
            <Button onClick={() => window.location.reload()} size="sm">Retry</Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Student Dashboard View
  if (user?.role === 'student') {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">Welcome back, {user.firstName}!</h1>
          <p className="text-muted-foreground">Here's what's happening with your courses today</p>
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <BookOpen className="h-4 w-4" />
                Enrolled Courses
              </CardDescription>
              <CardTitle className="text-2xl">{studentMetrics.enrolledCourses}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <Award className="h-4 w-4" />
                Current GPA
              </CardDescription>
              <CardTitle className="text-2xl">{studentMetrics.gpa}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <CheckCircle className="h-4 w-4" />
                Attendance Rate
              </CardDescription>
              <CardTitle className="text-2xl">{studentMetrics.attendanceRate}%</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <FileText className="h-4 w-4" />
                Pending Tasks
              </CardDescription>
              <CardTitle className="text-2xl">{studentMetrics.pendingTasks}</CardTitle>
            </CardHeader>
          </Card>
        </div>

        <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
          <Card>
            <CardHeader>
              <CardTitle>Course Progress</CardTitle>
              <CardDescription>Your progress in enrolled courses</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {studentMetrics.courseGrades.map((course, idx) => (
                <div key={idx} className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <div>
                      <span className="font-medium">{course.courseName || 'Course'}</span>
                      <span className="ml-2 text-muted-foreground">{course.courseCode || ''}</span>
                    </div>
                    <span className="font-medium">{Math.round(course.percentage || 0)}%</span>
                  </div>
                  <Progress value={course.percentage || 0} />
                </div>
              ))}
              {studentMetrics.courseGrades.length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-4">No course data available</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Upcoming Deadlines</CardTitle>
              <CardDescription>Don't miss these!</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {dashboardData?.upcomingEvents?.slice(0, 5).map((item) => (
                <div key={item.id} className="rounded-lg border p-3 space-y-1">
                  <div className="flex items-center justify-between">
                    <span className="font-medium text-sm">{item.title}</span>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <Calendar className="h-3 w-3" />
                    {format(new Date(item.start), 'MMM dd, yyyy')}
                  </div>
                  <div className="text-xs text-muted-foreground">{item.course}</div>
                </div>
              ))}
              {(!dashboardData?.upcomingEvents || dashboardData.upcomingEvents.length === 0) && (
                <p className="text-sm text-muted-foreground text-center py-4">No upcoming events</p>
              )}
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Recent Grades</CardTitle>
            <CardDescription>Your latest assessment results</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Course</TableHead>
                  <TableHead>Assessment</TableHead>
                  <TableHead>Score</TableHead>
                  <TableHead>Grade</TableHead>
                  <TableHead>Date</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {studentMetrics.recentGrades.map((grade, idx) => (
                  <TableRow key={idx}>
                    <TableCell className="font-medium">{grade.courseName || grade.courseCode || 'Course'}</TableCell>
                    <TableCell>{grade.assessmentName || grade.assessment_name || 'Assessment'}</TableCell>
                    <TableCell>{grade.obtainedMarks || grade.obtained_marks || 0}/{grade.totalMarks || grade.total_marks || 100}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{grade.letterGrade || grade.grade || 'N/A'}</Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {grade.gradedAt || grade.graded_at ? format(new Date(grade.gradedAt || grade.graded_at || ''), 'MMM dd, yyyy') : 'N/A'}
                    </TableCell>
                  </TableRow>
                ))}
                {studentMetrics.recentGrades.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center text-muted-foreground py-4">
                      No grades available yet
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Faculty Dashboard View
  if (user?.role === 'faculty') {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold">Welcome back, {user.firstName}!</h1>
          <p className="text-muted-foreground">Manage your courses and students</p>
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <BookOpen className="h-4 w-4" />
                Teaching Courses
              </CardDescription>
              <CardTitle className="text-2xl">{dashboardData?.metrics.totalCourses ?? 0}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                Total Students
              </CardDescription>
              <CardTitle className="text-2xl">{dashboardData?.metrics.totalStudents ?? 0}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <FileText className="h-4 w-4" />
                Pending Submissions
              </CardDescription>
              <CardTitle className="text-2xl">{dashboardData?.metrics?.pendingSubmissions ?? 0}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <AlertCircle className="h-4 w-4" />
                Announcements
              </CardDescription>
              <CardTitle className="text-2xl">{dashboardData?.metrics.announcements ?? 0}</CardTitle>
            </CardHeader>
          </Card>
        </div>

        <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
          <Card>
            <CardHeader>
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>Latest actions across your courses</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {dashboardData?.recentActivity?.slice(0, 5).map((activity) => (
                <div key={activity.id} className="flex items-start gap-3 pb-3 border-b last:border-0">
                  <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center">
                    <TrendingUp className="h-4 w-4 text-primary" />
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-medium">{activity.message}</p>
                    <p className="text-xs text-muted-foreground">{activity.entity}</p>
                  </div>
                  <span className="text-xs text-muted-foreground">
                    {format(new Date(activity.timestamp), 'MMM dd, HH:mm')}
                  </span>
                </div>
              ))}
              {(!dashboardData?.recentActivity || dashboardData.recentActivity.length === 0) && (
                <p className="text-sm text-muted-foreground text-center py-4">No recent activity</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Upcoming Events</CardTitle>
              <CardDescription>Your schedule for the next few days</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {dashboardData?.upcomingEvents?.slice(0, 5).map((event) => (
                <div key={event.id} className="rounded-lg border p-3 space-y-1">
                  <div className="flex items-center justify-between">
                    <span className="font-medium text-sm">{event.title}</span>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <Clock className="h-3 w-3" />
                    {format(new Date(event.start), 'MMM dd, HH:mm')}
                  </div>
                  <div className="text-xs text-muted-foreground">{event.course}</div>
                </div>
              ))}
              {(!dashboardData?.upcomingEvents || dashboardData.upcomingEvents.length === 0) && (
                <p className="text-sm text-muted-foreground text-center py-4">No upcoming events</p>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  // Admin Dashboard View
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Admin Dashboard</h1>
        <p className="text-muted-foreground">Overview of the entire institution</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              Total Students
            </CardDescription>
            <CardTitle className="text-2xl">{dashboardData?.metrics.totalStudents ?? 0}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <BookOpen className="h-4 w-4" />
              Total Courses
            </CardDescription>
            <CardTitle className="text-2xl">{dashboardData?.metrics.totalCourses ?? 0}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <GraduationCap className="h-4 w-4" />
              Faculty Members
            </CardDescription>
            <CardTitle className="text-2xl">{dashboardData?.metrics.totalFaculty ?? 0}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4" />
              Attendance Rate
            </CardDescription>
            <CardTitle className="text-2xl">{dashboardData?.metrics.attendanceRate ?? 0}%</CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>System Overview</CardTitle>
            <CardDescription>Key metrics and recent activity</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {dashboardData?.recentActivity?.slice(0, 5).map((activity) => (
              <div key={activity.id} className="flex items-start gap-3 pb-3 border-b last:border-0">
                <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center">
                  <TrendingUp className="h-4 w-4 text-primary" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-medium">{activity.message}</p>
                  <p className="text-xs text-muted-foreground">{activity.entity}</p>
                </div>
                <span className="text-xs text-muted-foreground">
                  {format(new Date(activity.timestamp), 'MMM dd, HH:mm')}
                </span>
              </div>
            ))}
            {(!dashboardData?.recentActivity || dashboardData.recentActivity.length === 0) && (
              <p className="text-sm text-muted-foreground text-center py-4">No recent activity</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Common administrative tasks</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <Button variant="outline" className="w-full justify-start" onClick={() => router.push('/students')}>
              <Users className="mr-2 h-4 w-4" />
              Manage Students
            </Button>
            <Button variant="outline" className="w-full justify-start" onClick={() => router.push('/courses')}>
              <BookOpen className="mr-2 h-4 w-4" />
              Manage Courses
            </Button>
            <Button variant="outline" className="w-full justify-start" onClick={() => router.push('/announcements')}>
              <AlertCircle className="mr-2 h-4 w-4" />
              Post Announcement
            </Button>
            <Button variant="outline" className="w-full justify-start" onClick={() => router.push('/users')}>
              <Users className="mr-2 h-4 w-4" />
              Manage Users
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
