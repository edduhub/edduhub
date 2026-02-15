"use client";

import { useState, useEffect, useCallback, useMemo } from 'react';
import { useAuth } from '@/lib/auth-context';
import { logger } from '@/lib/logger';
import { api, endpoints } from '@/lib/api-client';
import {
    Briefcase,
    TrendingUp,
    Users,
    Building2,
    Plus,
    Search,
    ExternalLink,
    DollarSign,
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
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';

type PlacementStatus = 'offered' | 'accepted' | 'rejected' | 'on-hold';

type Placement = {
    id: number;
    studentId: number;
    studentName?: string;
    companyName: string;
    jobTitle: string;
    package: number;
    placementDate: string;
    status: PlacementStatus;
};

type PlacementApi = {
    id?: number;
    student_id?: number;
    studentId?: number;
    company_name?: string;
    companyName?: string;
    job_title?: string;
    jobTitle?: string;
    package?: number;
    placement_date?: string;
    placementDate?: string;
    status?: string;
};

type PlacementStats = {
    totalPlacements: number;
    averagePackage: number;
    highestPackage: number;
    placementRate: number;
    totalCompanies: number;
    offeredCount: number;
};

type PlacementStatsApi = {
    total_placements?: number;
    average_package?: number;
    highest_package?: number;
    placement_rate?: number;
    unique_companies?: number;
    offered_count?: number;
};

type CompanyStats = {
    companyName: string;
    totalOffers: number;
    acceptedOffers: number;
    averagePackage: number;
    highestPackage: number;
    lowestPackage: number;
};

type CompanyStatsApi = {
    company_name?: string;
    companyName?: string;
    total_offers?: number;
    totalOffers?: number;
    accepted_offers?: number;
    acceptedOffers?: number;
    average_package?: number;
    averagePackage?: number;
    highest_package?: number;
    highestPackage?: number;
    lowest_package?: number;
    lowestPackage?: number;
};

type CompanyDetail = {
    companyName: string;
    placements: Placement[];
    totalOffers: number;
    acceptedOffers: number;
    averagePackage: number;
    highestPackage: number;
    lowestPackage: number;
};

type CompanyDetailApi = {
    company_name?: string;
    companyName?: string;
    placements?: PlacementApi[];
    total_offers?: number;
    totalOffers?: number;
    accepted_offers?: number;
    acceptedOffers?: number;
    average_package?: number;
    averagePackage?: number;
    highest_package?: number;
    highestPackage?: number;
    lowest_package?: number;
    lowestPackage?: number;
};

type StudentApi = {
    student_id?: number;
    id?: number;
    roll_no?: string;
    rollNo?: string;
    first_name?: string;
    firstName?: string;
    last_name?: string;
    lastName?: string;
};

type StudentOption = {
    id: number;
    label: string;
};

const normalizePlacementStatus = (value?: string): PlacementStatus => {
    const normalized = (value || '').toLowerCase();
    if (normalized === 'accepted' || normalized === 'rejected' || normalized === 'on-hold' || normalized === 'offered') {
        return normalized;
    }
    return 'offered';
};

const formatPlacementStatus = (status: PlacementStatus): string => {
    if (status === 'on-hold') return 'On-Hold';
    return `${status.charAt(0).toUpperCase()}${status.slice(1)}`;
};

const normalizePlacement = (placement: PlacementApi): Placement => ({
    id: placement.id ?? 0,
    studentId: placement.student_id ?? placement.studentId ?? 0,
    companyName: placement.company_name ?? placement.companyName ?? '',
    jobTitle: placement.job_title ?? placement.jobTitle ?? '',
    package: placement.package ?? 0,
    placementDate: placement.placement_date ?? placement.placementDate ?? new Date().toISOString(),
    status: normalizePlacementStatus(placement.status),
});

const normalizeStudentOption = (student: StudentApi): StudentOption => {
    const id = student.id ?? student.student_id ?? 0;
    const first = student.firstName ?? student.first_name ?? '';
    const last = student.lastName ?? student.last_name ?? '';
    const fullName = `${first} ${last}`.trim();
    const rollNo = student.rollNo ?? student.roll_no;

    if (fullName && rollNo) {
        return { id, label: `${fullName} (${rollNo})` };
    }
    if (fullName) {
        return { id, label: fullName };
    }
    if (rollNo) {
        return { id, label: `Roll ${rollNo}` };
    }
    return { id, label: `Student ID ${id}` };
};

export default function PlacementsPage() {
    const { user } = useAuth();
    const [placements, setPlacements] = useState<Placement[]>([]);
    const [stats, setStats] = useState<PlacementStats | null>(null);
    const [companyStats, setCompanyStats] = useState<CompanyStats[]>([]);
    const [selectedCompany, setSelectedCompany] = useState<CompanyDetail | null>(null);
    const [students, setStudents] = useState<StudentOption[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [isCompanyDialogOpen, setIsCompanyDialogOpen] = useState(false);
    const [activeTab, setActiveTab] = useState<'placements' | 'companies'>('placements');
    const [editingPlacement, setEditingPlacement] = useState<Placement | null>(null);

    const [formData, setFormData] = useState({
        studentId: '',
        companyName: '',
        jobTitle: '',
        package: '',
        placementDate: new Date().toISOString().split('T')[0],
        status: 'offered' as PlacementStatus,
    });

    const canViewStats = user?.role === 'admin' || user?.role === 'faculty';
    const canManage = user?.role === 'admin';

    const fetchData = useCallback(async () => {
        if (!user) {
            setIsLoading(false);
            return;
        }

        try {
            setIsLoading(true);
            setError(null);

            const placementsData = await api.get<PlacementApi[]>(endpoints.placements.list);
            const normalizedPlacements = Array.isArray(placementsData)
                ? placementsData.map(normalizePlacement)
                : [];
            setPlacements(normalizedPlacements);

            if (canViewStats) {
                const statsData = await api.get<PlacementStatsApi>(endpoints.placements.stats);
                setStats({
                    totalPlacements: statsData?.total_placements ?? 0,
                    averagePackage: statsData?.average_package ?? 0,
                    highestPackage: statsData?.highest_package ?? 0,
                    placementRate: statsData?.placement_rate ?? 0,
                    totalCompanies: statsData?.unique_companies ?? 0,
                    offeredCount: statsData?.offered_count ?? 0,
                });
            } else {
                setStats(null);
            }

            if (canViewStats) {
                const companyStatsData = await api.get<CompanyStatsApi[]>(endpoints.placements.companyStats);
                setCompanyStats(
                    Array.isArray(companyStatsData)
                        ? companyStatsData.map((cs) => ({
                            companyName: cs.company_name ?? cs.companyName ?? '',
                            totalOffers: cs.total_offers ?? cs.totalOffers ?? 0,
                            acceptedOffers: cs.accepted_offers ?? cs.acceptedOffers ?? 0,
                            averagePackage: cs.average_package ?? cs.averagePackage ?? 0,
                            highestPackage: cs.highest_package ?? cs.highestPackage ?? 0,
                            lowestPackage: cs.lowest_package ?? cs.lowestPackage ?? 0,
                        }))
                        : []
                );
            } else {
                setCompanyStats([]);
            }

            if (canViewStats) {
                const studentData = await api.get<StudentApi[]>(endpoints.students.list);
                setStudents(Array.isArray(studentData) ? studentData.map(normalizeStudentOption) : []);
            } else {
                setStudents([]);
            }
        } catch (fetchError) {
            logger.error('Failed to fetch placement data:', fetchError as Error);
            setError('Failed to load placement data');
        } finally {
            setIsLoading(false);
        }
    }, [canViewStats, user]);

    useEffect(() => {
        void fetchData();
    }, [fetchData]);

    const studentNameById = useMemo(() => {
        const mapping = new Map<number, string>();
        for (const student of students) {
            mapping.set(student.id, student.label);
        }
        return mapping;
    }, [students]);

    const handleSubmit = async (event: React.FormEvent) => {
        event.preventDefault();

        if (!canManage) {
            setError('Only admins can modify placement records');
            return;
        }

        if (!formData.studentId || !formData.companyName || !formData.jobTitle || !formData.package) {
            setError('All required fields must be filled');
            return;
        }

        try {
            setError(null);
            const payload = {
                student_id: Number(formData.studentId),
                company_name: formData.companyName.trim(),
                job_title: formData.jobTitle.trim(),
                package: Number(formData.package),
                placement_date: new Date(formData.placementDate).toISOString(),
                status: formData.status,
            };

            if (editingPlacement) {
                await api.put(endpoints.placements.update(editingPlacement.id), payload);
            } else {
                await api.post(endpoints.placements.create, payload);
            }

            setIsDialogOpen(false);
            setEditingPlacement(null);
            await fetchData();
        } catch (submitError) {
            logger.error('Failed to save placement:', submitError as Error);
            setError('Failed to save placement record');
        }
    };

    const handleDelete = async (id: number) => {
        if (!canManage) {
            setError('Only admins can delete placement records');
            return;
        }

        if (!confirm('Are you sure you want to delete this placement record?')) {
            return;
        }

        try {
            setError(null);
            await api.delete(endpoints.placements.delete(id));
            await fetchData();
        } catch (deleteError) {
            logger.error('Failed to delete placement:', deleteError as Error);
            setError('Failed to delete placement record');
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
            status: 'offered',
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
            status: placement.status,
        });
        setIsDialogOpen(true);
    };

    const handleCompanyClick = async (companyName: string) => {
        try {
            const companyData = await api.get<CompanyDetailApi>(endpoints.placements.company(encodeURIComponent(companyName)));
            setSelectedCompany({
                companyName: companyData.company_name ?? companyData.companyName ?? companyName,
                placements: Array.isArray(companyData.placements) 
                    ? companyData.placements.map(normalizePlacement) 
                    : [],
                totalOffers: companyData.total_offers ?? companyData.totalOffers ?? 0,
                acceptedOffers: companyData.accepted_offers ?? companyData.acceptedOffers ?? 0,
                averagePackage: companyData.average_package ?? companyData.averagePackage ?? 0,
                highestPackage: companyData.highest_package ?? companyData.highestPackage ?? 0,
                lowestPackage: companyData.lowest_package ?? companyData.lowestPackage ?? 0,
            });
            setIsCompanyDialogOpen(true);
        } catch (fetchError) {
            logger.error('Failed to fetch company details:', fetchError as Error);
            setError('Failed to load company details');
        }
    };

    const filteredPlacements = placements.filter((placement) => {
        const studentLabel = studentNameById.get(placement.studentId) || placement.studentName || '';
        const query = searchTerm.toLowerCase();

        return (
            placement.companyName.toLowerCase().includes(query) ||
            placement.jobTitle.toLowerCase().includes(query) ||
            studentLabel.toLowerCase().includes(query)
        );
    });

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
                {canManage && (
                    <Button onClick={openAddDialog} className="shadow-lg hover:shadow-primary/20">
                        <Plus className="w-4 h-4 mr-2" />
                        Add Record
                    </Button>
                )}
            </div>

            {error && (
                <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
                    {error}
                </div>
            )}

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
                            {(stats?.placementRate || 0).toFixed(1)}% Placement Rate
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
                            Hiring Partners
                        </div>
                    </CardContent>
                </Card>

                <Card className="border-none bg-amber-500/5 shadow-sm overflow-hidden relative">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        <Briefcase className="w-16 h-16" />
                    </div>
                    <CardHeader className="pb-2">
                        <CardDescription className="font-medium">Open Offers</CardDescription>
                        <CardTitle className="text-3xl font-bold">{stats?.offeredCount || 0}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center text-xs text-amber-600 font-semibold">
                            <Clock className="w-3 h-3 mr-1" />
                            Awaiting acceptance
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Tab Navigation */}
            {canViewStats && (
                <div className="flex gap-2 border-b border-border/50 pb-1">
                    <Button
                        variant={activeTab === 'placements' ? 'default' : 'ghost'}
                        onClick={() => setActiveTab('placements')}
                        className="rounded-b-none"
                    >
                        <Users className="w-4 h-4 mr-2" />
                        Placement Records
                    </Button>
                    <Button
                        variant={activeTab === 'companies' ? 'default' : 'ghost'}
                        onClick={() => setActiveTab('companies')}
                        className="rounded-b-none"
                    >
                        <Building2 className="w-4 h-4 mr-2" />
                        Company Statistics
                    </Button>
                </div>
            )}

            <Card className="border-none shadow-xl bg-background/50 backdrop-blur-sm">
                <CardHeader>
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                        <div>
                            <CardTitle>{activeTab === 'placements' ? 'Placement Records' : 'Company Statistics'}</CardTitle>
                            <CardDescription>
                                {activeTab === 'placements' 
                                    ? 'A detailed list of all student placement achievements'
                                    : 'Hiring statistics by company'}
                            </CardDescription>
                        </div>
                        {activeTab === 'placements' && (
                            <div className="relative w-full md:w-72">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                                <Input
                                    placeholder="Search by company, role, student..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        )}
                    </div>
                </CardHeader>
                <CardContent>
                    {activeTab === 'placements' ? (
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
                                    {canManage && <TableHead className="text-right">Actions</TableHead>}
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filteredPlacements.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={canManage ? 7 : 6} className="h-32 text-center text-muted-foreground">
                                            No placement records found.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    filteredPlacements.map((placement) => (
                                        <TableRow key={placement.id} className="hover:bg-muted/30 transition-colors">
                                            <TableCell className="font-medium">
                                                <div>
                                                    {placement.studentName || studentNameById.get(placement.studentId) || `Student ID: ${placement.studentId}`}
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
                                                        placement.status === 'accepted' ? 'default' :
                                                            placement.status === 'offered' ? 'secondary' :
                                                                placement.status === 'rejected' ? 'destructive' : 'outline'
                                                    }
                                                    className="font-semibold"
                                                >
                                                    {formatPlacementStatus(placement.status)}
                                                </Badge>
                                            </TableCell>
                                            <TableCell className="text-muted-foreground text-xs">
                                                {new Date(placement.placementDate).toLocaleDateString()}
                                            </TableCell>
                                            {canManage && (
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
                    ) : (
                        <div className="rounded-xl border border-border/50 overflow-hidden">
                            {companyStats.length === 0 ? (
                                <div className="h-32 flex items-center justify-center text-muted-foreground">
                                    No company statistics available.
                                </div>
                            ) : (
                                <Table>
                                    <TableHeader className="bg-muted/50">
                                        <TableRow>
                                            <TableHead>Company</TableHead>
                                            <TableHead>Total Offers</TableHead>
                                            <TableHead>Accepted</TableHead>
                                            <TableHead>Avg Package</TableHead>
                                            <TableHead>Highest</TableHead>
                                            <TableHead>Lowest</TableHead>
                                            <TableHead className="text-right">Action</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {companyStats.map((company) => (
                                            <TableRow key={company.companyName} className="hover:bg-muted/30 transition-colors cursor-pointer" onClick={() => handleCompanyClick(company.companyName)}>
                                                <TableCell className="font-medium">
                                                    <div className="flex items-center gap-2">
                                                        <Building2 className="w-4 h-4 text-muted-foreground" />
                                                        {company.companyName}
                                                    </div>
                                                </TableCell>
                                                <TableCell className="font-semibold">{company.totalOffers}</TableCell>
                                                <TableCell>
                                                    <Badge variant="default" className="font-semibold">
                                                        {company.acceptedOffers}
                                                    </Badge>
                                                </TableCell>
                                                <TableCell className="font-mono text-xs font-semibold">
                                                    ₹{((company.averagePackage) / 100000).toFixed(1)} LPA
                                                </TableCell>
                                                <TableCell className="font-mono text-xs text-green-600 font-semibold">
                                                    ₹{((company.highestPackage) / 100000).toFixed(1)} LPA
                                                </TableCell>
                                                <TableCell className="font-mono text-xs text-muted-foreground">
                                                    ₹{((company.lowestPackage) / 100000).toFixed(1)} LPA
                                                </TableCell>
                                                <TableCell className="text-right">
                                                    <Button variant="ghost" size="sm" className="text-primary">
                                                        View Details
                                                    </Button>
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            )}
                        </div>
                    )}
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
                                onValueChange={(value) => setFormData((prev) => ({ ...prev, studentId: value }))}
                            >
                                <SelectTrigger id="studentId">
                                    <SelectValue placeholder="Select student" />
                                </SelectTrigger>
                                <SelectContent>
                                    {students.map((student) => (
                                        <SelectItem key={student.id} value={student.id.toString()}>
                                            {student.label}
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
                                    onChange={(e) => setFormData((prev) => ({ ...prev, companyName: e.target.value }))}
                                    placeholder="e.g. Google, Microsoft"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="jobTitle">Job Title</Label>
                                <Input
                                    id="jobTitle"
                                    value={formData.jobTitle}
                                    onChange={(e) => setFormData((prev) => ({ ...prev, jobTitle: e.target.value }))}
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
                                    onChange={(e) => setFormData((prev) => ({ ...prev, package: e.target.value }))}
                                    placeholder="e.g. 1200000"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="status">Status</Label>
                                <Select
                                    value={formData.status}
                                    onValueChange={(value) => setFormData((prev) => ({ ...prev, status: value as PlacementStatus }))}
                                >
                                    <SelectTrigger id="status">
                                        <SelectValue placeholder="Select status" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="offered">Offered</SelectItem>
                                        <SelectItem value="accepted">Accepted</SelectItem>
                                        <SelectItem value="rejected">Rejected</SelectItem>
                                        <SelectItem value="on-hold">On-Hold</SelectItem>
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
                                onChange={(e) => setFormData((prev) => ({ ...prev, placementDate: e.target.value }))}
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

            <Dialog open={isCompanyDialogOpen} onOpenChange={setIsCompanyDialogOpen}>
                <DialogContent className="sm:max-w-[700px] max-h-[80vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <Building2 className="w-5 h-5" />
                            {selectedCompany?.companyName}
                        </DialogTitle>
                        <DialogDescription>
                            Detailed placement statistics and records for this company
                        </DialogDescription>
                    </DialogHeader>
                    {selectedCompany && (
                        <div className="space-y-4">
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                <div className="p-3 bg-primary/5 rounded-lg">
                                    <div className="text-xs text-muted-foreground">Total Offers</div>
                                    <div className="text-2xl font-bold">{selectedCompany.totalOffers}</div>
                                </div>
                                <div className="p-3 bg-green-500/5 rounded-lg">
                                    <div className="text-xs text-muted-foreground">Accepted</div>
                                    <div className="text-2xl font-bold text-green-600">{selectedCompany.acceptedOffers}</div>
                                </div>
                                <div className="p-3 bg-blue-500/5 rounded-lg">
                                    <div className="text-xs text-muted-foreground">Avg Package</div>
                                    <div className="text-2xl font-bold text-blue-600">₹{((selectedCompany.averagePackage) / 100000).toFixed(1)} LPA</div>
                                </div>
                                <div className="p-3 bg-amber-500/5 rounded-lg">
                                    <div className="text-xs text-muted-foreground">Highest</div>
                                    <div className="text-2xl font-bold text-amber-600">₹{((selectedCompany.highestPackage) / 100000).toFixed(1)} LPA</div>
                                </div>
                            </div>

                            <div className="rounded-lg border border-border/50 overflow-hidden">
                                <Table>
                                    <TableHeader className="bg-muted/50">
                                        <TableRow>
                                            <TableHead>Student</TableHead>
                                            <TableHead>Role</TableHead>
                                            <TableHead>Package</TableHead>
                                            <TableHead>Status</TableHead>
                                            <TableHead>Date</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {selectedCompany.placements.length === 0 ? (
                                            <TableRow>
                                                <TableCell colSpan={5} className="h-32 text-center text-muted-foreground">
                                                    No placement records found.
                                                </TableCell>
                                            </TableRow>
                                        ) : (
                                            selectedCompany.placements.map((placement) => (
                                                <TableRow key={placement.id}>
                                                    <TableCell className="font-medium">
                                                        {placement.studentName || studentNameById.get(placement.studentId) || `Student ID: ${placement.studentId}`}
                                                    </TableCell>
                                                    <TableCell>{placement.jobTitle}</TableCell>
                                                    <TableCell className="font-mono text-xs font-semibold">
                                                        ₹{placement.package.toLocaleString()}
                                                    </TableCell>
                                                    <TableCell>
                                                        <Badge
                                                            variant={
                                                                placement.status === 'accepted' ? 'default' :
                                                                    placement.status === 'offered' ? 'secondary' :
                                                                        placement.status === 'rejected' ? 'destructive' : 'outline'
                                                            }
                                                            className="font-semibold"
                                                        >
                                                            {formatPlacementStatus(placement.status)}
                                                        </Badge>
                                                    </TableCell>
                                                    <TableCell className="text-muted-foreground text-xs">
                                                        {new Date(placement.placementDate).toLocaleDateString()}
                                                    </TableCell>
                                                </TableRow>
                                            ))
                                        )}
                                    </TableBody>
                                </Table>
                            </div>
                        </div>
                    )}
                </DialogContent>
            </Dialog>
        </div>
    );
}
