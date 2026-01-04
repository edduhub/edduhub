"use client";

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { 
  Users, 
  Send, 
  Clock, 
  CheckSquare,
  FileText,
  Calendar,
  Filter,
  Search,
  Plus
} from 'lucide-react';
import type { Announcement, Course } from '@/lib/types';

export default function FacultyToolsPage() {
  return (
    <div className="min-h-screen bg-muted/10">
      <div className="container mx-auto p-6 space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <CheckSquare className="w-8 h-8" />
            Faculty Productivity Tools
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage announcements, grading rubrics, and office hours
          </p>
        </div>

        {/* Tools Tabs */}
        <Tabs defaultValue="announcements" className="space-y-4">
          <TabsList className="grid w-full grid-cols-3 lg:w-auto">
            <TabsTrigger value="announcements">
              <Send className="w-4 h-4 mr-2" />
              Bulk Announcements
            </TabsTrigger>
            <TabsTrigger value="rubrics">
              <FileText className="w-4 h-4 mr-2" />
              Grading Rubrics
            </TabsTrigger>
            <TabsTrigger value="officehours">
              <Clock className="w-4 h-4 mr-2" />
              Office Hours
            </TabsTrigger>
          </TabsList>

          <TabsContent value="announcements" className="space-y-4">
            <BulkAnnouncementsTool />
          </TabsContent>

          <TabsContent value="rubrics" className="space-y-4">
            <GradingRubricsTool />
          </TabsContent>

          <TabsContent value="officehours" className="space-y-4">
            <OfficeHoursTool />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

function BulkAnnouncementsTool() {
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [newAnnouncement, setNewAnnouncement] = useState({
    title: '',
    content: '',
    priority: 'normal' as 'low' | 'normal' | 'high' | 'urgent',
    targetAudience: [] as string[],
    courseId: undefined as number | undefined,
  });

  const handleCreateAnnouncement = () => {
    // API call to create announcement
    setAnnouncements([...announcements, {
      id: Date.now(),
      title: newAnnouncement.title,
      content: newAnnouncement.content,
      priority: newAnnouncement.priority,
      targetAudience: newAnnouncement.targetAudience,
      courseId: newAnnouncement.courseId,
      publishedAt: new Date().toISOString(),
      authorId: 'faculty1',
      authorName: 'Faculty Member',
      collegeId: 'college1',
    }]);
    setIsDialogOpen(false);
    setNewAnnouncement({
      title: '',
      content: '',
      priority: 'normal',
      targetAudience: [],
      courseId: undefined,
    });
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
          {/* Quick Actions */}
          <div className="grid gap-4 md:grid-cols-3 mb-6">
            <Button variant="outline" className="h-auto py-4">
              <div className="text-left">
                <div className="font-semibold">Send to All Courses</div>
                <div className="text-sm text-muted-foreground">
                  Broadcast to all your courses
                </div>
              </div>
            </Button>
            <Button variant="outline" className="h-auto py-4">
              <div className="text-left">
                <div className="font-semibold">Department Alert</div>
                <div className="text-sm text-muted-foreground">
                  Send to department students
                </div>
              </div>
            </Button>
            <Button variant="outline" className="h-auto py-4">
              <div className="text-left">
                <div className="font-semibold">Urgent Notice</div>
                <div className="text-sm text-muted-foreground">
                  High priority alert
                </div>
              </div>
            </Button>
          </div>

          {/* Recent Announcements */}
          <h3 className="font-semibold mb-3">Recent Announcements</h3>
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
                      <p className="text-sm text-muted-foreground mb-2">
                        {announcement.content}
                      </p>
                      <div className="flex items-center gap-4 text-xs text-muted-foreground">
                        <span>By {announcement.authorName}</span>
                        <span>{new Date(announcement.publishedAt).toLocaleString()}</span>
                        <span className="flex items-center gap-1">
                          <Users className="w-3 h-3" />
                          {announcement.targetAudience.length} recipients
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

      {/* Create Announcement Dialog */}
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
                    onChange={(e) => setNewAnnouncement({ ...newAnnouncement, priority: e.target.value as any })}
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
                    onChange={(e) => setNewAnnouncement({ ...newAnnouncement, courseId: e.target.value ? parseInt(e.target.value) : undefined })}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    <option value="">All Courses</option>
                    <option value="1">Computer Science 101</option>
                    <option value="2">Mathematics 201</option>
                  </select>
                </div>
              </div>

              <div className="space-y-2">
                <Label>Target Audience</Label>
                <div className="grid gap-2 md:grid-cols-2">
                  {['all', 'students', 'faculty', 'department'].map((audience) => (
                    <label key={audience} className="flex items-center gap-2 p-2 border rounded cursor-pointer hover:bg-muted/50">
                      <input
                        type="checkbox"
                        checked={newAnnouncement.targetAudience.includes(audience)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setNewAnnouncement({ ...newAnnouncement, targetAudience: [...newAnnouncement.targetAudience, audience] });
                          } else {
                            setNewAnnouncement({ ...newAnnouncement, targetAudience: newAnnouncement.targetAudience.filter(a => a !== audience) });
                          }
                        }}
                        className="rounded"
                      />
                      <span className="capitalize">{audience}</span>
                    </label>
                  ))}
                </div>
              </div>

              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={handleCreateAnnouncement}>
                  <Send className="w-4 h-4 mr-2" />
                  Send Announcement
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
  const [rubrics, setRubrics] = useState([
    {
      id: 1,
      name: 'Assignment Rubric',
      criteria: [
        { name: 'Content', weight: 40, description: 'Quality and depth of content' },
        { name: 'Organization', weight: 30, description: 'Structure and flow' },
        { name: 'Grammar', weight: 20, description: 'Language usage' },
        { name: 'Formatting', weight: 10, description: 'Adherence to guidelines' },
      ],
      createdAt: new Date().toISOString(),
    }
  ]);

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <FileText className="w-5 h-5" />
            Grading Rubrics
          </CardTitle>
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            New Rubric
          </Button>
        </div>
        <CardDescription>
          Create and manage rubrics for consistent grading
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {rubrics.map((rubric) => (
            <div key={rubric.id} className="p-4 border rounded-lg">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-semibold text-lg">{rubric.name}</h4>
                <Badge>{new Date(rubric.createdAt).toLocaleDateString()}</Badge>
              </div>
              <div className="space-y-2">
                <h5 className="font-medium text-sm text-muted-foreground">Criteria</h5>
                <div className="grid gap-2 md:grid-cols-2">
                  {rubric.criteria.map((criterion, idx) => (
                    <div key={idx} className="p-3 bg-muted/50 rounded">
                      <div className="flex items-center justify-between mb-1">
                        <span className="font-medium">{criterion.name}</span>
                        <Badge variant="outline">{criterion.weight}%</Badge>
                      </div>
                      <p className="text-sm text-muted-foreground">{criterion.description}</p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function OfficeHoursTool() {
  const [officeHours, setOfficeHours] = useState([
    {
      id: 1,
      day: 'Monday',
      startTime: '09:00',
      endTime: '11:00',
      location: 'Room 301, Building A',
      isVirtual: false,
      meetingLink: '',
    },
    {
      id: 2,
      day: 'Wednesday',
      startTime: '14:00',
      endTime: '16:00',
      location: 'Room 301, Building A',
      isVirtual: false,
      meetingLink: '',
    },
    {
      id: 3,
      day: 'Friday',
      startTime: '10:00',
      endTime: '12:00',
      location: '',
      isVirtual: true,
      meetingLink: 'https://meet.google.com/xyz',
    },
  ]);

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Clock className="w-5 h-5" />
            Office Hours
          </CardTitle>
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Add Slot
          </Button>
        </div>
        <CardDescription>
          Manage your availability for student consultations
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {officeHours.map((slot) => (
            <div key={slot.id} className="p-4 border rounded-lg">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <h4 className="font-semibold text-lg">{slot.day}</h4>
                    <Badge variant={slot.isVirtual ? 'default' : 'secondary'}>
                      {slot.isVirtual ? 'Virtual' : 'In-Person'}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <div className="flex items-center gap-1">
                      <Clock className="w-4 h-4" />
                      {slot.startTime} - {slot.endTime}
                    </div>
                    {!slot.isVirtual && (
                      <div className="flex items-center gap-1">
                        <Calendar className="w-4 h-4" />
                        {slot.location}
                      </div>
                    )}
                  </div>
                  {slot.isVirtual && slot.meetingLink && (
                    <a 
                      href={slot.meetingLink}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-primary hover:underline mt-2 inline-block"
                    >
                      Join Meeting
                    </a>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
