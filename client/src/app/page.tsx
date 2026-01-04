"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api-client";
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
import { DashboardResponse } from "@/lib/types";
import { logger } from "@/lib/logger";

export default function DashboardPage() {
  const { user, isLoading } = useAuth();
  const router = useRouter();
  const [dashboardData, setDashboardData] = useState<DashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [studentData, setStudentData] = useState<any>({
    enrolledCourses: 0,
    gpa: 0,
    attendanceRate: 0,
    pendingTasks: 0,
    courseGrades: [],
    recentGrades: []
  });
  const [studentLoading, setStudentLoading] = useState(false);
  const [studentError, setStudentError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      if (user) {
        try {
          setLoading(true);
          setError(null);
          const data = await api.get<any>("/api/dashboard");
          const normalized: DashboardResponse = {
            metrics: {
              totalStudents: data?.metrics?.totalStudents ?? 0,
              totalCourses: data?.metrics?.totalCourses ?? 0,
              attendanceRate: data?.metrics?.attendanceRate ?? 0,
              announcements: Array.isArray(data?.announcements) ? data.announcements.length : 0,
            },
            upcomingEvents: Array.isArray(data?.upcomingEvents) ? data.upcomingEvents : [],
            recentActivity: Array.isArray(data?.recentActivity) ? data.recentActivity : [],
          };
          setDashboardData(normalized);
        } catch (err: any) {
          logger.error('Dashboard fetch error:', err as Error);
          setError(err?.message || "Failed to fetch dashboard");
        } finally {
          setLoading(false);
        }
      }
    };

    fetchData();
  }, [user]);

  useEffect(() => {
    if (!isLoading && !user) {
      router.push("/auth/login");
    }
  }, [user, isLoading, router]);

  useEffect(() => {
    const fetchStudentData = async () => {
      if (user?.role !== 'student') {
        return;
      }

      setStudentLoading(true);
      setStudentError(null);

      try {
        const [courseGrades, attendance, assignments, grades] = await Promise.all([
          api.get<any[]>('/api/grades/courses'),
          api.get<any[]>('/api/attendance/stats/courses'),
          api.get<any[]>('/api/assignments'),
          api.get<any[]>('/api/grades')
        ]);

        const enrolledCourses = courseGrades?.length || 0;

        let gpa = 0;
        if (courseGrades && courseGrades.length > 0) {
          const gradePoints: Record<string, number> = {
            'A+': 4.0, 'A': 3.7, 'A-': 3.3,
            'B+': 3.0, 'B': 2.7, 'B-': 2.3,
            'C+': 2.0, 'C': 1.7, 'C-': 1.3,
            'D': 1.0, 'F': 0.0
          };
          const totalCredits = courseGrades.reduce((acc: number, c: any) => acc + (c.credits || 3), 0);
          if (totalCredits > 0) {
            const totalPoints = courseGrades.reduce((acc: number, c: any) => {
              const grade = c.letterGrade || c.grade || 'C';
              return acc + (gradePoints[grade] || 0) * (c.credits || 3);
            }, 0);
            gpa = totalPoints / totalCredits;
          }
        }

        let attendanceRate = 0;
        if (attendance && attendance.length > 0) {
          const totalSessions = attendance.reduce((acc: number, c: any) => acc + (c.totalSessions || c.total || 0), 0);
          const presentSessions = attendance.reduce((acc: number, c: any) => acc + (c.presentCount || c.present || 0), 0);
          if (totalSessions > 0) {
            attendanceRate = Math.round((presentSessions / totalSessions) * 100);
          }
        }

        const pendingTasks = assignments?.filter((a: any) => a.status === 'pending').length || 0;

        setStudentData({
          enrolledCourses,
          gpa: gpa.toFixed(2),
          attendanceRate,
          pendingTasks,
          courseGrades: (courseGrades || []).slice(0, 4),
          recentGrades: (grades || []).slice(0, 3)
        });
      } catch (err: any) {
        logger.error('Failed to fetch student data:', err as Error);
        setStudentError(err?.message || "Failed to load student data");
      } finally {
        setStudentLoading(false);
      }
    };

    fetchStudentData();
  }, [user?.role]);

  if (user.role === 'student') {
    if (studentLoading) {
      return (
        <div className="flex h-screen items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
        </div>
      );
    }

    if (studentError) {
      return (
        <div className="p-6">
          <Card>
            <CardHeader>
              <CardTitle>Student Dashboard</CardTitle>
              <CardDescription>Unable to load student data</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-sm text-destructive mb-4">{studentError}</div>
              <Button onClick={() => window.location.reload()} size="sm">Retry</Button>
            </CardContent>
          </Card>
        </div>
      );
    }

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
              <CardTitle className="text-2xl">{studentData.enrolledCourses}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <Award className="h-4 w-4" />
                Current GPA
              </CardDescription>
              <CardTitle className="text-2xl">{studentData.gpa}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <CheckCircle className="h-4 w-4" />
                Attendance Rate
              </CardDescription>
              <CardTitle className="text-2xl">{studentData.attendanceRate}%</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <FileText className="h-4 w-4" />
                Pending Tasks
              </CardDescription>
              <CardTitle className="text-2xl">{studentData.pendingTasks}</CardTitle>
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
              {studentData.courseGrades.map((course: any, idx: number) => (
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
              {studentData.courseGrades.length === 0 && (
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
                {studentData.recentGrades.map((grade: any, idx: number) => (
                  <TableRow key={idx}>
                    <TableCell className="font-medium">{grade.courseName || grade.courseCode || 'Course'}</TableCell>
                    <TableCell>{grade.assessmentName || grade.assessment_name || 'Assessment'}</TableCell>
                    <TableCell>{grade.obtainedMarks || grade.obtained_marks || 0}/{grade.totalMarks || grade.total_marks || 100}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{grade.letterGrade || grade.grade || 'N/A'}</Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {grade.gradedAt || grade.graded_at ? format(new Date(grade.gradedAt || grade.graded_at), 'MMM dd, yyyy') : 'N/A'}
                    </TableCell>
                  </TableRow>
                ))}
                {studentData.recentGrades.length === 0 && (
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

  // Faculty Dashboard
  if (user.role === 'faculty') {
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
                <Clock className="h-4 w-4" />
                Lectures Today
              </CardDescription>
              <CardTitle className="text-2xl">{dashboardData?.upcomingEvents?.length ?? 0}</CardTitle>
            </CardHeader>
          </Card>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Today's Schedule</CardTitle>
              <CardDescription>Your lectures for today</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {dashboardData?.upcomingEvents?.slice(0, 3).map((event) => (
                <div key={event.id} className="rounded-lg border p-4 space-y-2">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{event.title}</div>
                      <div className="text-sm text-muted-foreground">{event.course || 'Course'}</div>
                    </div>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {format(new Date(event.start), 'hh:mm a')}
                    </div>
                  </div>
                  <Button size="sm" className="w-full">Mark Attendance</Button>
                </div>
              ))}
              {(!dashboardData?.upcomingEvents || dashboardData.upcomingEvents.length === 0) && (
                <p className="text-sm text-muted-foreground text-center py-4">No lectures scheduled today</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>Latest system activity</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {dashboardData?.recentActivity?.slice(0, 3).map((activity) => (
                <div key={activity.id} className="rounded-lg border p-3 space-y-2">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium text-sm">{activity.message}</div>
                      <div className="text-xs text-muted-foreground capitalize">{activity.entity}</div>
                    </div>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {format(new Date(activity.timestamp), 'MMM dd, hh:mm a')}
                  </div>
                </div>
              ))}
              {(!dashboardData?.recentActivity || dashboardData.recentActivity.length === 0) && (
                <p className="text-sm text-muted-foreground text-center py-4">No recent activity</p>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  // Admin Dashboard
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Admin Dashboard</h1>
        <p className="text-muted-foreground">College-wide overview and management</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <GraduationCap className="h-4 w-4" />
              Total Students
            </CardDescription>
            <CardTitle className="text-2xl">{dashboardData?.metrics.totalStudents ?? 'N/A'}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              Faculty Members
            </CardDescription>
            <CardTitle className="text-2xl">{dashboardData?.metrics?.totalFaculty ?? 0}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <BookOpen className="h-4 w-4" />
              Active Courses
            </CardDescription>
            <CardTitle className="text-2xl">{dashboardData?.metrics.totalCourses ?? 'N/A'}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4" />
              Avg Attendance
            </CardDescription>
            <CardTitle className="text-2xl">{dashboardData?.metrics.attendanceRate ? Math.round(dashboardData.metrics.attendanceRate) : 'N/A'}%</CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Upcoming Events</CardTitle>
            <CardDescription>Scheduled college events</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {dashboardData?.upcomingEvents?.slice(0, 5).map((event) => (
                <div key={event.id} className="rounded-lg border p-3">
                  <div className="flex items-center justify-between">
                    <div className="font-medium text-sm">{event.title}</div>
                    <div className="text-xs text-muted-foreground">
                      {format(new Date(event.start), 'MMM dd')}
                    </div>
                  </div>
                  {event.course && (
                    <div className="text-xs text-muted-foreground mt-1">{event.course}</div>
                  )}
                </div>
              ))}
              {(!dashboardData?.upcomingEvents || dashboardData.upcomingEvents.length === 0) && (
                <p className="text-sm text-muted-foreground text-center py-4">No upcoming events</p>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>Latest system activity</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {dashboardData?.recentActivity?.slice(0, 5).map((activity) => (
              <div key={activity.id} className="rounded-lg border p-3">
                <div className="flex items-start gap-2">
                  <AlertCircle className="h-4 w-4 mt-0.5 text-blue-600" />
                  <div className="flex-1">
                    <p className="text-sm">{activity.message}</p>
                    <p className="text-xs text-muted-foreground capitalize mt-1">{activity.entity}</p>
                  </div>
                </div>
                <div className="text-xs text-muted-foreground mt-1">
                  {format(new Date(activity.timestamp), 'MMM dd, hh:mm a')}
                </div>
              </div>
            ))}
            {(!dashboardData?.recentActivity || dashboardData.recentActivity.length === 0) && (
              <p className="text-sm text-muted-foreground text-center py-4">No recent activity</p>
            )}
          </CardContent>
        </Card>
      </div>

    </div>
  );
}
