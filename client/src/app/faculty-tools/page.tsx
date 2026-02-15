"use client";

import { useEffect, useMemo, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { api, endpoints } from '@/lib/api-client';
import {
  useAnnouncements,
  useCreateAnnouncement,
  useFacultyRubrics,
  useCreateRubric,
  useDeleteRubric,
  useOfficeHours,
  useCreateOfficeHour,
  useDeleteOfficeHour,
  useBookings,
  useCreateBooking,
  useUpdateBookingStatus,
} from '@/lib/api-hooks';
import { useAuth } from '@/lib/auth-context';
import { logger } from '@/lib/logger';
import type { CreateOfficeHourBookingInput, CreateOfficeHourInput, CreateRubricInput, OfficeHourBooking, OfficeHourSlot } from '@/lib/types';
import {
  Users,
  Send,
  Clock,
  CheckSquare,
  FileText,
  Calendar,
  Plus,
  BookOpen,
  Loader2,
} from 'lucide-react';

type CourseOption = {
  id: number;
  name?: string;
  title?: string;
};

type DraftCriterion = {
  id: string;
  name: string;
  weight: number;
  description: string;
  maxScore: number;
};

const DAY_OPTIONS = [
  { value: 0, label: 'Sunday' },
  { value: 1, label: 'Monday' },
  { value: 2, label: 'Tuesday' },
  { value: 3, label: 'Wednesday' },
  { value: 4, label: 'Thursday' },
  { value: 5, label: 'Friday' },
  { value: 6, label: 'Saturday' },
];

const dayLabelByValue = DAY_OPTIONS.reduce<Record<number, string>>((acc, day) => {
  acc[day.value] = day.label;
  return acc;
}, {});

function toLocalDateInput(date: Date): string {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, '0');
  const day = `${date.getDate()}`.padStart(2, '0');
  return `${year}-${month}-${day}`;
}

