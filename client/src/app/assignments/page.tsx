"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Calendar, Clock, FileText, CheckCircle, XCircle, Loader2 } from "lucide-react";
import { format } from "date-fns";

type Assignment = {
  id: number;
  title: string;
  courseName: string;
  dueDate: string;
  maxScore: number;
  status: 'pending' | 'submitted' | 'graded';
  score?: number;
  description: string;
};

export default function AssignmentsPage() {
  const { user } = useAuth();
  const [assignments, setAssignments] = useState<Assignment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchAssignments = async () => {
      try {
        setLoading(true);
        const response = await api.get('/api/assignments');
        setAssignments(Array.isArray(response) ? response : []);
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

  const isOverdue = (dueDate: string) => {
    return new Date(dueDate) < new Date();
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
        {user?.role === 'faculty' && (
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Create Assignment
          </Button>
        )}
      </div>

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
                {getStatusBadge(assignment.status)}
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
                <Button variant={assignment.status === 'pending' ? 'default' : 'outline'}>
                  {assignment.status === 'pending' ? 'Submit' : 'View Details'}
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
