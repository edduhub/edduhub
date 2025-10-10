"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Plus, Search, UserPlus, Download, Filter } from "lucide-react";

type Student = {
  id: number;
  name: string;
  rollNo: string;
  email: string;
  department: string;
  semester: number;
  gpa: number;
  attendance: number;
  enrolledCourses: number;
  status: 'active' | 'inactive' | 'suspended';
  avatar?: string;
};

export default function StudentsPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedDepartment, setSelectedDepartment] = useState<string>("all");
  
  const [students] = useState<Student[]>([
    {
      id: 1,
      name: "Aarav Kumar",
      rollNo: "CS-2023-001",
      email: "aarav.kumar@college.edu",
      department: "Computer Science",
      semester: 6,
      gpa: 3.85,
      attendance: 92,
      enrolledCourses: 5,
      status: 'active'
    },
    {
      id: 2,
      name: "Mira Singh",
      rollNo: "CS-2023-002",
      email: "mira.singh@college.edu",
      department: "Computer Science",
      semester: 6,
      gpa: 3.92,
      attendance: 95,
      enrolledCourses: 5,
      status: 'active'
    },
    {
      id: 3,
      name: "Rahul Patel",
      rollNo: "EC-2023-015",
      email: "rahul.patel@college.edu",
      department: "Electronics",
      semester: 4,
      gpa: 3.65,
      attendance: 88,
      enrolledCourses: 6,
      status: 'active'
    },
    {
      id: 4,
      name: "Priya Sharma",
      rollNo: "ME-2023-022",
      email: "priya.sharma@college.edu",
      department: "Mechanical",
      semester: 8,
      gpa: 3.78,
      attendance: 90,
      enrolledCourses: 4,
      status: 'active'
    },
    {
      id: 5,
      name: "Arjun Verma",
      rollNo: "CS-2023-010",
      email: "arjun.verma@college.edu",
      department: "Computer Science",
      semester: 2,
      gpa: 3.45,
      attendance: 75,
      enrolledCourses: 6,
      status: 'active'
    }
  ]);

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

  const filteredStudents = students.filter(student => {
    const matchesSearch = 
      student.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      student.rollNo.toLowerCase().includes(searchQuery.toLowerCase()) ||
      student.email.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesDepartment = 
      selectedDepartment === "all" || 
      student.department === selectedDepartment;
    
    return matchesSearch && matchesDepartment;
  });

  const departments = ["all", ...Array.from(new Set(students.map(s => s.department)))];

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
          <Button variant="outline">
            <Download className="mr-2 h-4 w-4" />
            Export
          </Button>
          <Button>
            <UserPlus className="mr-2 h-4 w-4" />
            Add Student
          </Button>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Students
            </CardTitle>
            <div className="text-2xl font-bold">{students.length}</div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Average GPA
            </CardTitle>
            <div className="text-2xl font-bold">
              {(students.reduce((acc, s) => acc + s.gpa, 0) / students.length).toFixed(2)}
            </div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Avg Attendance
            </CardTitle>
            <div className="text-2xl font-bold">
              {Math.round(students.reduce((acc, s) => acc + s.attendance, 0) / students.length)}%
            </div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active Students
            </CardTitle>
            <div className="text-2xl font-bold text-green-600">
              {students.filter(s => s.status === 'active').length}
            </div>
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
                          {student.name.split(' ').map(n => n[0]).join('').toUpperCase()}
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
                  <TableCell>
                    <Button variant="outline" size="sm">
                      View
                    </Button>
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