"use client";

import { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { api, endpoints } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Plus, Calendar, Clock, FileText, CheckCircle, Loader2, Trash2 } from "lucide-react";
import { format } from "date-fns";
import { logger } from "@/lib/logger";

type Assignment = {
  id: number;
  title: string;
  courseId?: number;
  courseName?: string;
  dueDate: string;
  maxScore: number;
  status?: "pending" | "submitted" | "graded";
  score?: number;
  description: string;
};

type ApiAssignment = {
  id: number;
  title?: string;
  description?: string;
  courseId?: number;
  course_id?: number;
  courseName?: string;
  course_name?: string;
  dueDate?: string;
  due_date?: string;
  maxScore?: number;
  max_points?: number;
  status?: string;
  score?: number;
};

type CourseOption = {
  id: number;
  name?: string;
  title?: string;
};

const normalizeStatus = (status?: string): Assignment["status"] => {
  if (status === "submitted" || status === "graded") {
    return status;
  }
  return "pending";
};

const normalizeAssignment = (
  item: ApiAssignment,
  fallbackCourseName?: string
): Assignment => {
  const courseId = item.courseId ?? item.course_id;
  const courseName = item.courseName ?? item.course_name ?? fallbackCourseName;

  return {
    id: item.id,
    title: item.title ?? "Untitled assignment",
    courseId,
    courseName: courseName ?? (courseId ? `Course ${courseId}` : undefined),
    dueDate: item.dueDate ?? item.due_date ?? new Date().toISOString(),
    maxScore: item.maxScore ?? item.max_points ?? 100,
    status: normalizeStatus(item.status),
    score: item.score,
    description: item.description ?? "",
  };
};

