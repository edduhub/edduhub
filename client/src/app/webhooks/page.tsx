"use client";

import { useState, useEffect } from "react";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Plus, Loader2, AlertCircle, CheckCircle, Trash2, Edit, Send } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { logger } from '@/lib/logger';

type Webhook = {
  id: number;
  url: string;
  event: string;
  active: boolean;
  secret: string;
  createdAt: string;
  updatedAt: string;
};

type WebhookApi = {
  id?: number;
  url?: string;
  event?: string;
  event_type?: string;
  active?: boolean;
  is_active?: boolean;
  secret?: string;
  created_at?: string;
  updated_at?: string;
};

const EVENT_TYPES = [
  { value: "student.created", label: "Student Created" },
  { value: "student.updated", label: "Student Updated" },
  { value: "student.deleted", label: "Student Deleted" },
  { value: "course.created", label: "Course Created" },
  { value: "course.updated", label: "Course Updated" },
  { value: "grade.submitted", label: "Grade Submitted" },
  { value: "assignment.created", label: "Assignment Created" },
  { value: "attendance.marked", label: "Attendance Marked" },
];

const normalizeWebhook = (webhook: WebhookApi): Webhook => ({
  id: webhook.id ?? 0,
  url: webhook.url ?? "",
  event: webhook.event ?? webhook.event_type ?? "",
  active: webhook.active ?? webhook.is_active ?? true,
  secret: webhook.secret ?? "",
  createdAt: webhook.created_at ?? new Date().toISOString(),
  updatedAt: webhook.updated_at ?? new Date().toISOString(),
});

