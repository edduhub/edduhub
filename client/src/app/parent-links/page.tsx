"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Trash2, Link2, Loader2, Users } from "lucide-react";
import { api, endpoints } from "@/lib/api-client";
import { useStudents } from "@/lib/api-hooks";
import { logger } from "@/lib/logger";

type ParentRelationship = {
  id: number;
  parentUserId: number;
  parentName: string;
  parentEmail: string;
  studentId: number;
  studentRollNo: string;
  studentName: string;
  relation: string;
  isPrimaryContact: boolean;
  receiveNotifications: boolean;
  isVerified: boolean;
  createdAt: string;
};

type User = {
  id: number;
  name: string;
  email: string;
  role: string;
  is_active?: boolean;
};

export default function ParentLinksPage() {
  const router = useRouter();
  const { user, isLoading: authLoading } = useAuth();
  const [relationships, setRelationships] = useState<ParentRelationship[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [form, setForm] = useState({
    parentUserId: 0,
    studentId: 0,
    relation: "guardian" as "father" | "mother" | "guardian",
    isPrimaryContact: true,
    receiveNotifications: true,
  });

  const { data: students = [] } = useStudents();

  const loadRelationships = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.get<{ relationships: ParentRelationship[] }>(
        endpoints.parentRelationships.list
      );
      setRelationships(data?.relationships ?? []);
    } catch (err) {
      logger.error("Failed to load parent links", err as Error);
      setError("Failed to load parent links.");
      setRelationships([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const loadUsers = useCallback(async () => {
    try {
      const data = await api.get<User[]>(endpoints.users.list);
      setUsers(Array.isArray(data) ? data : []);
    } catch (err) {
      logger.error("Failed to load users", err as Error);
      setUsers([]);
    }
  }, []);

  useEffect(() => {
    if (!authLoading && user?.role !== "admin") {
      router.push("/");
      return;
    }
    if (user?.role === "admin") {
      loadRelationships();
      loadUsers();
    }
  }, [authLoading, user?.role, router, loadRelationships, loadUsers]);

  const parentUsers = users.filter((u) => u.role === "parent");

  const handleCreate = async () => {
    if (!form.parentUserId || !form.studentId) {
      setCreateError("Please select both parent and student.");
      return;
    }
    try {
      setIsCreating(true);
      setCreateError(null);
      await api.post(endpoints.parentRelationships.create, {
        parentUserId: form.parentUserId,
        studentId: form.studentId,
        relation: form.relation,
        isPrimaryContact: form.isPrimaryContact,
        receiveNotifications: form.receiveNotifications,
      });
      setShowCreate(false);
      setForm({
        parentUserId: 0,
        studentId: 0,
        relation: "guardian",
        isPrimaryContact: true,
        receiveNotifications: true,
      });
      loadRelationships();
    } catch (err) {
      logger.error("Failed to create link", err as Error);
      setCreateError(
        err instanceof Error ? err.message : "Failed to create parent-student link."
      );
    } finally {
      setIsCreating(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Remove this parent-student link?")) return;
    try {
      setError(null);
      await api.delete(endpoints.parentRelationships.delete(id));
      loadRelationships();
    } catch (err) {
      logger.error("Failed to delete link", err as Error);
      setError("Failed to remove link.");
    }
  };

  if (authLoading || (user && user.role !== "admin")) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-12 w-12 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="container mx-auto space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Parent-Student Links</h1>
          <p className="text-muted-foreground">
            Link parent accounts to student profiles for the parent portal.
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)}>
          <Link2 className="mr-2 h-4 w-4" />
          Add Link
        </Button>
      </div>

      {error && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-destructive">
          {error}
        </div>
      )}

      {/* Create form */}
      {showCreate && (
        <Card>
          <CardHeader>
            <CardTitle>Add Parent-Student Link</CardTitle>
            <CardDescription>
              Select a parent user and student to link. The parent will be able
              to view this student&apos;s data in the parent portal.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {createError && (
              <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
                {createError}
              </div>
            )}
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <Label>Parent</Label>
                <Select
                  value={form.parentUserId ? String(form.parentUserId) : ""}
                  onValueChange={(v) =>
                    setForm((prev) => ({ ...prev, parentUserId: parseInt(v, 10) || 0 }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select parent..." />
                  </SelectTrigger>
                  <SelectContent>
                    {parentUsers.map((u) => (
                      <SelectItem key={u.id} value={String(u.id)}>
                        {u.name} ({u.email})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Student</Label>
                <Select
                  value={form.studentId ? String(form.studentId) : ""}
                  onValueChange={(v) =>
                    setForm((prev) => ({ ...prev, studentId: parseInt(v, 10) || 0 }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select student..." />
                  </SelectTrigger>
                  <SelectContent>
                    {students.map((s) => (
                      <SelectItem key={s.id} value={String(s.id)}>
                        {s.firstName} {s.lastName} ({s.rollNo})
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label>Relation</Label>
                <Select
                  value={form.relation}
                  onValueChange={(v) =>
                    setForm((prev) => ({
                      ...prev,
                      relation: v as "father" | "mother" | "guardian",
                    }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="father">Father</SelectItem>
                    <SelectItem value="mother">Mother</SelectItem>
                    <SelectItem value="guardian">Guardian</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="flex gap-2">
              <Button onClick={handleCreate} disabled={isCreating}>
                {isCreating ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <Link2 className="mr-2 h-4 w-4" />
                )}
                Create Link
              </Button>
              <Button
                variant="outline"
                onClick={() => {
                  setShowCreate(false);
                  setCreateError(null);
                }}
              >
                Cancel
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Relationships table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            Existing Links
          </CardTitle>
          <CardDescription>
            {relationships.length} parent-student link{relationships.length !== 1 ? "s" : ""} configured.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : relationships.length === 0 ? (
            <div className="py-12 text-center text-muted-foreground">
              No parent-student links yet. Click &quot;Add Link&quot; to create one.
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Parent</TableHead>
                  <TableHead>Student</TableHead>
                  <TableHead>Relation</TableHead>
                  <TableHead>Linked</TableHead>
                  <TableHead className="w-[80px]" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {relationships.map((r) => (
                  <TableRow key={r.id}>
                    <TableCell>
                      <div>
                        <div className="font-medium">{r.parentName}</div>
                        <div className="text-sm text-muted-foreground">
                          {r.parentEmail}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div>
                        <div className="font-medium">{r.studentName}</div>
                        <div className="text-sm text-muted-foreground">
                          {r.studentRollNo}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell className="capitalize">{r.relation}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(r.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(r.id)}
                        className="text-destructive hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
