"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Clock, BookOpen, Trophy, Play } from "lucide-react";
import { fetchQuizzes } from "@/lib/api-client";

type Quiz = {
  id: number;
  title: string;
  description: string;
  course_id: number;
  time_limit_minutes: number;
  due_date: string;
  created_at: string;
  updated_at: string;
  course?: {
    id: number;
    name: string;
  };
  questions?: any[];
};

type DisplayQuiz = {
  id: number;
  title: string;
  courseName: string;
  duration: number;
  totalMarks: number;
  questionsCount: number;
  status: 'not_started' | 'in_progress' | 'completed';
  score?: number;
  attempts: number;
  maxAttempts: number;
  startTime?: string;
  endTime?: string;
};

export default function QuizzesPage() {
  const { user } = useAuth();
  const [quizzes, setQuizzes] = useState<DisplayQuiz[]>([]);

  useEffect(() => {
    const loadQuizzes = async () => {
      const data = await fetchQuizzes();
      const displayQuizzes: DisplayQuiz[] = data.map((quiz: Quiz) => ({
        id: quiz.id,
        title: quiz.title,
        courseName: quiz.course?.name || `Course ${quiz.course_id}`,
        duration: quiz.time_limit_minutes,
        totalMarks: (quiz.questions?.length || 0) * 10,
        questionsCount: quiz.questions?.length || 0,
        status: 'not_started' as const,
        attempts: 0,
        maxAttempts: 1,
        endTime: quiz.due_date,
      }));
      setQuizzes(displayQuizzes);
    };

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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Quizzes</h1>
          <p className="text-muted-foreground">
            {user?.role === 'student' ? 'Take quizzes and view your scores' : 'Manage course quizzes'}
          </p>
        </div>
        {user?.role === 'faculty' && (
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Create Quiz
          </Button>
        )}
      </div>

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
                return `${Math.round((avgScore / firstQuiz.totalMarks) * 100)}%`;
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
        {quizzes.map((quiz) => (
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
                  <span>{quiz.questionsCount} questions</span>
                </div>
                <div className="flex items-center gap-2">
                  <Trophy className="h-4 w-4 text-muted-foreground" />
                  <span>{quiz.totalMarks} marks</span>
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
              <Button 
                className="w-full" 
                disabled={quiz.attempts >= quiz.maxAttempts && quiz.status === 'completed'}
              >
                {quiz.status === 'completed' ? (
                  quiz.attempts < quiz.maxAttempts ? 'Retry Quiz' : 'View Results'
                ) : (
                  <>
                    <Play className="mr-2 h-4 w-4" />
                    Start Quiz
                  </>
                )}
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}