"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api, endpoints } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { QrCode, CheckCircle, XCircle, Clock, Calendar, Loader2, Filter, Search, Camera, Download } from "lucide-react";
import { format } from "date-fns";
import { logger } from '@/lib/logger';

type AttendanceRecord = {
  id: number;
  courseName: string;
  courseId: number;
  date: string;
  status: 'present' | 'absent' | 'late' | 'excused';
  markedBy?: string;
};

type CourseAttendance = {
  courseId: number;
  courseName: string;
  present: number;
  total: number;
  percentage: number;
};

type QRCodeData = {
  data?: string;
};

const normalizeStatus = (status?: string): AttendanceRecord['status'] => {
  const normalized = (status || '').toLowerCase();
  if (normalized === 'present' || normalized === 'absent' || normalized === 'late' || normalized === 'excused') {
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
  const [filter, setFilter] = useState<string>('all');

  // Faculty QR generation state
  const [showQRForm, setShowQRForm] = useState(false);
  const [courseId, setCourseId] = useState("");
  const [lectureId, setLectureId] = useState("");
  const [qrLoading, setQrLoading] = useState(false);
  const [qrImageUrl, setQrImageUrl] = useState<string | null>(null);

  // Student QR marking state
  const [marking, setMarking] = useState(false);
  const [markingSuccess, setMarkingSuccess] = useState(false);

  useEffect(() => {
    const fetchAttendance = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch individual attendance records
        try {
          const recordsResponse = await api.get<ApiAttendanceRecord[]>(endpoints.attendance.myAttendance);
          const normalizedRecords = (Array.isArray(recordsResponse) ? recordsResponse : []).map<AttendanceRecord>((record, index) => ({
            id: record.id ?? index,
            courseName: record.courseName ?? (record.courseId ? `Course ${record.courseId}` : 'Unknown Course'),
            courseId: record.courseId ?? 0,
            date: record.date ?? new Date().toISOString(),
            status: normalizeStatus(record.status),
          }));
          setRecords(normalizedRecords);
        } catch (err) {
          logger.warn('Failed to fetch attendance records:', { error: err });
        }

        // Try to fetch course attendance stats
        try {
          const statsResponse = await api.get<ApiCourseStat[]>(endpoints.attendance.myCourseStats);
          const normalizedStats = (Array.isArray(statsResponse) ? statsResponse : []).map<CourseAttendance>((stat) => ({
            courseName: stat.courseName ?? (stat.courseId ? `Course ${stat.courseId}` : 'Unknown Course'),
            courseId: stat.courseId ?? 0,
            present: stat.present ?? 0,
            total: stat.total ?? 0,
            percentage: stat.percentage ?? 0,
          }));
          setCourseStats(normalizedStats);
        } catch (err) {
          logger.warn('Failed to fetch attendance stats:', { error: err });
        }
      } catch (err) {
        logger.error('Failed to fetch attendance:', err as Error);
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
      late: { icon: Clock, className: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400', label: 'Late' },
      excused: { icon: CheckCircle, className: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400', label: 'Excused' }
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
      
      if (!resp.ok) {
        throw new Error('Failed to generate QR code');
      }

      // Get the image as blob
      const blob = await resp.blob();
      const url = URL.createObjectURL(blob);
      setQrImageUrl(url);
    } catch (err) {
      logger.error('QR generation error:', err as Error);
      setError(err instanceof Error ? err.message : 'Failed to generate QR code');
    } finally {
      setQrLoading(false);
    }
  };

  const markAttendanceWithQR = async (qrData: string) => {
    try {
      setMarking(true);
      setError(null);
      setMarkingSuccess(false);

      const response = await api.post(endpoints.attendance.processQR, {
        qrcode_data: qrData
      });

      setMarkingSuccess(true);
      
      // Refresh attendance data
      setTimeout(() => {
        window.location.reload();
      }, 2000);
    } catch (err) {
      logger.error('Attendance marking error:', err as Error);
      setError(err instanceof Error ? err.message : 'Failed to mark attendance');
    } finally {
      setMarking(false);
    }
  };

  const filteredRecords = records.filter(record => {
    if (filter === 'all') return true;
    return record.status === filter;
  });

  if (loading) {
    return (
      <div className="flex h-[60vh] items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin" />
        <span className="ml-2">Loading attendance...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Attendance</h1>
          <p className="text-muted-foreground">
            {user?.role === 'student' ? 'View and manage your attendance' : 'Generate QR codes and manage attendance'}
          </p>
        </div>
        
        {user?.role === 'faculty' && (
          <Dialog open={showQRForm} onOpenChange={setShowQRForm}>
            <DialogTrigger asChild>
              <Button>
                <QrCode className="mr-2 h-4 w-4" />
                Generate QR Code
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Generate Attendance QR Code</DialogTitle>
                <DialogDescription>
                  Enter course and lecture details to generate a QR code for attendance marking.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div>
                  <Label htmlFor="courseId">Course ID</Label>
                  <Input
                    id="courseId"
                    type="number"
                    value={courseId}
                    onChange={(e) => setCourseId(e.target.value)}
                    placeholder="Enter course ID"
                  />
                </div>
                <div>
                  <Label htmlFor="lectureId">Lecture ID</Label>
                  <Input
                    id="lectureId"
                    type="number"
                    value={lectureId}
                    onChange={(e) => setLectureId(e.target.value)}
                    placeholder="Enter lecture ID"
                  />
                </div>
                <Button 
                  onClick={generateQR} 
                  disabled={qrLoading || !courseId || !lectureId}
                  className="w-full"
                >
                  {qrLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Generating...
                    </>
                  ) : (
                    <>
                      <QrCode className="mr-2 h-4 w-4" />
                      Generate QR Code
                    </>
                  )}
                </Button>
                
                {qrImageUrl && (
                  <div className="text-center">
                    <h3 className="text-lg font-semibold mb-2">Generated QR Code</h3>
                    <img src={qrImageUrl} alt="Attendance QR Code" className="mx-auto border rounded-lg" />
                    <Button 
                      onClick={() => {
                        const link = document.createElement('a');
                        link.href = qrImageUrl;
                        link.download = `qr-code-course-${courseId}-lecture-${lectureId}.png`;
                        link.click();
                      }}
                      className="mt-2"
                      variant="outline"
                    >
                      <Download className="mr-2 h-4 w-4" />
                      Download QR Code
                    </Button>
                  </div>
                )}
              </div>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {error && (
        <Card className="border-red-200 bg-red-50 dark:border-red-800 dark:bg-red-950">
          <CardContent className="pt-6">
            <p className="text-red-800 dark:text-red-200">{error}</p>
          </CardContent>
        </Card>
      )}

      {markingSuccess && (
        <Card className="border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950">
          <CardContent className="pt-6">
            <p className="text-green-800 dark:text-green-200">Attendance marked successfully!</p>
          </CardContent>
        </Card>
      )}

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="records">Records</TabsTrigger>
          {user?.role === 'student' && (
            <TabsTrigger value="scan">Scan QR</TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Overall Attendance</CardTitle>
                <CheckCircle className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{overallAttendance}%</div>
                <Progress value={overallAttendance} className="mt-2" />
                <p className="text-xs text-muted-foreground mt-2">
                  {courseStats.reduce((acc, c) => acc + c.present, 0)} of {courseStats.reduce((acc, c) => acc + c.total, 0)} sessions
                </p>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Courses</CardTitle>
                <Calendar className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{courseStats.length}</div>
                <p className="text-xs text-muted-foreground">
                  Active courses this semester
                </p>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Present</CardTitle>
                <CheckCircle className="h-4 w-4 text-green-600" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">
                  {records.filter(r => r.status === 'present').length}
                </div>
                <p className="text-xs text-muted-foreground">
                  Sessions attended
                </p>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Absent</CardTitle>
                <XCircle className="h-4 w-4 text-red-600" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">
                  {records.filter(r => r.status === 'absent').length}
                </div>
                <p className="text-xs text-muted-foreground">
                  Sessions missed
                </p>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Course-wise Attendance</CardTitle>
              <CardDescription>Your attendance percentage for each course</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {courseStats.map((course) => (
                  <div key={course.courseId} className="flex items-center space-x-4">
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium truncate">{course.courseName}</p>
                      <p className="text-sm text-muted-foreground">
                        {course.present} of {course.total} sessions
                      </p>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Progress value={course.percentage} className="w-20" />
                      <span className="text-sm font-medium">{course.percentage}%</span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="records" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Attendance Records</CardTitle>
              <CardDescription>Your detailed attendance history</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center space-x-2 mb-4">
                <Filter className="h-4 w-4" />
                <Select value={filter} onValueChange={setFilter}>
                  <SelectTrigger className="w-48">
                    <SelectValue placeholder="Filter by status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Records</SelectItem>
                    <SelectItem value="present">Present</SelectItem>
                    <SelectItem value="absent">Absent</SelectItem>
                    <SelectItem value="late">Late</SelectItem>
                    <SelectItem value="excused">Excused</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Course</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Marked By</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredRecords.map((record) => (
                    <TableRow key={record.id}>
                      <TableCell className="font-medium">
                        {format(new Date(record.date), 'MMM dd, yyyy')}
                      </TableCell>
                      <TableCell>{record.courseName}</TableCell>
                      <TableCell>{getStatusBadge(record.status)}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {record.markedBy || 'System'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              
              {filteredRecords.length === 0 && (
                <div className="text-center py-8 text-muted-foreground">
                  No attendance records found
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {user?.role === 'student' && (
          <TabsContent value="scan" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Scan QR Code</CardTitle>
                <CardDescription>
                  Use your device camera to scan the QR code provided by your instructor
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="text-center">
                  <Camera className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
                  <p className="text-sm text-muted-foreground mb-4">
                    Point your camera at the QR code displayed by your instructor
                  </p>
                  <Button onClick={() => {
                    // In a real implementation, this would open camera and scan QR
                    const qrData = prompt("Enter QR code data (demo):");
                    if (qrData) {
                      markAttendanceWithQR(qrData);
                    }
                  }} disabled={marking}>
                    {marking ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Marking Attendance...
                      </>
                    ) : (
                      <>
                        <Camera className="mr-2 h-4 w-4" />
                        Scan & Mark Attendance
                      </>
                    )}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        )}
      </Tabs>
    </div>
  );
}