"use client";

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/lib/auth-context';
import { api, endpoints } from '@/lib/api-client';
import { logger } from '@/lib/logger';
import type { Role, Permission, User } from '@/lib/types';
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
    Users
} from 'lucide-react';
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
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

export default function RolesPermissionsPage() {
    const { user } = useAuth();
    const [roles, setRoles] = useState<Role[]>([]);
    const [permissions, setPermissions] = useState<Permission[]>([]);
    const [allUsers, setAllUsers] = useState<User[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [activeTab, setActiveTab] = useState('roles');

    const [isRoleDialogOpen, setIsRoleDialogOpen] = useState(false);
    const [editingRole, setEditingRole] = useState<Role | null>(null);

    // Form state for role
    const [roleFormData, setRoleFormData] = useState({
        name: '',
        description: '',
    });

    const fetchData = useCallback(async () => {
        try {
            setIsLoading(true);
            const [rolesData, permissionsData, usersData] = await Promise.all([
                api.get<Role[]>(endpoints.roles.list),
                api.get<Permission[]>(endpoints.permissions.list),
                api.get<User[]>(endpoints.users.list)
            ]);
            setRoles(rolesData || []);
            setPermissions(permissionsData || []);
            setAllUsers(usersData || []);
        } catch (error) {
            logger.error('Failed to fetch roles/permissions data:', error as Error);
        } finally {
            setIsLoading(false);
        }
    }, []);

    useEffect(() => {
        if (user?.role === 'admin' || user?.role === 'super_admin') {
            fetchData();
        }
    }, [user, fetchData]);

    const handleRoleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            if (editingRole) {
                await api.patch(endpoints.roles.update(editingRole.id), roleFormData);
            } else {
                await api.post(endpoints.roles.create, roleFormData);
            }
            setIsRoleDialogOpen(false);
            fetchData();
        } catch (error) {
            logger.error('Failed to save role:', error as Error);
        }
    };

    const handleDeleteRole = async (id: number) => {
        if (confirm('Are you sure you want to delete this role? This might affect users assigned to it.')) {
            try {
                await api.delete(endpoints.roles.delete(id));
                fetchData();
            } catch (error) {
                logger.error('Failed to delete role:', error as Error);
            }
        }
    };

    const openAddRoleDialog = () => {
        setEditingRole(null);
        setRoleFormData({ name: '', description: '' });
        setIsRoleDialogOpen(true);
    };

    if (isLoading) {
        return (
            <div className="container mx-auto p-6 flex flex-col items-center justify-center min-h-[400px] space-y-4">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary" />
                <p className="text-muted-foreground animate-pulse">Loading security configuration...</p>
            </div>
        );
    }

    return (
        <div className="container mx-auto p-6 space-y-8">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-primary/10 rounded-2xl">
                        <Shield className="w-8 h-8 text-primary" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">Roles & Permissions</h1>
                        <p className="text-muted-foreground">Configure access control levels and system permissions</p>
                    </div>
                </div>
                <Button onClick={openAddRoleDialog} className="shadow-lg hover:shadow-primary/20">
                    <Plus className="w-4 h-4 mr-2" />
                    Create Role
                </Button>
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
                <TabsList className="bg-background border shadow-sm">
                    <TabsTrigger value="roles" className="gap-2">
                        <Lock className="w-4 h-4" />
                        Roles Matrix
                    </TabsTrigger>
                    <TabsTrigger value="users" className="gap-2">
                        <Users className="w-4 h-4" />
                        User Assignments
                    </TabsTrigger>
                    <TabsTrigger value="permissions" className="gap-2">
                        <Key className="w-4 h-4" />
                        Available Permissions
                    </TabsTrigger>
                </TabsList>

                <TabsContent value="roles" className="space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {roles.map((role) => (
                            <Card key={role.id} className="border-none shadow-md bg-background/50 backdrop-blur-sm group hover:shadow-lg transition-all">
                                <CardHeader>
                                    <div className="flex items-start justify-between">
                                        <div className="space-y-1">
                                            <div className="flex items-center gap-2">
                                                <CardTitle className="text-xl">{role.name}</CardTitle>
                                                {role.isDefault && <Badge variant="secondary" className="text-[10px]">Default</Badge>}
                                            </div>
                                            <CardDescription className="line-clamp-2 min-h-[40px]">{role.description || 'No description provided'}</CardDescription>
                                        </div>
                                        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <Button variant="ghost" size="icon" className="h-8 w-8 text-primary" onClick={() => {
                                                setEditingRole(role);
                                                setRoleFormData({ name: role.name, description: role.description || '' });
                                                setIsRoleDialogOpen(true);
                                            }}>
                                                <Edit2 className="w-4 h-4" />
                                            </Button>
                                            {!role.isDefault && (
                                                <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => handleDeleteRole(role.id)}>
                                                    <Trash2 className="w-4 h-4" />
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
                                        <div className="flex flex-wrap gap-1.5 min-h-[60px]">
                                            {(role.permissions || []).slice(0, 6).map((perm) => (
                                                <Badge key={perm.id} variant="outline" className="bg-primary/5 text-[10px] border-primary/20">
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
                                    <Button variant="secondary" className="w-full text-xs font-bold gap-2 bg-muted/50 hover:bg-muted">
                                        <Settings2 className="w-3 h-3" />
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
                                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                                    <Input placeholder="Filter by name or email..." className="pl-10" />
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
                                    {allUsers.map((u) => (
                                        <TableRow key={u.id}>
                                            <TableCell className="font-medium text-sm">{u.firstName} {u.lastName}</TableCell>
                                            <TableCell className="text-sm text-muted-foreground">{u.email}</TableCell>
                                            <TableCell>
                                                <Badge className="font-semibold uppercase tracking-tight text-[10px]">
                                                    {u.role}
                                                </Badge>
                                            </TableCell>
                                            <TableCell className="text-right">
                                                <Button variant="ghost" size="sm" className="gap-2">
                                                    <UserPlus className="w-3.5 h-3.5" />
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
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                                {permissions.map((perm) => (
                                    <div key={perm.id} className="p-3 rounded-xl border bg-muted/30 hover:bg-muted/50 transition-colors space-y-2">
                                        <div className="flex items-center justify-between">
                                            <Badge variant="outline" className="font-mono text-[10px] bg-background">
                                                {perm.action}
                                            </Badge>
                                            <Badge variant="secondary" className="text-[10px] uppercase">
                                                {perm.resource}
                                            </Badge>
                                        </div>
                                        <p className="text-xs text-muted-foreground leading-relaxed">{perm.description || 'No specialized description'}</p>
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
                                onChange={(e) => setRoleFormData(prev => ({ ...prev, name: e.target.value }))}
                                placeholder="e.g. Content Manager"
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="description">Description</Label>
                            <Input
                                id="description"
                                value={roleFormData.description}
                                onChange={(e) => setRoleFormData(prev => ({ ...prev, description: e.target.value }))}
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
        </div>
    );
}
