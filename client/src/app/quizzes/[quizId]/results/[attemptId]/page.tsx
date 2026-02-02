"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";

type AnswerOption = {
  id: number;
  text: string;
  is_correct?: boolean;
};

type Question = {
  id: number;
  text: string;
  type: "MultipleChoice" | "TrueFalse" | "ShortAnswer";
  points: number;
  options?: AnswerOption[];
};

type Quiz = {
  id: number;
  title: string;
  questions?: Question[];
};

type Answer = {
  question_id: number;
  selected_option_id?: number[] | null;
  answer_text?: string;
  points_awarded?: number | null;
  is_correct?: boolean | null;
};

type AttemptResponse = {
  id: number;
  status: string;
  score?: number;
  quiz?: Quiz;
  answers?: Answer[];
};

export default function QuizResultsPage() {
  const params = useParams<{ quizId: string; attemptId: string }>();
  const router = useRouter();
  const attemptId = Number(params.attemptId);

  const [attempt, setAttempt] = useState<AttemptResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const questions: Question[] = useMemo(() => attempt?.quiz?.questions || [], [attempt]);
  const answersByQ = useMemo(() => {
    const map = new Map<number, Answer>();
    (attempt?.answers || []).forEach((a) => map.set(a.question_id, a));
    return map;
  }, [attempt]);

  useEffect(() => {
    const loadAttempt = async () => {
      try {
        setLoading(true);
        setError(null);
        const data = await api.get<AttemptResponse>(`/api/attempts/${attemptId}`);
        setAttempt(data);
      } catch (error) {
        setError(error instanceof Error ? error.message : "Failed to load results");
      } finally {
        setLoading(false);
      }
    };
    if (attemptId) loadAttempt();
  }, [attemptId]);

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  if (error) {
    return <div className="text-sm text-destructive">{error}</div>;
  }

  if (!attempt || !attempt.quiz) {
    return <div className="text-sm text-muted-foreground">No results available.</div>;
  }

  const totalPoints = questions.reduce((sum, q) => sum + (q.points || 0), 0);
  const score = attempt.score ?? 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{attempt.quiz.title}</h1>
          <div className="text-muted-foreground">Attempt #{attempt.id}</div>
        </div>
        <div className="flex items-center gap-3">
          <Badge className="bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
            Score: {score}/{totalPoints}
          </Badge>
          <Button variant="outline" onClick={() => router.push("/quizzes")}>Back to Quizzes</Button>
        </div>
      </div>

      {questions.map((q, idx) => {
        const ans = answersByQ.get(q.id);
        const isCorrect = ans?.is_correct === true;
        const partial = ans?.points_awarded && ans.points_awarded > 0 && !isCorrect;
        return (
          <Card key={q.id}>
            <CardHeader>
              <CardTitle className="text-base">
                {idx + 1}. {q.text} <span className="text-muted-foreground">({q.points} pts)</span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="mb-2">
                {isCorrect ? (
                  <Badge className="bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">Correct</Badge>
                ) : partial ? (
                  <Badge className="bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400">Partial</Badge>
                ) : (
                  <Badge className="bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400">Incorrect</Badge>
                )}
                {typeof ans?.points_awarded === "number" ? (
                  <span className="ml-3 text-sm">Points awarded: {ans.points_awarded}</span>
                ) : null}
              </div>

              {q.type === "ShortAnswer" ? (
                <div className="space-y-1 text-sm">
                  <div className="text-muted-foreground">Your answer</div>
                  <div className="rounded-md border p-2">{ans?.answer_text || "—"}</div>
                </div>
              ) : (
                <div className="space-y-2">
                  {(q.options || []).map((opt) => {
                    const chosen = (ans?.selected_option_id || []).includes(opt.id);
                    const correct = !!opt.is_correct;
                    return (
                      <div key={opt.id} className="text-sm">
                        <span
                          className={
                            correct ? "text-green-600" : chosen ? "text-red-600" : "text-foreground"
                          }
                        >
                          {chosen ? "● " : "○ "}{opt.text}
                        </span>
                        {correct ? <span className="ml-2 text-xs text-green-600">(correct)</span> : null}
                      </div>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}


