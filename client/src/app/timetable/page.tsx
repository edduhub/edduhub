"use client";

import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/lib/auth-context';
import { logger } from '@/lib/logger';
import { api, endpoints } from '@/lib/api-client';
import type { TimetableBlock, Course } from '@/lib/types';
import {
    Calendar,
    Clock,
    MapPin,
    User as UserIcon,
    Plus,
    Edit2,
    Trash2
} from 'lucide-react';
import {
    Card,
    CardContent,
    CardHeader,
    CardTitle
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

type TimetableBlockApi = {
    id?: number;
    college_id?: number;
    course_id?: number;
    day_of_week?: number;
    start_time?: unknown;
    end_time?: unknown;
    room_number?: string;
    faculty_id?: string;
    created_at?: string;
    updated_at?: string;
    // defensive support for existing camelCase payloads
    courseId?: number;
    dayOfWeek?: number;
    startTime?: unknown;
    endTime?: unknown;
    room?: string;
    facultyId?: string;
};

type CourseApi = {
    id?: number;
    code?: string;
    name?: string;
    // defensive support for snake_case variants
    course_id?: number;
    course_code?: string;
    course_name?: string;
};

const DAYS = [
    { id: 1, name: 'Monday' },
    { id: 2, name: 'Tuesday' },
    { id: 3, name: 'Wednesday' },
    { id: 4, name: 'Thursday' },
    { id: 5, name: 'Friday' },
    { id: 6, name: 'Saturday' },
    { id: 0, name: 'Sunday' },
];

const toInputTime = (value: string): string => {
    if (!value) return '';
    if (value.length >= 5) return value.slice(0, 5);
    return value;
};

const toApiTime = (value: string): string => {
    if (!value) return '00:00:00';
    if (value.length === 5) return `${value}:00`;
    return value;
};

const normalizePgTime = (value: unknown): string => {
    if (typeof value === 'string') {
        return toInputTime(value);
    }

    if (value && typeof value === 'object') {
        const record = value as Record<string, unknown>;

        const textValue = record.String ?? record.string ?? record.Time ?? record.time;
        if (typeof textValue === 'string' && textValue.trim().length > 0) {
            return toInputTime(textValue);
        }

        const micros = record.Microseconds ?? record.microseconds;
        if (typeof micros === 'number' && Number.isFinite(micros)) {
            const totalSeconds = Math.floor(micros / 1_000_000);
            const hours = Math.floor(totalSeconds / 3600) % 24;
            const minutes = Math.floor((totalSeconds % 3600) / 60);
            const hh = String(hours).padStart(2, '0');
            const mm = String(minutes).padStart(2, '0');
            return `${hh}:${mm}`;
        }
    }

    return '';
};

const normalizeCourse = (item: CourseApi): Course => {
    const id = item.id ?? item.course_id ?? 0;
    return {
        id,
        code: item.code ?? item.course_code ?? `CRS-${id}`,
        name: item.name ?? item.course_name ?? `Course ${id}`,
        credits: 0,
        semester: '',
        departmentId: 0,
        collegeId: '',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
    };
};

const normalizeBlock = (item: TimetableBlockApi, coursesById: Map<number, Course>): TimetableBlock => {
    const courseId = item.courseId ?? item.course_id ?? 0;
    const matchedCourse = coursesById.get(courseId);

    return {
        id: item.id ?? 0,
        collegeId: String(item.college_id ?? ''),
        courseId,
        courseName: matchedCourse?.name ?? `Course ${courseId}`,
        courseCode: matchedCourse?.code ?? `CRS-${courseId}`,
        dayOfWeek: item.dayOfWeek ?? item.day_of_week ?? 0,
        startTime: normalizePgTime(item.startTime ?? item.start_time),
        endTime: normalizePgTime(item.endTime ?? item.end_time),
        room: item.room ?? item.room_number,
        instructorId: item.facultyId ?? item.faculty_id,
        instructorName: item.faculty_id || 'Not Assigned',
        createdAt: item.created_at ?? new Date().toISOString(),
        updatedAt: item.updated_at ?? new Date().toISOString(),
    };
};

export default function TimetablePage() {
    const { user } = useAuth();
    const [blocks, setBlocks] = useState<TimetableBlock[]>([]);
    const [courses, setCourses] = useState<Course[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingBlock, setEditingBlock] = useState<TimetableBlock | null>(null);

    const [formData, setFormData] = useState({
        courseId: '',
        dayOfWeek: '1',
        startTime: '09:00',
        endTime: '10:00',
        room: '',
    });

    const fetchData = useCallback(async () => {
        if (!user) {
            setIsLoading(false);
            return;
        }

        try {
            setIsLoading(true);
            setError(null);

            const timetableEndpoint = user.role === 'student'
                ? endpoints.timetable.myTimetable
                : endpoints.timetable.list;

            const [rawBlocks, rawCourses] = await Promise.all([
                api.get<TimetableBlockApi[]>(timetableEndpoint),
                api.get<CourseApi[]>(endpoints.courses.list).catch(() => []),
            ]);

            const normalizedCourses = Array.isArray(rawCourses)
                ? rawCourses.map(normalizeCourse)
                : [];
            const coursesById = new Map(normalizedCourses.map((course) => [course.id, course]));

            const normalizedBlocks = Array.isArray(rawBlocks)
                ? rawBlocks.map((item) => normalizeBlock(item, coursesById))
                : [];

            setCourses(normalizedCourses);
            setBlocks(normalizedBlocks);
        } catch (fetchError) {
            logger.error('Failed to fetch timetable data:', fetchError as Error);
            setError('Failed to load timetable data');
        } finally {
            setIsLoading(false);
        }
    }, [user]);

    useEffect(() => {
        void fetchData();
    }, [fetchData]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!formData.courseId) {
            setError('Please select a course');
            return;
        }

        try {
            setError(null);

            const payload = {
                course_id: Number(formData.courseId),
                day_of_week: Number(formData.dayOfWeek),
                start_time: toApiTime(formData.startTime),
                end_time: toApiTime(formData.endTime),
                room_number: formData.room || null,
            };

            if (editingBlock) {
                await api.patch(endpoints.timetable.update(editingBlock.id), payload);
            } else {
                await api.post(endpoints.timetable.create, payload);
            }

            setIsDialogOpen(false);
            setEditingBlock(null);
            await fetchData();
        } catch (submitError) {
            logger.error('Failed to save timetable block:', submitError as Error);
            setError('Failed to save timetable block');
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm('Are you sure you want to delete this class?')) {
            return;
        }

        try {
            setError(null);
            await api.delete(endpoints.timetable.delete(id));
            await fetchData();
        } catch (deleteError) {
            logger.error('Failed to delete block:', deleteError as Error);
            setError('Failed to delete timetable block');
        }
    };

    const openAddDialog = () => {
        setEditingBlock(null);
        setFormData({
            courseId: '',
            dayOfWeek: '1',
            startTime: '09:00',
            endTime: '10:00',
            room: '',
        });
        setIsDialogOpen(true);
    };

    const openEditDialog = (block: TimetableBlock) => {
        setEditingBlock(block);
        setFormData({
            courseId: block.courseId.toString(),
            dayOfWeek: block.dayOfWeek.toString(),
            startTime: toInputTime(block.startTime),
            endTime: toInputTime(block.endTime),
            room: block.room || '',
        });
        setIsDialogOpen(true);
    };

    if (isLoading) {
        return (
            <div className="container mx-auto p-6 flex items-center justify-center min-h-[400px]">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary" />
            </div>
        );
    }

    const canManage = user?.role === 'admin' || user?.role === 'faculty';

    return (
        <div className="container mx-auto p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10 rounded-xl">
                        <Calendar className="w-8 h-8 text-primary" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">Timetable</h1>
                        <p className="text-muted-foreground">Manage and view your weekly academic schedule</p>
                    </div>
                </div>
                {canManage && (
                    <Button onClick={openAddDialog} className="shadow-lg hover:shadow-primary/20">
                        <Plus className="w-4 h-4 mr-2" />
                        Add Class
                    </Button>
                )}
            </div>

            {error && (
                <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
                    {error}
                </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-7 gap-4">
                {DAYS.filter((day) => day.id !== 0).map((day) => (
                    <Card key={day.id} className="border-none shadow-sm bg-muted/30">
                        <CardHeader className="p-3 text-center border-b bg-background/50 rounded-t-xl">
                            <CardTitle className="text-sm font-bold uppercase tracking-wider">{day.name}</CardTitle>
                        </CardHeader>
                        <CardContent className="p-2 space-y-3">
                            {blocks.filter((block) => block.dayOfWeek === day.id).length === 0 ? (
                                <div className="py-8 text-center text-xs text-muted-foreground italic">No classes</div>
                            ) : (
                                blocks
                                    .filter((block) => block.dayOfWeek === day.id)
                                    .sort((left, right) => left.startTime.localeCompare(right.startTime))
                                    .map((block) => (
                                        <div
                                            key={block.id}
                                            className="group p-3 rounded-lg border bg-background shadow-sm hover:shadow-md transition-all relative overflow-hidden border-l-4 border-l-primary"
                                        >
                                            <div className="space-y-2">
                                                <div className="flex items-start justify-between">
                                                    <Badge variant="outline" className="text-[10px] font-mono px-1.5 h-5 flex items-center gap-1">
                                                        <Clock className="w-3 h-3" />
                                                        {toInputTime(block.startTime)}
                                                    </Badge>
                                                    {canManage && (
                                                        <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                                            <button onClick={() => openEditDialog(block)} className="p-1 hover:text-primary transition-colors">
                                                                <Edit2 className="w-3 h-3" />
                                                            </button>
                                                            <button onClick={() => handleDelete(block.id)} className="p-1 hover:text-destructive transition-colors">
                                                                <Trash2 className="w-3 h-3" />
                                                            </button>
                                                        </div>
                                                    )}
                                                </div>

                                                <div>
                                                    <h4 className="text-sm font-bold leading-tight line-clamp-2">{block.courseName || `Course ${block.courseId}`}</h4>
                                                    <p className="text-[10px] text-muted-foreground font-medium uppercase mt-0.5">{block.courseCode || '-'}</p>
                                                </div>

                                                <div className="space-y-1">
                                                    <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                                                        <MapPin className="w-3 h-3 shrink-0" />
                                                        <span className="truncate">{block.room || 'TBA'}</span>
                                                    </div>
                                                    <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                                                        <UserIcon className="w-3 h-3 shrink-0" />
                                                        <span className="truncate">{block.instructorName || 'Not Assigned'}</span>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    ))
                            )}
                        </CardContent>
                    </Card>
                ))}
            </div>

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>{editingBlock ? 'Edit Class' : 'Add New Class'}</DialogTitle>
                        <DialogDescription>
                            Enter the details for the timetable block below.
                        </DialogDescription>
                    </DialogHeader>
                    <form onSubmit={handleSubmit} className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="course">Course</Label>
                            <Select
                                value={formData.courseId}
                                onValueChange={(value) => setFormData((prev) => ({ ...prev, courseId: value }))}
                            >
                                <SelectTrigger id="course">
                                    <SelectValue placeholder="Select a course" />
                                </SelectTrigger>
                                <SelectContent>
                                    {courses.map((course) => (
                                        <SelectItem key={course.id} value={course.id.toString()}>
                                            {course.name} ({course.code})
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="day">Day</Label>
                            <Select
                                value={formData.dayOfWeek}
                                onValueChange={(value) => setFormData((prev) => ({ ...prev, dayOfWeek: value }))}
                            >
                                <SelectTrigger id="day">
                                    <SelectValue placeholder="Select day" />
                                </SelectTrigger>
                                <SelectContent>
                                    {DAYS.map((day) => (
                                        <SelectItem key={day.id} value={day.id.toString()}>
                                            {day.name}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="startTime">Start Time</Label>
                                <Input
                                    id="startTime"
                                    type="time"
                                    value={formData.startTime}
                                    onChange={(event) => setFormData((prev) => ({ ...prev, startTime: event.target.value }))}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="endTime">End Time</Label>
                                <Input
                                    id="endTime"
                                    type="time"
                                    value={formData.endTime}
                                    onChange={(event) => setFormData((prev) => ({ ...prev, endTime: event.target.value }))}
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="room">Room</Label>
                            <Input
                                id="room"
                                placeholder="e.g. Room 302, Virtual"
                                value={formData.room}
                                onChange={(event) => setFormData((prev) => ({ ...prev, room: event.target.value }))}
                            />
                        </div>

                        <DialogFooter>
                            <Button type="button" variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                            <Button type="submit">
                                {editingBlock ? 'Update' : 'Create'}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>
        </div>
    );
}