export default function WebhooksPage() {
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingWebhook, setEditingWebhook] = useState<Webhook | null>(null);
  const [formData, setFormData] = useState({
    url: "",
    event: "",
    active: true,
    secret: "",
  });
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    void fetchWebhooks();
  }, []);

  const fetchWebhooks = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get<WebhookApi[]>('/api/webhooks');
      setWebhooks(Array.isArray(response) ? response.map(normalizeWebhook) : []);
    } catch (fetchError) {
      logger.error('Failed to fetch webhooks:', fetchError as Error);
      setError('Failed to load webhooks');
      setWebhooks([]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    try {
      setSubmitting(true);
      setError(null);

      const payload: Record<string, string | boolean> = {
        url: formData.url,
        event: formData.event,
        active: formData.active,
      };
      if (formData.secret.trim()) {
        payload.secret = formData.secret.trim();
      }

      await api.post('/api/webhooks', payload);

      setSuccess('Webhook created successfully');
      setDialogOpen(false);
      resetForm();
      await fetchWebhooks();
    } catch (createError) {
      setError(createError instanceof Error ? createError.message : 'Failed to create webhook');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdate = async () => {
    if (!editingWebhook) return;

    try {
      setSubmitting(true);
      setError(null);

      await api.patch(`/api/webhooks/${editingWebhook.id}`, {
        url: formData.url,
        event: formData.event,
        active: formData.active,
        secret: formData.secret,
      });

      setSuccess('Webhook updated successfully');
      setDialogOpen(false);
      resetForm();
      await fetchWebhooks();
    } catch (updateError) {
      setError(updateError instanceof Error ? updateError.message : 'Failed to update webhook');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this webhook?')) return;

    try {
      setError(null);
      await api.delete(`/api/webhooks/${id}`);
      setSuccess('Webhook deleted successfully');
      await fetchWebhooks();
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : 'Failed to delete webhook');
    }
  };

  const handleTest = async (id: number) => {
    try {
      setError(null);
      await api.post(`/api/webhooks/${id}/test`, {});
      setSuccess('Test event sent successfully');
    } catch (testError) {
      setError(testError instanceof Error ? testError.message : 'Failed to send test event');
    }
  };

  const handleToggleActive = async (webhook: Webhook) => {
    try {
      setError(null);
      await api.patch(`/api/webhooks/${webhook.id}`, {
        url: webhook.url,
        event: webhook.event,
        active: !webhook.active,
        secret: webhook.secret,
      });
      await fetchWebhooks();
    } catch (toggleError) {
      setError(toggleError instanceof Error ? toggleError.message : 'Failed to update webhook');
    }
  };

  const openCreateDialog = () => {
    resetForm();
    setEditingWebhook(null);
    setDialogOpen(true);
  };

  const openEditDialog = (webhook: Webhook) => {
    setFormData({
      url: webhook.url,
      event: webhook.event,
      active: webhook.active,
      secret: webhook.secret,
    });
    setEditingWebhook(webhook);
    setDialogOpen(true);
  };

  const resetForm = () => {
    setFormData({
      url: "",
      event: "",
      active: true,
      secret: "",
    });
    setEditingWebhook(null);
  };

  const getEventTypeLabel = (type: string) => {
    return EVENT_TYPES.find((eventType) => eventType.value === type)?.label || type;
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Webhook Management</h1>
          <p className="text-muted-foreground">
            Configure webhooks to receive real-time event notifications
          </p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button onClick={openCreateDialog}>
              <Plus className="mr-2 h-4 w-4" />
              Add Webhook
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>
                {editingWebhook ? 'Edit Webhook' : 'Create Webhook'}
              </DialogTitle>
              <DialogDescription>
                {editingWebhook
                  ? 'Update the webhook configuration'
                  : 'Add a new webhook to receive event notifications'}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="url">Webhook URL</Label>
                <Input
                  id="url"
                  placeholder="https://example.com/webhook"
                  value={formData.url}
                  onChange={(event) => setFormData({ ...formData, url: event.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="event">Event</Label>
                <Select
                  value={formData.event}
                  onValueChange={(value) => setFormData({ ...formData, event: value })}
                >
                  <SelectTrigger id="event">
                    <SelectValue placeholder="Select an event type" />
                  </SelectTrigger>
                  <SelectContent>
                    {EVENT_TYPES.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="secret">Secret (optional)</Label>
                <Input
                  id="secret"
                  placeholder="Optional signing secret"
                  value={formData.secret}
                  onChange={(event) => setFormData({ ...formData, secret: event.target.value })}
                />
              </div>
              <div className="flex items-center justify-between">
                <Label htmlFor="active">Active</Label>
                <Switch
                  id="active"
                  checked={formData.active}
                  onCheckedChange={(checked) => setFormData({ ...formData, active: checked })}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={editingWebhook ? handleUpdate : handleCreate}
                disabled={submitting || !formData.url || !formData.event}
              >
                {submitting ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : null}
                {editingWebhook ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {error && (
        <div className="flex items-center gap-2 rounded-lg bg-destructive/10 p-4 text-sm text-destructive">
          <AlertCircle className="h-4 w-4" />
          {error}
        </div>
      )}

      {success && (
        <div className="flex items-center gap-2 rounded-lg bg-green-50 dark:bg-green-900/20 p-4 text-sm text-green-800 dark:text-green-400">
          <CheckCircle className="h-4 w-4" />
          {success}
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Webhooks
            </CardTitle>
            <div className="text-2xl font-bold">{webhooks.length}</div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active
            </CardTitle>
            <div className="text-2xl font-bold text-green-600">
              {webhooks.filter((webhook) => webhook.active).length}
            </div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Inactive
            </CardTitle>
            <div className="text-2xl font-bold text-red-600">
              {webhooks.filter((webhook) => !webhook.active).length}
            </div>
          </CardHeader>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Webhooks</CardTitle>
          <CardDescription>
            Manage your webhook subscriptions and event notifications
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="h-6 w-6 animate-spin" />
            </div>
          ) : webhooks.length === 0 ? (
            <div className="text-center py-16">
              <p className="text-muted-foreground mb-4">No webhooks configured yet</p>
              <Button onClick={openCreateDialog}>
                <Plus className="mr-2 h-4 w-4" />
                Create Your First Webhook
              </Button>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>URL</TableHead>
                  <TableHead>Event</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Updated</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {webhooks.map((webhook) => (
                  <TableRow key={webhook.id}>
                    <TableCell className="font-mono text-sm max-w-xs truncate">
                      {webhook.url}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {getEventTypeLabel(webhook.event)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Switch
                          checked={webhook.active}
                          onCheckedChange={() => handleToggleActive(webhook)}
                        />
                        <Badge className={webhook.active ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400'}>
                          {webhook.active ? 'Active' : 'Inactive'}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(webhook.createdAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {new Date(webhook.updatedAt).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleTest(webhook.id)}
                          title="Send test event"
                        >
                          <Send className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openEditDialog(webhook)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(webhook.id)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Webhook Information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <h3 className="font-medium mb-2">Event Payload Format</h3>
            <div className="rounded-lg bg-muted p-4 font-mono text-xs">
              <pre>{JSON.stringify({
                event: "event.type",
                timestamp: "2024-01-01T00:00:00Z",
                data: {
                  // Event-specific data
                }
              }, null, 2)}</pre>
            </div>
          </div>
          <div>
            <h3 className="font-medium mb-2">Security</h3>
            <p className="text-sm text-muted-foreground">
              If configured, each request includes a signature in the <code className="bg-muted px-1 rounded">X-Webhook-Signature</code> header.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
