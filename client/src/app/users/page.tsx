"use client";

import { useState, useEffect, useMemo } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Search, UserPlus, Shield, GraduationCap, Users as UsersIcon, Loader2 } from "lucide-react";
import { api, endpoints } from "@/lib/api-client";
import { logger } from "@/lib/logger";

type User = {
  id: number;
  name: string;
  email: string;
  role: "student" | "faculty" | "admin";
  is_active: boolean;
  created_at: string;
  kratos_identity_id: string;
};

type UserApi = Partial<User> & {
  first_name?: string;
  last_name?: string;
};

const roleCycle: User["role"][] = ["student", "faculty", "admin"];

export default function UsersPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedRole, setSelectedRole] = useState<string>("all");
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [newUser, setNewUser] = useState({
    name: "",
    email: "",
    role: "student" as User["role"],
    kratosIdentityId: "",
    isActive: true,
  });

  const normalizeUser = (item: UserApi): User => {
    const id = Number(item.id ?? 0);
    const fallbackName = [item.first_name, item.last_name].filter(Boolean).join(" ").trim();
    const role = item.role === "faculty" || item.role === "admin" ? item.role : "student";

    return {
      id,
      name: (item.name || fallbackName || `User ${id}`).trim(),
      email: item.email || "",
      role,
      is_active: Boolean(item.is_active),
      created_at: item.created_at || new Date().toISOString(),
      kratos_identity_id: item.kratos_identity_id || "",
    };
  };

  const loadUsers = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.get<UserApi[]>(endpoints.users.list);
      setUsers((Array.isArray(data) ? data : []).map(normalizeUser));
    } catch (err) {
      logger.error("Failed to load users:", err as Error);
      setError("Failed to load users.");
      setUsers([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadUsers();
  }, []);

  const getRoleIcon = (role: string) => {
    switch (role) {
      case "student":
        return <GraduationCap className="h-4 w-4" />;
      case "faculty":
        return <UsersIcon className="h-4 w-4" />;
      case "admin":
        return <Shield className="h-4 w-4" />;
      default:
        return null;
    }
  };

  const getRoleBadge = (role: string) => {
    const styles = {
      student: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400",
      faculty: "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400",
      admin: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
    };

    return (
      <Badge className={styles[role as keyof typeof styles] || styles.student}>
        {getRoleIcon(role)}
        <span className="ml-1 capitalize">{role}</span>
      </Badge>
    );
  };

  const getStatusBadge = (isActive: boolean) => {
    const status = isActive ? "active" : "inactive";
    const styles = {
      active: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
      inactive: "bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400",
    };
    return <Badge className={styles[status as keyof typeof styles]}>{status}</Badge>;
  };

  const filteredUsers = useMemo(() => {
    return users.filter((user) => {
      const matchesSearch =
        user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        user.email.toLowerCase().includes(searchQuery.toLowerCase());

      const matchesRole = selectedRole === "all" || user.role === selectedRole;
      return matchesSearch && matchesRole;
    });
  }, [users, searchQuery, selectedRole]);

  const userCounts = {
    total: users.length,
    student: users.filter((u) => u.role === "student").length,
    faculty: users.filter((u) => u.role === "faculty").length,
    admin: users.filter((u) => u.role === "admin").length,
    active: users.filter((u) => u.is_active).length,
  };

  const toggleActive = async (u: User) => {
    try {
      setError(null);
      await api.patch(endpoints.users.updateStatus(u.id), { is_active: !u.is_active });
      setUsers((prev) => prev.map((item) => (item.id === u.id ? { ...item, is_active: !u.is_active } : item)));
    } catch (err) {
      logger.error("Failed to update user status:", err as Error);
      setError("Failed to update user status.");
    }
  };

  const cycleRole = async (u: User) => {
    const next = roleCycle[(roleCycle.indexOf(u.role) + 1) % roleCycle.length];
    try {
      setError(null);
      await api.patch(endpoints.users.updateRole(u.id), { role: next });
      setUsers((prev) => prev.map((item) => (item.id === u.id ? { ...item, role: next } : item)));
    } catch (err) {
      logger.error("Failed to update user role:", err as Error);
      setError("Failed to update user role.");
    }
  };

  const handleCreateUser = async () => {
    setCreateError(null);

    const name = newUser.name.trim();
    const email = newUser.email.trim();
    const kratosIdentityId = newUser.kratosIdentityId.trim();

    if (!name) {
      setCreateError("Name is required.");
      return;
    }
    if (!/^\S+@\S+\.\S+$/.test(email)) {
      setCreateError("A valid email is required.");
      return;
    }
    if (!kratosIdentityId) {
      setCreateError("Kratos identity ID is required.");
      return;
    }

    try {
      setIsCreating(true);
      await api.post(endpoints.users.create, {
        name,
        email,
        role: newUser.role,
        kratos_identity_id: kratosIdentityId,
        is_active: newUser.isActive,
      });

      setIsCreateOpen(false);
      setNewUser({
        name: "",
        email: "",
        role: "student",
        kratosIdentityId: "",
        isActive: true,
      });
      await loadUsers();
    } catch (err) {
      logger.error("Failed to create user:", err as Error);
      setCreateError("Failed to create user. Verify values and try again.");
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold">User Management</h1>
            <p className="text-muted-foreground">Manage users, roles, and permissions</p>
          </div>
          <Button
            onClick={() => {
              setCreateError(null);
              setIsCreateOpen(true);
            }}
          >
            <UserPlus className="mr-2 h-4 w-4" />
            Add User
          </Button>
        </div>

        {error && <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">{error}</div>}

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">Total Users</CardTitle>
              <div className="text-2xl font-bold">{userCounts.total}</div>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">Students</CardTitle>
              <div className="text-2xl font-bold text-blue-600">{userCounts.student}</div>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">Faculty</CardTitle>
              <div className="text-2xl font-bold text-purple-600">{userCounts.faculty}</div>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">Admins</CardTitle>
              <div className="text-2xl font-bold text-red-600">{userCounts.admin}</div>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">Active</CardTitle>
              <div className="text-2xl font-bold text-green-600">{userCounts.active}</div>
            </CardHeader>
          </Card>
        </div>

        <div className="flex items-center gap-4">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search users..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
          <div className="flex gap-2">
            {["all", "student", "faculty", "admin"].map((role) => (
              <Button
                key={role}
                variant={selectedRole === role ? "default" : "outline"}
                size="sm"
                onClick={() => setSelectedRole(role)}
              >
                {role === "all" ? "All" : role.charAt(0).toUpperCase() + role.slice(1)}
              </Button>
            ))}
          </div>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>All Users</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="flex items-center justify-center py-16">
                <Loader2 className="h-6 w-6 animate-spin" />
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>Role</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Joined</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredUsers.map((user) => (
                    <TableRow key={user.id}>
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <Avatar className="h-9 w-9">
                            <AvatarImage src={undefined} />
                            <AvatarFallback>
                              {user.name
                                .split(" ")
                                .filter(Boolean)
                                .map((part) => part[0])
                                .join("")
                                .slice(0, 2)
                                .toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <span className="font-medium">{user.name}</span>
                        </div>
                      </TableCell>
                      <TableCell className="text-muted-foreground">{user.email}</TableCell>
                      <TableCell>{getRoleBadge(user.role)}</TableCell>
                      <TableCell>{getStatusBadge(user.is_active)}</TableCell>
                      <TableCell className="text-muted-foreground">{new Date(user.created_at).toLocaleDateString()}</TableCell>
                      <TableCell className="space-x-2">
                        <Button variant="outline" size="sm" onClick={() => cycleRole(user)}>
                          Cycle Role
                        </Button>
                        <Button variant="outline" size="sm" onClick={() => toggleActive(user)}>
                          {user.is_active ? "Deactivate" : "Activate"}
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                  {filteredUsers.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={6} className="h-24 text-center text-sm text-muted-foreground">
                        No users found.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>

      {isCreateOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 p-4 flex items-center justify-center">
          <Card className="w-full max-w-xl">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Create User</CardTitle>
                <Button variant="ghost" size="sm" onClick={() => setIsCreateOpen(false)}>
                  Close
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="newUserName">Name</label>
                  <Input
                    id="newUserName"
                    value={newUser.name}
                    onChange={(e) => setNewUser((prev) => ({ ...prev, name: e.target.value }))}
                    placeholder="Full name"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="newUserEmail">Email</label>
                  <Input
                    id="newUserEmail"
                    type="email"
                    value={newUser.email}
                    onChange={(e) => setNewUser((prev) => ({ ...prev, email: e.target.value }))}
                    placeholder="name@college.edu"
                  />
                </div>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="newUserRole">Role</label>
                  <select
                    id="newUserRole"
                    value={newUser.role}
                    onChange={(e) => setNewUser((prev) => ({ ...prev, role: e.target.value as User["role"] }))}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="student">Student</option>
                    <option value="faculty">Faculty</option>
                    <option value="admin">Admin</option>
                  </select>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium" htmlFor="newUserStatus">Status</label>
                  <select
                    id="newUserStatus"
                    value={newUser.isActive ? "active" : "inactive"}
                    onChange={(e) => setNewUser((prev) => ({ ...prev, isActive: e.target.value === "active" }))}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                  </select>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="newUserKratosId">Kratos Identity ID</label>
                <Input
                  id="newUserKratosId"
                  value={newUser.kratosIdentityId}
                  onChange={(e) => setNewUser((prev) => ({ ...prev, kratosIdentityId: e.target.value }))}
                  placeholder="identity UUID"
                />
              </div>

              {createError && <p className="text-sm text-destructive">{createError}</p>}

              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setIsCreateOpen(false)} disabled={isCreating}>
                  Cancel
                </Button>
                <Button onClick={handleCreateUser} disabled={isCreating}>
                  {isCreating ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    <>
                      <UserPlus className="mr-2 h-4 w-4" />
                      Create User
                    </>
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </>
  );
}
