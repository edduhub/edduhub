"use client";

import { useState, useMemo } from "react";
import { useAuth } from "@/lib/auth-context";
import { useCourses } from "@/lib/api-hooks";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Plus, Search, Users, BookOpen, Clock, Calendar, Loader2 } from "lucide-react";

export default function CoursesPage() {
  const { user } = useAuth();
  const [searchQuery, setSearchQuery] = useState("");

  // React Query hook
  const { data: courses = [], isLoading: loading } = useCourses();

  // Filter courses using useMemo
  const filteredCourses = useMemo(() => {
    return courses.filter(course =>
      course.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      course.code.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (course.instructorName?.toLowerCase().includes(searchQuery.toLowerCase()) ?? false)
    );
  }, [courses, searchQuery]);

  // Calculate statistics using useMemo
  const stats = useMemo(() => {
    if (courses.length === 0) {
      return {
        total: 0,
        totalStudents: 0,
        avgProgress: 0,
        totalCredits: 0,
      };
    }

    return {
      total: courses.length,
      totalStudents: courses.reduce((acc, c) => acc + (c.enrollmentCount || 0), 0),
      totalCredits: courses.reduce((acc, c) => acc + c.credits, 0),
    };
  }, [courses]);

  const enrollmentPercentage = (enrolled: number, max: number) => {
    return Math.round((enrolled / max) * 100);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Courses</h1>
          <p className="text-muted-foreground">
            {user?.role === 'student' 
              ? 'Your enrolled courses and learning progress' 
              : user?.role === 'faculty'
              ? 'Manage your teaching courses'
              : 'All courses across departments'}
          </p>
        </div>
        {(user?.role === 'faculty' || user?.role === 'admin') && (
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Add Course
          </Button>
        )}
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Courses</CardDescription>
            <CardTitle className="text-2xl">{stats.total}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Students</CardDescription>
            <CardTitle className="text-2xl">{stats.totalStudents}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Avg Progress</CardDescription>
            <CardTitle className="text-2xl">{stats.avgProgress}%</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Credits</CardDescription>
            <CardTitle className="text-2xl">{stats.totalCredits}</CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search courses..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {filteredCourses.map((course) => (
            <Card key={course.id} className="hover:shadow-md transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">{course.code}</Badge>
                      <Badge>{course.credits} Credits</Badge>
                    </div>
                    <CardTitle className="text-xl">{course.name}</CardTitle>
                    <CardDescription>{course.instructorName || 'No instructor'}</CardDescription>
                  </div>
                  <BookOpen className="h-5 w-5 text-muted-foreground" />
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-sm text-muted-foreground">{course.description}</p>
                
                {user?.role === 'student' && (course as { progress?: number }).progress !== undefined && (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Course Progress</span>
                      <span className="font-medium">{(course as { progress?: number }).progress}%</span>
                    </div>
                    <Progress value={(course as { progress?: number }).progress} />
                  </div>
                )}

                <div className="space-y-2 text-sm">
                  <div className="flex items-center gap-2">
                    <Users className="h-4 w-4 text-muted-foreground" />
                    <span>{course.enrollmentCount || 0}/{course.maxEnrollment || 0} students</span>
                    <Badge 
                      variant="secondary"
                      className="ml-auto text-xs"
                    >
                      {enrollmentPercentage(course.enrollmentCount || 0, course.maxEnrollment || 1)}% full
                    </Badge>
                  </div>
                  {(course as { nextLecture?: string }).nextLecture && (
                    <div className="flex items-center gap-2">
                      <Clock className="h-4 w-4 text-muted-foreground" />
                      <span>Next: {(course as { nextLecture?: string }).nextLecture}</span>
                    </div>
                  )}
                  <div className="flex items-center gap-2">
                    <Calendar className="h-4 w-4 text-muted-foreground" />
                    <span>{course.semester}</span>
                  </div>
                </div>

                <Button className="w-full" variant="outline">
                  View Details
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
