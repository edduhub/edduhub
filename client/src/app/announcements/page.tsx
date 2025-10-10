"use client";

import { useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Plus, Search, Pin, Clock, AlertCircle } from "lucide-react";
import { format } from "date-fns";

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

  const [announcements] = useState<Announcement[]>([
    {
      id: 1,
      title: "Campus-wide Internet Maintenance",
      content: "The campus internet will undergo scheduled maintenance on Saturday, March 23rd from 2 AM to 6 AM. Services will be unavailable during this time.",
      priority: 'urgent',
      author: "IT Department",
      authorRole: "Admin",
      publishedAt: new Date().toISOString(),
      targetAudience: ['all'],
      isPinned: true
    },
    {
      id: 2,
      title: "Guest Lecture on AI Ethics",
      content: "Dr. Sarah Johnson from MIT will be delivering a guest lecture on AI Ethics and Responsible Computing. All students are encouraged to attend.",
      priority: 'high',
      author: "Dr. Rajesh Kumar",
      authorRole: "Faculty",
      publishedAt: new Date(Date.now() - 3600000).toISOString(),
      courseName: "CS401 - Machine Learning",
      targetAudience: ['students', 'faculty']
    },
    {
      id: 3,
      title: "Assignment Submission Deadline Extended",
      content: "The submission deadline for Assignment 3 has been extended to March 25th due to technical issues with the submission portal.",
      priority: 'normal',
      author: "Prof. Priya Sharma",
      authorRole: "Faculty",
      publishedAt: new Date(Date.now() - 7200000).toISOString(),
      courseName: "CS201 - Data Structures",
      targetAudience: ['students']
    },
    {
      id: 4,
      title: "Hackathon Registration Open",
      content: "Annual college hackathon registration is now open! Team size 2-4 members. Amazing prizes to be won. Register before March 30th.",
      priority: 'normal',
      author: "Student Council",
      authorRole: "Student",
      publishedAt: new Date(Date.now() - 86400000).toISOString(),
      departmentName: "Computer Science",
      targetAudience: ['students'],
      isPinned: true
    },
    {
      id: 5,
      title: "Library Timings Update",
      content: "The library will now be open until 11 PM on weekdays starting next week. Weekend timings remain unchanged.",
      priority: 'low',
      author: "Library Administration",
      authorRole: "Admin",
      publishedAt: new Date(Date.now() - 172800000).toISOString(),
      targetAudience: ['all']
    }
  ]);

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
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            New Announcement
          </Button>
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