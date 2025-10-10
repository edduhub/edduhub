"use client";

import { useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Calendar as CalendarIcon, Clock, MapPin, BookOpen } from "lucide-react";
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameMonth, isToday, parseISO } from "date-fns";

type CalendarEvent = {
  id: number;
  title: string;
  type: 'lecture' | 'exam' | 'event' | 'holiday' | 'deadline';
  start: string;
  end: string;
  courseName?: string;
  location?: string;
  description?: string;
};

export default function CalendarPage() {
  const { user } = useAuth();
  const [currentDate, setCurrentDate] = useState(new Date());
  const [selectedDate, setSelectedDate] = useState<Date | null>(null);

  const [events] = useState<CalendarEvent[]>([
    {
      id: 1,
      title: "Data Structures Lecture",
      type: 'lecture',
      start: new Date().toISOString(),
      end: new Date(Date.now() + 5400000).toISOString(),
      courseName: "CS201",
      location: "Room 301"
    },
    {
      id: 2,
      title: "Database Systems Midterm",
      type: 'exam',
      start: new Date(Date.now() + 2 * 86400000).toISOString(),
      end: new Date(Date.now() + 2 * 86400000 + 7200000).toISOString(),
      courseName: "CS305",
      location: "Exam Hall A"
    },
    {
      id: 3,
      title: "ML Assignment Due",
      type: 'deadline',
      start: new Date(Date.now() + 3 * 86400000).toISOString(),
      end: new Date(Date.now() + 3 * 86400000).toISOString(),
      courseName: "CS401"
    },
    {
      id: 4,
      title: "Tech Fest 2024",
      type: 'event',
      start: new Date(Date.now() + 7 * 86400000).toISOString(),
      end: new Date(Date.now() + 9 * 86400000).toISOString(),
      location: "Main Campus"
    },
    {
      id: 5,
      title: "Guest Lecture: AI Ethics",
      type: 'event',
      start: new Date(Date.now() + 5 * 86400000).toISOString(),
      end: new Date(Date.now() + 5 * 86400000 + 3600000).toISOString(),
      location: "Auditorium"
    }
  ]);

  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(currentDate);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });

  const getEventBadge = (type: string) => {
    const config = {
      lecture: { className: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400', label: 'Lecture', icon: BookOpen },
      exam: { className: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400', label: 'Exam', icon: CalendarIcon },
      event: { className: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400', label: 'Event', icon: CalendarIcon },
      holiday: { className: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400', label: 'Holiday', icon: CalendarIcon },
      deadline: { className: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400', label: 'Deadline', icon: Clock }
    };
    const { className, label, icon: Icon } = config[type as keyof typeof config];
    return (
      <Badge className={className}>
        <Icon className="mr-1 h-3 w-3" />
        {label}
      </Badge>
    );
  };

  const getEventsForDate = (date: Date) => {
    return events.filter(event => {
      const eventStart = new Date(event.start);
      return eventStart.toDateString() === date.toDateString();
    });
  };

  const upcomingEvents = events
    .filter(event => new Date(event.start) >= new Date())
    .sort((a, b) => new Date(a.start).getTime() - new Date(b.start).getTime())
    .slice(0, 5);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Calendar</h1>
          <p className="text-muted-foreground">
            View your schedule and upcoming events
          </p>
        </div>
        {(user?.role === 'faculty' || user?.role === 'admin') && (
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Add Event
          </Button>
        )}
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Events</CardDescription>
            <CardTitle className="text-2xl">{events.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Lectures</CardDescription>
            <CardTitle className="text-2xl text-blue-600">
              {events.filter(e => e.type === 'lecture').length}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Exams</CardDescription>
            <CardTitle className="text-2xl text-red-600">
              {events.filter(e => e.type === 'exam').length}
            </CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Deadlines</CardDescription>
            <CardTitle className="text-2xl text-orange-600">
              {events.filter(e => e.type === 'deadline').length}
            </CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>{format(currentDate, 'MMMM yyyy')}</CardTitle>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentDate(new Date(currentDate.setMonth(currentDate.getMonth() - 1)))}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentDate(new Date())}
                >
                  Today
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentDate(new Date(currentDate.setMonth(currentDate.getMonth() + 1)))}
                >
                  Next
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-7 gap-2">
              {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map(day => (
                <div key={day} className="text-center text-sm font-medium text-muted-foreground p-2">
                  {day}
                </div>
              ))}
              {daysInMonth.map((day, idx) => {
                const dayEvents = getEventsForDate(day);
                const hasEvents = dayEvents.length > 0;
                const isCurrentDay = isToday(day);
                
                return (
                  <button
                    key={idx}
                    onClick={() => setSelectedDate(day)}
                    className={`
                      min-h-[80px] p-2 rounded-lg border text-left transition-colors
                      ${!isSameMonth(day, currentDate) ? 'opacity-40' : ''}
                      ${isCurrentDay ? 'border-primary bg-primary/5' : 'border-border'}
                      ${hasEvents ? 'bg-accent/50' : ''}
                      hover:bg-accent cursor-pointer
                    `}
                  >
                    <div className={`text-sm font-medium mb-1 ${isCurrentDay ? 'text-primary' : ''}`}>
                      {format(day, 'd')}
                    </div>
                    {dayEvents.slice(0, 2).map(event => (
                      <div
                        key={event.id}
                        className="text-xs truncate mb-1 px-1 py-0.5 rounded bg-primary/10"
                      >
                        {event.title}
                      </div>
                    ))}
                    {dayEvents.length > 2 && (
                      <div className="text-xs text-muted-foreground">
                        +{dayEvents.length - 2} more
                      </div>
                    )}
                  </button>
                );
              })}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Upcoming Events</CardTitle>
            <CardDescription>Your next scheduled events</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {upcomingEvents.map((event) => (
                <div key={event.id} className="space-y-2 p-3 rounded-lg border">
                  <div className="flex items-start justify-between gap-2">
                    <h4 className="font-medium text-sm">{event.title}</h4>
                    {getEventBadge(event.type)}
                  </div>
                  <div className="space-y-1 text-xs text-muted-foreground">
                    <div className="flex items-center gap-2">
                      <CalendarIcon className="h-3 w-3" />
                      {format(parseISO(event.start), 'MMM dd, yyyy')}
                    </div>
                    <div className="flex items-center gap-2">
                      <Clock className="h-3 w-3" />
                      {format(parseISO(event.start), 'hh:mm a')} - {format(parseISO(event.end), 'hh:mm a')}
                    </div>
                    {event.location && (
                      <div className="flex items-center gap-2">
                        <MapPin className="h-3 w-3" />
                        {event.location}
                      </div>
                    )}
                    {event.courseName && (
                      <div className="flex items-center gap-2">
                        <BookOpen className="h-3 w-3" />
                        {event.courseName}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}