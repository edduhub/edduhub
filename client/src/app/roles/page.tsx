"use client";

import { useState, useEffect, useCallback, useMemo } from 'react';
import { useAuth } from '@/lib/auth-context';
import { api, endpoints } from '@/lib/api-client';
import { logger } from '@/lib/logger';
import type { Role, Permission } from '@/lib/types';
import {
  Shield,
  Lock,
  UserPlus,
  Key,
  Plus,
  Edit2,
  Trash2,
  Search,
  Settings2,
  Users,
  Loader2,
} from 'lucide-react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs';

type PermissionApi = {
  id?: number;
  resource?: string;
  action?: string;
  description?: string | null;
  createdAt?: string;
  created_at?: string;
};

type RoleApi = {
  id?: number;
  collegeId?: string;
  college_id?: number;
  name?: string;
  description?: string | null;
  permissions?: PermissionApi[];
  isDefault?: boolean;
  is_system_role?: boolean;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
};

type UserApi = {
  id?: number | string;
  name?: string;
  first_name?: string;
  last_name?: string;
  email?: string;
  role?: string;
};

type UserRow = {
  id: number;
  name: string;
  email: string;
  role: string;
};

const normalizePermission = (perm: PermissionApi): Permission => ({
  id: Number(perm.id ?? 0),
  resource: perm.resource || 'unknown',
  action: perm.action || 'unknown',
  description: perm.description || undefined,
  createdAt: perm.createdAt || perm.created_at || new Date().toISOString(),
});

const normalizeRole = (role: RoleApi): Role => ({
  id: Number(role.id ?? 0),
  collegeId: String(role.collegeId ?? role.college_id ?? ''),
  name: role.name || 'Unnamed Role',
  description: role.description || undefined,
  permissions: (role.permissions || []).map(normalizePermission),
  isDefault: Boolean(role.isDefault ?? role.is_system_role),
  createdAt: role.createdAt || role.created_at || new Date().toISOString(),
  updatedAt: role.updatedAt || role.updated_at || new Date().toISOString(),
});

const normalizeUser = (user: UserApi): UserRow | null => {
  const id = Number(user.id ?? 0);
  if (!Number.isFinite(id) || id <= 0) {
    return null;
  }

  const name = user.name || [user.first_name, user.last_name].filter(Boolean).join(' ').trim() || `User ${id}`;

  return {
    id,
    name,
    email: user.email || '',
    role: user.role || 'student',
  };
};

