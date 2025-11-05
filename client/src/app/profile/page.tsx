"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/lib/auth-context";
import { api, endpoints, fetchProfile } from "@/lib/api-client";
import { Profile } from "@/lib/types";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Camera, Mail, Phone, MapPin, Calendar, Edit2 } from "lucide-react";
import { logger } from '@/lib/logger';

type ApiProfile = {
  phone_number?: string;
  date_of_birth?: string;
  address?: string;
  bio?: string;
  profile_image?: string;
};

type UpdateProfilePayload = {
  bio?: string;
  phoneNumber?: string;
  address?: string;
  dateOfBirth?: string;
};

const formatDateForInput = (value?: string) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return date.toISOString().split("T")[0];
};

export default function ProfilePage() {
  const { user } = useAuth();
  const [isEditing, setIsEditing] = useState(false);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [profileData, setProfileData] = useState({
    firstName: user?.firstName || "",
    lastName: user?.lastName || "",
    email: user?.email || "",
    phone: "",
    dateOfBirth: "",
    address: "",
    bio: ""
  });

  useEffect(() => {
    const loadProfile = async () => {
      try {
        setLoading(true);
        setError(null);
        const data = await fetchProfile();
        setProfile(data);
        const apiProfile = data as Profile & ApiProfile;
        setProfileData({
          firstName: user?.firstName || "",
          lastName: user?.lastName || "",
          email: user?.email || "",
          phone: apiProfile.phone_number || "",
          dateOfBirth: formatDateForInput(apiProfile.date_of_birth),
          address: apiProfile.address || "",
          bio: apiProfile.bio || ""
        });
      } catch (err) {
        logger.error('Error occurred', err as Error);
        setError("Failed to load profile");
      } finally {
        setLoading(false);
      }
    };
    loadProfile();
  }, [user]);

  const userInitials = user
    ? `${user.firstName[0]}${user.lastName[0]}`.toUpperCase()
    : 'U';

  const handleSave = async () => {
    try {
      setLoading(true);
      setError(null);

      const payload: UpdateProfilePayload = {
        bio: profileData.bio || undefined,
        phoneNumber: profileData.phone || undefined,
        address: profileData.address || undefined,
        dateOfBirth: profileData.dateOfBirth || undefined,
      };

      await api.patch(endpoints.auth.profile, payload);

      const updated = await fetchProfile();
      setProfile(updated);
      const updatedProfile = updated as Profile & ApiProfile;
      setProfileData((prev) => ({
        ...prev,
        phone: updatedProfile.phone_number || prev.phone,
        dateOfBirth: formatDateForInput(updatedProfile.date_of_birth) || prev.dateOfBirth,
        address: updatedProfile.address || prev.address,
        bio: updatedProfile.bio || prev.bio,
      }));

      setIsEditing(false);
    } catch (err) {
      logger.error('Error occurred', err as Error);
      setError("Failed to update profile");
    } finally {
      setLoading(false);
    }
  };

  const getRoleBadge = (role?: string) => {
    if (!role) return null;
    const styles = {
      student: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
      faculty: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
      admin: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    };
    return (
      <Badge className={styles[role as keyof typeof styles]}>
        {role.toUpperCase()}
      </Badge>
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Profile</h1>
        <p className="text-muted-foreground">
          Manage your personal information
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[300px_1fr]">
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col items-center space-y-4">
              <div className="relative">
                <Avatar className="h-32 w-32">
                  <AvatarImage src={user?.avatar} />
                  <AvatarFallback className="text-2xl">{userInitials}</AvatarFallback>
                </Avatar>
                <Button
                  size="icon"
                  variant="outline"
                  className="absolute bottom-0 right-0 rounded-full"
                >
                  <Camera className="h-4 w-4" />
                </Button>
              </div>
              <div className="text-center space-y-1">
                <h2 className="text-xl font-bold">
                  {user?.firstName} {user?.lastName}
                </h2>
                <p className="text-sm text-muted-foreground">{user?.email}</p>
                {getRoleBadge(user?.role)}
              </div>
              <Separator />
              <div className="w-full space-y-2">
                <div className="flex items-center gap-2 text-sm">
                  <Mail className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">Email Verified</span>
                </div>
                <div className="flex items-center gap-2 text-sm">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span className="text-muted-foreground">
                    Joined {profile?.joined_at ? new Date(profile.joined_at).getFullYear() : 'N/A'}
                  </span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Personal Information</CardTitle>
                <CardDescription>
                  Update your personal details
                </CardDescription>
              </div>
              <Button
                variant={isEditing ? "default" : "outline"}
                disabled={loading}
                onClick={() => isEditing ? handleSave() : setIsEditing(true)}
              >
                {isEditing ? (loading ? "Saving..." : "Save Changes") : (
                  <>
                    <Edit2 className="mr-2 h-4 w-4" />
                    Edit
                  </>
                )}
              </Button>
            </div>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="firstName">First Name</Label>
                <Input
                  id="firstName"
                  value={profileData.firstName}
                  onChange={(e) => setProfileData(prev => ({ ...prev, firstName: e.target.value }))}
                  disabled={!isEditing}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lastName">Last Name</Label>
                <Input
                  id="lastName"
                  value={profileData.lastName}
                  onChange={(e) => setProfileData(prev => ({ ...prev, lastName: e.target.value }))}
                  disabled={!isEditing}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={profileData.email}
                onChange={(e) => setProfileData(prev => ({ ...prev, email: e.target.value }))}
                disabled={!isEditing}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="phone">Phone Number</Label>
              <Input
                id="phone"
                type="tel"
                value={profileData.phone}
                onChange={(e) => setProfileData(prev => ({ ...prev, phone: e.target.value }))}
                disabled={!isEditing}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="dob">Date of Birth</Label>
              <Input
                id="dob"
                type="date"
                value={profileData.dateOfBirth}
                onChange={(e) => setProfileData(prev => ({ ...prev, dateOfBirth: e.target.value }))}
                disabled={!isEditing}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="address">Address</Label>
              <Input
                id="address"
                value={profileData.address}
                onChange={(e) => setProfileData(prev => ({ ...prev, address: e.target.value }))}
                disabled={!isEditing}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="bio">Bio</Label>
              <textarea
                id="bio"
                value={profileData.bio}
                onChange={(e) => setProfileData(prev => ({ ...prev, bio: e.target.value }))}
                disabled={!isEditing}
                className="flex min-h-[100px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>

            {isEditing && (
              <div className="flex justify-end gap-4">
                <Button variant="outline" onClick={() => setIsEditing(false)} disabled={loading}>
                  Cancel
                </Button>
                <Button onClick={handleSave} disabled={loading}>
                  {loading ? "Saving..." : "Save Changes"}
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {user?.role === 'student' && (
        <Card>
          <CardHeader>
            <CardTitle>Academic Information</CardTitle>
            <CardDescription>Your academic details</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">College ID</p>
                <p className="font-medium">{profile?.college_id || user.collegeId}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Department</p>
                <p className="font-medium">N/A</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Semester</p>
                <p className="font-medium">N/A</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">GPA</p>
                <p className="font-medium">N/A</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}