"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Plus, Clock, BookOpen, Trophy, Play, Loader2, Trash2, Pencil, Check, HelpCircle } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { api, endpoints } from "@/lib/api-client";
import { logger } from "@/lib/logger";

type DisplayQuiz = {
  id: number;
  title: string;
  courseId?: number;
  courseName?: string;
  duration?: number;
  totalMarks?: number;
  questionsCount?: number;
  status: "not_started" | "in_progress" | "completed";
  score?: number;
  attempts: number;
  maxAttempts: number;
  dueDate?: string;
};

type ApiQuiz = {
  id: number;
  title?: string;
  description?: string;
  courseId?: number;
  course_id?: number;
  courseName?: string;
  course_name?: string;
  duration?: number;
  time_limit_minutes?: number;
  dueDate?: string;
  due_date?: string;
  status?: string;
  attempts?: number;
  maxAttempts?: number;
  allowedAttempts?: number;
  allowed_attempts?: number;
  score?: number;
  questions?: Array<{ id: number; marks?: number; points?: number }>;
};

type CourseOption = {
  id: number;
  name?: string;
  title?: string;
};

type QuestionType = "multiple_choice" | "true_false" | "short_answer";

type QuestionOption = {
  id?: number;
  text: string;
  isCorrect: boolean;
};

type DisplayQuestion = {
  id: number;
  text: string;
  type: QuestionType;
  options?: QuestionOption[];
  correctAnswer?: string;
  points: number;
};

type ApiQuestion = {
  id: number;
  question_text?: string;
  text?: string;
  question_type?: string;
  type?: string;
  options?: Array<{
    id?: number;
    option_text?: string;
    text?: string;
    is_correct?: boolean;
    isCorrect?: boolean;
  }>;
  correct_answer?: string;
  correctAnswer?: string;
  points?: number;
  marks?: number;
};

type StartAttemptResponse = {
  id?: number;
};

const normalizeStatus = (status?: string): DisplayQuiz["status"] => {
  if (status === "completed" || status === "Completed") {
    return "completed";
  }
  if (status === "in_progress" || status === "InProgress") {
    return "in_progress";
  }
  return "not_started";
};

const normalizeQuiz = (item: ApiQuiz, fallbackCourseName?: string): DisplayQuiz => {
  const courseId = item.courseId ?? item.course_id;
  const courseName = item.courseName ?? item.course_name ?? fallbackCourseName;

  return {
    id: item.id,
    title: item.title ?? "Untitled Quiz",
    courseId,
    courseName: courseName ?? (courseId ? `Course ${courseId}` : undefined),
    duration: item.duration ?? item.time_limit_minutes,
    totalMarks:
      item.questions?.reduce((sum, question) => sum + (question?.marks ?? question?.points ?? 0), 0) ?? undefined,
    questionsCount: item.questions?.length,
    status: normalizeStatus(item.status),
    attempts: item.attempts ?? 0,
    maxAttempts: item.maxAttempts ?? item.allowedAttempts ?? item.allowed_attempts ?? 1,
    score: item.score,
    dueDate: item.dueDate ?? item.due_date,
  };
};

const normalizeQuestion = (item: ApiQuestion): DisplayQuestion => {
  const text = item.question_text ?? item.text ?? "";
  const type = (item.question_type ?? item.type ?? "multiple_choice") as QuestionType;
  
  let options: QuestionOption[] | undefined;
  if (item.options && item.options.length > 0) {
    options = item.options.map((opt) => ({
      id: opt.id,
      text: opt.option_text ?? opt.text ?? "",
      isCorrect: opt.is_correct ?? opt.isCorrect ?? false,
    }));
  }

  const correctAnswer = item.correct_answer ?? item.correctAnswer;

  return {
    id: item.id,
    text,
    type,
    options,
    correctAnswer,
    points: item.points ?? item.marks ?? 1,
  };
};

