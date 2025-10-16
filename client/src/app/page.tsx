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
  Loader2
} from "lucide-react";
import { format } from "date-fns";
import { DashboardResponse } from "@/lib/api";

export default function DashboardPage() {
  const { user, isLoading } = useAuth();
  const router = useRouter();
  const [dashboardData, setDashboardData] = useState<DashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      if (user) {
        try {
          const data = await api.get<DashboardResponse>('/api/dashboard');
          setDashboardData(data);
        } catch (error) {
          console.error('Failed to fetch dashboard:', error);
        } finally {
          setLoading(false);
        }
      }
    };

    fetchData();
  }, [user]);

  useEffect(() => {
    if (!isLoading && !user) {
      router.push('/auth/login');
    }
  }, [user, isLoading, router]);

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  // Student Dashboard
  if (user.role === 'student') {
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
              <CardTitle className="text-2xl">5</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <Award className="h-4 w-4" />
                Current GPA
              </CardDescription>
              <CardTitle className="text-2xl">3.85</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <CheckCircle className="h-4 w-4" />
                Attendance Rate
              </CardDescription>
              <CardTitle className="text-2xl">92%</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <FileText className="h-4 w-4" />
                Pending Tasks
              </CardDescription>
              <CardTitle className="text-2xl">3</CardTitle>
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
              {[
                { name: "Data Structures", code: "CS201", progress: 75 },
                { name: "Database Systems", code: "CS305", progress: 60 },
                { name: "Machine Learning", code: "CS401", progress: 45 },
                { name: "Web Development", code: "CS302", progress: 80 }
              ].map((course) => (
                <div key={course.code} className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <div>
                      <span className="font-medium">{course.name}</span>
                      <span className="ml-2 text-muted-foreground">{course.code}</span>
                    </div>
                    <span className="font-medium">{course.progress}%</span>
                  </div>
                  <Progress value={course.progress} />
                </div>
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Upcoming Deadlines</CardTitle>
              <CardDescription>Don't miss these!</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {dashboardData?.upcomingEvents?.map((item) => (
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
                {[
                  { course: "CS201", assessment: "Midterm Quiz", score: "45/50", grade: "A", date: "2024-03-15" },
                  { course: "CS305", assessment: "Assignment 2", score: "92/100", grade: "A+", date: "2024-03-12" },
                  { course: "CS401", assessment: "Project", score: "88/100", grade: "A", date: "2024-03-10" }
                ].map((grade, idx) => (
                  <TableRow key={idx}>
                    <TableCell className="font-medium">{grade.course}</TableCell>
                    <TableCell>{grade.assessment}</TableCell>
                    <TableCell>{grade.score}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{grade.grade}</Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{grade.date}</TableCell>
                  </TableRow>
                ))}
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
              <CardTitle className="text-2xl">4</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                Total Students
              </CardDescription>
              <CardTitle className="text-2xl">315</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <FileText className="h-4 w-4" />
                Pending Submissions
              </CardDescription>
              <CardTitle className="text-2xl">42</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription className="flex items-center gap-2">
                <Clock className="h-4 w-4" />
                Lectures Today
              </CardDescription>
              <CardTitle className="text-2xl">3</CardTitle>
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
              {[
                { course: "CS201", title: "Data Structures", time: "10:00 AM - 11:30 AM", room: "Room 301", students: 82 },
                { course: "CS305", title: "Database Systems", time: "2:00 PM - 3:30 PM", room: "Room 205", students: 76 },
                { course: "CS401", title: "Machine Learning", time: "4:00 PM - 5:30 PM", room: "Lab 102", students: 65 }
              ].map((lecture, idx) => (
                <div key={idx} className="rounded-lg border p-4 space-y-2">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{lecture.title}</div>
                      <div className="text-sm text-muted-foreground">{lecture.course}</div>
                    </div>
                    <Badge variant="outline">{lecture.students} students</Badge>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {lecture.time}
                    </div>
                    <div>{lecture.room}</div>
                  </div>
                  <Button size="sm" className="w-full">Mark Attendance</Button>
                </div>
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Pending Grading</CardTitle>
              <CardDescription>Submissions waiting for review</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {[
                { title: "Assignment 3", course: "CS201", submissions: 18, dueDate: "2 days ago" },
                { title: "Quiz 2", course: "CS305", submissions: 12, dueDate: "1 day ago" },
                { title: "Project Phase 2", course: "CS401", submissions: 8, dueDate: "Today" }
              ].map((item, idx) => (
                <div key={idx} className="rounded-lg border p-3 space-y-2">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium text-sm">{item.title}</div>
                      <div className="text-xs text-muted-foreground">{item.course}</div>
                    </div>
                    <Badge>{item.submissions} pending</Badge>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">Due: {item.dueDate}</span>
                    <Button size="sm" variant="outline">Review</Button>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Course Performance</CardTitle>
            <CardDescription>Average performance across your courses</CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Course</TableHead>
                  <TableHead>Students</TableHead>
                  <TableHead>Avg Grade</TableHead>
                  <TableHead>Attendance</TableHead>
                  <TableHead>Completion</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {[
                  { code: "CS201", name: "Data Structures", students: 82, avgGrade: "B+", attendance: 88, completion: 65 },
                  { code: "CS305", name: "Database Systems", students: 76, avgGrade: "A-", attendance: 92, completion: 55 },
                  { code: "CS401", name: "Machine Learning", students: 65, avgGrade: "A", attendance: 90, completion: 45 }
                ].map((course, idx) => (
                  <TableRow key={idx}>
                    <TableCell>
                      <div>
                        <div className="font-medium">{course.name}</div>
                        <div className="text-sm text-muted-foreground">{course.code}</div>
                      </div>
                    </TableCell>
                    <TableCell>{course.students}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{course.avgGrade}</Badge>
                    </TableCell>
                    <TableCell>{course.attendance}%</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Progress value={course.completion} className="w-20" />
                        <span className="text-sm">{course.completion}%</span>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
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
          <CardContent>
            <div className="flex items-center gap-1 text-xs text-green-600">
              <TrendingUp className="h-3 w-3" />
              <span>+12% from last year</span>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <Users className="h-4 w-4" />
              Faculty Members
            </CardDescription>
            <CardTitle className="text-2xl">118</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-1 text-xs text-green-600">
              <TrendingUp className="h-3 w-3" />
              <span>+5% from last year</span>
            </div>
          </CardContent>
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
            <CardTitle className="text-2xl">{dashboardData?.metrics.attendanceRate ?? 'N/A'}%</CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Department Statistics</CardTitle>
            <CardDescription>Student enrollment by department</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {[
                { name: "Computer Science", students: 450, faculty: 25, percentage: 21 },
                { name: "Electronics & Communication", students: 380, faculty: 22, percentage: 18 },
                { name: "Mechanical Engineering", students: 420, faculty: 28, percentage: 20 },
                { name: "Civil Engineering", students: 350, faculty: 20, percentage: 16 },
                { name: "Business Administration", students: 545, faculty: 23, percentage: 25 }
              ].map((dept, idx) => (
                <div key={idx} className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium">{dept.name}</span>
                    <span className="text-muted-foreground">{dept.students} students</span>
                  </div>
                  <Progress value={dept.percentage * 2} />
                  <div className="text-xs text-muted-foreground">{dept.faculty} faculty members</div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Alerts</CardTitle>
            <CardDescription>Items requiring attention</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {[
              { type: "warning", message: "15 students below 75% attendance threshold", action: "View Students" },
              { type: "info", message: "3 courses with low enrollment (< 50%)", action: "View Courses" },
              { type: "warning", message: "8 pending faculty leave requests", action: "Review Requests" },
              { type: "info", message: "Database backup completed successfully", action: "View Logs" }
            ].map((alert, idx) => (
              <div key={idx} className="rounded-lg border p-3 space-y-2">
                <div className="flex items-start gap-2">
                  <AlertCircle className={`h-4 w-4 mt-0.5 ${alert.type === 'warning' ? 'text-yellow-600' : 'text-blue-600'}`} />
                  <div className="flex-1">
                    <p className="text-sm">{alert.message}</p>
                  </div>
                </div>
                <Button size="sm" variant="outline" className="w-full">{alert.action}</Button>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>Latest actions across the system</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Activity</TableHead>
                <TableHead>User</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Time</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {dashboardData?.recentActivity?.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="font-medium">{item.message}</TableCell>
                  <TableCell>System</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="capitalize">{item.entity}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground">{format(new Date(item.timestamp), 'PPpp')}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
