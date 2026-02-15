"use client";

import { useState, useEffect, useCallback, useRef } from 'react';
import { useAuth } from '@/lib/auth-context';
import { logger } from '@/lib/logger';
import { api, endpoints } from '@/lib/api-client';
import type { Notification } from '@/lib/types';
import {
    Bell,
    Check,
    CheckCheck,
    Trash2,
    Info,
    AlertTriangle,
    CheckCircle2,
    XCircle,
    MoreVertical,
    Settings,
    Filter
} from 'lucide-react';
import {
    Card,
    CardContent,
    CardHeader
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { cn, formatDistanceToNow } from '@/lib/utils';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

export default function NotificationsPage() {
    const { user } = useAuth();
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [unreadCount, setUnreadCount] = useState(0);
    const [isLoading, setIsLoading] = useState(true);
    const [activeTab, setActiveTab] = useState('all');
    const [typeFilter, setTypeFilter] = useState<'all' | Notification['type']>('all');
    const wsRef = useRef<WebSocket | null>(null);

    const fetchNotifications = useCallback(async () => {
        try {
            const data = await api.get<Notification[]>(endpoints.notifications.list);
            setNotifications(data || []);

            const countData = await api.get<{ unread_count?: number; unreadCount?: number }>(endpoints.notifications.unreadCount);
            setUnreadCount(countData.unread_count ?? countData.unreadCount ?? 0);
        } catch (error) {
            logger.error('Failed to fetch notifications:', error as Error);
        } finally {
            setIsLoading(false);
        }
    }, []);

    const setupWebSocket = useCallback(() => {
        if (!user) return;

        const wsUrl = `${process.env.NEXT_PUBLIC_API_URL?.replace(/^http/, 'ws')}/api/notifications/ws`;

        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onmessage = (event) => {
            try {
                const newNotification = JSON.parse(event.data) as Notification;
                setNotifications(prev => [newNotification, ...prev]);
                setUnreadCount(prev => prev + 1);
            } catch (err) {
                logger.error('WebSocket message parsing error:', err as Error);
            }
        };

        ws.onclose = () => {
            logger.info('WebSocket connection closed. Reconnecting...');
            setTimeout(setupWebSocket, 3000);
        };

        ws.onerror = (err) => {
            logger.error('WebSocket error:', err as unknown as Error);
        };

        return () => {
            ws.close();
        };
    }, [user]);

    useEffect(() => {
        if (user) {
            fetchNotifications();
            const cleanup = setupWebSocket();
            return cleanup;
        }
        return undefined;
    }, [user, fetchNotifications, setupWebSocket]);

    const markAsRead = async (id: number) => {
        try {
            await api.patch(endpoints.notifications.markAsRead(id), {});
            setNotifications(prev =>
                prev.map(n => n.id === id ? { ...n, isRead: true } : n)
            );
            setUnreadCount(prev => Math.max(0, prev - 1));
        } catch (error) {
            logger.error('Failed to mark notification as read:', error as Error);
        }
    };

    const markAllAsRead = async () => {
        try {
            await api.post(endpoints.notifications.markAllAsRead, {});
            setNotifications(prev => prev.map(n => ({ ...n, isRead: true })));
            setUnreadCount(0);
        } catch (error) {
            logger.error('Failed to mark all as read:', error as Error);
        }
    };

    const deleteNotification = async (id: number) => {
        try {
            await api.delete(endpoints.notifications.delete(id));
            setNotifications(prev => {
                const n = prev.find(item => item.id === id);
                if (n && !n.isRead) {
                    setUnreadCount(count => Math.max(0, count - 1));
                }
                return prev.filter(item => item.id !== id);
            });
        } catch (error) {
            logger.error('Failed to delete notification:', error as Error);
        }
    };

    const filteredNotifications = activeTab === 'unread'
        ? notifications.filter(n => !n.isRead)
        : notifications;
    const displayedNotifications = typeFilter === 'all'
        ? filteredNotifications
        : filteredNotifications.filter((notification) => notification.type === typeFilter);
    const cycleFilter = () => {
        const order: Array<'all' | Notification['type']> = ['all', 'info', 'success', 'warning', 'error'];
        const next = order[(order.indexOf(typeFilter) + 1) % order.length];
        setTypeFilter(next);
    };

    const getTypeIcon = (type: Notification['type']) => {
        switch (type) {
            case 'success': return <CheckCircle2 className="w-5 h-5 text-green-500" />;
            case 'warning': return <AlertTriangle className="w-5 h-5 text-yellow-500" />;
            case 'error': return <XCircle className="w-5 h-5 text-red-500" />;
            default: return <Info className="w-5 h-5 text-blue-500" />;
        }
    };

    if (isLoading) {
        return (
            <div className="container mx-auto p-6 space-y-4">
                <div className="h-8 w-48 bg-muted animate-pulse rounded" />
                <div className="h-[400px] w-full bg-muted animate-pulse rounded" />
            </div>
        );
    }

    return (
        <div className="container mx-auto p-6 max-w-4xl space-y-6">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10 rounded-full">
                        <Bell className="w-6 h-6 text-primary" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">Notifications</h1>
                        <p className="text-muted-foreground">Keep track of your latest updates and alerts</p>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    {unreadCount > 0 && (
                        <Button variant="outline" size="sm" onClick={markAllAsRead}>
                            <CheckCheck className="w-4 h-4 mr-2" />
                            Mark all as read
                        </Button>
                    )}
                    <Button variant="ghost" size="icon" onClick={fetchNotifications} title="Refresh notifications">
                        <Settings className="w-5 h-5" />
                    </Button>
                </div>
            </div>

            <Card className="border-none shadow-md bg-background/50 backdrop-blur-sm">
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                            <div className="flex items-center justify-between w-full">
                                <TabsList className="grid w-[300px] grid-cols-2">
                                    <TabsTrigger value="all" className="relative">
                                        All
                                        {notifications.length > 0 && (
                                            <Badge variant="secondary" className="ml-2 bg-muted/50">
                                                {notifications.length}
                                            </Badge>
                                        )}
                                    </TabsTrigger>
                                    <TabsTrigger value="unread">
                                        Unread
                                        {unreadCount > 0 && (
                                            <Badge variant="default" className="ml-2 animate-pulse">
                                                {unreadCount}
                                            </Badge>
                                        )}
                                    </TabsTrigger>
                                </TabsList>
                                <div className="flex items-center gap-2">
                                    <Button variant="ghost" size="sm" className="text-xs" onClick={cycleFilter}>
                                        <Filter className="w-3 h-3 mr-2" />
                                        Filter: {typeFilter}
                                    </Button>
                                </div>
                            </div>
                        </Tabs>
                    </div>
                </CardHeader>
                <CardContent className="px-0">
                    <div className="divide-y divide-border/50">
                        {displayedNotifications.length === 0 ? (
                            <div className="flex flex-col items-center justify-center py-20 text-center space-y-4">
                                <div className="p-4 bg-muted/20 rounded-full">
                                    <Bell className="w-12 h-12 text-muted-foreground/50" />
                                </div>
                                <div>
                                    <h3 className="text-lg font-medium">No notifications</h3>
                                    <p className="text-sm text-muted-foreground max-w-xs mx-auto">
                                        {activeTab === 'unread'
                                            ? "You've caught up with everything! No unread notifications found."
                                            : typeFilter !== 'all'
                                                ? `No ${typeFilter} notifications found.`
                                                : "When you receive notifications, they'll appear here."}
                                    </p>
                                </div>
                            </div>
                        ) : (
                            displayedNotifications.map((notification) => (
                                <div
                                    key={notification.id}
                                    className={cn(
                                        "flex items-start gap-4 p-4 transition-colors hover:bg-muted/30 group",
                                        !notification.isRead && "bg-primary/5"
                                    )}
                                >
                                    <div className="mt-1 shrink-0">
                                        {getTypeIcon(notification.type)}
                                    </div>
                                    <div className="flex-1 space-y-1 min-w-0">
                                        <div className="flex items-center justify-between gap-2">
                                            <h4 className={cn(
                                                "text-sm font-semibold truncate",
                                                !notification.isRead ? "text-primary" : "text-foreground"
                                            )}>
                                                {notification.title}
                                            </h4>
                                            <span className="text-[10px] text-muted-foreground whitespace-nowrap">
                                                {formatDistanceToNow(notification.createdAt)}
                                            </span>
                                        </div>
                                        <p className="text-sm text-muted-foreground line-clamp-2 leading-relaxed">
                                            {notification.message}
                                        </p>
                                        {notification.metadata && Object.keys(notification.metadata).length > 0 && (
                                            <div className="flex flex-wrap gap-2 mt-2">
                                                {notification.metadata.courseName && (
                                                    <Badge variant="outline" className="text-[10px] py-0">
                                                        {notification.metadata.courseName}
                                                    </Badge>
                                                )}
                                                {notification.category && (
                                                    <Badge variant="secondary" className="text-[10px] py-0 capitalize">
                                                        {notification.category}
                                                    </Badge>
                                                )}
                                            </div>
                                        )}
                                    </div>
                                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0">
                                        {!notification.isRead && (
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-8 w-8 hover:bg-primary/10 hover:text-primary"
                                                onClick={() => markAsRead(notification.id)}
                                                title="Mark as read"
                                            >
                                                <Check className="w-4 h-4" />
                                            </Button>
                                        )}
                                        <DropdownMenu>
                                            <DropdownMenuTrigger asChild>
                                                <Button variant="ghost" size="icon" className="h-8 w-8">
                                                    <MoreVertical className="w-4 h-4" />
                                                </Button>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent align="end">
                                                <DropdownMenuItem
                                                    className="text-destructive focus:bg-destructive/10 focus:text-destructive"
                                                    onClick={() => deleteNotification(notification.id)}
                                                >
                                                    <Trash2 className="w-4 h-4 mr-2" />
                                                    Delete
                                                </DropdownMenuItem>
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                    </div>
                                </div>
                            ))
                        )}
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
