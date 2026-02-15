"use client";

import { useState, useEffect, useRef } from "react";
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
import { Camera, Mail, Calendar, Edit2, Loader2 } from "lucide-react";
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
  phone_number?: string;
  address?: string;
  date_of_birth?: string;
};

type StudentDashboardApi = {
  student?: {
    semester?: number;
    department?: number;
    departmentName?: string;
    department_name?: string;
  };
  academicOverview?: {
    gpa?: number;
    enrolledCourses?: number;
    enrolled_courses?: number;
  };
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
  const [success, setSuccess] = useState<string | null>(null);
  const [uploadingImage, setUploadingImage] = useState(false);
  const [profileImageUrl, setProfileImageUrl] = useState("");
  const imageInputRef = useRef<HTMLInputElement>(null);
  const [academicInfo, setAcademicInfo] = useState({
    department: "N/A",
    semester: "N/A",
    gpa: "N/A",
    enrolledCourses: "N/A",
  });
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
        setProfileImageUrl(apiProfile.profile_image || user?.avatar || "");
        setProfileData({
          firstName: user?.firstName || "",
          lastName: user?.lastName || "",
          email: user?.email || "",
          phone: apiProfile.phone_number || "",
          dateOfBirth: formatDateForInput(apiProfile.date_of_birth),
          address: apiProfile.address || "",
          bio: apiProfile.bio || ""
        });

        if (user?.role === "student") {
          try {
            const dashboard = await api.get<StudentDashboardApi>("/api/student/dashboard");
            const semester = dashboard?.student?.semester;
            const department =
              dashboard?.student?.departmentName ||
              dashboard?.student?.department_name ||
              (dashboard?.student?.department ? `Department ${dashboard.student.department}` : undefined);
            const gpa = dashboard?.academicOverview?.gpa;
            const enrolledCourses = dashboard?.academicOverview?.enrolledCourses ?? dashboard?.academicOverview?.enrolled_courses;

            setAcademicInfo({
              department: department || "N/A",
              semester: typeof semester === "number" ? String(semester) : "N/A",
              gpa: typeof gpa === "number" ? gpa.toFixed(2) : "N/A",
              enrolledCourses: typeof enrolledCourses === "number" ? String(enrolledCourses) : "N/A",
            });
          } catch (dashboardError) {
            logger.warn("Failed to load student academic info for profile:", { error: dashboardError });
            setAcademicInfo({
              department: "N/A",
              semester: "N/A",
              gpa: "N/A",
              enrolledCourses: "N/A",
            });
          }
        }
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
      setSuccess(null);

      const payload: UpdateProfilePayload = {
        bio: profileData.bio || undefined,
        phone_number: profileData.phone || undefined,
        address: profileData.address || undefined,
        date_of_birth: profileData.dateOfBirth
          ? new Date(`${profileData.dateOfBirth}T00:00:00Z`).toISOString()
          : undefined,
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
      setSuccess("Profile updated successfully");
    } catch (err) {
      logger.error('Error occurred', err as Error);
      setError("Failed to update profile");
    } finally {
      setLoading(false);
    }
  };

  const resetFormFromProfile = () => {
    const apiProfile = profile as (Profile & ApiProfile) | null;
    setProfileData({
      firstName: user?.firstName || "",
      lastName: user?.lastName || "",
      email: user?.email || "",
      phone: apiProfile?.phone_number || "",
      dateOfBirth: formatDateForInput(apiProfile?.date_of_birth),
      address: apiProfile?.address || "",
      bio: apiProfile?.bio || "",
    });
  };

  const handleCancelEdit = () => {
    resetFormFromProfile();
    setIsEditing(false);
  };

  const handleProfileImageUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    try {
      setUploadingImage(true);
      setError(null);
      setSuccess(null);

      const formData = new FormData();
      formData.append("image", file);

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/api/profile/upload-image`,
        {
          method: "POST",
          credentials: "include",
          body: formData,
        }
      );

      const result = await response.json().catch(() => ({}));
      if (!response.ok) {
        const message = result?.error || result?.message || "Failed to upload profile image";
        throw new Error(message);
      }

      const imageUrl = result?.data?.url || result?.url;
      if (typeof imageUrl === "string" && imageUrl.length > 0) {
        setProfileImageUrl(imageUrl);
        setProfile((prev) => (prev ? { ...prev, profile_image: imageUrl } : prev));
      }

      setSuccess("Profile image updated successfully");
    } catch (uploadError) {
      logger.error("Failed to upload profile image", uploadError as Error);
      setError(uploadError instanceof Error ? uploadError.message : "Failed to upload profile image");
    } finally {
      setUploadingImage(false);
      if (imageInputRef.current) {
        imageInputRef.current.value = "";
      }
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
      {error && (
        <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-lg bg-green-500/10 p-3 text-sm text-green-700">
          {success}
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-[300px_1fr]">
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col items-center space-y-4">
              <div className="relative">
                <Avatar className="h-32 w-32">
                  <AvatarImage src={profileImageUrl || user?.avatar} />
                  <AvatarFallback className="text-2xl">{userInitials}</AvatarFallback>
                </Avatar>
                <Button
                  size="icon"
                  variant="outline"
                  className="absolute bottom-0 right-0 rounded-full"
                  onClick={() => imageInputRef.current?.click()}
                  disabled={uploadingImage}
                >
                  {uploadingImage ? <Loader2 className="h-4 w-4 animate-spin" /> : <Camera className="h-4 w-4" />}
                </Button>
                <input
                  ref={imageInputRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={handleProfileImageUpload}
                />
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
                  disabled
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lastName">Last Name</Label>
                <Input
                  id="lastName"
                  value={profileData.lastName}
                  onChange={(e) => setProfileData(prev => ({ ...prev, lastName: e.target.value }))}
                  disabled
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
                disabled
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
                <Button variant="outline" onClick={handleCancelEdit} disabled={loading}>
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
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">College ID</p>
                <p className="font-medium">{profile?.college_id || user.collegeId}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Department</p>
                <p className="font-medium">{academicInfo.department}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Semester</p>
                <p className="font-medium">{academicInfo.semester}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">GPA</p>
                <p className="font-medium">{academicInfo.gpa}</p>
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">Enrolled Courses</p>
                <p className="font-medium">{academicInfo.enrolledCourses}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
