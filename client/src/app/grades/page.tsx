"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Progress } from "@/components/ui/progress";
import { Award, TrendingUp, FileText, Loader2 } from "lucide-react";

type Grade = {
  id: number;
  courseName: string;
  courseCode: string;
  assessmentType: string;
  assessmentName: string;
  score: number;
  maxScore: number;
  percentage: number;
  date: string;
};

type CourseGrade = {
  courseName: string;
  courseCode: string;
  totalScore: number;
  maxScore: number;
  percentage: number;
  grade: string;
  credits: number;
};

export default function GradesPage() {
  const { user } = useAuth();
  const [grades, setGrades] = useState<Grade[]>([]);
  const [courseGrades, setCourseGrades] = useState<CourseGrade[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchGrades = async () => {
      try {
        setLoading(true);
        // Try to fetch individual grades
        try {
          const gradesResponse = await api.get('/api/grades');
          setGrades(Array.isArray(gradesResponse) ? gradesResponse : []);
        } catch (err) {
          console.warn('Failed to fetch individual grades:', err);
        }

        // Try to fetch course grades
        try {
          const courseGradesResponse = await api.get('/api/grades/courses');
          setCourseGrades(Array.isArray(courseGradesResponse) ? courseGradesResponse : []);
        } catch (err) {
          console.warn('Failed to fetch course grades:', err);
        }
      } catch (err) {
        console.error('Failed to fetch grades:', err);
        setError('Failed to load grades');
      } finally {
        setLoading(false);
      }
    };

    fetchGrades();
  }, []);

  const calculateGPA = () => {
    const gradePoints: Record<string, number> = {
      'A+': 4.0, 'A': 3.7, 'A-': 3.3,
      'B+': 3.0, 'B': 2.7, 'B-': 2.3,
      'C+': 2.0, 'C': 1.7, 'C-': 1.3,
      'D': 1.0, 'F': 0.0
    };
    
    const totalPoints = courseGrades.reduce(
      (acc, course) => acc + (gradePoints[course.grade] || 0) * course.credits,
      0
    );
    const totalCredits = courseGrades.reduce((acc, course) => acc + course.credits, 0);
    
    return (totalPoints / totalCredits).toFixed(2);
  };

  const getGradeColor = (percentage: number) => {
    if (percentage >= 90) return 'text-green-600';
    if (percentage >= 80) return 'text-blue-600';
    if (percentage >= 70) return 'text-yellow-600';
    return 'text-red-600';
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Grades</h1>
          <p className="text-muted-foreground">
            {user?.role === 'student' ? 'View your academic performance' : 'Manage student grades'}
          </p>
        </div>
        <Button>
          <FileText className="mr-2 h-4 w-4" />
          Generate Report
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <Award className="h-4 w-4" />
              Overall GPA
            </CardDescription>
            <CardTitle className="text-3xl">{calculateGPA()}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">Out of 4.0</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4" />
              Average Score
            </CardDescription>
            <CardTitle className="text-3xl">
              {Math.round(courseGrades.reduce((acc, c) => acc + c.percentage, 0) / courseGrades.length)}%
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">Across all courses</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Credits</CardDescription>
            <CardTitle className="text-3xl">
              {courseGrades.reduce((acc, c) => acc + c.credits, 0)}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">Enrolled credits</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Course Grades</CardTitle>
          <CardDescription>Overall performance in each course</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {courseGrades.map((course) => (
              <div key={course.courseCode} className="space-y-2">
                <div className="flex items-center justify-between">
                  <div>
                    <span className="font-medium">{course.courseName}</span>
                    <span className="ml-2 text-sm text-muted-foreground">{course.courseCode}</span>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className={`text-lg font-bold ${getGradeColor(course.percentage)}`}>
                      {course.grade}
                    </span>
                    <span className="text-sm text-muted-foreground">
                      {course.percentage}%
                    </span>
                  </div>
                </div>
                <Progress value={course.percentage} />
                <p className="text-xs text-muted-foreground">
                  {course.totalScore}/{course.maxScore} • {course.credits} credits
                </p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Recent Grades</CardTitle>
          <CardDescription>Individual assessment scores</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Course</TableHead>
                <TableHead>Assessment</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Score</TableHead>
                <TableHead>Percentage</TableHead>
                <TableHead>Date</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {grades.map((grade) => (
                <TableRow key={grade.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{grade.courseName}</div>
                      <div className="text-sm text-muted-foreground">{grade.courseCode}</div>
                    </div>
                  </TableCell>
                  <TableCell>{grade.assessmentName}</TableCell>
                  <TableCell>
                    <span className="rounded-full bg-primary/10 px-2 py-1 text-xs font-medium">
                      {grade.assessmentType}
                    </span>
                  </TableCell>
                  <TableCell className="font-medium">
                    {grade.score}/{grade.maxScore}
                  </TableCell>
                  <TableCell>
                    <span className={`font-medium ${getGradeColor(grade.percentage)}`}>
                      {grade.percentage}%
                    </span>
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
