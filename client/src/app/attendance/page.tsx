"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { QrCode, CheckCircle, XCircle, Clock, Calendar, Loader2 } from "lucide-react";
import { format } from "date-fns";

type AttendanceRecord = {
  id: number;
  courseName: string;
  courseCode: string;
  date: string;
  status: 'present' | 'absent' | 'late';
  markedBy?: string;
};

type CourseAttendance = {
  courseName: string;
  courseCode: string;
  present: number;
  total: number;
  percentage: number;
};

export default function AttendancePage() {
  const { user } = useAuth();
  const [records, setRecords] = useState<AttendanceRecord[]>([]);
  const [courseStats, setCourseStats] = useState<CourseAttendance[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchAttendance = async () => {
      try {
        setLoading(true);
        // Fetch individual attendance records
        try {
          const recordsResponse = await api.get('/api/attendance/student/me');
          setRecords(Array.isArray(recordsResponse) ? recordsResponse : []);
        } catch (err) {
          console.warn('Failed to fetch attendance records:', err);
        }

        // Try to fetch course attendance stats
        try {
          const statsResponse = await api.get('/api/attendance/stats/courses');
          setCourseStats(Array.isArray(statsResponse) ? statsResponse : []);
        } catch (err) {
          console.warn('Failed to fetch attendance stats:', err);
        }
      } catch (err) {
        console.error('Failed to fetch attendance:', err);
        setError('Failed to load attendance data');
      } finally {
        setLoading(false);
      }
    };

    fetchAttendance();
  }, []);

  const overallAttendance = Math.round(
    (courseStats.reduce((acc, c) => acc + c.present, 0) / 
     courseStats.reduce((acc, c) => acc + c.total, 0)) * 100
  );

  const getStatusBadge = (status: string) => {
    const config = {
      present: { icon: CheckCircle, className: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400', label: 'Present' },
      absent: { icon: XCircle, className: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400', label: 'Absent' },
      late: { icon: Clock, className: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400', label: 'Late' }
    };
    const { icon: Icon, className, label } = config[status as keyof typeof config];
    return (
      <Badge className={className}>
        <Icon className="mr-1 h-3 w-3" />
        {label}
      </Badge>
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Attendance</h1>
          <p className="text-muted-foreground">
            {user?.role === 'student' ? 'Track your attendance across courses' : 'Mark and manage student attendance'}
          </p>
        </div>
        {user?.role === 'faculty' && (
          <Button>
            <QrCode className="mr-2 h-4 w-4" />
            Generate QR Code
          </Button>
        )}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Overall Attendance</CardTitle>
          <CardDescription>Your attendance across all courses</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Attendance Rate</span>
              <span className="text-2xl font-bold">{overallAttendance}%</span>
            </div>
            <Progress value={overallAttendance} className="h-3" />
            <p className="text-xs text-muted-foreground">
              {courseStats.reduce((acc, c) => acc + c.present, 0)} present out of{' '}
              {courseStats.reduce((acc, c) => acc + c.total, 0)} total classes
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Course-wise Attendance</CardTitle>
          <CardDescription>Attendance percentage for each course</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {courseStats.map((course) => (
              <div key={course.courseCode} className="space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <div>
                    <span className="font-medium">{course.courseName}</span>
                    <span className="ml-2 text-muted-foreground">{course.courseCode}</span>
                  </div>
                  <span className="font-medium">{course.percentage.toFixed(1)}%</span>
                </div>
                <Progress value={course.percentage} />
                <p className="text-xs text-muted-foreground">
                  {course.present}/{course.total} classes
                </p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Recent Attendance</CardTitle>
          <CardDescription>Your recent attendance records</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Date</TableHead>
                <TableHead>Course</TableHead>
                <TableHead>Status</TableHead>
                {user?.role === 'student' && <TableHead>Marked By</TableHead>}
              </TableRow>
            </TableHeader>
            <TableBody>
              {records.map((record) => (
                <TableRow key={record.id}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Calendar className="h-4 w-4 text-muted-foreground" />
                      {format(new Date(record.date), 'MMM dd, yyyy')}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div>
                      <div className="font-medium">{record.courseName}</div>
                      <div className="text-sm text-muted-foreground">{record.courseCode}</div>
                    </div>
                  </TableCell>
                  <TableCell>{getStatusBadge(record.status)}</TableCell>
                  {user?.role === 'student' && <TableCell>{record.markedBy}</TableCell>}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
