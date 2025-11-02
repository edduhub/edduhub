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
import { Plus, Loader2, AlertCircle, CheckCircle, Trash2, Edit, Send, Power } from "lucide-react";
import { Switch } from "@/components/ui/switch";

type Webhook = {
  id: number;
  url: string;
  event_type: string;
  is_active: boolean;
  secret: string;
  created_at: string;
  last_triggered_at?: string;
  failure_count: number;
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

export default function WebhooksPage() {
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Create/Edit dialog state
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingWebhook, setEditingWebhook] = useState<Webhook | null>(null);
  const [formData, setFormData] = useState({
    url: "",
    event_type: "",
    is_active: true,
  });
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchWebhooks();
  }, []);

  const fetchWebhooks = async () => {
    try {
      setLoading(true);
      const response = await api.get('/api/webhooks');
      setWebhooks(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to fetch webhooks:', err);
      setError('Failed to load webhooks');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    try {
      setSubmitting(true);
      setError(null);
      await api.post('/api/webhooks', formData);
      setSuccess('Webhook created successfully');
      setDialogOpen(false);
      resetForm();
      await fetchWebhooks();
    } catch (err: any) {
      setError(err.message || 'Failed to create webhook');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdate = async () => {
    if (!editingWebhook) return;

    try {
      setSubmitting(true);
      setError(null);
      await api.put(`/api/webhooks/${editingWebhook.id}`, formData);
      setSuccess('Webhook updated successfully');
      setDialogOpen(false);
      resetForm();
      await fetchWebhooks();
    } catch (err: any) {
      setError(err.message || 'Failed to update webhook');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this webhook?')) return;

    try {
      await api.delete(`/api/webhooks/${id}`);
      setSuccess('Webhook deleted successfully');
      await fetchWebhooks();
    } catch (err: any) {
      setError(err.message || 'Failed to delete webhook');
    }
  };

  const handleTest = async (id: number) => {
    try {
      setError(null);
      await api.post(`/api/webhooks/${id}/test`);
      setSuccess('Test event sent successfully');
    } catch (err: any) {
      setError(err.message || 'Failed to send test event');
    }
  };

  const handleToggleActive = async (webhook: Webhook) => {
    try {
      await api.put(`/api/webhooks/${webhook.id}`, {
        ...webhook,
        is_active: !webhook.is_active,
      });
      await fetchWebhooks();
    } catch (err: any) {
      setError(err.message || 'Failed to update webhook');
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
      event_type: webhook.event_type,
      is_active: webhook.is_active,
    });
    setEditingWebhook(webhook);
    setDialogOpen(true);
  };

  const resetForm = () => {
    setFormData({
      url: "",
      event_type: "",
      is_active: true,
    });
    setEditingWebhook(null);
  };

  const getEventTypeLabel = (type: string) => {
    return EVENT_TYPES.find(et => et.value === type)?.label || type;
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
                  onChange={(e) => setFormData({ ...formData, url: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="event_type">Event Type</Label>
                <Select
                  value={formData.event_type}
                  onValueChange={(value) => setFormData({ ...formData, event_type: value })}
                >
                  <SelectTrigger id="event_type">
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
              <div className="flex items-center justify-between">
                <Label htmlFor="is_active">Active</Label>
                <Switch
                  id="is_active"
                  checked={formData.is_active}
                  onCheckedChange={(checked) => setFormData({ ...formData, is_active: checked })}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={editingWebhook ? handleUpdate : handleCreate}
                disabled={submitting || !formData.url || !formData.event_type}
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
              {webhooks.filter(w => w.is_active).length}
            </div>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Failed Deliveries
            </CardTitle>
            <div className="text-2xl font-bold text-red-600">
              {webhooks.reduce((acc, w) => acc + w.failure_count, 0)}
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
                  <TableHead>Event Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Failures</TableHead>
                  <TableHead>Last Triggered</TableHead>
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
                        {getEventTypeLabel(webhook.event_type)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Switch
                          checked={webhook.is_active}
                          onCheckedChange={() => handleToggleActive(webhook)}
                        />
                        <Badge className={webhook.is_active ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400'}>
                          {webhook.is_active ? 'Active' : 'Inactive'}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell>
                      {webhook.failure_count > 0 ? (
                        <span className="text-red-600 font-medium">{webhook.failure_count}</span>
                      ) : (
                        <span className="text-muted-foreground">0</span>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {webhook.last_triggered_at
                        ? new Date(webhook.last_triggered_at).toLocaleDateString()
                        : 'Never'}
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
              Each webhook includes a secret token for verification. The secret is sent in the <code className="bg-muted px-1 rounded">X-Webhook-Secret</code> header with each request.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
