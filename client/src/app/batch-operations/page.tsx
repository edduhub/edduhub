"use client";

import { useState, useRef } from "react";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Download, Upload, Loader2, AlertCircle, CheckCircle, FileText, Users, GraduationCap, BookOpen } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { logger } from '@/lib/logger';

type ImportResult = {
  success: number;
  failed: number;
  errors?: string[];
  message?: string;
};

type Course = {
  id: number;
  name: string;
  code: string;
};

export default function BatchOperationsPage() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [importResult, setImportResult] = useState<ImportResult | null>(null);

  // File upload refs
  const studentFileRef = useRef<HTMLInputElement>(null);
  const gradeFileRef = useRef<HTMLInputElement>(null);
  const enrollmentFileRef = useRef<HTMLInputElement>(null);

  // Grade import state
  const [selectedCourse, setSelectedCourse] = useState<string>("");
  const [courses, setCourses] = useState<Course[]>([]);
  const [loadingCourses, setLoadingCourses] = useState(false);

  // Enrollment import state
  const [enrollmentCourse, setEnrollmentCourse] = useState<string>("");

  const downloadCsvTemplate = (filename: string, header: string, sampleRow: string) => {
    const csv = `${header}\n${sampleRow}\n`;
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  };

  const handleStudentTemplateDownload = () => {
    downloadCsvTemplate(
      'students_import_template.csv',
      'roll_no,first_name,last_name,email',
      'CS001,John,Doe,john.doe@example.edu'
    );
    setSuccess('Student CSV template downloaded');
  };

  const handleGradeTemplateDownload = () => {
    const courseHint = selectedCourse ? `for course ${selectedCourse}` : '';
    downloadCsvTemplate(
      'grades_import_template.csv',
      'roll_no,grade,marks',
      'CS001,A,92'
    );
    setSuccess(`Grade CSV template downloaded ${courseHint}`.trim());
  };

  const handleEnrollmentTemplateDownload = () => {
    downloadCsvTemplate(
      'enrollment_import_template.csv',
      'roll_no',
      'CS001'
    );
    setSuccess('Enrollment CSV template downloaded');
  };

  // Load courses for grade import
  const loadCourses = async () => {
    try {
      setLoadingCourses(true);
      const response = await api.get('/api/courses');
      setCourses(Array.isArray(response) ? response : []);
    } catch (err) {
      logger.error('Failed to fetch courses:', err as Error);
    } finally {
      setLoadingCourses(false);
    }
  };

  // Import students from CSV
  const handleStudentImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      setLoading(true);
      setError(null);
      setSuccess(null);
      setImportResult(null);

      const formData = new FormData();
      formData.append('file', file);

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/batch/students/import`, {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Import failed');
      }

      setImportResult(result.data);
      setSuccess(`Successfully imported ${result.data.success} students`);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to import students');
    } finally {
      setLoading(false);
      if (studentFileRef.current) {
        studentFileRef.current.value = '';
      }
    }
  };

  // Export students to CSV
  const handleStudentExport = async () => {
    try {
      setLoading(true);
      setError(null);
      setSuccess(null);

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/batch/students/export`, {
        method: 'GET',
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Export failed');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `students_export_${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);

      setSuccess('Students exported successfully');
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to export students');
    } finally {
      setLoading(false);
    }
  };

  // Import grades from CSV
  const handleGradeImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !selectedCourse) {
      setError('Please select a course first');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      setSuccess(null);
      setImportResult(null);

      const formData = new FormData();
      formData.append('file', file);
      formData.append('course_id', selectedCourse);

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/batch/grades/import`, {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Import failed');
      }

      setImportResult(result.data);
      setSuccess(`Successfully imported ${result.data.success} grades`);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to import grades');
    } finally {
      setLoading(false);
      if (gradeFileRef.current) {
        gradeFileRef.current.value = '';
      }
    }
  };

  // Export grades to CSV
  const handleGradeExport = async () => {
    if (!selectedCourse) {
      setError('Please select a course first');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      setSuccess(null);

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/batch/grades/export?course_id=${selectedCourse}`, {
        method: 'GET',
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Export failed');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `grades_course_${selectedCourse}_${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);

      setSuccess('Grades exported successfully');
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to export grades');
    } finally {
      setLoading(false);
    }
  };

  // Import enrollments from CSV
  const handleEnrollmentImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !enrollmentCourse) {
      setError('Please select a course first');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      setSuccess(null);
      setImportResult(null);

      const formData = new FormData();
      formData.append('file', file);
      formData.append('course_id', enrollmentCourse);

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/batch/enroll`, {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Enrollment import failed');
      }

      setImportResult(result.data);
      setSuccess(`Successfully enrolled ${result.data.success} students`);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to enroll students');
    } finally {
      setLoading(false);
      if (enrollmentFileRef.current) {
        enrollmentFileRef.current.value = '';
      }
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Batch Operations</h1>
        <p className="text-muted-foreground">
          Import and export students and grades in bulk
        </p>
      </div>

      {error && (
        <div className="flex items-center gap-2 rounded-lg bg-destructive/10 p-4 text-sm text-destructive">
          <AlertCircle className="h-4 w-4" />
          {error}
        </div>
      )}

      {success && (
        <div className="flex items-center gap-2 rounded-lg bg-green-50 dark:bg-green-900/20 p-4 text-sm text-green-800 dark:text-green-400">
          <CheckCircle className="h-4 w-4" />
          {success}
        </div>
      )}

      {importResult && (
        <Card>
          <CardHeader>
            <CardTitle>Import Results</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Successfully imported:</span>
                <span className="font-semibold text-green-600">{importResult.success}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Failed:</span>
                <span className="font-semibold text-red-600">{importResult.failed}</span>
              </div>
              {importResult.errors && importResult.errors.length > 0 && (
                <div className="mt-4">
                  <p className="text-sm font-medium mb-2">Errors:</p>
                  <ul className="list-disc list-inside text-sm text-muted-foreground space-y-1">
                    {importResult.errors.map((err, idx) => (
                      <li key={idx}>{err}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      <Tabs defaultValue="students" className="space-y-4">
        <TabsList>
          <TabsTrigger value="students" className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            Students
          </TabsTrigger>
          <TabsTrigger value="grades" className="flex items-center gap-2" onClick={() => loadCourses()}>
            <GraduationCap className="h-4 w-4" />
            Grades
          </TabsTrigger>
          <TabsTrigger value="enrollment" className="flex items-center gap-2" onClick={() => loadCourses()}>
            <BookOpen className="h-4 w-4" />
            Enrollment
          </TabsTrigger>
        </TabsList>

        <TabsContent value="students" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Upload className="h-5 w-5" />
                  Import Students
                </CardTitle>
                <CardDescription>
                  Upload a CSV file to import students in bulk
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="student-file">CSV File</Label>
                  <Input
                    id="student-file"
                    ref={studentFileRef}
                    type="file"
                    accept=".csv"
                    onChange={handleStudentImport}
                    disabled={loading}
                  />
                </div>
                <div className="rounded-lg bg-muted p-3 text-sm">
                  <p className="font-medium mb-2">CSV Format:</p>
                  <code className="block text-xs">roll_no,first_name,last_name,email</code>
                  <p className="text-muted-foreground mt-2">
                    The first row should contain column headers
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="w-full"
                  onClick={handleStudentTemplateDownload}
                >
                  <FileText className="mr-2 h-4 w-4" />
                  Download Template
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Download className="h-5 w-5" />
                  Export Students
                </CardTitle>
                <CardDescription>
                  Download all students as a CSV file
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-sm text-muted-foreground">
                  Export all student data including roll numbers, names, emails, and enrollment information.
                </p>
                <Button
                  onClick={handleStudentExport}
                  disabled={loading}
                  className="w-full"
                >
                  {loading ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Download className="mr-2 h-4 w-4" />
                  )}
                  Export Students
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="grades" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Select Course</CardTitle>
              <CardDescription>
                Choose a course to import or export grades
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <Label htmlFor="course-select">Course</Label>
                <Select value={selectedCourse} onValueChange={setSelectedCourse}>
                  <SelectTrigger id="course-select">
                    <SelectValue placeholder="Select a course" />
                  </SelectTrigger>
                  <SelectContent>
                    {loadingCourses ? (
                      <div className="p-2 text-center text-sm text-muted-foreground">
                        Loading courses...
                      </div>
                    ) : courses.length === 0 ? (
                      <div className="p-2 text-center text-sm text-muted-foreground">
                        No courses available
                      </div>
                    ) : (
                      courses.map((course) => (
                        <SelectItem key={course.id} value={course.id.toString()}>
                          {course.code} - {course.name}
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Upload className="h-5 w-5" />
                  Import Grades
                </CardTitle>
                <CardDescription>
                  Upload a CSV file to import grades in bulk
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="grade-file">CSV File</Label>
                  <Input
                    id="grade-file"
                    ref={gradeFileRef}
                    type="file"
                    accept=".csv"
                    onChange={handleGradeImport}
                    disabled={loading || !selectedCourse}
                  />
                </div>
                <div className="rounded-lg bg-muted p-3 text-sm">
                  <p className="font-medium mb-2">CSV Format:</p>
                  <code className="block text-xs">roll_no,grade,marks</code>
                  <p className="text-muted-foreground mt-2">
                    The first row should contain column headers
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="w-full"
                  disabled={!selectedCourse}
                  onClick={handleGradeTemplateDownload}
                >
                  <FileText className="mr-2 h-4 w-4" />
                  Download Template
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Download className="h-5 w-5" />
                  Export Grades
                </CardTitle>
                <CardDescription>
                  Download grades for the selected course
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-sm text-muted-foreground">
                  Export all grade data for the selected course including student information and scores.
                </p>
                <Button
                  onClick={handleGradeExport}
                  disabled={loading || !selectedCourse}
                  className="w-full"
                >
                  {loading ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Download className="mr-2 h-4 w-4" />
                  )}
                  Export Grades
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="enrollment" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Select Course</CardTitle>
              <CardDescription>
                Choose a course to enroll students in
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <Label htmlFor="enrollment-course-select">Course</Label>
                <Select value={enrollmentCourse} onValueChange={setEnrollmentCourse}>
                  <SelectTrigger id="enrollment-course-select">
                    <SelectValue placeholder="Select a course" />
                  </SelectTrigger>
                  <SelectContent>
                    {loadingCourses ? (
                      <div className="p-2 text-center text-sm text-muted-foreground">
                        Loading courses...
                      </div>
                    ) : courses.length === 0 ? (
                      <div className="p-2 text-center text-sm text-muted-foreground">
                        No courses available
                      </div>
                    ) : (
                      courses.map((course) => (
                        <SelectItem key={course.id} value={course.id.toString()}>
                          {course.code} - {course.name}
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Upload className="h-5 w-5" />
                Enroll Students
              </CardTitle>
              <CardDescription>
                Upload a CSV file with student roll numbers to enroll them in the selected course
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="enrollment-file">CSV File</Label>
                <Input
                  id="enrollment-file"
                  ref={enrollmentFileRef}
                  type="file"
                  accept=".csv"
                  onChange={handleEnrollmentImport}
                  disabled={loading || !enrollmentCourse}
                />
              </div>
              <div className="rounded-lg bg-muted p-3 text-sm">
                <p className="font-medium mb-2">CSV Format:</p>
                <code className="block text-xs">roll_no</code>
                <p className="text-muted-foreground mt-2">
                  The first row should contain column headers. Only roll_no column is required.
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                disabled={!enrollmentCourse}
                onClick={handleEnrollmentTemplateDownload}
              >
                <FileText className="mr-2 h-4 w-4" />
                Download Template
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