export default function FacultyToolsPage() {
  const { user } = useAuth();
  const role = user?.role;
  const canManage = role === 'faculty' || role === 'admin';
  const tabs = useMemo(
    () =>
      canManage
        ? [
            { key: 'announcements', label: 'Bulk Announcements', icon: Send },
            { key: 'rubrics', label: 'Grading Rubrics', icon: FileText },
            { key: 'officehours', label: 'Office Hours', icon: Clock },
            { key: 'bookings', label: 'Bookings', icon: Calendar },
          ]
        : [
            { key: 'officehours', label: 'Office Hours', icon: Clock },
            { key: 'bookings', label: 'Bookings', icon: Calendar },
          ],
    [canManage],
  );

  const [tab, setTab] = useState<string>(tabs[0]?.key ?? 'officehours');

  useEffect(() => {
    if (!tabs.find((item) => item.key === tab)) {
      setTab(tabs[0]?.key ?? 'officehours');
    }
  }, [tab, tabs]);

  return (
    <div className="min-h-screen bg-muted/10">
      <div className="container mx-auto p-6 space-y-6">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <CheckSquare className="w-8 h-8" />
            Faculty Productivity Tools
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage announcements, grading workflows, office hours, and student bookings
          </p>
        </div>

        <Tabs value={tab} onValueChange={setTab} className="space-y-4">
          <TabsList className={`grid w-full ${canManage ? 'grid-cols-4' : 'grid-cols-2'} lg:w-auto`}>
            {tabs.map(({ key, label, icon: Icon }) => (
              <TabsTrigger key={key} value={key}>
                <Icon className="w-4 h-4 mr-2" />
                {label}
              </TabsTrigger>
            ))}
          </TabsList>

          {canManage && (
            <TabsContent value="announcements" className="space-y-4">
              <BulkAnnouncementsTool />
            </TabsContent>
          )}

          {canManage && (
            <TabsContent value="rubrics" className="space-y-4">
              <GradingRubricsTool />
            </TabsContent>
          )}

          <TabsContent value="officehours" className="space-y-4">
            <OfficeHoursTool canManage={canManage} />
          </TabsContent>

          <TabsContent value="bookings" className="space-y-4">
            <BookingsTool canManage={canManage} />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

function BulkAnnouncementsTool() {
  const { data: announcements = [] } = useAnnouncements();
  const [courses, setCourses] = useState<CourseOption[]>([]);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [newAnnouncement, setNewAnnouncement] = useState({
    title: '',
    content: '',
    priority: 'normal' as 'low' | 'normal' | 'high' | 'urgent',
    targetAudience: [] as string[],
    courseId: undefined as number | undefined,
  });
  const createAnnouncement = useCreateAnnouncement();

  useEffect(() => {
    let mounted = true;
    const loadCourses = async () => {
      try {
        const data = await api.get<CourseOption[]>(endpoints.courses.list);
        if (!mounted) return;
        setCourses(Array.isArray(data) ? data : []);
      } catch (error) {
        logger.error('Failed to load courses for announcements:', error as Error);
      }
    };
    void loadCourses();
    return () => {
      mounted = false;
    };
  }, []);

  const handleCreateAnnouncement = async () => {
    setFormError(null);
    if (!newAnnouncement.title.trim() || !newAnnouncement.content.trim()) {
      setFormError('Title and content are required.');
      return;
    }

    setIsSubmitting(true);
    try {
      const content = newAnnouncement.targetAudience.length > 0
        ? `${newAnnouncement.content.trim()}\n\nTarget Audience: ${newAnnouncement.targetAudience.join(', ')}`
        : newAnnouncement.content.trim();

      await createAnnouncement.mutateAsync({
        title: newAnnouncement.title.trim(),
        content,
        priority: newAnnouncement.priority,
        course_id: newAnnouncement.courseId,
        is_published: true,
        published_at: new Date().toISOString(),
      });

      setIsDialogOpen(false);
      setNewAnnouncement({
        title: '',
        content: '',
        priority: 'normal',
        targetAudience: [],
        courseId: undefined,
      });
    } catch (error) {
      logger.error('Failed to create announcement:', error as Error);
      setFormError('Failed to create announcement. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Send className="w-5 h-5" />
              Bulk Announcements
            </CardTitle>
            <Button onClick={() => setIsDialogOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              New Announcement
            </Button>
          </div>
          <CardDescription>
            Send announcements to multiple courses or the entire college
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {announcements.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No announcements yet. Create your first one!
              </div>
            ) : (
              announcements.map((announcement) => (
                <div key={announcement.id} className="p-4 border rounded-lg">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <h4 className="font-semibold">{announcement.title}</h4>
                        <Badge
                          variant={
                            announcement.priority === 'urgent' ? 'destructive' :
                            announcement.priority === 'high' ? 'default' :
                            'secondary'
                          }
                        >
                          {announcement.priority}
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mb-2">{announcement.content}</p>
                      <div className="flex items-center gap-4 text-xs text-muted-foreground">
                        <span>By {announcement.authorName}</span>
                        <span>{new Date(announcement.publishedAt).toLocaleString()}</span>
                        <span className="flex items-center gap-1">
                          <Users className="w-3 h-3" />
                          {(announcement.targetAudience?.length ?? 0)} recipients
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </CardContent>
      </Card>

      {isDialogOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Create Announcement</CardTitle>
                <Button variant="ghost" size="sm" onClick={() => setIsDialogOpen(false)}>
                  ×
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="title">Title</Label>
                <Input
                  id="title"
                  value={newAnnouncement.title}
                  onChange={(e) => setNewAnnouncement({ ...newAnnouncement, title: e.target.value })}
                  placeholder="Announcement title"
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="content">Content</Label>
                <textarea
                  id="content"
                  value={newAnnouncement.content}
                  onChange={(e) => setNewAnnouncement({ ...newAnnouncement, content: e.target.value })}
                  placeholder="Announcement content..."
                  className="flex min-h-[150px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  required
                />
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="priority">Priority</Label>
                  <select
                    id="priority"
                    value={newAnnouncement.priority}
                    onChange={(e) => setNewAnnouncement({ ...newAnnouncement, priority: e.target.value as 'low' | 'normal' | 'high' | 'urgent' })}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="low">Low</option>
                    <option value="normal">Normal</option>
                    <option value="high">High</option>
                    <option value="urgent">Urgent</option>
                  </select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="course">Course (Optional)</Label>
                  <select
                    id="course"
                    value={newAnnouncement.courseId || ''}
                    onChange={(e) => setNewAnnouncement({ ...newAnnouncement, courseId: e.target.value ? parseInt(e.target.value, 10) : undefined })}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="">All Courses</option>
                    {courses.map((course) => (
                      <option key={course.id} value={course.id}>
                        {course.name || course.title || `Course ${course.id}`}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              {formError && <p className="text-sm text-destructive">{formError}</p>}

              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setIsDialogOpen(false)} disabled={isSubmitting}>
                  Cancel
                </Button>
                <Button onClick={handleCreateAnnouncement} disabled={isSubmitting || !newAnnouncement.title || !newAnnouncement.content}>
                  {isSubmitting ? 'Sending...' : (
                    <>
                      <Send className="w-4 h-4 mr-2" />
                      Send Announcement
                    </>
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

function GradingRubricsTool() {
  const { data: rubrics = [], isLoading } = useFacultyRubrics();
  const createRubric = useCreateRubric();
  const deleteRubric = useDeleteRubric();

  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [rubricName, setRubricName] = useState('');
  const [maxScore, setMaxScore] = useState(100);
  const [criteria, setCriteria] = useState<DraftCriterion[]>([
    { id: 'criterion-1', name: '', weight: 100, description: '', maxScore: 100 },
  ]);

  const resetForm = () => {
    setRubricName('');
    setMaxScore(100);
    setCriteria([{ id: `criterion-${Date.now()}`, name: '', weight: 100, description: '', maxScore: 100 }]);
    setFormError(null);
  };

  const addCriterion = () => {
    setCriteria((prev) => [...prev, { id: `criterion-${Date.now()}`, name: '', weight: 0, description: '', maxScore: 0 }]);
  };

  const updateCriterion = (criterionId: string, patch: Partial<DraftCriterion>) => {
    setCriteria((prev) => prev.map((item) => (item.id === criterionId ? { ...item, ...patch } : item)));
  };

  const removeCriterion = (criterionId: string) => {
    setCriteria((prev) => (prev.length > 1 ? prev.filter((item) => item.id !== criterionId) : prev));
  };

  const handleCreate = async () => {
    setFormError(null);
    const name = rubricName.trim();
    const normalizedCriteria = criteria
      .map((criterion, index) => ({
        name: criterion.name.trim(),
        description: criterion.description.trim() || undefined,
        weight: criterion.weight,
        max_score: criterion.maxScore,
        sort_order: index + 1,
      }))
      .filter((criterion) => criterion.name !== '');

    if (!name) {
      setFormError('Rubric name is required.');
      return;
    }
    if (normalizedCriteria.length === 0) {
      setFormError('Add at least one criterion.');
      return;
    }

    const payload: CreateRubricInput = {
      name,
      is_template: false,
      is_active: true,
      max_score: maxScore,
      criteria: normalizedCriteria,
    };

    try {
      await createRubric.mutateAsync(payload);
      setIsDialogOpen(false);
      resetForm();
    } catch (error) {
      logger.error('Failed to create rubric:', error as Error);
      setFormError('Failed to save rubric.');
    }
  };

  const handleDelete = async (rubricId: number) => {
    try {
      await deleteRubric.mutateAsync(rubricId);
    } catch (error) {
      logger.error('Failed to delete rubric:', error as Error);
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <FileText className="w-5 h-5" />
              Grading Rubrics
            </CardTitle>
            <Button onClick={() => setIsDialogOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              New Rubric
            </Button>
          </div>
          <CardDescription>Create and manage reusable grading rubrics.</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading rubrics...</div>
          ) : rubrics.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">No rubrics yet. Create your first rubric.</div>
          ) : (
            <div className="space-y-4">
              {rubrics.map((rubric) => (
                <div key={rubric.id} className="p-4 border rounded-lg">
                  <div className="flex items-center justify-between mb-3">
                    <div>
                      <h4 className="font-semibold text-lg">{rubric.name}</h4>
                      <p className="text-sm text-muted-foreground">Max score: {rubric.maxScore}</p>
                    </div>
                    <Button variant="outline" size="sm" onClick={() => handleDelete(rubric.id)}>
                      Delete
                    </Button>
                  </div>
                  <div className="grid gap-2 md:grid-cols-2">
                    {rubric.criteria.map((criterion) => (
                      <div key={criterion.id} className="p-3 bg-muted/50 rounded">
                        <div className="flex items-center justify-between mb-1">
                          <span className="font-medium">{criterion.name}</span>
                          <Badge variant="outline">{criterion.weight}%</Badge>
                        </div>
                        <p className="text-sm text-muted-foreground">{criterion.description || 'No description'}</p>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {isDialogOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-3xl max-h-[90vh] overflow-y-auto">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Create Rubric</CardTitle>
                <Button variant="ghost" size="sm" onClick={() => {
                  setIsDialogOpen(false);
                  resetForm();
                }}>
                  ×
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="rubricName">Rubric Name</Label>
                  <Input
                    id="rubricName"
                    value={rubricName}
                    onChange={(event) => setRubricName(event.target.value)}
                    placeholder="e.g., Final Project Rubric"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="maxScore">Max Score</Label>
                  <Input
                    id="maxScore"
                    type="number"
                    min={1}
                    value={maxScore}
                    onChange={(event) => setMaxScore(Number.parseInt(event.target.value || '100', 10))}
                  />
                </div>
              </div>

              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label>Criteria</Label>
                  <Button type="button" variant="outline" size="sm" onClick={addCriterion}>
                    <Plus className="w-4 h-4 mr-2" />
                    Add Criterion
                  </Button>
                </div>

                {criteria.map((criterion) => (
                  <div key={criterion.id} className="p-3 border rounded-md space-y-3">
                    <div className="grid gap-3 md:grid-cols-[2fr_1fr_1fr]">
                      <Input
                        value={criterion.name}
                        onChange={(event) => updateCriterion(criterion.id, { name: event.target.value })}
                        placeholder="Criterion name"
                      />
                      <Input
                        type="number"
                        min={1}
                        value={criterion.maxScore}
                        onChange={(event) => updateCriterion(criterion.id, { maxScore: Number.parseInt(event.target.value || '0', 10) })}
                        placeholder="Max score"
                      />
                      <Input
                        type="number"
                        min={0}
                        max={100}
                        value={criterion.weight}
                        onChange={(event) => updateCriterion(criterion.id, { weight: Number.parseInt(event.target.value || '0', 10) })}
                        placeholder="Weight %"
                      />
                    </div>
                    <TextareaCriterion
                      value={criterion.description}
                      onChange={(value) => updateCriterion(criterion.id, { description: value })}
                    />
                    <div className="flex justify-end">
                      <Button type="button" variant="outline" size="sm" onClick={() => removeCriterion(criterion.id)} disabled={criteria.length === 1}>
                        Remove
                      </Button>
                    </div>
                  </div>
                ))}
              </div>

              {formError && <p className="text-sm text-destructive">{formError}</p>}

              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => {
                  setIsDialogOpen(false);
                  resetForm();
                }}>
                  Cancel
                </Button>
                <Button onClick={handleCreate} disabled={createRubric.isPending}>
                  {createRubric.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Save Rubric'}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

function OfficeHoursTool({ canManage }: { canManage: boolean }) {
  const { data: officeHours = [], isLoading } = useOfficeHours({ activeOnly: !canManage });
  const createOfficeHour = useCreateOfficeHour();
  const deleteOfficeHour = useDeleteOfficeHour();

  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [newSlot, setNewSlot] = useState<CreateOfficeHourInput>({
    day_of_week: 1,
    start_time: '09:00',
    end_time: '10:00',
    location: '',
    is_virtual: false,
    virtual_link: '',
    max_students: 1,
    is_active: true,
  });

  const resetForm = () => {
    setNewSlot({
      day_of_week: 1,
      start_time: '09:00',
      end_time: '10:00',
      location: '',
      is_virtual: false,
      virtual_link: '',
      max_students: 1,
      is_active: true,
    });
    setFormError(null);
  };

  const handleCreate = async () => {
    setFormError(null);
    if (newSlot.end_time <= newSlot.start_time) {
      setFormError('End time must be later than start time.');
      return;
    }
    if (!newSlot.is_virtual && !newSlot.location?.trim()) {
      setFormError('Location is required for in-person slots.');
      return;
    }
    if (newSlot.is_virtual && !newSlot.virtual_link?.trim()) {
      setFormError('Meeting link is required for virtual slots.');
      return;
    }

    try {
      await createOfficeHour.mutateAsync({
        ...newSlot,
        location: newSlot.is_virtual ? undefined : newSlot.location?.trim(),
        virtual_link: newSlot.is_virtual ? newSlot.virtual_link?.trim() : undefined,
      });
      setIsDialogOpen(false);
      resetForm();
    } catch (error) {
      logger.error('Failed to create office hour:', error as Error);
      setFormError('Failed to save office-hour slot.');
    }
  };

  const handleDelete = async (officeHourId: number) => {
    try {
      await deleteOfficeHour.mutateAsync(officeHourId);
    } catch (error) {
      logger.error('Failed to delete office hour:', error as Error);
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Clock className="w-5 h-5" />
              Office Hours
            </CardTitle>
            {canManage && (
              <Button onClick={() => setIsDialogOpen(true)}>
                <Plus className="w-4 h-4 mr-2" />
                Add Slot
              </Button>
            )}
          </div>
          <CardDescription>
            {canManage ? 'Manage your availability for student consultations.' : 'Browse available office-hour slots.'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading office hours...</div>
          ) : officeHours.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">No office-hour slots configured.</div>
          ) : (
            <div className="space-y-4">
              {officeHours.map((slot) => (
                <OfficeHourSlotCard key={slot.id} slot={slot} canManage={canManage} onDelete={handleDelete} />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {canManage && isDialogOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-2xl">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle>Add Office-Hour Slot</CardTitle>
                <Button variant="ghost" size="sm" onClick={() => {
                  setIsDialogOpen(false);
                  resetForm();
                }}>
                  ×
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-4">
                <div className="space-y-2 md:col-span-2">
                  <Label htmlFor="slotDay">Day</Label>
                  <select
                    id="slotDay"
                    value={newSlot.day_of_week}
                    onChange={(event) => setNewSlot((prev) => ({ ...prev, day_of_week: Number.parseInt(event.target.value, 10) }))}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    {DAY_OPTIONS.map((day) => (
                      <option key={day.value} value={day.value}>{day.label}</option>
                    ))}
                  </select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="startTime">Start</Label>
                  <Input id="startTime" type="time" value={newSlot.start_time} onChange={(event) => setNewSlot((prev) => ({ ...prev, start_time: event.target.value }))} />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="endTime">End</Label>
                  <Input id="endTime" type="time" value={newSlot.end_time} onChange={(event) => setNewSlot((prev) => ({ ...prev, end_time: event.target.value }))} />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="maxStudents">Max Students</Label>
                <Input
                  id="maxStudents"
                  type="number"
                  min={1}
                  value={newSlot.max_students}
                  onChange={(event) => setNewSlot((prev) => ({ ...prev, max_students: Number.parseInt(event.target.value || '1', 10) }))}
                />
              </div>

              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={newSlot.is_virtual}
                  onChange={(event) =>
                    setNewSlot((prev) => ({
                      ...prev,
                      is_virtual: event.target.checked,
                      location: event.target.checked ? '' : prev.location,
                      virtual_link: event.target.checked ? prev.virtual_link : '',
                    }))
                  }
                />
                Virtual session
              </label>

              {newSlot.is_virtual ? (
                <div className="space-y-2">
                  <Label htmlFor="meetingLink">Meeting Link</Label>
                  <Input
                    id="meetingLink"
                    value={newSlot.virtual_link || ''}
                    onChange={(event) => setNewSlot((prev) => ({ ...prev, virtual_link: event.target.value }))}
                    placeholder="https://meet.example.com/..."
                  />
                </div>
              ) : (
                <div className="space-y-2">
                  <Label htmlFor="location">Location</Label>
                  <Input
                    id="location"
                    value={newSlot.location || ''}
                    onChange={(event) => setNewSlot((prev) => ({ ...prev, location: event.target.value }))}
                    placeholder="Room 301, Building A"
                  />
                </div>
              )}

              {formError && <p className="text-sm text-destructive">{formError}</p>}

              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => {
                  setIsDialogOpen(false);
                  resetForm();
                }}>
                  Cancel
                </Button>
                <Button onClick={handleCreate} disabled={createOfficeHour.isPending}>
                  {createOfficeHour.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Save Slot'}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

function OfficeHourSlotCard({
  slot,
  canManage,
  onDelete,
}: {
  slot: OfficeHourSlot;
  canManage: boolean;
  onDelete: (officeHourId: number) => Promise<void>;
}) {
  return (
    <div className="p-4 border rounded-lg">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <h4 className="font-semibold text-lg">{dayLabelByValue[slot.dayOfWeek] ?? `Day ${slot.dayOfWeek}`}</h4>
            <Badge variant={slot.isVirtual ? 'default' : 'secondary'}>
              {slot.isVirtual ? 'Virtual' : 'In-Person'}
            </Badge>
            {!slot.isActive && <Badge variant="outline">Inactive</Badge>}
          </div>
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-1">
              <Clock className="w-4 h-4" />
              {slot.startTime} - {slot.endTime}
            </div>
            {!slot.isVirtual && slot.location && (
              <div className="flex items-center gap-1">
                <Calendar className="w-4 h-4" />
                {slot.location}
              </div>
            )}
            {slot.facultyName && <span>Faculty: {slot.facultyName}</span>}
            <span>Capacity: {slot.maxStudents}</span>
          </div>
          {slot.isVirtual && slot.virtualLink && (
            <a href={slot.virtualLink} target="_blank" rel="noopener noreferrer" className="text-sm text-primary hover:underline mt-2 inline-block">
              Join Meeting
            </a>
          )}
        </div>
        {canManage && (
          <Button variant="outline" size="sm" onClick={() => void onDelete(slot.id)}>
            Remove
          </Button>
        )}
      </div>
    </div>
  );
}

function BookingsTool({ canManage }: { canManage: boolean }) {
  const { data: officeHours = [] } = useOfficeHours({ activeOnly: true });
  const { data: bookings = [], isLoading } = useBookings();
  const createBooking = useCreateBooking();
  const updateBookingStatus = useUpdateBookingStatus();

  const [selectedOfficeHourID, setSelectedOfficeHourID] = useState<number | null>(null);
  const [bookingDate, setBookingDate] = useState(toLocalDateInput(new Date()));
  const [purpose, setPurpose] = useState('');
  const [notes, setNotes] = useState<Record<number, string>>({});
  const [statusDraft, setStatusDraft] = useState<Record<number, OfficeHourBooking['status']>>({});
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const initialStatus: Record<number, OfficeHourBooking['status']> = {};
    const initialNotes: Record<number, string> = {};
    bookings.forEach((booking) => {
      initialStatus[booking.id] = booking.status;
      initialNotes[booking.id] = booking.notes || '';
    });
    setStatusDraft(initialStatus);
    setNotes(initialNotes);
  }, [bookings]);

  const availableSlots = useMemo(() => officeHours.filter((slot) => slot.isActive), [officeHours]);

  const handleBook = async (officeHourID: number) => {
    setError(null);
    const payload: CreateOfficeHourBookingInput = {
      office_hour_id: officeHourID,
      booking_date: bookingDate,
      purpose: purpose.trim() || undefined,
    };

    try {
      await createBooking.mutateAsync(payload);
      setSelectedOfficeHourID(null);
      setPurpose('');
    } catch (bookingError) {
      logger.error('Failed to create booking:', bookingError as Error);
      setError('Failed to create booking.');
    }
  };

  const handleCancel = async (bookingId: number) => {
    try {
      await updateBookingStatus.mutateAsync({ bookingId, status: 'cancelled' });
    } catch (cancelError) {
      logger.error('Failed to cancel booking:', cancelError as Error);
      setError('Failed to cancel booking.');
    }
  };

  const handleUpdateStatus = async (bookingId: number) => {
    const nextStatus = statusDraft[bookingId] || 'confirmed';
    try {
      await updateBookingStatus.mutateAsync({ bookingId, status: nextStatus, notes: notes[bookingId] || undefined });
    } catch (statusError) {
      logger.error('Failed to update booking status:', statusError as Error);
      setError('Failed to update booking status.');
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BookOpen className="w-5 h-5" />
            {canManage ? 'Booking Management' : 'Reserve Office-Hour Slot'}
          </CardTitle>
          <CardDescription>
            {canManage
              ? 'Track bookings and update consultation outcomes.'
              : 'Choose an available slot and reserve your meeting.'}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!canManage && (
            <>
              <div className="grid gap-4 md:grid-cols-3">
                <div className="space-y-2 md:col-span-1">
                  <Label htmlFor="bookingDate">Booking Date</Label>
                  <Input
                    id="bookingDate"
                    type="date"
                    value={bookingDate}
                    min={toLocalDateInput(new Date())}
                    onChange={(event) => setBookingDate(event.target.value)}
                  />
                </div>
                <div className="space-y-2 md:col-span-2">
                  <Label htmlFor="purpose">Purpose (optional)</Label>
                  <Input
                    id="purpose"
                    value={purpose}
                    onChange={(event) => setPurpose(event.target.value)}
                    placeholder="What would you like to discuss?"
                  />
                </div>
              </div>

              {availableSlots.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">No active office-hour slots are currently available.</div>
              ) : (
                <div className="space-y-3">
                  {availableSlots.map((slot) => (
                    <div key={slot.id} className="p-4 border rounded-lg flex items-center justify-between gap-4">
                      <div>
                        <div className="font-semibold">{dayLabelByValue[slot.dayOfWeek] ?? `Day ${slot.dayOfWeek}`} {slot.startTime}-{slot.endTime}</div>
                        <p className="text-sm text-muted-foreground">
                          {slot.isVirtual ? `Virtual ${slot.virtualLink ? `(${slot.virtualLink})` : ''}` : slot.location || 'In-person'}
                        </p>
                      </div>
                      <Button
                        onClick={() => {
                          setSelectedOfficeHourID(slot.id);
                          void handleBook(slot.id);
                        }}
                        disabled={createBooking.isPending && selectedOfficeHourID === slot.id}
                      >
                        {createBooking.isPending && selectedOfficeHourID === slot.id ? (
                          <Loader2 className="w-4 h-4 animate-spin" />
                        ) : 'Book Slot'}
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </>
          )}

          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading bookings...</div>
          ) : bookings.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">No bookings found.</div>
          ) : (
            <div className="space-y-3">
              {bookings.map((booking) => (
                <div key={booking.id} className="p-4 border rounded-lg">
                  <div className="flex items-center justify-between gap-3 mb-2">
                    <div>
                      <h4 className="font-semibold">Booking #{booking.id}</h4>
                      <p className="text-sm text-muted-foreground">
                        {new Date(booking.bookingDate).toLocaleDateString()} {booking.startTime}-{booking.endTime}
                      </p>
                    </div>
                    <Badge
                      variant={booking.status === 'cancelled' ? 'secondary' : booking.status === 'completed' ? 'default' : booking.status === 'no_show' ? 'destructive' : 'outline'}
                    >
                      {booking.status}
                    </Badge>
                  </div>

                  {booking.purpose && <p className="text-sm mb-2">Purpose: {booking.purpose}</p>}

                  {canManage ? (
                    <div className="grid gap-3 md:grid-cols-[180px_1fr_auto]">
                      <select
                        value={statusDraft[booking.id] || booking.status}
                        onChange={(event) => setStatusDraft((prev) => ({ ...prev, [booking.id]: event.target.value as OfficeHourBooking['status'] }))}
                        className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      >
                        <option value="confirmed">confirmed</option>
                        <option value="completed">completed</option>
                        <option value="no_show">no_show</option>
                        <option value="cancelled">cancelled</option>
                      </select>
                      <Input
                        value={notes[booking.id] || ''}
                        onChange={(event) => setNotes((prev) => ({ ...prev, [booking.id]: event.target.value }))}
                        placeholder="Add notes"
                      />
                      <Button onClick={() => void handleUpdateStatus(booking.id)} disabled={updateBookingStatus.isPending}>
                        Update
                      </Button>
                    </div>
                  ) : (
                    booking.status === 'confirmed' && (
                      <Button variant="outline" onClick={() => void handleCancel(booking.id)} disabled={updateBookingStatus.isPending}>
                        Cancel Booking
                      </Button>
                    )
                  )}
                </div>
              ))}
            </div>
          )}

          {error && <p className="text-sm text-destructive">{error}</p>}
        </CardContent>
      </Card>
    </div>
  );
}

function TextareaCriterion({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <textarea
      value={value}
      onChange={(event) => onChange(event.target.value)}
      placeholder="Description"
      className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
    />
  );
}
