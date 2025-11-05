"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api, endpoints } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Search, Pin, Clock, AlertCircle, Loader2 } from "lucide-react";
import { format } from "date-fns";
import { logger } from '@/lib/logger';

type Announcement = {
  id: number;
  title: string;
  content: string;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  author: string;
  authorRole: string;
  publishedAt: string;
  courseName?: string;
  departmentName?: string;
  isPinned?: boolean;
  targetAudience: string[];
};

export default function AnnouncementsPage() {
  const { user } = useAuth();
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedPriority, setSelectedPriority] = useState<string>("all");
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [formData, setFormData] = useState({
    title: "",
    content: "",
    priority: "normal",
  });

  useEffect(() => {
    const fetchAnnouncements = async () => {
      try {
        setLoading(true);
        const response = await api.get('/api/announcements');
        setAnnouncements(Array.isArray(response) ? response : []);
      } catch (err) {
        logger.error('Failed to fetch announcements:', err as Error);
        setError('Failed to load announcements');
      } finally {
        setLoading(false);
      }
    };

    fetchAnnouncements();
  }, []);

  const handleCreateAnnouncement = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.title.trim() || !formData.content.trim()) {
      return;
    }

    try {
      setCreating(true);
      const newAnnouncement = await api.post(endpoints.announcements.create, {
        title: formData.title,
        content: formData.content,
        priority: formData.priority,
        is_published: true,
        published_at: new Date().toISOString(),
      }) as Announcement;

      setAnnouncements(prev => [newAnnouncement, ...prev]);
      setFormData({ title: "", content: "", priority: "normal" });
      setDialogOpen(false);
    } catch (err) {
      logger.error('Failed to create announcement:', err as Error);
      setError('Failed to create announcement');
    } finally {
      setCreating(false);
    }
  };

  const getPriorityBadge = (priority: string) => {
    const config = {
      low: { className: 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400', label: 'Low' },
      normal: { className: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400', label: 'Normal' },
      high: { className: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400', label: 'High' },
      urgent: { className: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400', label: 'Urgent' }
    };
    const { className, label } = config[priority as keyof typeof config];
    return (
      <Badge className={className}>
        {priority === 'urgent' && <AlertCircle className="mr-1 h-3 w-3" />}
        {label}
      </Badge>
    );
  };

  const filteredAnnouncements = announcements.filter(announcement => {
    const matchesSearch = 
      announcement.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      announcement.content.toLowerCase().includes(searchQuery.toLowerCase()) ||
      announcement.author.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesPriority = 
      selectedPriority === "all" || 
      announcement.priority === selectedPriority;
    
    return matchesSearch && matchesPriority;
  });

  const pinnedAnnouncements = filteredAnnouncements.filter(a => a.isPinned);
  const regularAnnouncements = filteredAnnouncements.filter(a => !a.isPinned);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Announcements</h1>
          <p className="text-muted-foreground">
            Stay updated with the latest news and updates
          </p>
        </div>
        {(user?.role === 'faculty' || user?.role === 'admin') && (
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                New Announcement
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[525px]">
              <form onSubmit={handleCreateAnnouncement}>
                <DialogHeader>
                  <DialogTitle>Create Announcement</DialogTitle>
                  <DialogDescription>
                    Share important information with students and faculty
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="title">Title</Label>
                    <Input
                      id="title"
                      value={formData.title}
                      onChange={(e) => setFormData(prev => ({ ...prev, title: e.target.value }))}
                      placeholder="Enter announcement title"
                      required
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="content">Content</Label>
                    <Textarea
                      id="content"
                      value={formData.content}
                      onChange={(e) => setFormData(prev => ({ ...prev, content: e.target.value }))}
                      placeholder="Enter announcement content"
                      required
                      rows={5}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="priority">Priority</Label>
                    <Select
                      value={formData.priority}
                      onValueChange={(value) => setFormData(prev => ({ ...prev, priority: value }))}
                    >
                      <SelectTrigger id="priority">
                        <SelectValue placeholder="Select priority" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="low">Low</SelectItem>
                        <SelectItem value="normal">Normal</SelectItem>
                        <SelectItem value="high">High</SelectItem>
                        <SelectItem value="urgent">Urgent</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setDialogOpen(false)} disabled={creating}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={creating}>
                    {creating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    {creating ? 'Creating...' : 'Create Announcement'}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        )}
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Announcements</CardDescription>
            <CardTitle className="text-2xl">{announcements.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Urgent</CardDescription>
            <CardTitle className="text-2xl text-red-600">
              {announcements.filter(a => a.priority === 'urgent').length}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Pinned</CardDescription>
            <CardTitle className="text-2xl text-blue-600">
              {announcements.filter(a => a.isPinned).length}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Today</CardDescription>
            <CardTitle className="text-2xl">
              {announcements.filter(a => {
                const today = new Date();
                const published = new Date(a.publishedAt);
                return today.toDateString() === published.toDateString();
              }).length}
            </CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search announcements..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex gap-2">
          {["all", "urgent", "high", "normal", "low"].map((priority) => (
            <Button
              key={priority}
              variant={selectedPriority === priority ? "default" : "outline"}
              size="sm"
              onClick={() => setSelectedPriority(priority)}
            >
              {priority.charAt(0).toUpperCase() + priority.slice(1)}
            </Button>
          ))}
        </div>
      </div>

      {pinnedAnnouncements.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-sm font-medium">
            <Pin className="h-4 w-4" />
            Pinned Announcements
          </div>
          {pinnedAnnouncements.map((announcement) => (
            <Card key={announcement.id} className="border-primary/20 bg-primary/5">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-3 flex-1">
                    <Avatar className="h-10 w-10">
                      <AvatarFallback>
                        {announcement.author.split(' ').map(n => n[0]).join('').toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <CardTitle className="text-lg">{announcement.title}</CardTitle>
                        {getPriorityBadge(announcement.priority)}
                        <Pin className="h-4 w-4 ml-auto text-muted-foreground" />
                      </div>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <span>{announcement.author}</span>
                        <span>•</span>
                        <span className="capitalize">{announcement.authorRole}</span>
                        {announcement.courseName && (
                          <>
                            <span>•</span>
                            <span>{announcement.courseName}</span>
                          </>
                        )}
                        <span>•</span>
                        <div className="flex items-center gap-1">
                          <Clock className="h-3 w-3" />
                          {format(new Date(announcement.publishedAt), 'MMM dd, HH:mm')}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">{announcement.content}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <div className="space-y-4">
        {pinnedAnnouncements.length > 0 && (
          <div className="text-sm font-medium">All Announcements</div>
        )}
        {regularAnnouncements.map((announcement) => (
          <Card key={announcement.id} className="hover:shadow-md transition-shadow">
            <CardHeader>
              <div className="flex items-start gap-3">
                <Avatar className="h-10 w-10">
                  <AvatarFallback>
                    {announcement.author.split(' ').map(n => n[0]).join('').toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <CardTitle className="text-lg">{announcement.title}</CardTitle>
                    {getPriorityBadge(announcement.priority)}
                  </div>
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <span>{announcement.author}</span>
                    <span>•</span>
                    <span className="capitalize">{announcement.authorRole}</span>
                    {announcement.courseName && (
                      <>
                        <span>•</span>
                        <span>{announcement.courseName}</span>
                      </>
                    )}
                    {announcement.departmentName && (
                      <>
                        <span>•</span>
                        <span>{announcement.departmentName}</span>
                      </>
                    )}
                    <span>•</span>
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      {format(new Date(announcement.publishedAt), 'MMM dd, HH:mm')}
                    </div>
                  </div>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">{announcement.content}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
