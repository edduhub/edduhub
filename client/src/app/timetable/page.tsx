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
import { cn } from '@/lib/utils';
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

const DAYS = [
    { id: 1, name: 'Monday' },
    { id: 2, name: 'Tuesday' },
    { id: 3, name: 'Wednesday' },
    { id: 4, name: 'Thursday' },
    { id: 5, name: 'Friday' },
    { id: 6, name: 'Saturday' },
    { id: 0, name: 'Sunday' },
];

export default function TimetablePage() {
    const { user } = useAuth();
    const [blocks, setBlocks] = useState<TimetableBlock[]>([]);
    const [courses, setCourses] = useState<Course[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingBlock, setEditingBlock] = useState<TimetableBlock | null>(null);

    // Form state
    const [formData, setFormData] = useState({
        courseId: '',
        dayOfWeek: '1',
        startTime: '09:00',
        endTime: '10:00',
        room: '',
        type: 'lecture' as TimetableBlock['type']
    });

    const fetchTimetable = useCallback(async () => {
        try {
            const endpoint = user?.role === 'student'
                ? endpoints.timetable.myTimetable
                : endpoints.timetable.list;
            const data = await api.get<TimetableBlock[]>(endpoint);
            setBlocks(data || []);
        } catch (error) {
            logger.error('Failed to fetch timetable:', error as Error);
        } finally {
            setIsLoading(false);
        }
    }, [user]);

    const fetchCourses = useCallback(async () => {
        if (user?.role === 'admin' || user?.role === 'faculty') {
            try {
                const data = await api.get<Course[]>(endpoints.courses.list);
                setCourses(data || []);
            } catch (error) {
                logger.error('Failed to fetch courses:', error as Error);
            }
        }
    }, [user]);

    useEffect(() => {
        if (user) {
            fetchTimetable();
            fetchCourses();
        }
    }, [user, fetchTimetable, fetchCourses]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            const payload = {
                courseId: parseInt(formData.courseId),
                dayOfWeek: parseInt(formData.dayOfWeek),
                startTime: formData.startTime,
                endTime: formData.endTime,
                room: formData.room,
                type: formData.type || 'lecture'
            };

            if (editingBlock) {
                await api.patch(endpoints.timetable.update(editingBlock.id), payload);
            } else {
                await api.post(endpoints.timetable.create, payload);
            }

            setIsDialogOpen(false);
            setEditingBlock(null);
            fetchTimetable();
        } catch (error) {
            logger.error('Failed to save timetable block:', error as Error);
        }
    };

    const handleDelete = async (id: number) => {
        if (confirm('Are you sure you want to delete this class?')) {
            try {
                await api.delete(endpoints.timetable.delete(id));
                fetchTimetable();
            } catch (error) {
                logger.error('Failed to delete block:', error as Error);
            }
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
            type: 'lecture'
        });
        setIsDialogOpen(true);
    };

    const openEditDialog = (block: TimetableBlock) => {
        setEditingBlock(block);
        setFormData({
            courseId: block.courseId.toString(),
            dayOfWeek: block.dayOfWeek.toString(),
            startTime: block.startTime,
            endTime: block.endTime,
            room: block.room || '',
            type: block.type || 'lecture'
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

    const isAdmin = user?.role === 'admin' || user?.role === 'faculty';

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
                {isAdmin && (
                    <Button onClick={openAddDialog} className="shadow-lg hover:shadow-primary/20">
                        <Plus className="w-4 h-4 mr-2" />
                        Add Class
                    </Button>
                )}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-7 gap-4">
                {DAYS.filter(d => d.id !== 0).map((day) => (
                    <Card key={day.id} className="border-none shadow-sm bg-muted/30">
                        <CardHeader className="p-3 text-center border-b bg-background/50 rounded-t-xl">
                            <CardTitle className="text-sm font-bold uppercase tracking-wider">{day.name}</CardTitle>
                        </CardHeader>
                        <CardContent className="p-2 space-y-3">
                            {blocks.filter(b => b.dayOfWeek === day.id).length === 0 ? (
                                <div className="py-8 text-center text-xs text-muted-foreground italic">No classes</div>
                            ) : (
                                blocks
                                    .filter(b => b.dayOfWeek === day.id)
                                    .sort((a, b) => a.startTime.localeCompare(b.startTime))
                                    .map((block) => (
                                        <div
                                            key={block.id}
                                            className={cn(
                                                "group p-3 rounded-lg border bg-background shadow-sm hover:shadow-md transition-all relative overflow-hidden",
                                                block.type === 'lab' ? "border-l-4 border-l-purple-500" :
                                                    block.type === 'tutorial' ? "border-l-4 border-l-amber-500" :
                                                        "border-l-4 border-l-primary"
                                            )}
                                        >
                                            <div className="space-y-2">
                                                <div className="flex items-start justify-between">
                                                    <Badge variant="outline" className="text-[10px] font-mono px-1.5 h-5 flex items-center gap-1">
                                                        <Clock className="w-3 h-3" />
                                                        {block.startTime}
                                                    </Badge>
                                                    {isAdmin && (
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
                                                    <h4 className="text-sm font-bold leading-tight line-clamp-2">{block.courseName}</h4>
                                                    <p className="text-[10px] text-muted-foreground font-medium uppercase mt-0.5">{block.courseCode}</p>
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
                                onValueChange={(v) => setFormData(prev => ({ ...prev, courseId: v }))}
                            >
                                <SelectTrigger id="course">
                                    <SelectValue placeholder="Select a course" />
                                </SelectTrigger>
                                <SelectContent>
                                    {courses.map(course => (
                                        <SelectItem key={course.id} value={course.id.toString()}>
                                            {course.name} ({course.code})
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="day">Day</Label>
                                <Select
                                    value={formData.dayOfWeek}
                                    onValueChange={(v) => setFormData(prev => ({ ...prev, dayOfWeek: v }))}
                                >
                                    <SelectTrigger id="day">
                                        <SelectValue placeholder="Select day" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {DAYS.map(day => (
                                            <SelectItem key={day.id} value={day.id.toString()}>
                                                {day.name}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="type">Type</Label>
                                <Select
                                    value={formData.type}
                                    onValueChange={(v: string) => setFormData(prev => ({ ...prev, type: v as typeof prev.type }))}
                                >
                                    <SelectTrigger id="type">
                                        <SelectValue placeholder="Select type" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="lecture">Lecture</SelectItem>
                                        <SelectItem value="lab">Laboratory</SelectItem>
                                        <SelectItem value="tutorial">Tutorial</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="startTime">Start Time</Label>
                                <Input
                                    id="startTime"
                                    type="time"
                                    value={formData.startTime}
                                    onChange={(e) => setFormData(prev => ({ ...prev, startTime: e.target.value }))}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="endTime">End Time</Label>
                                <Input
                                    id="endTime"
                                    type="time"
                                    value={formData.endTime}
                                    onChange={(e) => setFormData(prev => ({ ...prev, endTime: e.target.value }))}
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="room">Room</Label>
                            <Input
                                id="room"
                                placeholder="e.g. Room 302, Virtual"
                                value={formData.room}
                                onChange={(e) => setFormData(prev => ({ ...prev, room: e.target.value }))}
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
