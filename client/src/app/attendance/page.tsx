"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api, endpoints } from "@/lib/api-client";
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

type ApiAttendanceRecord = {
  id?: number;
  courseId?: number;
  courseName?: string;
  date?: string;
  status?: string;
};

type ApiCourseStat = {
  courseId?: number;
  courseName?: string;
  present?: number;
  total?: number;
  percentage?: number;
};

const normalizeStatus = (status?: string): AttendanceRecord['status'] => {
  const normalized = (status || '').toLowerCase();
  if (normalized === 'present' || normalized === 'absent' || normalized === 'late') {
    return normalized;
  }
  return 'absent';
};

export default function AttendancePage() {
  const { user } = useAuth();
  const [records, setRecords] = useState<AttendanceRecord[]>([]);
  const [courseStats, setCourseStats] = useState<CourseAttendance[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Faculty QR generation state
  const [showQRForm, setShowQRForm] = useState(false);
  const [courseId, setCourseId] = useState("");
  const [lectureId, setLectureId] = useState("");
  const [qrLoading, setQrLoading] = useState(false);
  const [qrImageUrl, setQrImageUrl] = useState<string | null>(null);

  // Student QR marking state
  const [qrToken, setQrToken] = useState("");
  const [marking, setMarking] = useState(false);

  useEffect(() => {
    const fetchAttendance = async () => {
      try {
        setLoading(true);
        // Fetch individual attendance records
        try {
          const recordsResponse = await api.get<ApiAttendanceRecord[]>(endpoints.attendance.myAttendance);
          const normalizedRecords = (Array.isArray(recordsResponse) ? recordsResponse : []).map<AttendanceRecord>((record, index) => ({
            id: record.id ?? index,
            courseName: record.courseName ?? (record.courseId ? `Course ${record.courseId}` : 'Unknown Course'),
            courseCode: record.courseId ? `COURSE-${record.courseId}` : 'COURSE',
            date: record.date ?? new Date().toISOString(),
            status: normalizeStatus(record.status),
          }));
          setRecords(normalizedRecords);
        } catch (err) {
          console.warn('Failed to fetch attendance records:', err);
        }

        // Try to fetch course attendance stats
        try {
          const statsResponse = await api.get<ApiCourseStat[]>(endpoints.attendance.myCourseStats);
          const normalizedStats = (Array.isArray(statsResponse) ? statsResponse : []).map<CourseAttendance>((stat) => ({
            courseName: stat.courseName ?? (stat.courseId ? `Course ${stat.courseId}` : 'Unknown Course'),
            courseCode: stat.courseId ? `COURSE-${stat.courseId}` : 'COURSE',
            present: stat.present ?? 0,
            total: stat.total ?? 0,
            percentage: stat.percentage ?? 0,
          }));
          setCourseStats(normalizedStats);
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

  const overallAttendance = (() => {
    const total = courseStats.reduce((acc, c) => acc + c.total, 0);
    if (!total) return 0;
    const present = courseStats.reduce((acc, c) => acc + c.present, 0);
    return Math.round((present / total) * 100);
  })();

  const getStatusBadge = (status: string) => {
    const config = {
      present: { icon: CheckCircle, className: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400', label: 'Present' },
      absent: { icon: XCircle, className: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400', label: 'Absent' },
      late: { icon: Clock, className: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400', label: 'Late' }
    } as const;
    const { icon: Icon, className, label } = config[status as keyof typeof config];
    return (
      <Badge className={className}>
        <Icon className="mr-1 h-3 w-3" />
        {label}
      </Badge>
    );
  };

  const generateQR = async () => {
    try {
      setQrLoading(true);
      setError(null);
      const cid = Number(courseId);
      const lid = Number(lectureId);
      if (!cid || !lid) throw new Error('Course ID and Lecture ID are required');
      const resp = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/attendance/course/${cid}/lecture/${lid}/qrcode`, {
        method: 'GET',
        credentials: 'include',
      });
      if (!resp.ok) throw new Error('Failed to generate QR');
      const blob = await resp.blob();
      const url = URL.createObjectURL(blob);
      setQrImageUrl(url);
    } catch (e: any) {
      console.error(e);
      setError(e?.message || 'Failed to generate QR');
    } finally {
      setQrLoading(false);
    }
  };

  const processQR = async () => {
    try {
      setMarking(true);
      setError(null);
      await api.post('/api/attendance/process-qr', { qrcode_data: qrToken });
      // Refresh records after marking
      const refreshed = await api.get<ApiAttendanceRecord[]>(endpoints.attendance.myAttendance);
      const normalizedRecords = (Array.isArray(refreshed) ? refreshed : []).map<AttendanceRecord>((record, index) => ({
        id: record.id ?? index,
        courseName: record.courseName ?? (record.courseId ? `Course ${record.courseId}` : 'Unknown Course'),
        courseCode: record.courseId ? `COURSE-${record.courseId}` : 'COURSE',
        date: record.date ?? new Date().toISOString(),
        status: normalizeStatus(record.status),
      }));
      setRecords(normalizedRecords);
      setQrToken("");
    } catch (e: any) {
      console.error(e);
      setError(e?.message || 'Failed to mark attendance');
    } finally {
      setMarking(false);
    }
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
          <Button onClick={() => setShowQRForm(v => !v)}>
            <QrCode className="mr-2 h-4 w-4" />
            {showQRForm ? 'Close' : 'Generate QR Code'}
          </Button>
        )}
      </div>

      {user?.role === 'faculty' && showQRForm && (
        <Card>
          <CardHeader>
            <CardTitle>Generate Attendance QR</CardTitle>
            <CardDescription>Provide course and lecture IDs</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <label className="text-sm font-medium">Course ID</label>
                <input className="w-full rounded-md border px-3 py-2" value={courseId} onChange={e => setCourseId(e.target.value)} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Lecture ID</label>
                <input className="w-full rounded-md border px-3 py-2" value={lectureId} onChange={e => setLectureId(e.target.value)} />
              </div>
            </div>
            <div className="mt-4 flex items-center gap-4">
              <Button onClick={generateQR} disabled={qrLoading}>
                {qrLoading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <QrCode className="mr-2 h-4 w-4" />}
                Generate
              </Button>
              {qrImageUrl && (
                <img src={qrImageUrl} alt="Attendance QR" className="h-40 w-40 border" />
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {user?.role === 'student' && (
        <Card>
          <CardHeader>
            <CardTitle>Mark Attendance via QR</CardTitle>
            <CardDescription>Paste scanned QR token if camera scan is unavailable</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <input className="flex-1 rounded-md border px-3 py-2" placeholder="QR token" value={qrToken} onChange={e => setQrToken(e.target.value)} />
              <Button onClick={processQR} disabled={marking || !qrToken}>
                {marking ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <CheckCircle className="mr-2 h-4 w-4" />}
                Mark
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

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
