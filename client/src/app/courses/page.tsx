"use client";

import { useState, useMemo } from "react";
import { useAuth } from "@/lib/auth-context";
import { useCourses, useCreateCourse } from "@/lib/api-hooks";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Plus, Search, Users, BookOpen, Clock, Calendar, Loader2 } from "lucide-react";
import { logger } from "@/lib/logger";

type CourseCard = {
  id: number;
  code: string;
  name: string;
  description: string;
  credits: number;
  semester: string;
  instructorName: string;
  enrollmentCount: number;
  maxEnrollment: number;
  progress?: number;
  nextLecture?: string;
};

export default function CoursesPage() {
  const { user } = useAuth();
  const [searchQuery, setSearchQuery] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [selectedCourse, setSelectedCourse] = useState<CourseCard | null>(null);
  const [newCourse, setNewCourse] = useState({
    name: "",
    description: "",
    credits: "3",
    instructorId: user?.role === "faculty" ? String(Number.parseInt(user.id, 10) || "") : "",
  });

  const { data: courses = [], isLoading: loading } = useCourses();
  const createCourse = useCreateCourse();

  const normalizedCourses = useMemo<CourseCard[]>(() => {
    return courses
      .map<CourseCard | null>((course) => {
        const raw = course as typeof course & {
          instructor?: string;
          enrolledStudents?: number;
          maxStudents?: number;
          progress?: number;
          nextLecture?: string;
          semester?: string;
          code?: string;
        };

        const id = raw.id;
        if (!id) return null;

        return {
          id,
          code: raw.code || `COURSE-${id}`,
          name: raw.name,
          description: raw.description || "No description available.",
          credits: raw.credits,
          semester: raw.semester || "Current",
          instructorName: raw.instructorName || raw.instructor || "Unknown",
          enrollmentCount: raw.enrollmentCount ?? raw.enrolledStudents ?? 0,
          maxEnrollment: raw.maxEnrollment ?? raw.maxStudents ?? 100,
          progress: raw.progress,
          nextLecture: raw.nextLecture,
        } as CourseCard;
      })
      .filter((course): course is CourseCard => course !== null);
  }, [courses]);

  const filteredCourses = useMemo(() => {
    const query = searchQuery.toLowerCase();
    return normalizedCourses.filter((course) =>
      course.name.toLowerCase().includes(query) ||
      course.code.toLowerCase().includes(query) ||
      course.instructorName.toLowerCase().includes(query)
    );
  }, [normalizedCourses, searchQuery]);

  const stats = useMemo(() => {
    if (normalizedCourses.length === 0) {
      return {
        total: 0,
        totalStudents: 0,
        avgProgress: 0,
        totalCredits: 0,
      };
    }

    const progressValues = normalizedCourses
      .map((course) => course.progress)
      .filter((value): value is number => typeof value === "number");

    return {
      total: normalizedCourses.length,
      totalStudents: normalizedCourses.reduce((acc, course) => acc + course.enrollmentCount, 0),
      avgProgress:
        progressValues.length > 0
          ? Math.round(progressValues.reduce((acc, value) => acc + value, 0) / progressValues.length)
          : 0,
      totalCredits: normalizedCourses.reduce((acc, course) => acc + course.credits, 0),
    };
  }, [normalizedCourses]);

  const enrollmentPercentage = (enrolled: number, max: number) => {
    if (max <= 0) return 0;
    return Math.min(100, Math.round((enrolled / max) * 100));
  };

  const handleCreateCourse = async () => {
    setCreateError(null);

    const name = newCourse.name.trim();
    const description = newCourse.description.trim();
    const credits = Number.parseInt(newCourse.credits, 10);
    const instructorId = Number.parseInt(newCourse.instructorId, 10);

    if (!name) {
      setCreateError("Course name is required.");
      return;
    }
    if (!Number.isFinite(credits) || credits < 1 || credits > 5) {
      setCreateError("Credits must be between 1 and 5.");
      return;
    }
    if (!Number.isFinite(instructorId) || instructorId <= 0) {
      setCreateError("A valid instructor ID is required.");
      return;
    }

    try {
      await createCourse.mutateAsync({
        name,
        description,
        credits,
        instructor_id: instructorId,
      } as unknown as Parameters<typeof createCourse.mutateAsync>[0]);

      setIsCreateOpen(false);
      setNewCourse({
        name: "",
        description: "",
        credits: "3",
        instructorId: user?.role === "faculty" ? String(Number.parseInt(user.id, 10) || "") : "",
      });
    } catch (error) {
      logger.error("Failed to create course:", error as Error);
      setCreateError("Failed to create course. Please verify the inputs and try again.");
    }
  };

  return (
    <>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">Courses</h1>
            <p className="text-muted-foreground">
              {user?.role === "student"
                ? "Your enrolled courses and learning progress"
                : user?.role === "faculty"
                  ? "Manage your teaching courses"
                  : "All courses across departments"}
            </p>
          </div>
          {(user?.role === "faculty" || user?.role === "admin") && (
            <Button onClick={() => {
              setCreateError(null);
              setIsCreateOpen(true);
            }}>
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
                      <CardDescription>{course.instructorName}</CardDescription>
                    </div>
                    <BookOpen className="h-5 w-5 text-muted-foreground" />
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <p className="text-sm text-muted-foreground">{course.description}</p>

                  {user?.role === "student" && typeof course.progress === "number" && (
                    <div className="space-y-2">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">Course Progress</span>
                        <span className="font-medium">{course.progress}%</span>
                      </div>
                      <Progress value={course.progress} />
                    </div>
                  )}

                  <div className="space-y-2 text-sm">
                    <div className="flex items-center gap-2">
                      <Users className="h-4 w-4 text-muted-foreground" />
                      <span>{course.enrollmentCount}/{course.maxEnrollment} students</span>
                      <Badge variant="secondary" className="ml-auto text-xs">
                        {enrollmentPercentage(course.enrollmentCount, course.maxEnrollment)}% full
                      </Badge>
                    </div>
                    {course.nextLecture && (
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4 text-muted-foreground" />
                        <span>Next: {course.nextLecture}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2">
                      <Calendar className="h-4 w-4 text-muted-foreground" />
                      <span>{course.semester}</span>
                    </div>
                  </div>

                  <Button className="w-full" variant="outline" onClick={() => setSelectedCourse(course)}>
                    View Details
                  </Button>
                </CardContent>
              </Card>
            ))}

            {filteredCourses.length === 0 && (
              <div className="md:col-span-2 rounded-lg border p-8 text-center text-sm text-muted-foreground">
                No courses found for the current filters.
              </div>
            )}
          </div>
        )}
      </div>

      {isCreateOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-xl">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Create Course</CardTitle>
                <Button variant="ghost" size="sm" onClick={() => setIsCreateOpen(false)}>
                  Close
                </Button>
              </div>
              <CardDescription>Add a new course to your college.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="courseName">Course Name</label>
                <Input
                  id="courseName"
                  value={newCourse.name}
                  onChange={(e) => setNewCourse((prev) => ({ ...prev, name: e.target.value }))}
                  placeholder="e.g., Operating Systems"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="courseDescription">Description</label>
                <Input
                  id="courseDescription"
                  value={newCourse.description}
                  onChange={(e) => setNewCourse((prev) => ({ ...prev, description: e.target.value }))}
                  placeholder="Optional course description"
                />
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="courseCredits">Credits</label>
                  <Input
                    id="courseCredits"
                    type="number"
                    min={1}
                    max={5}
                    value={newCourse.credits}
                    onChange={(e) => setNewCourse((prev) => ({ ...prev, credits: e.target.value }))}
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="instructorId">Instructor ID</label>
                  <Input
                    id="instructorId"
                    type="number"
                    min={1}
                    value={newCourse.instructorId}
                    onChange={(e) => setNewCourse((prev) => ({ ...prev, instructorId: e.target.value }))}
                    placeholder="Numeric user ID"
                  />
                </div>
              </div>

              {createError && <p className="text-sm text-destructive">{createError}</p>}

              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setIsCreateOpen(false)} disabled={createCourse.isPending}>
                  Cancel
                </Button>
                <Button onClick={handleCreateCourse} disabled={createCourse.isPending}>
                  {createCourse.isPending ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    <>
                      <Plus className="mr-2 h-4 w-4" />
                      Create Course
                    </>
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {selectedCourse && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-2xl">
            <CardHeader>
              <div className="flex items-center justify-between gap-3">
                <div>
                  <CardTitle>{selectedCourse.name}</CardTitle>
                  <CardDescription>{selectedCourse.code}</CardDescription>
                </div>
                <Button variant="ghost" size="sm" onClick={() => setSelectedCourse(null)}>
                  Close
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Instructor</p>
                  <p className="font-medium">{selectedCourse.instructorName}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Semester</p>
                  <p className="font-medium">{selectedCourse.semester}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Credits</p>
                  <p className="font-medium">{selectedCourse.credits}</p>
                </div>
                <div>
                  <p className="text-xs uppercase text-muted-foreground">Enrollment</p>
                  <p className="font-medium">
                    {selectedCourse.enrollmentCount}/{selectedCourse.maxEnrollment}
                  </p>
                </div>
              </div>

              <div>
                <p className="text-xs uppercase text-muted-foreground">Description</p>
                <p className="text-sm mt-1">{selectedCourse.description}</p>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </>
  );
}