export default function AssignmentsPage() {
  const { user } = useAuth();
  const searchParams = useSearchParams();
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [courses, setCourses] = useState<CourseOption[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [managedAssignment, setManagedAssignment] = useState<Assignment | null>(null);
  const [focusedAssignmentId, setFocusedAssignmentId] = useState<number | null>(null);

  const [newAssignment, setNewAssignment] = useState({
    courseId: "",
    title: "",
    description: "",
    dueDate: "",
    maxScore: 100,
  });

  const isStudent = user?.role === "student";

  const loadCourses = useCallback(async () => {
    if (isStudent) {
      setCourses([]);
      return [] as CourseOption[];
    }

    try {
      const courseData = await api.get<CourseOption[]>(endpoints.courses.list);
      const normalizedCourses = Array.isArray(courseData) ? courseData : [];
      setCourses(normalizedCourses);
      return normalizedCourses;
    } catch (err) {
      logger.error("Failed to fetch courses for assignments:", err as Error);
      setCourses([]);
      return [];
    }
  }, [isStudent]);

  const loadAssignments = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      if (isStudent) {
        const response = await api.get<ApiAssignment[]>(endpoints.assignments.list);
        const normalized = (Array.isArray(response) ? response : []).map((item) => normalizeAssignment(item));
        setAssignments(normalized);
        return;
      }

      const courseList = await loadCourses();
      if (courseList.length === 0) {
        setAssignments([]);
        return;
      }

      const listByCourseResponses = await Promise.all(
        courseList.map(async (course) => {
          const items = await api.get<ApiAssignment[]>(endpoints.assignments.listByCourse(course.id));
          const courseName = course.name || course.title || `Course ${course.id}`;
          return (Array.isArray(items) ? items : []).map((item) => normalizeAssignment(item, courseName));
        })
      );

      const merged = listByCourseResponses
        .flat()
        .sort((a, b) => new Date(a.dueDate).getTime() - new Date(b.dueDate).getTime());
      setAssignments(merged);
    } catch (err) {
      logger.error("Failed to fetch assignments:", err as Error);
      setError("Failed to load assignments");
      setAssignments([]);
    } finally {
      setLoading(false);
    }
  }, [isStudent, loadCourses]);

  useEffect(() => {
    if (!user) {
      setLoading(false);
      return;
    }
    loadAssignments();
  }, [user, loadAssignments]);

  useEffect(() => {
    const focusParam = searchParams.get("focus");
    const focusId = Number.parseInt(focusParam || "", 10);
    if (Number.isFinite(focusId) && focusId > 0) {
      setFocusedAssignmentId(focusId);
      const timeoutId = window.setTimeout(() => setFocusedAssignmentId(null), 3000);
      return () => window.clearTimeout(timeoutId);
    }
    return undefined;
  }, [searchParams]);

  const getStatusBadge = (status: string) => {
    const styles = {
      pending: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400",
      submitted: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400",
      graded: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
    };
    return <Badge className={styles[status as keyof typeof styles]}>{status}</Badge>;
  };

  const isOverdue = (dueDate: string) => new Date(dueDate) < new Date();

  const handleCreate = async () => {
    try {
      setCreating(true);
      setError(null);

      const courseIdNum = Number.parseInt(newAssignment.courseId, 10);
      if (!courseIdNum || courseIdNum <= 0) {
        throw new Error("Course ID is required");
      }
      if (!newAssignment.title.trim()) {
        throw new Error("Title is required");
      }

      const dueDate = new Date(`${newAssignment.dueDate}T23:59:59`);
      if (Number.isNaN(dueDate.getTime())) {
        throw new Error("A valid due date is required");
      }

      await api.post(endpoints.assignments.create(courseIdNum), {
        title: newAssignment.title.trim(),
        description: newAssignment.description.trim(),
        due_date: dueDate.toISOString(),
        max_points: Number(newAssignment.maxScore),
      });

      setShowCreate(false);
      setNewAssignment({ courseId: "", title: "", description: "", dueDate: "", maxScore: 100 });
      await loadAssignments();
    } catch (err) {
      logger.error("Error occurred", err instanceof Error ? err : new Error(String(err)));
      setError(err instanceof Error ? err.message : "Failed to create assignment");
    } finally {
      setCreating(false);
    }
  };

  const handleSubmit = async (assignment: Assignment) => {
    try {
      setError(null);
      if (!assignment.courseId) {
        setError("Course ID not available for this assignment");
        return;
      }
      await api.post(endpoints.assignments.submit(assignment.courseId, assignment.id), { content_text: "Submitted via portal" });
      setAssignments((prev) => prev.map((item) => (item.id === assignment.id ? { ...item, status: "submitted" } : item)));
    } catch (err) {
      logger.error("Error occurred", err as Error);
      setError("Failed to submit assignment");
    }
  };

  const handleDeleteAssignment = async (assignment: Assignment) => {
    if (!assignment.courseId) {
      setError("Course ID not available; cannot delete assignment.");
      return;
    }

    try {
      setDeletingId(assignment.id);
      setError(null);
      await api.delete(endpoints.assignments.delete(assignment.courseId, assignment.id));
      setAssignments((prev) => prev.filter((item) => item.id !== assignment.id));
      setManagedAssignment(null);
    } catch (err) {
      logger.error("Failed to delete assignment:", err as Error);
      setError("Failed to delete assignment.");
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Assignments</h1>
            <p className="text-muted-foreground">
              {isStudent ? "View and submit your assignments" : "Manage course assignments"}
            </p>
          </div>
          {!isStudent && (
            <Button onClick={() => setShowCreate((value) => !value)}>
              <Plus className="mr-2 h-4 w-4" />
              {showCreate ? "Close" : "Create Assignment"}
            </Button>
          )}
        </div>

        {showCreate && (
          <Card>
            <CardHeader>
              <CardTitle>New Assignment</CardTitle>
              <CardDescription>Provide course and assignment details</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium">Course</label>
                  {courses.length > 0 ? (
                    <select
                      value={newAssignment.courseId}
                      onChange={(event) => setNewAssignment({ ...newAssignment, courseId: event.target.value })}
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    >
                      <option value="">Select course</option>
                      {courses.map((course) => (
                        <option key={course.id} value={String(course.id)}>
                          {course.name || course.title || `Course ${course.id}`}
                        </option>
                      ))}
                    </select>
                  ) : (
                    <Input
                      value={newAssignment.courseId}
                      onChange={(event) => setNewAssignment({ ...newAssignment, courseId: event.target.value })}
                      placeholder="Course ID"
                    />
                  )}
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Title</label>
                  <Input
                    value={newAssignment.title}
                    onChange={(event) => setNewAssignment({ ...newAssignment, title: event.target.value })}
                  />
                </div>
                <div className="space-y-2 sm:col-span-2">
                  <label className="text-sm font-medium">Description</label>
                  <Input
                    value={newAssignment.description}
                    onChange={(event) => setNewAssignment({ ...newAssignment, description: event.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Due Date</label>
                  <Input
                    type="date"
                    value={newAssignment.dueDate}
                    onChange={(event) => setNewAssignment({ ...newAssignment, dueDate: event.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Max Score</label>
                  <Input
                    type="number"
                    value={newAssignment.maxScore}
                    onChange={(event) => setNewAssignment({ ...newAssignment, maxScore: Number(event.target.value || 100) })}
                  />
                </div>
              </div>
              <div className="mt-4 flex justify-end">
                <Button onClick={handleCreate} disabled={creating}>
                  {creating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Plus className="mr-2 h-4 w-4" />}
                  Create
                </Button>
              </div>
            </CardContent>
          </Card>
        )}

        {error && <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">{error}</div>}

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Card>
            <CardHeader className="pb-3">
              <CardDescription>Total Assignments</CardDescription>
              <CardTitle className="text-2xl">{assignments.length}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription>Pending</CardDescription>
              <CardTitle className="text-2xl text-yellow-600">
                {assignments.filter((assignment) => assignment.status === "pending").length}
              </CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardDescription>Graded</CardDescription>
              <CardTitle className="text-2xl text-green-600">
                {assignments.filter((assignment) => assignment.status === "graded").length}
              </CardTitle>
            </CardHeader>
          </Card>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : (
          <div className="space-y-4">
            {assignments.map((assignment) => (
              <Card
                key={assignment.id}
                className={[
                  "hover:shadow-md transition-shadow",
                  focusedAssignmentId === assignment.id ? "ring-2 ring-primary/60" : "",
                ].join(" ")}
              >
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="space-y-1">
                      <CardTitle className="text-xl">{assignment.title}</CardTitle>
                      <CardDescription className="flex items-center gap-2">
                        <FileText className="h-4 w-4" />
                        {assignment.courseName}
                      </CardDescription>
                    </div>
                    {assignment.status && getStatusBadge(assignment.status)}
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground mb-4">{assignment.description}</p>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-6 text-sm">
                      <div className="flex items-center gap-2">
                        <Calendar className="h-4 w-4 text-muted-foreground" />
                        <span className={isOverdue(assignment.dueDate) && assignment.status === "pending" ? "text-destructive font-medium" : ""}>
                          Due: {format(new Date(assignment.dueDate), "MMM dd, yyyy")}
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4 text-muted-foreground" />
                        <span>Max Score: {assignment.maxScore}</span>
                      </div>
                      {assignment.score !== undefined && (
                        <div className="flex items-center gap-2">
                          <CheckCircle className="h-4 w-4 text-green-600" />
                          <span className="font-medium">Score: {assignment.score}/{assignment.maxScore}</span>
                        </div>
                      )}
                    </div>
                    <div className="space-x-2">
                      {isStudent ? (
                        <Button variant={assignment.status === "pending" ? "default" : "outline"} onClick={() => handleSubmit(assignment)}>
                          {assignment.status === "pending" ? "Submit" : "View"}
                        </Button>
                      ) : (
                        <Button variant="outline" onClick={() => setManagedAssignment(assignment)}>
                          Manage
                        </Button>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}

            {assignments.length === 0 && (
              <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
                No assignments found.
              </div>
            )}
          </div>
        )}
      </div>

      {managedAssignment && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-xl">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Manage Assignment</CardTitle>
                <Button variant="ghost" size="sm" onClick={() => setManagedAssignment(null)}>
                  Close
                </Button>
              </div>
              <CardDescription>{managedAssignment.title}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-3">
                <p className="text-sm"><span className="font-medium">Course:</span> {managedAssignment.courseName || "N/A"}</p>
                <p className="text-sm"><span className="font-medium">Due:</span> {format(new Date(managedAssignment.dueDate), "MMM dd, yyyy")}</p>
                <p className="text-sm"><span className="font-medium">Max Score:</span> {managedAssignment.maxScore}</p>
              </div>

              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setManagedAssignment(null)}>
                  Cancel
                </Button>
                <Button
                  variant="destructive"
                  onClick={() => handleDeleteAssignment(managedAssignment)}
                  disabled={deletingId === managedAssignment.id}
                >
                  {deletingId === managedAssignment.id ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Deleting...
                    </>
                  ) : (
                    <>
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete Assignment
                    </>
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </>
  );
}