export default function RolesPermissionsPage() {
  const { user } = useAuth();
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [allUsers, setAllUsers] = useState<UserRow[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('roles');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [isRoleDialogOpen, setIsRoleDialogOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);

  const [isPermissionsDialogOpen, setIsPermissionsDialogOpen] = useState(false);
  const [permissionsRole, setPermissionsRole] = useState<Role | null>(null);
  const [selectedPermissionIds, setSelectedPermissionIds] = useState<number[]>([]);
  const [isSavingPermissions, setIsSavingPermissions] = useState(false);

  const [isAssignDialogOpen, setIsAssignDialogOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<UserRow | null>(null);
  const [selectedRoleId, setSelectedRoleId] = useState<string>('');
  const [isAssigningRole, setIsAssigningRole] = useState(false);
  const [userSearch, setUserSearch] = useState('');

  const [roleFormData, setRoleFormData] = useState({
    name: '',
    description: '',
  });

  const fetchData = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      const [rolesData, permissionsData, usersData] = await Promise.all([
        api.get<RoleApi[]>(endpoints.roles.list),
        api.get<PermissionApi[]>(endpoints.permissions.list),
        api.get<UserApi[]>(endpoints.users.list),
      ]);

      const normalizedRoles = (Array.isArray(rolesData) ? rolesData : [])
        .map(normalizeRole)
        .filter((role) => role.id > 0);
      const normalizedPermissions = (Array.isArray(permissionsData) ? permissionsData : [])
        .map(normalizePermission)
        .filter((permission) => permission.id > 0);
      const normalizedUsers = (Array.isArray(usersData) ? usersData : [])
        .map(normalizeUser)
        .filter((value): value is UserRow => Boolean(value));

      setRoles(normalizedRoles);
      setPermissions(normalizedPermissions);
      setAllUsers(normalizedUsers);
    } catch (fetchError) {
      logger.error('Failed to fetch roles/permissions data:', fetchError as Error);
      setError('Failed to load roles and permissions data');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (user?.role === 'admin' || user?.role === 'super_admin') {
      void fetchData();
    }
  }, [user, fetchData]);

  const filteredUsers = useMemo(() => {
    const query = userSearch.trim().toLowerCase();
    if (!query) {
      return allUsers;
    }

    return allUsers.filter((u) =>
      u.name.toLowerCase().includes(query) || u.email.toLowerCase().includes(query)
    );
  }, [allUsers, userSearch]);

  const handleRoleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!roleFormData.name.trim()) {
      setError('Role name is required');
      return;
    }

    try {
      setError(null);
      setSuccess(null);

      if (editingRole) {
        await api.patch(endpoints.roles.update(editingRole.id), {
          name: roleFormData.name.trim(),
          description: roleFormData.description.trim() || null,
        });
        setSuccess('Role updated successfully');
      } else {
        await api.post(endpoints.roles.create, {
          name: roleFormData.name.trim(),
          description: roleFormData.description.trim() || null,
        });
        setSuccess('Role created successfully');
      }

      setIsRoleDialogOpen(false);
      await fetchData();
    } catch (saveError) {
      logger.error('Failed to save role:', saveError as Error);
      setError('Failed to save role');
    }
  };

  const handleDeleteRole = async (id: number) => {
    if (!confirm('Are you sure you want to delete this role? This might affect users assigned to it.')) {
      return;
    }

    try {
      setError(null);
      setSuccess(null);
      await api.delete(endpoints.roles.delete(id));
      setSuccess('Role deleted successfully');
      await fetchData();
    } catch (deleteError) {
      logger.error('Failed to delete role:', deleteError as Error);
      setError('Failed to delete role');
    }
  };

  const openAddRoleDialog = () => {
    setEditingRole(null);
    setRoleFormData({ name: '', description: '' });
    setIsRoleDialogOpen(true);
  };

  const openPermissionsDialog = (role: Role) => {
    setPermissionsRole(role);
    setSelectedPermissionIds((role.permissions || []).map((permission) => permission.id));
    setIsPermissionsDialogOpen(true);
  };

  const togglePermissionSelection = (permissionId: number) => {
    setSelectedPermissionIds((prev) =>
      prev.includes(permissionId)
        ? prev.filter((id) => id !== permissionId)
        : [...prev, permissionId]
    );
  };

  const handleSavePermissions = async () => {
    if (!permissionsRole) {
      return;
    }

    if (selectedPermissionIds.length === 0) {
      setError('Select at least one permission');
      return;
    }

    try {
      setIsSavingPermissions(true);
      setError(null);
      setSuccess(null);

      await api.post(endpoints.roles.assignPermissions(permissionsRole.id), {
        permission_ids: selectedPermissionIds,
      });

      setSuccess('Permissions updated successfully');
      setIsPermissionsDialogOpen(false);
      await fetchData();
    } catch (saveError) {
      logger.error('Failed to assign permissions:', saveError as Error);
      setError('Failed to assign permissions');
    } finally {
      setIsSavingPermissions(false);
    }
  };

  const openAssignRoleDialog = (targetUser: UserRow) => {
    setSelectedUser(targetUser);

    const matchingRole = roles.find(
      (role) => role.name.toLowerCase() === targetUser.role.toLowerCase()
    );
    setSelectedRoleId(matchingRole ? String(matchingRole.id) : '');
    setIsAssignDialogOpen(true);
  };

  const handleAssignRole = async () => {
    if (!selectedUser || !selectedRoleId) {
      setError('Please select a role');
      return;
    }

    try {
      setIsAssigningRole(true);
      setError(null);
      setSuccess(null);

      const roleId = Number.parseInt(selectedRoleId, 10);
      if (!roleId || roleId <= 0) {
        throw new Error('Invalid role selected');
      }

      await api.post(endpoints.userRoles.assign, {
        user_id: selectedUser.id,
        role_id: roleId,
      });

      const selectedRole = roles.find((role) => role.id === roleId);
      if (selectedRole) {
        setAllUsers((prev) =>
          prev.map((entry) =>
            entry.id === selectedUser.id
              ? { ...entry, role: selectedRole.name }
              : entry
          )
        );
      }

      setSuccess('Role assigned successfully');
      setIsAssignDialogOpen(false);
    } catch (assignError) {
      logger.error('Failed to assign role:', assignError as Error);
      setError(assignError instanceof Error ? assignError.message : 'Failed to assign role');
    } finally {
      setIsAssigningRole(false);
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto flex min-h-[400px] flex-col items-center justify-center space-y-4 p-6">
        <div className="h-12 w-12 animate-spin rounded-full border-b-2 border-primary" />
        <p className="animate-pulse text-muted-foreground">Loading security configuration...</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto space-y-8 p-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="rounded-2xl bg-primary/10 p-3">
            <Shield className="h-8 w-8 text-primary" />
          </div>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Roles & Permissions</h1>
            <p className="text-muted-foreground">Configure access control levels and system permissions</p>
          </div>
        </div>
        <Button onClick={openAddRoleDialog} className="shadow-lg hover:shadow-primary/20">
          <Plus className="mr-2 h-4 w-4" />
          Create Role
        </Button>
      </div>

      {error && (
        <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-lg bg-green-500/10 p-3 text-sm text-green-700">
          {success}
        </div>
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
        <TabsList className="border bg-background shadow-sm">
          <TabsTrigger value="roles" className="gap-2">
            <Lock className="h-4 w-4" />
            Roles Matrix
          </TabsTrigger>
          <TabsTrigger value="users" className="gap-2">
            <Users className="h-4 w-4" />
            User Assignments
          </TabsTrigger>
          <TabsTrigger value="permissions" className="gap-2">
            <Key className="h-4 w-4" />
            Available Permissions
          </TabsTrigger>
        </TabsList>

        <TabsContent value="roles" className="space-y-6">
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
            {roles.map((role) => (
              <Card key={role.id} className="group border-none bg-background/50 shadow-md backdrop-blur-sm transition-all hover:shadow-lg">
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <CardTitle className="text-xl">{role.name}</CardTitle>
                        {role.isDefault && <Badge variant="secondary" className="text-[10px]">System</Badge>}
                      </div>
                      <CardDescription className="line-clamp-2 min-h-[40px]">{role.description || 'No description provided'}</CardDescription>
                    </div>
                    <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-primary"
                        onClick={() => {
                          setEditingRole(role);
                          setRoleFormData({ name: role.name, description: role.description || '' });
                          setIsRoleDialogOpen(true);
                        }}
                      >
                        <Edit2 className="h-4 w-4" />
                      </Button>
                      {!role.isDefault && (
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => handleDeleteRole(role.id)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                      <span>Enabled Permissions</span>
                      <span>{role.permissions?.length || 0} Total</span>
                    </div>
                    <div className="flex min-h-[60px] flex-wrap gap-1.5">
                      {(role.permissions || []).slice(0, 6).map((perm) => (
                        <Badge key={perm.id} variant="outline" className="border-primary/20 bg-primary/5 text-[10px]">
                          {perm.action}:{perm.resource}
                        </Badge>
                      ))}
                      {(role.permissions?.length || 0) > 6 && (
                        <Badge variant="outline" className="text-[10px]">
                          +{(role.permissions?.length || 0) - 6} more
                        </Badge>
                      )}
                    </div>
                  </div>
                  <Button
                    variant="secondary"
                    className="w-full gap-2 bg-muted/50 text-xs font-bold hover:bg-muted"
                    onClick={() => openPermissionsDialog(role)}
                  >
                    <Settings2 className="h-3 w-3" />
                    Manage Permissions
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="users">
          <Card className="border-none shadow-xl">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>User Role Assignments</CardTitle>
                  <CardDescription>View and manage roles assigned to specific users</CardDescription>
                </div>
                <div className="relative w-72">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder="Filter by name or email..."
                    className="pl-10"
                    value={userSearch}
                    onChange={(e) => setUserSearch(e.target.value)}
                  />
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>Active Roles</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredUsers.map((u) => (
                    <TableRow key={u.id}>
                      <TableCell className="text-sm font-medium">{u.name}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{u.email}</TableCell>
                      <TableCell>
                        <Badge className="text-[10px] font-semibold uppercase tracking-tight">
                          {u.role}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <Button variant="ghost" size="sm" className="gap-2" onClick={() => openAssignRoleDialog(u)}>
                          <UserPlus className="h-3.5 w-3.5" />
                          Modify Role
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="permissions">
          <Card className="border-none shadow-xl">
            <CardHeader>
              <CardTitle>System Capabilities</CardTitle>
              <CardDescription>Reference list of all available system actions and resources</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
                {permissions.map((perm) => (
                  <div key={perm.id} className="space-y-2 rounded-xl border bg-muted/30 p-3 transition-colors hover:bg-muted/50">
                    <div className="flex items-center justify-between">
                      <Badge variant="outline" className="bg-background font-mono text-[10px]">
                        {perm.action}
                      </Badge>
                      <Badge variant="secondary" className="text-[10px] uppercase">
                        {perm.resource}
                      </Badge>
                    </div>
                    <p className="text-xs leading-relaxed text-muted-foreground">{perm.description || 'No specialized description'}</p>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Dialog open={isRoleDialogOpen} onOpenChange={setIsRoleDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>{editingRole ? 'Update Role' : 'Create New Role'}</DialogTitle>
            <DialogDescription>
              Define the role identity. Permissions are assigned after creation.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleRoleSubmit} className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">Role Name</Label>
              <Input
                id="name"
                value={roleFormData.name}
                onChange={(e) => setRoleFormData((prev) => ({ ...prev, name: e.target.value }))}
                placeholder="e.g. Content Manager"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                value={roleFormData.description}
                onChange={(e) => setRoleFormData((prev) => ({ ...prev, description: e.target.value }))}
                placeholder="Briefly describe responsibilities"
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setIsRoleDialogOpen(false)}>Cancel</Button>
              <Button type="submit">{editingRole ? 'Update Role' : 'Create Role'}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={isPermissionsDialogOpen} onOpenChange={setIsPermissionsDialogOpen}>
        <DialogContent className="sm:max-w-[560px]">
          <DialogHeader>
            <DialogTitle>Manage Permissions</DialogTitle>
            <DialogDescription>
              {permissionsRole ? `Assign permissions for ${permissionsRole.name}` : 'Select permissions for this role'}
            </DialogDescription>
          </DialogHeader>
          <div className="max-h-[340px] space-y-3 overflow-auto py-2">
            {permissions.map((permission) => (
              <label
                key={permission.id}
                className="flex cursor-pointer items-start gap-3 rounded-md border p-3 hover:bg-muted/20"
              >
                <input
                  type="checkbox"
                  className="mt-1"
                  checked={selectedPermissionIds.includes(permission.id)}
                  onChange={() => togglePermissionSelection(permission.id)}
                />
                <div>
                  <div className="font-medium">{permission.action}:{permission.resource}</div>
                  <div className="text-xs text-muted-foreground">{permission.description || 'No description'}</div>
                </div>
              </label>
            ))}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsPermissionsDialogOpen(false)}
              disabled={isSavingPermissions}
            >
              Cancel
            </Button>
            <Button onClick={handleSavePermissions} disabled={isSavingPermissions || !permissionsRole}>
              {isSavingPermissions ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Save Permissions
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={isAssignDialogOpen} onOpenChange={setIsAssignDialogOpen}>
        <DialogContent className="sm:max-w-[420px]">
          <DialogHeader>
            <DialogTitle>Modify User Role</DialogTitle>
            <DialogDescription>
              {selectedUser ? `Select a role for ${selectedUser.name}` : 'Assign a role'}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2 py-2">
            <Label htmlFor="assign-role">Role</Label>
            <select
              id="assign-role"
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              value={selectedRoleId}
              onChange={(e) => setSelectedRoleId(e.target.value)}
            >
              <option value="">Select role</option>
              {roles.map((role) => (
                <option key={role.id} value={role.id}>
                  {role.name}
                </option>
              ))}
            </select>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsAssignDialogOpen(false)} disabled={isAssigningRole}>
              Cancel
            </Button>
            <Button onClick={handleAssignRole} disabled={isAssigningRole || !selectedRoleId || !selectedUser}>
              {isAssigningRole ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Assign Role
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
