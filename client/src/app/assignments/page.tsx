"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api, endpoints } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Plus, Calendar, Clock, FileText, CheckCircle, Loader2 } from "lucide-react";
import { format } from "date-fns";

type Assignment = {
  id: number;
  title: string;
  courseId?: number;
  courseName?: string;
  dueDate: string;
  maxScore: number;
  status?: 'pending' | 'submitted' | 'graded';
  score?: number;
  description: string;
};

type ApiAssignment = {
  id: number;
  title?: string;
  description?: string;
  courseId?: number;
  courseName?: string;
  dueDate?: string;
  maxScore?: number;
  status?: string;
  score?: number;
};

const normalizeStatus = (status?: string): Assignment['status'] => {
  if (status === 'submitted' || status === 'graded') {
    return status;
  }
  return 'pending';
};

export default function AssignmentsPage() {
  const { user } = useAuth();
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Create form
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newAssignment, setNewAssignment] = useState({
    courseId: '',
    title: '',
    description: '',
    dueDate: '',
    maxScore: 100,
  });

  useEffect(() => {
    const fetchAssignments = async () => {
      try {
        setLoading(true);
        const response = await api.get<ApiAssignment[]>(endpoints.assignments.list);
        const normalized = (Array.isArray(response) ? response : []).map<Assignment>((item) => ({
          id: item.id,
          title: item.title ?? 'Untitled assignment',
          courseId: item.courseId,
          courseName: item.courseName ?? (item.courseId ? `Course ${item.courseId}` : undefined),
          dueDate: item.dueDate ?? new Date().toISOString(),
          maxScore: item.maxScore ?? 100,
          status: normalizeStatus(item.status),
          score: item.score,
          description: item.description ?? '',
        }));
        setAssignments(normalized);
      } catch (err) {
        console.error('Failed to fetch assignments:', err);
        setError('Failed to load assignments');
      } finally {
        setLoading(false);
      }
    };

    fetchAssignments();
  }, []);

  const getStatusBadge = (status: string) => {
    const styles = {
      pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
      submitted: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
      graded: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    };
    return <Badge className={styles[status as keyof typeof styles]}>{status}</Badge>;
  };

  const isOverdue = (dueDate: string) => new Date(dueDate) < new Date();

  const handleCreate = async () => {
    try {
      setCreating(true);
      setError(null);
      const courseIdNum = Number(newAssignment.courseId);
      if (!courseIdNum) throw new Error('Course ID is required');
      await api.post(endpoints.assignments.create(courseIdNum), {
        courseId: courseIdNum,
        title: newAssignment.title,
        description: newAssignment.description,
        dueDate: newAssignment.dueDate,
        maxScore: Number(newAssignment.maxScore),
      });
      const refreshed = await api.get<ApiAssignment[]>(endpoints.assignments.list);
      const normalized = (Array.isArray(refreshed) ? refreshed : []).map<Assignment>((item) => ({
        id: item.id,
        title: item.title ?? 'Untitled assignment',
        courseId: item.courseId,
        courseName: item.courseName ?? (item.courseId ? `Course ${item.courseId}` : undefined),
        dueDate: item.dueDate ?? new Date().toISOString(),
        maxScore: item.maxScore ?? 100,
        status: normalizeStatus(item.status),
        score: item.score,
        description: item.description ?? '',
      }));
      setAssignments(normalized);
      setShowCreate(false);
      setNewAssignment({ courseId: '', title: '', description: '', dueDate: '', maxScore: 100 });
    } catch (e: any) {
      console.error(e);
      setError(e?.message || 'Failed to create assignment');
    } finally {
      setCreating(false);
    }
  };

  const handleSubmit = async (a: Assignment) => {
    try {
      if (!a.courseId) {
        setError('Course ID not available for this assignment');
        return;
      }
      await api.post(endpoints.assignments.submit(a.courseId, a.id), { content: 'Submitted via portal' });
      setAssignments(prev => prev.map(x => x.id === a.id ? { ...x, status: 'submitted' } : x));
    } catch (e) {
      console.error(e);
      setError('Failed to submit assignment');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Assignments</h1>
          <p className="text-muted-foreground">
            {user?.role === 'student' ? 'View and submit your assignments' : 'Manage course assignments'}
          </p>
        </div>
        {user?.role !== 'student' && (
          <Button onClick={() => setShowCreate(v => !v)}>
            <Plus className="mr-2 h-4 w-4" />
            {showCreate ? 'Close' : 'Create Assignment'}
          </Button>
        )}
      </div>

      {showCreate && (
        <Card>
          <CardHeader>
            <CardTitle>New Assignment</CardTitle>
            <CardDescription>Provide course ID and details</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <label className="text-sm font-medium">Course ID</label>
                <Input value={newAssignment.courseId} onChange={e => setNewAssignment({ ...newAssignment, courseId: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Title</label>
                <Input value={newAssignment.title} onChange={e => setNewAssignment({ ...newAssignment, title: e.target.value })} />
              </div>
              <div className="space-y-2 sm:col-span-2">
                <label className="text-sm font-medium">Description</label>
                <Input value={newAssignment.description} onChange={e => setNewAssignment({ ...newAssignment, description: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Due Date</label>
                <Input type="date" value={newAssignment.dueDate} onChange={e => setNewAssignment({ ...newAssignment, dueDate: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Max Score</label>
                <Input type="number" value={newAssignment.maxScore} onChange={e => setNewAssignment({ ...newAssignment, maxScore: Number(e.target.value || 100) })} />
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

      {error && (
        <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">{error}</div>
      )}

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
              {assignments.filter(a => a.status === 'pending').length}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Graded</CardDescription>
            <CardTitle className="text-2xl text-green-600">
              {assignments.filter(a => a.status === 'graded').length}
            </CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="space-y-4">
        {assignments.map((assignment) => (
          <Card key={assignment.id} className="hover:shadow-md transition-shadow">
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
                    <span className={isOverdue(assignment.dueDate) && assignment.status === 'pending' ? 'text-destructive font-medium' : ''}>
                      Due: {format(new Date(assignment.dueDate), 'MMM dd, yyyy')}
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
                  {user?.role === 'student' ? (
                    <Button variant={assignment.status === 'pending' ? 'default' : 'outline'} onClick={() => handleSubmit(assignment)}>
                      {assignment.status === 'pending' ? 'Submit' : 'View'}
                    </Button>
                  ) : (
                    <Button variant="outline">Manage</Button>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
