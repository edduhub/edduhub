"use client";

import { useState, useMemo } from "react";
import { useStudents, useCreateStudent, useUpdateStudent } from "@/lib/api-hooks";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Plus, Search, UserPlus, Download, Loader2 } from "lucide-react";
import { logger } from '@/lib/logger';

export default function StudentsPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedDepartment, setSelectedDepartment] = useState<string>("all");
  const [showCreate, setShowCreate] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Create form state
  const [newStudent, setNewStudent] = useState({
    firstName: "",
    lastName: "",
    email: "",
    rollNo: "",
    department: "",
    semester: 1,
  });

  // React Query hooks
  const { data: students = [], isLoading: loading } = useStudents();
  const createStudent = useCreateStudent();
  const updateStudent = useUpdateStudent();

  // Filter students using useMemo
  const filteredStudents = useMemo(() => {
    return students.filter(student => {
      const fullName = `${student.firstName} ${student.lastName}`.toLowerCase();
      const matchesSearch = 
        fullName.includes(searchQuery.toLowerCase()) ||
        student.rollNo.toLowerCase().includes(searchQuery.toLowerCase()) ||
        student.email.toLowerCase().includes(searchQuery.toLowerCase());
      
      const matchesDepartment = 
        selectedDepartment === "all" || 
        student.departmentName === selectedDepartment;
      
      return matchesSearch && matchesDepartment;
    });
  }, [students, searchQuery, selectedDepartment]);

  // Get unique departments using useMemo
  const departments = useMemo(() => {
    return ["all", ...Array.from(new Set(students.map(s => s.departmentName).filter(Boolean)))];
  }, [students]);

  // Calculate statistics using useMemo
  const stats = useMemo(() => {
    if (students.length === 0) {
      return {
        total: 0,
        avgGpa: '0.00',
        avgAttendance: 0,
        activeCount: 0,
      };
    }

    const total = students.length;
    const avgGpa = (students.reduce((acc, s) => acc + (s.gpa ?? 0), 0) / total).toFixed(2);
    const activeCount = students.filter(s => s.status === 'active').length;

    return { total, avgGpa, avgAttendance: 0, activeCount };
  }, [students]);

  const getStatusBadge = (status: string) => {
    const styles = {
      active: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      inactive: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
      suspended: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    };
    return <Badge className={styles[status as keyof typeof styles]}>{status}</Badge>;
  };

  const getGPAColor = (gpa: number) => {
    if (gpa >= 3.7) return 'text-green-600';
    if (gpa >= 3.0) return 'text-blue-600';
    if (gpa >= 2.5) return 'text-yellow-600';
    return 'text-red-600';
  };

  const handleExport = async () => {
    try {
      setIsExporting(true);
      const resp = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/batch/students/export`, {
        method: 'GET',
        credentials: 'include',
      });
      if (!resp.ok) throw new Error('Export failed');
      const blob = await resp.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'students_export.csv';
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (e) {
      logger.error('Error occurred', e instanceof Error ? e : new Error(String(e)));
      setError('Export failed');
    } finally {
      setIsExporting(false);
    }
  };

  const handleCreate = async () => {
    try {
      setError(null);
      await createStudent.mutateAsync({
        first_name: newStudent.firstName,
        last_name: newStudent.lastName,
        email: newStudent.email,
        roll_no: newStudent.rollNo,
        department: newStudent.department,
        semester: Number(newStudent.semester),
      });
      setShowCreate(false);
      setNewStudent({ firstName: '', lastName: '', email: '', rollNo: '', department: '', semester: 1 });
    } catch (e) {
      logger.error('Error occurred', e instanceof Error ? e : new Error(String(e)));
      setError('Failed to create student');
    }
  };

  const toggleStatus = async (student: typeof students[0]) => {
    try {
      const nextStatus = student.status === 'active' ? 'inactive' : 'active';
      await updateStudent.mutateAsync({
        id: student.id,
        data: { status: nextStatus },
      });
    } catch (e) {
      logger.error('Error occurred', e instanceof Error ? e : new Error(String(e)));
      setError('Failed to update status');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Students</h1>
          <p className="text-muted-foreground">
            Manage student information and performance
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleExport} disabled={isExporting}>
            {isExporting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Download className="mr-2 h-4 w-4" />}
            Export
          </Button>
          <Button onClick={() => setShowCreate(v => !v)}>
            <UserPlus className="mr-2 h-4 w-4" />
            {showCreate ? 'Close' : 'Add Student'}
          </Button>
        </div>
      </div>

      {showCreate && (
        <Card>
          <CardHeader>
            <CardTitle>New Student</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <label className="text-sm font-medium">First Name</label>
                <Input value={newStudent.firstName} onChange={e => setNewStudent({ ...newStudent, firstName: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Last Name</label>
                <Input value={newStudent.lastName} onChange={e => setNewStudent({ ...newStudent, lastName: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Email</label>
                <Input type="email" value={newStudent.email} onChange={e => setNewStudent({ ...newStudent, email: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Roll No</label>
                <Input value={newStudent.rollNo} onChange={e => setNewStudent({ ...newStudent, rollNo: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Department</label>
                <Input value={newStudent.department} onChange={e => setNewStudent({ ...newStudent, department: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Semester</label>
                <Input type="number" min={1} max={12} value={newStudent.semester} onChange={e => setNewStudent({ ...newStudent, semester: Number(e.target.value || 1) })} />
              </div>
            </div>
            <div className="mt-4 flex justify-end">
              <Button onClick={handleCreate} disabled={createStudent.isPending}>
                {createStudent.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Plus className="mr-2 h-4 w-4" />}
                Create
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {error && (
        <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Students
            </CardTitle>
            <div className="text-2xl font-bold">{stats.total}</div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Average GPA
            </CardTitle>
            <div className="text-2xl font-bold">{stats.avgGpa}</div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Avg Attendance
            </CardTitle>
            <div className="text-2xl font-bold">{stats.avgAttendance}%</div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active Students
            </CardTitle>
            <div className="text-2xl font-bold text-green-600">{stats.activeCount}</div>
          </CardHeader>
        </Card>
      </div>

      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search students..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex gap-2">
          {departments.map((dept) => (
            <Button
              key={dept}
              variant={selectedDepartment === dept ? "default" : "outline"}
              size="sm"
              onClick={() => setSelectedDepartment(dept)}
            >
              {dept === "all" ? "All" : dept}
            </Button>
          ))}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Students</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="h-6 w-6 animate-spin" />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Student</TableHead>
                  <TableHead>Roll No</TableHead>
                  <TableHead>Department</TableHead>
                  <TableHead>Semester</TableHead>
                  <TableHead>GPA</TableHead>
                  <TableHead>Attendance</TableHead>
                  <TableHead>Courses</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredStudents.map((student) => (
                  <TableRow key={student.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar className="h-9 w-9">
                          <AvatarImage src={student.avatar} />
                          <AvatarFallback>
                            {student.name.split(' ').map((n: string) => n[0]).join('').toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="font-medium">{student.name}</div>
                          <div className="text-sm text-muted-foreground">{student.email}</div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-sm">{student.rollNo}</TableCell>
                    <TableCell>{student.department}</TableCell>
                    <TableCell>{student.semester}</TableCell>
                    <TableCell>
                      <span className={`font-medium ${getGPAColor(student.gpa)}`}>
                        {student.gpa.toFixed(2)}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className={student.attendance < 75 ? 'text-red-600 font-medium' : ''}>
                        {student.attendance}%
                      </span>
                    </TableCell>
                    <TableCell>{student.enrolledCourses}</TableCell>
                    <TableCell>{getStatusBadge(student.status)}</TableCell>
                    <TableCell className="space-x-2">
                      <Button variant="outline" size="sm">
                        View
                      </Button>
                      <Button variant="outline" size="sm" onClick={() => toggleStatus(student)} disabled={updateStudent.isPending}>
                        {student.status === 'active' ? 'Deactivate' : 'Activate'}
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
