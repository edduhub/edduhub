"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Users, BookOpen, Search, Building2, Loader2 } from "lucide-react";
import { api } from "@/lib/api-client";

type Department = {
  id: number;
  name: string;
  code: string;
  hodName?: string;
  studentCount?: number;
  facultyCount?: number;
  coursesCount?: number;
};

export default function DepartmentsPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [departments, setDepartments] = useState<Department[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);

  const [newDept, setNewDept] = useState({
    name: "",
    code: "",
    description: "",
  });

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        setError(null);
        const data = await api.get<Department[]>("/api/departments");
        setDepartments(Array.isArray(data) ? data : []);
      } catch (e) {
        console.error(e);
        setError("Failed to load departments");
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const filteredDepartments = departments.filter(
    dept => 
      dept.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      dept.code.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const totals = {
    students: departments.reduce((acc, d) => acc + (d.studentCount || 0), 0),
    faculty: departments.reduce((acc, d) => acc + (d.facultyCount || 0), 0),
    courses: departments.reduce((acc, d) => acc + (d.coursesCount || 0), 0),
  };

  const handleCreate = async () => {
    try {
      setCreating(true);
      await api.post("/api/departments", {
        name: newDept.name,
        code: newDept.code,
        description: newDept.description,
      });
      const data = await api.get<Department[]>("/api/departments");
      setDepartments(Array.isArray(data) ? data : []);
      setShowCreate(false);
      setNewDept({ name: "", code: "", description: "" });
    } catch (e) {
      console.error(e);
      setError("Failed to create department");
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Departments</h1>
          <p className="text-muted-foreground">
            Manage academic departments and their resources
          </p>
        </div>
        <Button onClick={() => setShowCreate(v => !v)}>
          <Plus className="mr-2 h-4 w-4" />
          {showCreate ? 'Close' : 'Add Department'}
        </Button>
      </div>

      {showCreate && (
        <Card>
          <CardHeader>
            <CardTitle>New Department</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-3">
              <div className="space-y-2">
                <label className="text-sm font-medium">Name</label>
                <Input value={newDept.name} onChange={e => setNewDept({ ...newDept, name: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Code</label>
                <Input value={newDept.code} onChange={e => setNewDept({ ...newDept, code: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Description</label>
                <Input value={newDept.description} onChange={e => setNewDept({ ...newDept, description: e.target.value })} />
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
        <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Departments</CardDescription>
            <CardTitle className="text-2xl">{departments.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Students</CardDescription>
            <CardTitle className="text-2xl">{totals.students.toLocaleString()}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Faculty</CardDescription>
            <CardTitle className="text-2xl">{totals.faculty}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Courses</CardDescription>
            <CardTitle className="text-2xl">{totals.courses}</CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search departments..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {loading ? (
          <div className="col-span-full flex items-center justify-center py-16">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : (
          filteredDepartments.map((dept) => (
            <Card key={dept.id} className="hover:shadow-md transition-shadow cursor-pointer">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="rounded-lg bg-primary/10 p-3">
                      <Building2 className="h-6 w-6 text-primary" />
                    </div>
                    <div>
                      <CardTitle className="text-lg">{dept.name}</CardTitle>
                      <CardDescription className="text-sm">{dept.code}</CardDescription>
                    </div>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-3 gap-4 text-center">
                  <div className="space-y-1">
                    <div className="flex items-center justify-center">
                      <Users className="h-4 w-4 text-muted-foreground" />
                    </div>
                    <div className="text-2xl font-bold">{dept.studentCount ?? 0}</div>
                    <div className="text-xs text-muted-foreground">Students</div>
                  </div>
                  <div className="space-y-1">
                    <div className="flex items-center justify-center">
                      <Users className="h-4 w-4 text-muted-foreground" />
                    </div>
                    <div className="text-2xl font-bold">{dept.facultyCount ?? 0}</div>
                    <div className="text-xs text-muted-foreground">Faculty</div>
                  </div>
                  <div className="space-y-1">
                    <div className="flex items-center justify-center">
                      <BookOpen className="h-4 w-4 text-muted-foreground" />
                    </div>
                    <div className="text-2xl font-bold">{dept.coursesCount ?? 0}</div>
                    <div className="text-xs text-muted-foreground">Courses</div>
                  </div>
                </div>
                <Button variant="outline" className="w-full">
                  View Details
                </Button>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}