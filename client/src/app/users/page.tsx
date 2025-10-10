"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Plus, Search, UserPlus, Shield, GraduationCap, Users as UsersIcon } from "lucide-react";

type User = {
  id: string;
  name: string;
  email: string;
  role: 'student' | 'faculty' | 'admin';
  status: 'active' | 'inactive' | 'suspended';
  department?: string;
  joinedDate: string;
  avatar?: string;
};

export default function UsersPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedRole, setSelectedRole] = useState<string>("all");
  
  const [users] = useState<User[]>([
    {
      id: "1",
      name: "Aarav Kumar",
      email: "aarav.kumar@college.edu",
      role: "student",
      status: "active",
      department: "Computer Science",
      joinedDate: "2023-08-15",
      avatar: ""
    },
    {
      id: "2",
      name: "Dr. Priya Sharma",
      email: "priya.sharma@college.edu",
      role: "faculty",
      status: "active",
      department: "Computer Science",
      joinedDate: "2020-06-01",
      avatar: ""
    },
    {
      id: "3",
      name: "Rahul Patel",
      email: "rahul.patel@college.edu",
      role: "admin",
      status: "active",
      department: "Administration",
      joinedDate: "2019-01-10",
      avatar: ""
    },
    {
      id: "4",
      name: "Mira Singh",
      email: "mira.singh@college.edu",
      role: "student",
      status: "active",
      department: "Electronics",
      joinedDate: "2023-08-15",
      avatar: ""
    },
    {
      id: "5",
      name: "Prof. Amit Verma",
      email: "amit.verma@college.edu",
      role: "faculty",
      status: "active",
      department: "Mechanical",
      joinedDate: "2018-07-20",
      avatar: ""
    }
  ]);

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'student': return <GraduationCap className="h-4 w-4" />;
      case 'faculty': return <UsersIcon className="h-4 w-4" />;
      case 'admin': return <Shield className="h-4 w-4" />;
      default: return null;
    }
  };

  const getRoleBadge = (role: string) => {
    const styles = {
      student: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
      faculty: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
      admin: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    };
    return (
      <Badge className={styles[role as keyof typeof styles]}>
        {getRoleIcon(role)}
        <span className="ml-1 capitalize">{role}</span>
      </Badge>
    );
  };

  const getStatusBadge = (status: string) => {
    const styles = {
      active: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      inactive: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400',
      suspended: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    };
    return <Badge className={styles[status as keyof typeof styles]}>{status}</Badge>;
  };

  const filteredUsers = users.filter(user => {
    const matchesSearch = 
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.department?.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesRole = selectedRole === "all" || user.role === selectedRole;
    
    return matchesSearch && matchesRole;
  });

  const userCounts = {
    total: users.length,
    student: users.filter(u => u.role === 'student').length,
    faculty: users.filter(u => u.role === 'faculty').length,
    admin: users.filter(u => u.role === 'admin').length,
    active: users.filter(u => u.status === 'active').length
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">User Management</h1>
          <p className="text-muted-foreground">
            Manage users, roles, and permissions
          </p>
        </div>
        <Button>
          <UserPlus className="mr-2 h-4 w-4" />
          Add User
        </Button>
      </div>

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
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>User</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Department</TableHead>
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
                        <AvatarImage src={user.avatar} />
                        <AvatarFallback>
                          {user.name.split(' ').map(n => n[0]).join('').toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      <span className="font-medium">{user.name}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-muted-foreground">{user.email}</TableCell>
                  <TableCell>{getRoleBadge(user.role)}</TableCell>
                  <TableCell>{user.department}</TableCell>
                  <TableCell>{getStatusBadge(user.status)}</TableCell>
                  <TableCell className="text-muted-foreground">{user.joinedDate}</TableCell>
                  <TableCell>
                    <Button variant="outline" size="sm">
                      Edit
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}