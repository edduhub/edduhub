"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Users, BookOpen, Search, Building2 } from "lucide-react";

type Department = {
  id: number;
  name: string;
  code: string;
  hodName?: string;
  studentCount: number;
  facultyCount: number;
  coursesCount: number;
};

export default function DepartmentsPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [departments] = useState<Department[]>([
    {
      id: 1,
      name: "Computer Science",
      code: "CS",
      hodName: "Dr. Rajesh Kumar",
      studentCount: 450,
      facultyCount: 25,
      coursesCount: 42
    },
    {
      id: 2,
      name: "Electronics & Communication",
      code: "EC",
      hodName: "Prof. Priya Sharma",
      studentCount: 380,
      facultyCount: 22,
      coursesCount: 38
    },
    {
      id: 3,
      name: "Mechanical Engineering",
      code: "ME",
      hodName: "Dr. Amit Patel",
      studentCount: 420,
      facultyCount: 28,
      coursesCount: 45
    },
    {
      id: 4,
      name: "Civil Engineering",
      code: "CE",
      hodName: "Prof. Neha Singh",
      studentCount: 350,
      facultyCount: 20,
      coursesCount: 35
    },
    {
      id: 5,
      name: "Business Administration",
      code: "BA",
      hodName: "Dr. Sanjay Mehta",
      studentCount: 520,
      facultyCount: 18,
      coursesCount: 30
    }
  ]);

  const filteredDepartments = departments.filter(
    dept => 
      dept.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      dept.code.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const totalStudents = departments.reduce((acc, dept) => acc + dept.studentCount, 0);
  const totalFaculty = departments.reduce((acc, dept) => acc + dept.facultyCount, 0);
  const totalCourses = departments.reduce((acc, dept) => acc + dept.coursesCount, 0);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Departments</h1>
          <p className="text-muted-foreground">
            Manage academic departments and their resources
          </p>
        </div>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Add Department
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Departments</CardDescription>
            <CardTitle className="text-2xl">{departments.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Students</CardDescription>
            <CardTitle className="text-2xl">{totalStudents.toLocaleString()}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Faculty</CardDescription>
            <CardTitle className="text-2xl">{totalFaculty}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Courses</CardDescription>
            <CardTitle className="text-2xl">{totalCourses}</CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search departments..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {filteredDepartments.map((dept) => (
          <Card key={dept.id} className="hover:shadow-md transition-shadow cursor-pointer">
            <CardHeader>
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-primary/10 p-3">
                    <Building2 className="h-6 w-6 text-primary" />
                  </div>
                  <div>
                    <CardTitle className="text-lg">{dept.name}</CardTitle>
                    <CardDescription className="text-sm">{dept.code}</CardDescription>
                  </div>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {dept.hodName && (
                <div className="text-sm">
                  <span className="text-muted-foreground">HOD: </span>
                  <span className="font-medium">{dept.hodName}</span>
                </div>
              )}
              <div className="grid grid-cols-3 gap-4 text-center">
                <div className="space-y-1">
                  <div className="flex items-center justify-center">
                    <Users className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <div className="text-2xl font-bold">{dept.studentCount}</div>
                  <div className="text-xs text-muted-foreground">Students</div>
                </div>
                <div className="space-y-1">
                  <div className="flex items-center justify-center">
                    <Users className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <div className="text-2xl font-bold">{dept.facultyCount}</div>
                  <div className="text-xs text-muted-foreground">Faculty</div>
                </div>
                <div className="space-y-1">
                  <div className="flex items-center justify-center">
                    <BookOpen className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <div className="text-2xl font-bold">{dept.coursesCount}</div>
                  <div className="text-xs text-muted-foreground">Courses</div>
                </div>
              </div>
              <Button variant="outline" className="w-full">
                View Details
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}