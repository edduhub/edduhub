"use client";

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { 
  Users, 
  BookOpen, 
  Calendar, 
  TrendingUp, 
  Bell,
  GraduationCap,
  FileText,
  AlertCircle
} from 'lucide-react';
import type { Student, Grade, Attendance, Announcement, Assignment } from '@/lib/types';

export default function ParentDashboard() {
  const [selectedStudent, setSelectedStudent] = useState<Student | null>(null);
  const [students, setStudents] = useState<Student[]>([]);
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Fetch parent's linked students and announcements
    const fetchData = async () => {
      try {
        // Mock data - replace with actual API calls
        setStudents([
          {
            id: 1,
            userId: 'student1',
            rollNo: 'CS2024001',
            firstName: 'John',
            lastName: 'Doe',
            email: 'john.doe@college.edu',
            semester: 3,
            departmentId: 1,
            departmentName: 'Computer Science',
            collegeId: 'college1',
            status: 'active',
            gpa: 3.7,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          }
        ]);

        setAnnouncements([
          {
            id: 1,
            title: 'Midterm Examination Schedule Released',
            content: 'Midterm exams will start from next Monday...',
            priority: 'high',
            targetAudience: ['all'],
            publishedAt: new Date().toISOString(),
            authorId: 'admin1',
            authorName: 'Admin',
            collegeId: 'college1',
          }
        ]);

        if (students.length > 0) {
          setSelectedStudent(students[0]);
        }
      } catch (error) {
        console.error('Failed to fetch data:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-muted/10">
      <div className="container mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Parent Portal</h1>
            <p className="text-muted-foreground mt-1">Monitor your child's academic progress</p>
          </div>
          <Button variant="outline">
            <Bell className="w-4 h-4 mr-2" />
            Notifications
          </Button>
        </div>

        {/* Student Selection */}
        {students.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="w-5 h-5" />
                Select Student
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex gap-2 flex-wrap">
                {students.map((student) => (
                  <Button
                    key={student.id}
                    variant={selectedStudent?.id === student.id ? 'default' : 'outline'}
                    onClick={() => setSelectedStudent(student)}
                  >
                    {student.firstName} {student.lastName}
                    {selectedStudent?.id === student.id && <Badge className="ml-2">Active</Badge>}
                  </Button>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {selectedStudent && (
          <>
            {/* Student Overview */}
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">
                    Current Semester
                  </CardTitle>
                  <GraduationCap className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">Semester {selectedStudent.semester}</div>
                  <p className="text-xs text-muted-foreground">
                    {selectedStudent.departmentName}
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">
                    GPA
                  </CardTitle>
                  <TrendingUp className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{selectedStudent.gpa?.toFixed(2) || 'N/A'}</div>
                  <p className="text-xs text-muted-foreground">
                    Current academic year
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">
                    Enrollment
                  </CardTitle>
                  <BookOpen className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {selectedStudent.enrolledCourses || 0} Courses
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Active enrollment
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">
                    Status
                  </CardTitle>
                  <AlertCircle className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold capitalize">{selectedStudent.status}</div>
                  <p className="text-xs text-muted-foreground">
                    Account status
                  </p>
                </CardContent>
              </Card>
            </div>

            {/* Detailed Tabs */}
            <Tabs defaultValue="grades" className="space-y-4">
              <TabsList className="grid w-full grid-cols-4 lg:w-auto">
                <TabsTrigger value="grades">Grades</TabsTrigger>
                <TabsTrigger value="attendance">Attendance</TabsTrigger>
                <TabsTrigger value="assignments">Assignments</TabsTrigger>
                <TabsTrigger value="announcements">Announcements</TabsTrigger>
              </TabsList>

              <TabsContent value="grades" className="space-y-4">
                <ParentGradesView studentId={selectedStudent.id} />
              </TabsContent>

              <TabsContent value="attendance" className="space-y-4">
                <ParentAttendanceView studentId={selectedStudent.id} />
              </TabsContent>

              <TabsContent value="assignments" className="space-y-4">
                <ParentAssignmentsView studentId={selectedStudent.id} />
              </TabsContent>

              <TabsContent value="announcements" className="space-y-4">
                <ParentAnnouncementsView announcements={announcements} />
              </TabsContent>
            </Tabs>
          </>
        )}

        {/* No students linked */}
        {students.length === 0 && (
          <Card>
            <CardHeader>
              <CardTitle>No Students Linked</CardTitle>
              <CardDescription>
                You haven't been linked to any student accounts yet. Please contact the college administration.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button>Contact Administration</Button>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}

// Sub-components for different views
function ParentGradesView({ studentId }: { studentId: number }) {
  const [grades, setGrades] = useState<Grade[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Fetch grades for the selected student
    // Mock data - replace with API call
    setTimeout(() => {
      setGrades([
        {
          id: 1,
          studentId: studentId,
          courseId: 101,
          assessmentType: 'Assignment',
          assessmentName: 'Mid-term Exam',
          score: 85,
          maxScore: 100,
          percentage: 85,
          gradedDate: new Date().toISOString(),
          collegeId: 'college1',
        }
      ]);
      setIsLoading(false);
    }, 500);
  }, [studentId]);

  if (isLoading) {
    return <div className="text-center py-8 text-muted-foreground">Loading grades...</div>;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="w-5 h-5" />
          Recent Grades
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {grades.map((grade) => (
            <div key={grade.id} className="flex items-center justify-between p-4 border rounded-lg">
              <div>
                <h4 className="font-semibold">{grade.assessmentName}</h4>
                <p className="text-sm text-muted-foreground">{grade.assessmentType}</p>
              </div>
              <div className="text-right">
                <div className="text-2xl font-bold">{grade.percentage}%</div>
                <Badge variant={grade.percentage >= 75 ? 'default' : 'destructive'}>
                  {grade.percentage >= 75 ? 'Good' : 'Needs Improvement'}
                </Badge>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function ParentAttendanceView({ studentId }: { studentId: number }) {
  const [attendance, setAttendance] = useState<Attendance[]>([]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Calendar className="w-5 h-5" />
          Attendance Overview
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-muted-foreground">
          Attendance data will be loaded here
        </div>
      </CardContent>
    </Card>
  );
}

function ParentAssignmentsView({ studentId }: { studentId: number }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <BookOpen className="w-5 h-5" />
          Upcoming Assignments
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-center py-8 text-muted-foreground">
          Assignments will be loaded here
        </div>
      </CardContent>
    </Card>
  );
}

function ParentAnnouncementsView({ announcements }: { announcements: Announcement[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Bell className="w-5 h-5" />
          School Announcements
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {announcements.map((announcement) => (
            <div key={announcement.id} className="p-4 border rounded-lg space-y-2">
              <div className="flex items-center justify-between">
                <h4 className="font-semibold">{announcement.title}</h4>
                <Badge 
                  variant={announcement.priority === 'urgent' ? 'destructive' : 
                            announcement.priority === 'high' ? 'default' : 'secondary'}
                >
                  {announcement.priority}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground">{announcement.content}</p>
              <p className="text-xs text-muted-foreground">
                {new Date(announcement.publishedAt).toLocaleDateString()}
              </p>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