export default function QuizzesPage() {
  const router = useRouter();
  const { user } = useAuth();

  const [quizzes, setQuizzes] = useState<DisplayQuiz[]>([]);
  const [courses, setCourses] = useState<CourseOption[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [startingQuizId, setStartingQuizId] = useState<number | null>(null);

  const [managedQuiz, setManagedQuiz] = useState<DisplayQuiz | null>(null);
  const [deletingQuizId, setDeletingQuizId] = useState<number | null>(null);

  const [showQuestions, setShowQuestions] = useState(false);
  const [questions, setQuestions] = useState<DisplayQuestion[]>([]);
  const [questionsLoading, setQuestionsLoading] = useState(false);
  const [questionsError, setQuestionsError] = useState<string | null>(null);

  const [showQuestionDialog, setShowQuestionDialog] = useState(false);
  const [editingQuestion, setEditingQuestion] = useState<DisplayQuestion | null>(null);
  const [savingQuestion, setSavingQuestion] = useState(false);
  const [deletingQuestionId, setDeletingQuestionId] = useState<number | null>(null);

  const [newQuestion, setNewQuestion] = useState({
    text: "",
    type: "multiple_choice" as QuestionType,
    options: [
      { text: "", isCorrect: false },
      { text: "", isCorrect: false },
      { text: "", isCorrect: false },
      { text: "", isCorrect: false },
    ],
    correctAnswer: "",
    points: 1,
  });

  const [newQuiz, setNewQuiz] = useState({
    courseId: "",
    title: "",
    description: "",
    duration: 30,
    dueDate: "",
  });

  const isStudent = user?.role === "student";

  const loadCourses = useCallback(async () => {
    if (isStudent) {
      setCourses([]);
      return [] as CourseOption[];
    }

    try {
      const data = await api.get<CourseOption[]>(endpoints.courses.list);
      const normalized = Array.isArray(data) ? data : [];
      setCourses(normalized);
      return normalized;
    } catch (err) {
      logger.error("Failed to load courses for quizzes", err as Error);
      setCourses([]);
      return [];
    }
  }, [isStudent]);

  const loadQuizzes = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      if (isStudent) {
        const data = await api.get<ApiQuiz[]>(endpoints.quizzes.myQuizzes);
        const display = (Array.isArray(data) ? data : []).map((item) => normalizeQuiz(item));
        setQuizzes(display);
        return;
      }

      const courseList = await loadCourses();
      if (courseList.length === 0) {
        setQuizzes([]);
        return;
      }

      const byCourse = await Promise.all(
        courseList.map(async (course) => {
          const response = await api.get<ApiQuiz[]>(endpoints.quizzes.listByCourse(course.id));
          const courseName = course.name || course.title || `Course ${course.id}`;
          return (Array.isArray(response) ? response : []).map((item) => normalizeQuiz(item, courseName));
        })
      );

      const merged = byCourse
        .flat()
        .sort((a, b) => a.title.localeCompare(b.title));
      setQuizzes(merged);
    } catch (err) {
      logger.error("Failed to load quizzes", err as Error);
      setError("Failed to load quizzes");
      setQuizzes([]);
    } finally {
      setLoading(false);
    }
  }, [isStudent, loadCourses]);

  useEffect(() => {
    if (!user) {
      setLoading(false);
      return;
    }
    void loadQuizzes();
  }, [user, loadQuizzes]);

  const getStatusBadge = (status: DisplayQuiz["status"]) => {
    const styles = {
      not_started: "bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400",
      in_progress: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400",
      completed: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
    };
    const labels = {
      not_started: "Available",
      in_progress: "In Progress",
      completed: "Completed",
    };
    return <Badge className={styles[status]}>{labels[status]}</Badge>;
  };

  const startQuiz = async (quizId: number) => {
    try {
      setStartingQuizId(quizId);
      const attempt = await api.post<StartAttemptResponse>(endpoints.quizAttempts.start(quizId), {});
      setQuizzes((prev) =>
        prev.map((quiz) =>
          quiz.id === quizId ? { ...quiz, status: "in_progress", attempts: quiz.attempts + 1 } : quiz
        )
      );

      if (attempt?.id) {
        router.push(`/quizzes/${quizId}/attempt/${attempt.id}`);
      }
    } catch (err) {
      logger.error("Failed to start quiz", err as Error);
      setError("Failed to start quiz");
    } finally {
      setStartingQuizId(null);
    }
  };

  const createQuiz = async () => {
    try {
      setCreating(true);
      setError(null);

      const courseIdNum = Number.parseInt(newQuiz.courseId, 10);
      if (!courseIdNum || courseIdNum <= 0) {
        throw new Error("Course ID is required");
      }
      if (!newQuiz.title.trim()) {
        throw new Error("Title is required");
      }

      const payload: {
        title: string;
        description: string;
        time_limit_minutes: number;
        due_date?: string;
      } = {
        title: newQuiz.title.trim(),
        description: newQuiz.description.trim(),
        time_limit_minutes: Number(newQuiz.duration),
      };

      if (newQuiz.dueDate) {
        const dueDate = new Date(`${newQuiz.dueDate}T23:59:59`);
        if (!Number.isNaN(dueDate.getTime())) {
          payload.due_date = dueDate.toISOString();
        }
      }

      await api.post(endpoints.quizzes.create(courseIdNum), payload);

      setShowCreate(false);
      setNewQuiz({
        courseId: "",
        title: "",
        description: "",
        duration: 30,
        dueDate: "",
      });
      await loadQuizzes();
    } catch (err) {
      logger.error("Failed to create quiz", err instanceof Error ? err : new Error(String(err)));
      setError(err instanceof Error ? err.message : "Failed to create quiz");
    } finally {
      setCreating(false);
    }
  };

  const handleDeleteQuiz = async (quiz: DisplayQuiz) => {
    try {
      if (!quiz.courseId) {
        throw new Error("Course ID is missing for this quiz");
      }

      setDeletingQuizId(quiz.id);
      setError(null);
      await api.delete(endpoints.quizzes.delete(quiz.courseId, quiz.id));
      await loadQuizzes();
      setManagedQuiz(null);
    } catch (err) {
      logger.error("Failed to delete quiz", err as Error);
      setError(err instanceof Error ? err.message : "Failed to delete quiz");
    } finally {
      setDeletingQuizId(null);
    }
  };

  const loadQuestions = useCallback(async (quizId: number) => {
    try {
      setQuestionsLoading(true);
      setQuestionsError(null);
      const data = await api.get<ApiQuestion[]>(endpoints.quizzes.questions(quizId));
      const normalized = (Array.isArray(data) ? data : []).map(normalizeQuestion);
      setQuestions(normalized);
    } catch (err) {
      logger.error("Failed to load questions", err as Error);
      setQuestionsError("Failed to load questions");
      setQuestions([]);
    } finally {
      setQuestionsLoading(false);
    }
  }, []);

  const openQuestions = (quiz: DisplayQuiz) => {
    setManagedQuiz(quiz);
    setShowQuestions(true);
    void loadQuestions(quiz.id);
  };

  const resetQuestionForm = () => {
    setNewQuestion({
      text: "",
      type: "multiple_choice",
      options: [
        { text: "", isCorrect: false },
        { text: "", isCorrect: false },
        { text: "", isCorrect: false },
        { text: "", isCorrect: false },
      ],
      correctAnswer: "",
      points: 1,
    });
    setEditingQuestion(null);
  };

  const openCreateQuestion = () => {
    resetQuestionForm();
    setShowQuestionDialog(true);
  };

  const openEditQuestion = (question: DisplayQuestion) => {
    setEditingQuestion(question);
    if (question.type === "multiple_choice" && question.options) {
      const options = question.options.map((opt) => ({
        text: opt.text,
        isCorrect: opt.isCorrect,
      }));
      while (options.length < 4) {
        options.push({ text: "", isCorrect: false });
      }
      setNewQuestion({
        text: question.text,
        type: question.type,
        options: options.slice(0, 4),
        correctAnswer: "",
        points: question.points,
      });
    } else {
      setNewQuestion({
        text: question.text,
        type: question.type,
        options: [
          { text: "", isCorrect: false },
          { text: "", isCorrect: false },
          { text: "", isCorrect: false },
          { text: "", isCorrect: false },
        ],
        correctAnswer: question.correctAnswer || "",
        points: question.points,
      });
    }
    setShowQuestionDialog(true);
  };

  const saveQuestion = async () => {
    if (!managedQuiz) return;

    try {
      setSavingQuestion(true);
      setQuestionsError(null);

      if (!newQuestion.text.trim()) {
        throw new Error("Question text is required");
      }

      if (newQuestion.type === "multiple_choice") {
        const hasCorrect = newQuestion.options?.some((opt) => opt.isCorrect && opt.text.trim());
        if (!hasCorrect) {
          throw new Error("At least one correct option is required");
        }
      } else if (!newQuestion.correctAnswer.trim()) {
        throw new Error("Correct answer is required");
      }

      const payload: {
        question_text: string;
        question_type: string;
        options?: Array<{ option_text: string; is_correct: boolean }>;
        correct_answer?: string;
        points: number;
      } = {
        question_text: newQuestion.text.trim(),
        question_type: newQuestion.type,
        points: Number(newQuestion.points) || 1,
      };

      if (newQuestion.type === "multiple_choice") {
        payload.options = newQuestion.options
          ?.filter((opt) => opt.text.trim())
          .map((opt) => ({
            option_text: opt.text.trim(),
            is_correct: opt.isCorrect,
          }));
      } else {
        payload.correct_answer = newQuestion.correctAnswer.trim();
      }

      if (editingQuestion) {
        await api.patch(endpoints.quizzes.questions(managedQuiz.id), {
          questionID: editingQuestion.id,
          ...payload,
        });
      } else {
        await api.post(endpoints.quizzes.questions(managedQuiz.id), payload);
      }

      setShowQuestionDialog(false);
      resetQuestionForm();
      await loadQuestions(managedQuiz.id);
      await loadQuizzes();
    } catch (err) {
      logger.error("Failed to save question", err as Error);
      setQuestionsError(err instanceof Error ? err.message : "Failed to save question");
    } finally {
      setSavingQuestion(false);
    }
  };

  const deleteQuestion = async (questionId: number) => {
    if (!managedQuiz) return;

    try {
      setDeletingQuestionId(questionId);
      setQuestionsError(null);
      await api.delete(endpoints.quizzes.questions(managedQuiz.id) + `/${questionId}`);
      await loadQuestions(managedQuiz.id);
      await loadQuizzes();
    } catch (err) {
      logger.error("Failed to delete question", err as Error);
      setQuestionsError(err instanceof Error ? err.message : "Failed to delete question");
    } finally {
      setDeletingQuestionId(null);
    }
  };

  const getQuestionTypeLabel = (type: QuestionType) => {
    switch (type) {
      case "multiple_choice":
        return "Multiple Choice";
      case "true_false":
        return "True/False";
      case "short_answer":
        return "Short Answer";
      default:
        return type;
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Quizzes</h1>
          <p className="text-muted-foreground">
            {isStudent ? "Take quizzes and view your scores" : "Manage course quizzes"}
          </p>
        </div>
        {!isStudent && (
          <Button
            onClick={() => {
              if (!showCreate) {
                void loadCourses();
              }
              setShowCreate((value) => !value);
            }}
          >
            <Plus className="mr-2 h-4 w-4" />
            {showCreate ? "Close" : "Create Quiz"}
          </Button>
        )}
      </div>

      {showCreate && (
        <Card>
          <CardHeader>
            <CardTitle>New Quiz</CardTitle>
            <CardDescription>Provide course and quiz details</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <label className="text-sm font-medium">Course ID</label>
                {courses.length > 0 ? (
                  <select
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    value={newQuiz.courseId}
                    onChange={(event) => setNewQuiz({ ...newQuiz, courseId: event.target.value })}
                  >
                    <option value="">Select a course</option>
                    {courses.map((course) => (
                      <option key={course.id} value={course.id}>
                        {(course.name || course.title || `Course ${course.id}`) + ` (ID: ${course.id})`}
                      </option>
                    ))}
                  </select>
                ) : (
                  <Input
                    value={newQuiz.courseId}
                    onChange={(event) => setNewQuiz({ ...newQuiz, courseId: event.target.value })}
                    placeholder="Course ID"
                  />
                )}
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Title</label>
                <Input
                  value={newQuiz.title}
                  onChange={(event) => setNewQuiz({ ...newQuiz, title: event.target.value })}
                />
              </div>
              <div className="space-y-2 sm:col-span-2">
                <label className="text-sm font-medium">Description</label>
                <Input
                  value={newQuiz.description}
                  onChange={(event) => setNewQuiz({ ...newQuiz, description: event.target.value })}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Duration (min)</label>
                <Input
                  type="number"
                  value={newQuiz.duration}
                  onChange={(event) => setNewQuiz({ ...newQuiz, duration: Number(event.target.value || 30) })}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Due Date (optional)</label>
                <Input
                  type="date"
                  value={newQuiz.dueDate}
                  onChange={(event) => setNewQuiz({ ...newQuiz, dueDate: event.target.value })}
                />
              </div>
            </div>
            <div className="mt-4 flex justify-end">
              <Button onClick={createQuiz} disabled={creating}>
                {creating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Plus className="mr-2 h-4 w-4" />}
                Create
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {error && <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">{error}</div>}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Quizzes</CardDescription>
            <CardTitle className="text-2xl">{quizzes.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Completed</CardDescription>
            <CardTitle className="text-2xl text-green-600">
              {quizzes.filter((quiz) => quiz.status === "completed").length}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Average Score</CardDescription>
            <CardTitle className="text-2xl">
              {(() => {
                const scoredQuizzes = quizzes.filter((quiz) => quiz.score !== undefined);
                if (scoredQuizzes.length === 0) {
                  return "0%";
                }
                const totalScore = scoredQuizzes.reduce((sum, quiz) => sum + (quiz.score || 0), 0);
                const avgScore = totalScore / scoredQuizzes.length;
                const firstQuiz = scoredQuizzes[0];
                return firstQuiz?.totalMarks
                  ? `${Math.round((avgScore / (firstQuiz.totalMarks || 1)) * 100)}%`
                  : `${Math.round(avgScore)}%`;
              })()}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Pending</CardDescription>
            <CardTitle className="text-2xl text-yellow-600">
              {quizzes.filter((quiz) => quiz.status === "not_started").length}
            </CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {loading ? (
          <div className="col-span-full flex items-center justify-center py-16">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : quizzes.length === 0 ? (
          <div className="col-span-full rounded-lg border border-dashed p-10 text-center text-muted-foreground">
            No quizzes found.
          </div>
        ) : (
          quizzes.map((quiz) => {
            const cannotAttempt = quiz.status === "completed" || quiz.attempts >= quiz.maxAttempts;

            return (
              <Card key={quiz.id} className="transition-shadow hover:shadow-md">
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="space-y-1">
                      <CardTitle className="text-lg">{quiz.title}</CardTitle>
                      <CardDescription>{quiz.courseName}</CardDescription>
                    </div>
                    {getStatusBadge(quiz.status)}
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2 text-sm">
                    <div className="flex items-center gap-2">
                      <Clock className="h-4 w-4 text-muted-foreground" />
                      <span>{quiz.duration ?? 0} minutes</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <BookOpen className="h-4 w-4 text-muted-foreground" />
                      <span>{quiz.questionsCount ?? 0} questions</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Trophy className="h-4 w-4 text-muted-foreground" />
                      <span>{quiz.totalMarks ?? 0} marks</span>
                    </div>
                    <div className="text-muted-foreground">
                      Attempts: {quiz.attempts}/{quiz.maxAttempts}
                    </div>
                    {quiz.dueDate && (
                      <div className="text-muted-foreground">
                        Due: {new Date(quiz.dueDate).toLocaleString()}
                      </div>
                    )}
                    {quiz.score !== undefined && (
                      <div className="font-medium text-green-600">
                        Score: {quiz.score}/{quiz.totalMarks}
                      </div>
                    )}
                  </div>

                  {isStudent ? (
                    <Button
                      className="w-full"
                      onClick={() => {
                        if (!cannotAttempt) {
                          void startQuiz(quiz.id);
                        }
                      }}
                      disabled={cannotAttempt || startingQuizId === quiz.id}
                      variant={cannotAttempt ? "outline" : "default"}
                    >
                      {startingQuizId === quiz.id ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Starting...
                        </>
                      ) : cannotAttempt ? (
                        "Completed"
                      ) : (
                        <>
                          <Play className="mr-2 h-4 w-4" />
                          Start Quiz
                        </>
                      )}
                    </Button>
                  ) : (
                    <div className="flex gap-2">
                      <Button
                        className="flex-1"
                        variant="outline"
                        onClick={() => openQuestions(quiz)}
                      >
                        <HelpCircle className="mr-2 h-4 w-4" />
                        Questions
                      </Button>
                      <Button
                        className="flex-1"
                        variant="outline"
                        onClick={() => setManagedQuiz(quiz)}
                      >
                        Manage
                      </Button>
                    </div>
                  )}
                </CardContent>
              </Card>
            );
          })
        )}
      </div>

      {managedQuiz && (
        <Card>
          <CardHeader className="flex flex-row items-start justify-between space-y-0">
            <div>
              <CardTitle>Manage Quiz</CardTitle>
              <CardDescription>{managedQuiz.title}</CardDescription>
            </div>
            <Button variant="ghost" size="sm" onClick={() => setManagedQuiz(null)}>
              Close
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="text-sm text-muted-foreground">
              Course: {managedQuiz.courseName ?? managedQuiz.courseId ?? "Unknown"}
            </div>
            <div className="flex flex-wrap gap-2">
              <Button variant="outline" onClick={() => setManagedQuiz(null)}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={() => {
                  void handleDeleteQuiz(managedQuiz);
                }}
                disabled={deletingQuizId === managedQuiz.id}
              >
                {deletingQuizId === managedQuiz.id ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Deleting...
                  </>
                ) : (
                  <>
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete Quiz
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {showQuestions && managedQuiz && (
        <Dialog open={showQuestions} onOpenChange={setShowQuestions}>
          <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Quiz Questions</DialogTitle>
              <DialogDescription>
                {managedQuiz.title} - {questions.length} questions ({questions.reduce((sum, q) => sum + q.points, 0)} total marks)
              </DialogDescription>
            </DialogHeader>

            {questionsError && (
              <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
                {questionsError}
              </div>
            )}

            {questionsLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin" />
              </div>
            ) : questions.length === 0 ? (
              <div className="rounded-lg border border-dashed p-8 text-center text-muted-foreground">
                No questions yet. Add questions to this quiz.
              </div>
            ) : (
              <div className="space-y-3">
                {questions.map((question, index) => (
                  <div
                    key={question.id}
                    className="rounded-lg border p-3 transition-colors hover:bg-muted/50"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex-1 space-y-1">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium text-muted-foreground">
                            Q{index + 1}
                          </span>
                          <Badge variant="outline">{getQuestionTypeLabel(question.type)}</Badge>
                          <Badge variant="secondary">{question.points} pts</Badge>
                        </div>
                        <p className="text-sm font-medium">{question.text}</p>
                        {question.type === "multiple_choice" && question.options && (
                          <div className="mt-2 space-y-1">
                            {question.options.map((opt, optIdx) => (
                              <div
                                key={optIdx}
                                className={`text-xs ${
                                  opt.isCorrect ? "font-medium text-green-600" : "text-muted-foreground"
                                }`}
                              >
                                {opt.isCorrect && <Check className="mr-1 inline h-3 w-3" />}
                                {String.fromCharCode(65 + optIdx)}. {opt.text}
                              </div>
                            ))}
                          </div>
                        )}
                        {(question.type === "true_false" || question.type === "short_answer") &&
                          question.correctAnswer && (
                            <div className="mt-1 text-xs text-green-600">
                              Answer: {question.correctAnswer}
                            </div>
                          )}
                      </div>
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openEditQuestion(question)}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => deleteQuestion(question.id)}
                          disabled={deletingQuestionId === question.id}
                        >
                          {deletingQuestionId === question.id ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Trash2 className="h-4 w-4 text-destructive" />
                          )}
                        </Button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}

            <DialogFooter>
              <Button variant="outline" onClick={() => setShowQuestions(false)}>
                Close
              </Button>
              <Button onClick={openCreateQuestion}>
                <Plus className="mr-2 h-4 w-4" />
                Add Question
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {showQuestionDialog && (
        <Dialog open={showQuestionDialog} onOpenChange={setShowQuestionDialog}>
          <DialogContent className="max-w-lg">
            <DialogHeader>
              <DialogTitle>{editingQuestion ? "Edit Question" : "Add Question"}</DialogTitle>
              <DialogDescription>
                {editingQuestion ? "Update the question details" : "Create a new question for this quiz"}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Question Text</label>
                <Input
                  value={newQuestion.text}
                  onChange={(e) => setNewQuestion({ ...newQuestion, text: e.target.value })}
                  placeholder="Enter your question"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium">Question Type</label>
                  <select
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    value={newQuestion.type}
                    onChange={(e) =>
                      setNewQuestion({
                        ...newQuestion,
                        type: e.target.value as QuestionType,
                        options:
                          e.target.value === "multiple_choice"
                            ? [
                                { text: "", isCorrect: false },
                                { text: "", isCorrect: false },
                                { text: "", isCorrect: false },
                                { text: "", isCorrect: false },
                              ]
                            : newQuestion.options,
                      })
                    }
                  >
                    <option value="multiple_choice">Multiple Choice</option>
                    <option value="true_false">True/False</option>
                    <option value="short_answer">Short Answer</option>
                  </select>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Points</label>
                  <Input
                    type="number"
                    min="1"
                    value={newQuestion.points}
                    onChange={(e) =>
                      setNewQuestion({ ...newQuestion, points: Number(e.target.value) || 1 })
                    }
                  />
                </div>
              </div>

              {newQuestion.type === "multiple_choice" && (
                <div className="space-y-2">
                  <label className="text-sm font-medium">Options</label>
                  <div className="space-y-2">
                    {newQuestion.options?.map((option, index) => (
                      <div key={index} className="flex items-center gap-2">
                        <input
                          type="radio"
                          name="correctOption"
                          checked={option.isCorrect}
                          onChange={() => {
                            const newOptions = newQuestion.options?.map((opt, i) => ({
                              ...opt,
                              isCorrect: i === index,
                            }));
                            setNewQuestion({ ...newQuestion, options: newOptions });
                          }}
                          className="h-4 w-4"
                        />
                        <Input
                          value={option.text}
                          onChange={(e) => {
                            const newOptions = [...(newQuestion.options || [])];
                            newOptions[index] = { ...newOptions[index], text: e.target.value };
                            setNewQuestion({ ...newQuestion, options: newOptions });
                          }}
                          placeholder={`Option ${String.fromCharCode(65 + index)}`}
                          className="flex-1"
                        />
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {(newQuestion.type === "true_false" || newQuestion.type === "short_answer") && (
                <div className="space-y-2">
                  <label className="text-sm font-medium">Correct Answer</label>
                  {newQuestion.type === "true_false" ? (
                    <select
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      value={newQuestion.correctAnswer}
                      onChange={(e) => setNewQuestion({ ...newQuestion, correctAnswer: e.target.value })}
                    >
                      <option value="">Select answer</option>
                      <option value="true">True</option>
                      <option value="false">False</option>
                    </select>
                  ) : (
                    <Input
                      value={newQuestion.correctAnswer}
                      onChange={(e) => setNewQuestion({ ...newQuestion, correctAnswer: e.target.value })}
                      placeholder="Enter correct answer"
                    />
                  )}
                </div>
              )}

              {questionsError && newQuestion.text && (
                <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
                  {questionsError}
                </div>
              )}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setShowQuestionDialog(false)}>
                Cancel
              </Button>
              <Button onClick={saveQuestion} disabled={savingQuestion}>
                {savingQuestion ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <Plus className="mr-2 h-4 w-4" />
                    {editingQuestion ? "Update" : "Add"} Question
                  </>
                )}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
