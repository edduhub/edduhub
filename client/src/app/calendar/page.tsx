"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Plus, Calendar as CalendarIcon, Clock, MapPin, BookOpen, Loader2 } from "lucide-react";
import { format, startOfMonth, endOfMonth, eachDayOfInterval, isSameMonth, isToday, parseISO } from "date-fns";
import { logger } from '@/lib/logger';

type CalendarEvent = {
  id: number;
  title: string;
  type: 'exam' | 'event' | 'holiday' | 'deadline' | 'other';
  start: string;
  end: string;
  courseName?: string;
  location?: string;
  description?: string;
};

type CalendarEventApi = {
  id?: number;
  title?: string;
  description?: string;
  event_type?: CalendarEvent['type'];
  type?: CalendarEvent['type'];
  date?: string;
  start?: string;
  start_time?: string;
  end?: string;
  end_time?: string;
  course_name?: string;
  courseName?: string;
  location?: string;
};

export default function CalendarPage() {
  const { user } = useAuth();
  const [currentDate, setCurrentDate] = useState(new Date());
  const [events, setEvents] = useState<CalendarEvent[]>([]);
  const [error, setError] = useState<string | null>(null);

  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newEvent, setNewEvent] = useState({
    title: '',
    type: 'event',
    start: '',
    end: '',
    location: '',
    description: '',
  });

  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const response = await api.get<CalendarEventApi[]>('/api/calendar');
        const normalized = Array.isArray(response)
          ? response.map((event) => {
              const start = event.start ?? event.start_time ?? event.date ?? new Date().toISOString();
              const end = event.end ?? event.end_time ?? start;
              const type = event.event_type ?? event.type ?? 'event';
              return {
                id: event.id ?? 0,
                title: event.title ?? 'Untitled Event',
                description: event.description,
                type: type === 'exam' || type === 'event' || type === 'holiday' || type === 'deadline' || type === 'other'
                  ? type
                  : 'event',
                start,
                end,
                courseName: event.courseName ?? event.course_name,
                location: event.location,
              };
            })
          : [];
        setEvents(normalized);
      } catch (err) {
        logger.error('Failed to fetch calendar events:', err as Error);
        setError('Failed to load calendar events');
      }
    };

    fetchEvents();
  }, []);

  const monthStart = startOfMonth(currentDate);
  const monthEnd = endOfMonth(currentDate);
  const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });

  const getEventBadge = (type: string) => {
    const config = {
      exam: { className: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400', label: 'Exam', icon: CalendarIcon },
      event: { className: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400', label: 'Event', icon: CalendarIcon },
      holiday: { className: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400', label: 'Holiday', icon: CalendarIcon },
      deadline: { className: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400', label: 'Deadline', icon: Clock },
      other: { className: 'bg-slate-100 text-slate-800 dark:bg-slate-900/30 dark:text-slate-400', label: 'Other', icon: CalendarIcon },
    };
    const selected = config[type as keyof typeof config] || config.other;
    const { className, label, icon: Icon } = selected;
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

  const createEvent = async () => {
    try {
      setCreating(true);
      setError(null);

      const startValue = newEvent.start || new Date().toISOString();
      await api.post('/api/calendar', {
        title: newEvent.title,
        description: newEvent.description,
        event_type: newEvent.type,
        date: new Date(startValue).toISOString(),
      });

      const data = await api.get<CalendarEventApi[]>('/api/calendar');
      const normalized = Array.isArray(data)
        ? data.map((event) => {
            const start = event.start ?? event.start_time ?? event.date ?? new Date().toISOString();
            const end = event.end ?? event.end_time ?? start;
            const type = event.event_type ?? event.type ?? 'event';
            return {
              id: event.id ?? 0,
              title: event.title ?? 'Untitled Event',
              description: event.description,
              type: type === 'exam' || type === 'event' || type === 'holiday' || type === 'deadline' || type === 'other'
                ? type
                : 'event',
              start,
              end,
              courseName: event.courseName ?? event.course_name,
              location: event.location,
            };
          })
        : [];
      setEvents(normalized);
      setShowCreate(false);
      setNewEvent({ title: '', type: 'event', start: '', end: '', location: '', description: '' });
    } catch (e) {
      logger.error('Error occurred', e as Error);
      setError('Failed to create event');
    } finally {
      setCreating(false);
    }
  };

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
          <Button onClick={() => setShowCreate(v => !v)}>
            <Plus className="mr-2 h-4 w-4" />
            {showCreate ? 'Close' : 'Add Event'}
          </Button>
        )}
      </div>

      {showCreate && (
        <Card>
          <CardHeader>
            <CardTitle>New Event</CardTitle>
            <CardDescription>Provide event details</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <label className="text-sm font-medium">Title</label>
                <input className="w-full rounded-md border px-3 py-2" value={newEvent.title} onChange={e => setNewEvent({ ...newEvent, title: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Type</label>
                <select
                  className="w-full rounded-md border px-3 py-2 bg-background"
                  value={newEvent.type}
                  onChange={e => setNewEvent({ ...newEvent, type: e.target.value as 'exam' | 'event' | 'holiday' | 'deadline' | 'other' })}
                >
                  <option value="event">Event</option>
                  <option value="exam">Exam</option>
                  <option value="holiday">Holiday</option>
                  <option value="deadline">Deadline</option>
                  <option value="other">Other</option>
                </select>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Start</label>
                <input type="datetime-local" className="w-full rounded-md border px-3 py-2" value={newEvent.start} onChange={e => setNewEvent({ ...newEvent, start: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">End</label>
                <input type="datetime-local" className="w-full rounded-md border px-3 py-2" value={newEvent.end} onChange={e => setNewEvent({ ...newEvent, end: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Location</label>
                <input className="w-full rounded-md border px-3 py-2" value={newEvent.location} onChange={e => setNewEvent({ ...newEvent, location: e.target.value })} />
              </div>
              <div className="space-y-2 sm:col-span-2">
                <label className="text-sm font-medium">Description</label>
                <input className="w-full rounded-md border px-3 py-2" value={newEvent.description} onChange={e => setNewEvent({ ...newEvent, description: e.target.value })} />
              </div>
            </div>
            <div className="mt-4 flex justify-end">
              <Button onClick={createEvent} disabled={creating}>
                {creating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Plus className="mr-2 h-4 w-4" />}
                Create
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {error && (
        <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">{error}</div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Total Events</CardDescription>
            <CardTitle className="text-2xl">{events.length}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardDescription>Events</CardDescription>
              <CardTitle className="text-2xl text-blue-600">
              {events.filter(e => e.type === 'event').length}
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
                  onClick={() => setCurrentDate((prev) => {
                    const next = new Date(prev);
                    next.setMonth(prev.getMonth() - 1);
                    return next;
                  })}
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
                  onClick={() => setCurrentDate((prev) => {
                    const next = new Date(prev);
                    next.setMonth(prev.getMonth() + 1);
                    return next;
                  })}
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
