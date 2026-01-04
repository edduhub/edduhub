"use client";

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/lib/auth-context';
import { api, endpoints } from '@/lib/api-client';
import type { Student } from '@/lib/types';
import {
    Briefcase,
    TrendingUp,
    Users,
    Building2,
    Plus,
    Search,
    Filter,
    MoreVertical,
    ExternalLink,
    DollarSign,
    Calendar,
    CheckCircle2,
    Clock,
    ArrowUpRight,
    Edit2,
    Trash2
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
    DialogTrigger,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';

// Define the Placement type locally if not in types.ts
type Placement = {
    id: number;
    studentId: number;
    studentName?: string;
    companyName: string;
    jobTitle: string;
    package: number;
    placementDate: string;
    status: 'Offered' | 'Accepted' | 'Rejected' | 'On-Hold';
    student?: Student;
};

type PlacementStats = {
    totalPlacements: number;
    averagePackage: number;
    highestPackage: number;
    placementRate: number;
    totalCompanies: number;
};

export default function PlacementsPage() {
    const { user } = useAuth();
    const [placements, setPlacements] = useState<Placement[]>([]);
    const [stats, setStats] = useState<PlacementStats | null>(null);
    const [students, setStudents] = useState<Student[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingPlacement, setEditingPlacement] = useState<Placement | null>(null);

    // Form state
    const [formData, setFormData] = useState({
        studentId: '',
        companyName: '',
        jobTitle: '',
        package: '',
        placementDate: new Date().toISOString().split('T')[0],
        status: 'Offered' as Placement['status']
    });

    const fetchData = useCallback(async () => {
        try {
            setIsLoading(true);
            const [placementsData, statsData] = await Promise.all([
                api.get<Placement[]>(endpoints.placements.list),
                api.get<PlacementStats>(endpoints.placements.stats)
            ]);
            setPlacements(placementsData || []);
            setStats(statsData);

            if (user?.role === 'admin' || user?.role === 'faculty') {
                const studentData = await api.get<Student[]>(endpoints.students.list);
                setStudents(studentData || []);
            }
        } catch (error) {
            console.error('Failed to fetch placement data:', error);
        } finally {
            setIsLoading(false);
        }
    }, [user]);

    useEffect(() => {
        if (user) {
            fetchData();
        }
    }, [user, fetchData]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            const payload = {
                ...formData,
                studentId: parseInt(formData.studentId),
                package: parseFloat(formData.package),
                placementDate: new Date(formData.placementDate).toISOString()
            };

            if (editingPlacement) {
                await api.put(endpoints.placements.update(editingPlacement.id), payload);
            } else {
                await api.post(endpoints.placements.create, payload);
            }

            setIsDialogOpen(false);
            setEditingPlacement(null);
            fetchData();
        } catch (error) {
            console.error('Failed to save placement:', error);
        }
    };

    const handleDelete = async (id: number) => {
        if (confirm('Are you sure you want to delete this placement record?')) {
            try {
                await api.delete(endpoints.placements.delete(id));
                fetchData();
            } catch (error) {
                console.error('Failed to delete placement:', error);
            }
        }
    };

    const openAddDialog = () => {
        setEditingPlacement(null);
        setFormData({
            studentId: '',
            companyName: '',
            jobTitle: '',
            package: '',
            placementDate: new Date().toISOString().split('T')[0],
            status: 'Offered'
        });
        setIsDialogOpen(true);
    };

    const openEditDialog = (placement: Placement) => {
        setEditingPlacement(placement);
        setFormData({
            studentId: placement.studentId.toString(),
            companyName: placement.companyName,
            jobTitle: placement.jobTitle,
            package: placement.package.toString(),
            placementDate: placement.placementDate.split('T')[0],
            status: placement.status
        });
        setIsDialogOpen(true);
    };

    const filteredPlacements = placements.filter(p =>
        p.companyName.toLowerCase().includes(searchTerm.toLowerCase()) ||
        p.jobTitle.toLowerCase().includes(searchTerm.toLowerCase()) ||
        p.studentName?.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const isAdmin = user?.role === 'admin' || user?.role === 'faculty';

    if (isLoading) {
        return (
            <div className="container mx-auto p-6 flex flex-col items-center justify-center min-h-[400px] space-y-4">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary" />
                <p className="text-muted-foreground animate-pulse">Loading placement data...</p>
            </div>
        );
    }

    return (
        <div className="container mx-auto p-6 space-y-8">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-primary/10 rounded-2xl">
                        <Briefcase className="w-8 h-8 text-primary" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">Placements</h1>
                        <p className="text-muted-foreground">Track and manage student career opportunities and successes</p>
                    </div>
                </div>
                {isAdmin && (
                    <Button onClick={openAddDialog} className="shadow-lg hover:shadow-primary/20">
                        <Plus className="w-4 h-4 mr-2" />
                        Add Record
                    </Button>
                )}
            </div>

            {/* Stats Overview */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <Card className="border-none bg-primary/5 shadow-sm overflow-hidden relative">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        <Users className="w-16 h-16" />
                    </div>
                    <CardHeader className="pb-2">
                        <CardDescription className="font-medium">Total Placed</CardDescription>
                        <CardTitle className="text-3xl font-bold">{stats?.totalPlacements || 0}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center text-xs text-primary font-semibold">
                            <TrendingUp className="w-3 h-3 mr-1" />
                            {stats?.placementRate || 0}% Placement Rate
                        </div>
                    </CardContent>
                </Card>

                <Card className="border-none bg-green-500/5 shadow-sm overflow-hidden relative">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        <DollarSign className="w-16 h-16" />
                    </div>
                    <CardHeader className="pb-2">
                        <CardDescription className="font-medium">Average Package</CardDescription>
                        <CardTitle className="text-3xl font-bold">₹{((stats?.averagePackage || 0) / 100000).toFixed(1)} LPA</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center text-xs text-green-600 font-semibold">
                            <ArrowUpRight className="w-3 h-3 mr-1" />
                            Top ₹{((stats?.highestPackage || 0) / 100000).toFixed(1)} LPA
                        </div>
                    </CardContent>
                </Card>

                <Card className="border-none bg-blue-500/5 shadow-sm overflow-hidden relative">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        <Building2 className="w-16 h-16" />
                    </div>
                    <CardHeader className="pb-2">
                        <CardDescription className="font-medium">Companies Participated</CardDescription>
                        <CardTitle className="text-3xl font-bold">{stats?.totalCompanies || 0}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center text-xs text-blue-600 font-semibold">
                            <ExternalLink className="w-3 h-3 mr-1" />
                            View Company Partners
                        </div>
                    </CardContent>
                </Card>

                <Card className="border-none bg-amber-500/5 shadow-sm overflow-hidden relative">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        <Briefcase className="w-16 h-16" />
                    </div>
                    <CardHeader className="pb-2">
                        <CardDescription className="font-medium">Active Drives</CardDescription>
                        <CardTitle className="text-3xl font-bold">12</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center text-xs text-amber-600 font-semibold">
                            <Clock className="w-3 h-3 mr-1" />
                            4 Starting this week
                        </div>
                    </CardContent>
                </Card>
            </div>

            <Card className="border-none shadow-xl bg-background/50 backdrop-blur-sm">
                <CardHeader>
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                        <div>
                            <CardTitle>Placement Records</CardTitle>
                            <CardDescription>A detailed list of all student placement achievements</CardDescription>
                        </div>
                        <div className="relative w-full md:w-72">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                            <Input
                                placeholder="Search by company, role, student..."
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                                className="pl-10"
                            />
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="rounded-xl border border-border/50 overflow-hidden">
                        <Table>
                            <TableHeader className="bg-muted/50">
                                <TableRow>
                                    <TableHead>Student</TableHead>
                                    <TableHead>Company</TableHead>
                                    <TableHead>Role</TableHead>
                                    <TableHead>Package</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Date</TableHead>
                                    {isAdmin && <TableHead className="text-right">Actions</TableHead>}
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filteredPlacements.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={isAdmin ? 7 : 6} className="h-32 text-center text-muted-foreground">
                                            No placement records found.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    filteredPlacements.map((placement) => (
                                        <TableRow key={placement.id} className="hover:bg-muted/30 transition-colors">
                                            <TableCell className="font-medium">
                                                <div>
                                                    {placement.studentName || `Student ID: ${placement.studentId}`}
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    <Building2 className="w-4 h-4 text-muted-foreground" />
                                                    {placement.companyName}
                                                </div>
                                            </TableCell>
                                            <TableCell>{placement.jobTitle}</TableCell>
                                            <TableCell className="font-mono text-xs font-semibold">
                                                ₹{placement.package.toLocaleString()}
                                            </TableCell>
                                            <TableCell>
                                                <Badge
                                                    variant={
                                                        placement.status === 'Accepted' ? 'default' :
                                                            placement.status === 'Offered' ? 'secondary' :
                                                                placement.status === 'Rejected' ? 'destructive' : 'outline'
                                                    }
                                                    className="font-semibold"
                                                >
                                                    {placement.status}
                                                </Badge>
                                            </TableCell>
                                            <TableCell className="text-muted-foreground text-xs">
                                                {new Date(placement.placementDate).toLocaleDateString()}
                                            </TableCell>
                                            {isAdmin && (
                                                <TableCell className="text-right">
                                                    <div className="flex justify-end gap-1">
                                                        <Button variant="ghost" size="icon" className="h-8 w-8 text-primary" onClick={() => openEditDialog(placement)}>
                                                            <Edit2 className="w-3.5 h-3.5" />
                                                        </Button>
                                                        <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => handleDelete(placement.id)}>
                                                            <Trash2 className="w-3.5 h-3.5" />
                                                        </Button>
                                                    </div>
                                                </TableCell>
                                            )}
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </div>
                </CardContent>
            </Card>

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle>{editingPlacement ? 'Edit Placement Record' : 'Add New Placement'}</DialogTitle>
                        <DialogDescription>
                            Record a new student placement success here.
                        </DialogDescription>
                    </DialogHeader>
                    <form onSubmit={handleSubmit} className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="studentId">Student</Label>
                            <Select
                                value={formData.studentId}
                                onValueChange={(v) => setFormData(prev => ({ ...prev, studentId: v }))}
                            >
                                <SelectTrigger id="studentId">
                                    <SelectValue placeholder="Select student" />
                                </SelectTrigger>
                                <SelectContent>
                                    {students.map(student => (
                                        <SelectItem key={student.id} value={student.id.toString()}>
                                            {student.firstName} {student.lastName} ({student.rollNo})
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="companyName">Company Name</Label>
                                <Input
                                    id="companyName"
                                    value={formData.companyName}
                                    onChange={(e) => setFormData(prev => ({ ...prev, companyName: e.target.value }))}
                                    placeholder="e.g. Google, Microsoft"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="jobTitle">Job Title</Label>
                                <Input
                                    id="jobTitle"
                                    value={formData.jobTitle}
                                    onChange={(e) => setFormData(prev => ({ ...prev, jobTitle: e.target.value }))}
                                    placeholder="e.g. SDE-1"
                                />
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="package">Package (Annual ₹)</Label>
                                <Input
                                    id="package"
                                    type="number"
                                    value={formData.package}
                                    onChange={(e) => setFormData(prev => ({ ...prev, package: e.target.value }))}
                                    placeholder="e.g. 1200000"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="status">Status</Label>
                                <Select
                                    value={formData.status}
                                    onValueChange={(v: any) => setFormData(prev => ({ ...prev, status: v }))}
                                >
                                    <SelectTrigger id="status">
                                        <SelectValue placeholder="Select status" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="Offered">Offered</SelectItem>
                                        <SelectItem value="Accepted">Accepted</SelectItem>
                                        <SelectItem value="Rejected">Rejected</SelectItem>
                                        <SelectItem value="On-Hold">On-Hold</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="placementDate">Placement Date</Label>
                            <Input
                                id="placementDate"
                                type="date"
                                value={formData.placementDate}
                                onChange={(e) => setFormData(prev => ({ ...prev, placementDate: e.target.value }))}
                            />
                        </div>

                        <DialogFooter>
                            <Button type="button" variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                            <Button type="submit">
                                {editingPlacement ? 'Update' : 'Save Record'}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>
        </div>
    );
}
