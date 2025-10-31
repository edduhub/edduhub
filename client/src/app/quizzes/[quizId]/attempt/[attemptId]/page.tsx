"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { api, endpoints } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import RadioGroup from "@/components/ui/radio-group";

type QuizType = "MultipleChoice" | "TrueFalse" | "ShortAnswer";

type AnswerOption = {
  id: number;
  text: string;
  is_correct?: boolean;
};

type Question = {
  id: number;
  text: string;
  type: QuizType;
  points: number;
  options?: AnswerOption[];
};

type Quiz = {
  id: number;
  title: string;
  questions?: Question[];
};

type StudentAnswer = {
  question_id: number; // wire format compatibility if backend is snake_case tolerant
  QuestionID?: number; // TS convenience to mirror Go struct tags when marshalled
  SelectedOptionID?: number[] | null;
  AnswerText?: string;
};

type Attempt = {
  id: number;
  status: string;
  score?: number;
  quiz?: Quiz;
  answers?: Array<{
    question_id?: number;
    QuestionID?: number;
    SelectedOptionID?: number[] | null;
    AnswerText?: string;
  }>;
};

type PageProps = {
  params: { quizId: string; attemptId: string };
};

export default function QuizAttemptPage({ params }: PageProps) {
  const router = useRouter();
  const quizId = Number(params.quizId);
  const attemptId = Number(params.attemptId);

  const [attempt, setAttempt] = useState<Attempt | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Local answers state keyed by questionId
  const [answers, setAnswers] = useState<Record<number, { optionId?: number; text?: string }>>({});

  const questions = useMemo(() => attempt?.quiz?.questions ?? [], [attempt]);

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        setError(null);
        const data = await api.get<Attempt>(endpoints.quizAttempts.get(quizId, attemptId));
        setAttempt(data);
        // Seed existing answers if any
        const prefilled: Record<number, { optionId?: number; text?: string }> = {};
        (data.answers || []).forEach((a) => {
          const qid = a.QuestionID ?? a.question_id;
          if (!qid) return;
          if (a.SelectedOptionID && a.SelectedOptionID.length > 0) {
            prefilled[qid] = { optionId: a.SelectedOptionID[0] };
          } else if (a.AnswerText) {
            prefilled[qid] = { text: a.AnswerText };
          }
        });
        setAnswers(prefilled);
      } catch (e: any) {
        setError(e?.message || "Failed to load attempt");
      } finally {
        setLoading(false);
      }
    };
    if (quizId && attemptId) load();
  }, [quizId, attemptId]);

  const buildPayload = (): { answers: StudentAnswer[] } => {
    const payload: StudentAnswer[] = (questions || []).map((q) => {
      const a = answers[q.id] || {};
      const base: StudentAnswer = { question_id: q.id, QuestionID: q.id };
      if (q.type === "MultipleChoice" || q.type === "TrueFalse") {
        return { ...base, SelectedOptionID: a.optionId ? [a.optionId] : [] };
      }
      return { ...base, AnswerText: a.text || "" };
    });
    return { answers: payload };
  };

  const onSubmit = async () => {
    try {
      setSubmitting(true);
      setError(null);
      const body = buildPayload();
      const res = await api.post<Attempt>(endpoints.quizAttempts.submit(quizId, attemptId), body);
      setAttempt(res);
    } catch (e: any) {
      setError(e?.message || "Failed to submit attempt");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-[60vh] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <Card>
          <CardHeader>
            <CardTitle>Quiz Attempt</CardTitle>
            <CardDescription>Error</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-destructive">{error}</div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!attempt) return null;

  const completed = attempt.status === "Completed" || attempt.status === "Graded";

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{attempt.quiz?.title || "Quiz"}</h1>
          <p className="text-muted-foreground">Attempt #{attempt.id}</p>
        </div>
        <div>
          <Badge variant={completed ? "secondary" : "default"}>{attempt.status}</Badge>
        </div>
      </div>

      {(questions || []).map((q, idx) => (
        <Card key={q.id}>
          <CardHeader>
            <CardTitle className="text-base">Q{idx + 1}. {q.text}</CardTitle>
            <CardDescription>{q.points} points • {q.type}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {q.type === "MultipleChoice" || q.type === "TrueFalse" ? (
              <RadioGroup
                name={`q-${q.id}`}
                direction="vertical"
                value={answers[q.id]?.optionId}
                onChange={(val) => setAnswers((prev) => ({ ...prev, [q.id]: { ...prev[q.id], optionId: Number(val) } }))}
                options={(q.options || []).map((opt) => ({ label: opt.text, value: opt.id }))}
                className="space-y-2"
              />
            ) : (
              <textarea
                className="min-h-[100px] w-full rounded-md border px-3 py-2"
                value={answers[q.id]?.text || ""}
                onChange={(e) => setAnswers((prev) => ({ ...prev, [q.id]: { ...prev[q.id], text: e.target.value } }))}
                placeholder="Type your answer"
              />
            )}
          </CardContent>
        </Card>
      ))}

      <div className="flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          {completed && typeof attempt.score === "number" ? (
            <>Final Score: <span className="font-medium">{attempt.score}</span></>
          ) : (
            <>Answer all questions and submit your attempt.</>
          )}
        </div>
        {!completed && (
          <Button onClick={onSubmit} disabled={submitting}>
            {submitting ? "Submitting..." : "Submit Attempt"}
          </Button>
        )}
      </div>
    </div>
  );
}


