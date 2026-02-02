"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Clock, BookOpen, Trophy, Play, Loader2 } from "lucide-react";
import { api, endpoints } from "@/lib/api-client";
import { logger } from '@/lib/logger';

 type DisplayQuiz = {
  id: number;
  title: string;
  courseName?: string;
  duration?: number;
  totalMarks?: number;
  questionsCount?: number;
  status: 'not_started' | 'in_progress' | 'completed';
  score?: number;
  attempts: number;
  maxAttempts: number;
  startTime?: string;
  endTime?: string;
};

type ApiQuiz = {
  id: number;
  title?: string;
  description?: string;
  courseId?: number;
  courseName?: string;
  duration?: number;
  dueDate?: string;
  status?: string;
  attempts?: number;
  maxAttempts?: number;
  score?: number;
  questions?: Array<{ id: number; text: string; options: string[]; correctOption: number; marks: number }>;
};

export default function QuizzesPage() {
  const router = useRouter();
  const { user } = useAuth();
  const [quizzes, setQuizzes] = useState<DisplayQuiz[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newQuiz, setNewQuiz] = useState({ courseId: '', title: '', description: '', duration: 30, totalMarks: 100 });

  const loadQuizzes = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.get<ApiQuiz[]>(endpoints.quizzes.myQuizzes);
      const display: DisplayQuiz[] = (Array.isArray(data) ? data : []).map((q) => ({
        id: q.id,
        title: q.title ?? 'Untitled Quiz',
        courseName: q.courseName,
        duration: q.duration,
        totalMarks: q.questions?.reduce((sum, question) => sum + (question?.marks ?? 0), 0) ?? undefined,
        questionsCount: q.questions?.length,
        status: q.status === 'completed' ? 'completed' : q.status === 'in_progress' ? 'in_progress' : 'not_started',
        attempts: q.attempts ?? 0,
        maxAttempts: q.maxAttempts ?? 1,
        endTime: q.dueDate,
      }));
      setQuizzes(display);
    } catch (e) {
      logger.error('Error occurred', e as Error);
      setError('Failed to load quizzes');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadQuizzes();
  }, []);

  const getStatusBadge = (status: string) => {
    const styles = {
      not_started: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
      in_progress: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
      completed: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    };
    const labels = {
      not_started: 'Available',
      in_progress: 'In Progress',
      completed: 'Completed'
    };
    return <Badge className={styles[status as keyof typeof styles]}>{labels[status as keyof typeof labels]}</Badge>;
  };

  const startQuiz = async (quizId: number) => {
    try {
      const attempt = await api.post<any>(endpoints.quizAttempts.start(quizId), {});
      setQuizzes(prev => prev.map(q => q.id === quizId ? { ...q, status: 'in_progress', attempts: q.attempts + 1 } : q));
      // Navigate to attempt page
      if (attempt?.id) {
        router.push(`/quizzes/${quizId}/attempt/${attempt.id}`);
      }
    } catch (e) {
      logger.error('Error occurred', e as Error);
      setError('Failed to start quiz');
    }
  };

  const createQuiz = async () => {
    try {
      setCreating(true);
      const courseIdNum = Number(newQuiz.courseId);
      if (!courseIdNum) throw new Error('Course ID is required');
      await api.post<any>(endpoints.quizzes.create(courseIdNum), {
        courseId: courseIdNum,
        title: newQuiz.title,
        description: newQuiz.description,
        duration: Number(newQuiz.duration),
        totalMarks: Number(newQuiz.totalMarks),
        allowedAttempts: 1,
      });
      await loadQuizzes();
      setShowCreate(false);
      setNewQuiz({ courseId: '', title: '', description: '', duration: 30, totalMarks: 100 });
    } catch (error) {
      logger.error('Error occurred', error instanceof Error ? error : new Error(String(error)));
      setError(error instanceof Error ? error.message : 'Failed to create quiz');
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Quizzes</h1>
          <p className="text-muted-foreground">
            {user?.role === 'student' ? 'Take quizzes and view your scores' : 'Manage course quizzes'}
          </p>
        </div>
        {user?.role !== 'student' && (
          <Button onClick={() => setShowCreate(v => !v)}>
            <Plus className="mr-2 h-4 w-4" />
            {showCreate ? 'Close' : 'Create Quiz'}
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
                <input className="w-full rounded-md border px-3 py-2" value={newQuiz.courseId} onChange={e => setNewQuiz({ ...newQuiz, courseId: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Title</label>
                <input className="w-full rounded-md border px-3 py-2" value={newQuiz.title} onChange={e => setNewQuiz({ ...newQuiz, title: e.target.value })} />
              </div>
              <div className="space-y-2 sm:col-span-2">
                <label className="text-sm font-medium">Description</label>
                <input className="w-full rounded-md border px-3 py-2" value={newQuiz.description} onChange={e => setNewQuiz({ ...newQuiz, description: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Duration (min)</label>
                <input type="number" className="w-full rounded-md border px-3 py-2" value={newQuiz.duration} onChange={e => setNewQuiz({ ...newQuiz, duration: Number(e.target.value || 30) })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Total Marks</label>
                <input type="number" className="w-full rounded-md border px-3 py-2" value={newQuiz.totalMarks} onChange={e => setNewQuiz({ ...newQuiz, totalMarks: Number(e.target.value || 100) })} />
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

      {error && (
        <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">{error}</div>
      )}

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
              {quizzes.filter(q => q.status === 'completed').length}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Average Score</CardDescription>
            <CardTitle className="text-2xl">
              {(() => {
                const scoredQuizzes = quizzes.filter(q => q.score !== undefined);
                if (scoredQuizzes.length === 0) return '0%';
                const totalScore = scoredQuizzes.reduce((acc, q) => acc + (q.score || 0), 0);
                const avgScore = totalScore / scoredQuizzes.length;
                const firstQuiz = scoredQuizzes[0];
                return firstQuiz?.totalMarks ? `${Math.round((avgScore / (firstQuiz.totalMarks || 1)) * 100)}%` : '0%';
              })()}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Pending</CardDescription>
            <CardTitle className="text-2xl text-yellow-600">
              {quizzes.filter(q => q.status === 'not_started').length}
            </CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {loading ? (
          <div className="col-span-full flex items-center justify-center py-16">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : (
          quizzes.map((quiz) => (
            <Card key={quiz.id} className="hover:shadow-md transition-shadow">
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
                    <span>{quiz.duration} minutes</span>
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
                  {quiz.score !== undefined && (
                    <div className="font-medium text-green-600">
                      Score: {quiz.score}/{quiz.totalMarks}
                    </div>
                  )}
                </div>
                {user?.role === 'student' ? (
                  <Button 
                    className="w-full"
                    onClick={() => quiz.status === 'completed' ? undefined : startQuiz(quiz.id)}
                    disabled={(quiz.status as DisplayQuiz['status']) === 'completed' || (quiz.attempts >= quiz.maxAttempts && (quiz.status as DisplayQuiz['status']) === 'completed')}
                    variant={quiz.status === 'completed' ? 'outline' : 'default'}
                  >
                    {quiz.status === 'completed' ? 'Completed' : (
                      <>
                        <Play className="mr-2 h-4 w-4" />
                        Start Quiz
                      </>
                    )}
                  </Button>
                ) : (
                  <Button className="w-full" variant="outline">Manage</Button>
                )}
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}