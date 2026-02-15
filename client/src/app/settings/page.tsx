"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Bell, Globe, Shield, Loader2 } from "lucide-react";
import { api, endpoints } from "@/lib/api-client";
import { logger } from "@/lib/logger";

interface UserSettings {
  email_notifications: boolean;
  push_notifications: boolean;
  assignment_reminders: boolean;
  grade_updates: boolean;
  announcement_alerts: boolean;
  theme: string;
  language: string;
  timezone: string;
  session_timeout: number;
}

export default function SettingsPage() {
  useAuth();
  const [settings, setSettings] = useState({
    emailNotifications: true,
    pushNotifications: false,
    assignmentReminders: true,
    gradeUpdates: true,
    announcementAlerts: true,
    twoFactorAuth: false,
    sessionTimeout: "30",
    language: "en",
    timezone: "Asia/Kolkata"
  });
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isPasswordDialogOpen, setIsPasswordDialogOpen] = useState(false);
  const [isChangingPassword, setIsChangingPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [passwordMessage, setPasswordMessage] = useState<string | null>(null);
  const [passwordForm, setPasswordForm] = useState({
    oldPassword: "",
    newPassword: "",
    confirmPassword: "",
  });

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await api.get<UserSettings>("/api/settings");
      if (data) {
        setSettings(prev => ({
          ...prev,
          emailNotifications: data.email_notifications,
          pushNotifications: data.push_notifications,
          assignmentReminders: data.assignment_reminders,
          gradeUpdates: data.grade_updates,
          announcementAlerts: data.announcement_alerts,
          language: data.language || "en",
          timezone: data.timezone || "UTC",
          sessionTimeout: String(data.session_timeout || 30),
        }));
      }
    } catch (err) {
      logger.error("Failed to fetch settings", err as Error);
      setError("Failed to load settings");
    } finally {
      setIsLoading(false);
    }
  };

  const handleToggle = (key: string) => {
    setSettings(prev => ({
      ...prev,
      [key]: !prev[key as keyof typeof prev]
    }));
  };

  const handleSave = async () => {
    setIsSaving(true);
    setError(null);
    setPasswordMessage(null);
    try {
      await api.put("/api/settings", {
        email_notifications: settings.emailNotifications,
        push_notifications: settings.pushNotifications,
        assignment_reminders: settings.assignmentReminders,
        grade_updates: settings.gradeUpdates,
        announcement_alerts: settings.announcementAlerts,
        theme: "system",
        language: settings.language,
        timezone: settings.timezone,
        session_timeout: parseInt(settings.sessionTimeout, 10),
      });
    } catch (err) {
      logger.error("Failed to save settings", err as Error);
      setError("Failed to save settings");
    } finally {
      setIsSaving(false);
    }
  };

  const handlePasswordChange = async () => {
    try {
      setIsChangingPassword(true);
      setError(null);
      setPasswordMessage(null);

      if (!passwordForm.oldPassword || !passwordForm.newPassword || !passwordForm.confirmPassword) {
        throw new Error("All password fields are required");
      }
      if (passwordForm.newPassword.length < 8) {
        throw new Error("New password must be at least 8 characters");
      }
      if (passwordForm.newPassword !== passwordForm.confirmPassword) {
        throw new Error("New password and confirmation do not match");
      }

      await api.post(endpoints.auth.changePassword, {
        old_password: passwordForm.oldPassword,
        new_password: passwordForm.newPassword,
      });

      setPasswordMessage("Password updated successfully");
      setPasswordForm({ oldPassword: "", newPassword: "", confirmPassword: "" });
      setIsPasswordDialogOpen(false);
    } catch (err) {
      logger.error("Failed to change password", err as Error);
      setError(err instanceof Error ? err.message : "Failed to update password");
    } finally {
      setIsChangingPassword(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground">
          Manage your account settings and preferences
        </p>
      </div>

      {error && (
        <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}
      {passwordMessage && (
        <div className="rounded-lg bg-green-500/10 p-3 text-sm text-green-700">
          {passwordMessage}
        </div>
      )}

      <div className="grid gap-6">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              <CardTitle>Notifications</CardTitle>
            </div>
            <CardDescription>
              Configure how you receive notifications
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Email Notifications</Label>
                <p className="text-sm text-muted-foreground">
                  Receive notifications via email
                </p>
              </div>
              <Switch
                checked={settings.emailNotifications}
                onCheckedChange={() => handleToggle('emailNotifications')}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Push Notifications</Label>
                <p className="text-sm text-muted-foreground">
                  Receive push notifications on your device
                </p>
              </div>
              <Switch
                checked={settings.pushNotifications}
                onCheckedChange={() => handleToggle('pushNotifications')}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Assignment Reminders</Label>
                <p className="text-sm text-muted-foreground">
                  Get reminded about upcoming assignments
                </p>
              </div>
              <Switch
                checked={settings.assignmentReminders}
                onCheckedChange={() => handleToggle('assignmentReminders')}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Grade Updates</Label>
                <p className="text-sm text-muted-foreground">
                  Notify when grades are posted
                </p>
              </div>
              <Switch
                checked={settings.gradeUpdates}
                onCheckedChange={() => handleToggle('gradeUpdates')}
              />
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Announcement Alerts</Label>
                <p className="text-sm text-muted-foreground">
                  Get notified about new announcements
                </p>
              </div>
              <Switch
                checked={settings.announcementAlerts}
                onCheckedChange={() => handleToggle('announcementAlerts')}
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Shield className="h-5 w-5" />
              <CardTitle>Security</CardTitle>
            </div>
            <CardDescription>
              Manage your security preferences
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Two-Factor Authentication</Label>
                <p className="text-sm text-muted-foreground">
                  Add an extra layer of security
                </p>
              </div>
              <Switch
                checked={settings.twoFactorAuth}
                onCheckedChange={() => handleToggle('twoFactorAuth')}
              />
            </div>
            <Separator />
            <div className="space-y-2">
              <Label htmlFor="session-timeout">Session Timeout (minutes)</Label>
              <Input
                id="session-timeout"
                type="number"
                value={settings.sessionTimeout}
                onChange={(e) => setSettings(prev => ({ ...prev, sessionTimeout: e.target.value }))}
                className="max-w-xs"
              />
              <p className="text-sm text-muted-foreground">
                Automatically log out after inactivity
              </p>
            </div>
            <Separator />
            <div className="space-y-2">
              <Label>Change Password</Label>
              <Button variant="outline" onClick={() => setIsPasswordDialogOpen(true)}>Update Password</Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Globe className="h-5 w-5" />
              <CardTitle>Preferences</CardTitle>
            </div>
            <CardDescription>
              Customize your experience
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="language">Language</Label>
              <select
                id="language"
                value={settings.language}
                onChange={(e) => setSettings(prev => ({ ...prev, language: e.target.value }))}
                className="flex h-10 w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="en">English</option>
                <option value="hi">Hindi</option>
                <option value="es">Spanish</option>
              </select>
            </div>
            <Separator />
            <div className="space-y-2">
              <Label htmlFor="timezone">Timezone</Label>
              <select
                id="timezone"
                value={settings.timezone}
                onChange={(e) => setSettings(prev => ({ ...prev, timezone: e.target.value }))}
                className="flex h-10 w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="Asia/Kolkata">Asia/Kolkata (IST)</option>
                <option value="America/New_York">America/New_York (EST)</option>
                <option value="Europe/London">Europe/London (GMT)</option>
                <option value="Asia/Tokyo">Asia/Tokyo (JST)</option>
              </select>
            </div>
          </CardContent>
        </Card>

        <div className="flex justify-end gap-4">
          <Button variant="outline" onClick={fetchSettings} disabled={isSaving}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isSaving}>
            {isSaving ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
            Save Changes
          </Button>
        </div>
      </div>

      <Dialog open={isPasswordDialogOpen} onOpenChange={setIsPasswordDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update Password</DialogTitle>
            <DialogDescription>
              Enter your current password and choose a new one.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="old-password">Current Password</Label>
              <Input
                id="old-password"
                type="password"
                value={passwordForm.oldPassword}
                onChange={(e) => setPasswordForm(prev => ({ ...prev, oldPassword: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="new-password">New Password</Label>
              <Input
                id="new-password"
                type="password"
                value={passwordForm.newPassword}
                onChange={(e) => setPasswordForm(prev => ({ ...prev, newPassword: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirm-password">Confirm New Password</Label>
              <Input
                id="confirm-password"
                type="password"
                value={passwordForm.confirmPassword}
                onChange={(e) => setPasswordForm(prev => ({ ...prev, confirmPassword: e.target.value }))}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsPasswordDialogOpen(false)}
              disabled={isChangingPassword}
            >
              Cancel
            </Button>
            <Button
              onClick={handlePasswordChange}
              disabled={isChangingPassword}
            >
              {isChangingPassword ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Update Password
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
