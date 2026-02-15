"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { api } from "@/lib/api-client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Loader2, AlertCircle, CheckCircle, Trash2, Download, Upload, FolderPlus, Folder, File, Tag, Clock } from "lucide-react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { logger } from '@/lib/logger';

type FileItem = {
  id: number;
  name: string;
  size: number;
  mime_type: string;
  category: string;
  description?: string;
  folder_id?: number;
  is_public: boolean;
  created_at: string;
  updated_at: string;
  current_version_id: number;
  version_number: number;
  tags?: string[];
  uploaded_by_name?: string;
};

type FolderItem = {
  id: number;
  name: string;
  parent_id?: number;
  created_at: string;
  file_count?: number;
};

type FileVersion = {
  id: number;
  file_id: number;
  version_number: number;
  size: number;
  mime_type: string;
  storage_path: string;
  comment?: string;
  uploaded_by: number;
  created_at: string;
  is_current: boolean;
};

export default function FilesPage() {
  const [files, setFiles] = useState<FileItem[]>([]);
  const [folders, setFolders] = useState<FolderItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Current folder
  const [currentFolderId, setCurrentFolderId] = useState<number | null>(null);
  const [folderPath, setFolderPath] = useState<Array<{ id: number; name: string }>>([]);

  // File upload dialog
  const [uploadDialogOpen, setUploadDialogOpen] = useState(false);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploadForm, setUploadForm] = useState({
    category: "document",
    description: "",
    tags: "",
  });

  // Folder creation dialog
  const [folderDialogOpen, setFolderDialogOpen] = useState(false);
  const [creatingFolder, setCreatingFolder] = useState(false);
  const [newFolderName, setNewFolderName] = useState("");

  // Version dialog
  const [versionDialogOpen, setVersionDialogOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<FileItem | null>(null);
  const [versions, setVersions] = useState<FileVersion[]>([]);
  const [loadingVersions, setLoadingVersions] = useState(false);

  // New version upload
  const [newVersionDialogOpen, setNewVersionDialogOpen] = useState(false);
  const [uploadingNewVersion, setUploadingNewVersion] = useState(false);
  const [versionComment, setVersionComment] = useState("");
  const newVersionInputRef = useRef<HTMLInputElement>(null);

  const fetchFiles = useCallback(async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (currentFolderId) params.append('folder_id', currentFolderId.toString());

      const response = await api.get(`/api/file-management?${params.toString()}`);
      setFiles(Array.isArray(response) ? response : []);
    } catch (err) {
      logger.error('Failed to fetch files:', err as Error);
      setError('Failed to load files');
    } finally {
      setLoading(false);
    }
  }, [currentFolderId]);

  const fetchFolders = useCallback(async () => {
    try {
      const params = new URLSearchParams();
      if (currentFolderId) params.append('parent_id', currentFolderId.toString());

      const response = await api.get(`/api/folders?${params.toString()}`);
      setFolders(Array.isArray(response) ? response : []);
    } catch (err) {
      logger.error('Failed to fetch folders:', err as Error);
    }
  }, [currentFolderId]);

  useEffect(() => {
    void fetchFiles();
    void fetchFolders();
  }, [fetchFiles, fetchFolders]);

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      setUploading(true);
      setError(null);

      const formData = new FormData();
      formData.append('file', file);
      formData.append('category', uploadForm.category);
      if (uploadForm.description) formData.append('description', uploadForm.description);
      if (currentFolderId) formData.append('folder_id', currentFolderId.toString());
      if (uploadForm.tags) formData.append('tags', JSON.stringify(uploadForm.tags.split(',').map(t => t.trim())));

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/file-management/upload`, {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      if (!response.ok) throw new Error('Upload failed');

      setSuccess('File uploaded successfully');
      setUploadDialogOpen(false);
      resetUploadForm();
      await fetchFiles();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to upload file');
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  const handleCreateFolder = async () => {
    try {
      setCreatingFolder(true);
      setError(null);

      await api.post('/api/folders', {
        name: newFolderName,
        parent_id: currentFolderId,
      });

      setSuccess('Folder created successfully');
      setFolderDialogOpen(false);
      setNewFolderName("");
      await fetchFolders();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to create folder');
    } finally {
      setCreatingFolder(false);
    }
  };

  const handleDeleteFile = async (id: number) => {
    if (!confirm('Are you sure you want to delete this file?')) return;

    try {
      await api.delete(`/api/file-management/${id}`);
      setSuccess('File deleted successfully');
      await fetchFiles();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to delete file');
    }
  };

  const handleViewVersions = async (file: FileItem) => {
    try {
      setSelectedFile(file);
      setVersionDialogOpen(true);
      setLoadingVersions(true);

      const response = await api.get(`/api/file-management/${file.id}/versions`);
      setVersions(Array.isArray(response) ? response : []);
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to load versions');
    } finally {
      setLoadingVersions(false);
    }
  };

  const handleUploadNewVersion = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !selectedFile) return;

    try {
      setUploadingNewVersion(true);
      setError(null);

      const formData = new FormData();
      formData.append('file', file);
      if (versionComment) formData.append('comment', versionComment);

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/file-management/${selectedFile.id}/versions`, {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      if (!response.ok) throw new Error('Upload failed');

      setSuccess('New version uploaded successfully');
      setNewVersionDialogOpen(false);
      setVersionComment("");
      await handleViewVersions(selectedFile);
      await fetchFiles();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to upload new version');
    } finally {
      setUploadingNewVersion(false);
      if (newVersionInputRef.current) newVersionInputRef.current.value = '';
    }
  };

  const handleSetCurrentVersion = async (versionId: number) => {
    if (!selectedFile) return;

    try {
      await api.patch(`/api/file-management/${selectedFile.id}/versions/${versionId}/current`, {});
      setSuccess('Current version updated successfully');
      await handleViewVersions(selectedFile);
      await fetchFiles();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to set current version');
    }
  };

  const handleDownload = async (fileId: number) => {
    try {
      const response = await api.get<{ url?: string }>(`/api/file-management/${fileId}/download`);
      if (response?.url) {
        window.open(response.url, '_blank');
      }
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to download file');
    }
  };

  const resetUploadForm = () => {
    setUploadForm({
      category: "document",
      description: "",
      tags: "",
    });
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
  };

  const handleOpenFolder = (folder: FolderItem) => {
    setFolderPath((prev) => [...prev, { id: folder.id, name: folder.name }]);
    setCurrentFolderId(folder.id);
  };

  const handleNavigateBack = () => {
    if (folderPath.length === 0) return;
    const next = folderPath.slice(0, -1);
    setFolderPath(next);
    setCurrentFolderId(next.length > 0 ? next[next.length - 1].id : null);
  };

  const handleGoToRoot = () => {
    setCurrentFolderId(null);
    setFolderPath([]);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">File Management</h1>
          <p className="text-muted-foreground">
            Manage files with versioning, folders, and tags
          </p>
          <div className="mt-2 flex items-center gap-2 text-sm text-muted-foreground">
            <span>Location:</span>
            <Badge variant="outline">Root</Badge>
            {folderPath.map((segment) => (
              <Badge key={segment.id} variant="outline">{segment.name}</Badge>
            ))}
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleNavigateBack} disabled={folderPath.length === 0}>
            Back
          </Button>
          <Button variant="outline" onClick={handleGoToRoot} disabled={currentFolderId === null}>
            Root
          </Button>
          <Dialog open={folderDialogOpen} onOpenChange={setFolderDialogOpen}>
            <DialogTrigger asChild>
              <Button variant="outline">
                <FolderPlus className="mr-2 h-4 w-4" />
                New Folder
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Folder</DialogTitle>
                <DialogDescription>Create a new folder to organize your files</DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <Label htmlFor="folder-name">Folder Name</Label>
                  <Input
                    id="folder-name"
                    value={newFolderName}
                    onChange={(e) => setNewFolderName(e.target.value)}
                    placeholder="Enter folder name"
                  />
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setFolderDialogOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={handleCreateFolder} disabled={creatingFolder || !newFolderName}>
                  {creatingFolder && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                  Create
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          <Dialog open={uploadDialogOpen} onOpenChange={setUploadDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <Upload className="mr-2 h-4 w-4" />
                Upload File
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Upload File</DialogTitle>
                <DialogDescription>Upload a new file with optional metadata</DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <Label htmlFor="file">File</Label>
                  <Input
                    id="file"
                    ref={fileInputRef}
                    type="file"
                    onChange={handleFileUpload}
                    disabled={uploading}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="category">Category</Label>
                  <Select
                    value={uploadForm.category}
                    onValueChange={(value) => setUploadForm({ ...uploadForm, category: value })}
                  >
                    <SelectTrigger id="category">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="document">Document</SelectItem>
                      <SelectItem value="image">Image</SelectItem>
                      <SelectItem value="video">Video</SelectItem>
                      <SelectItem value="assignment">Assignment</SelectItem>
                      <SelectItem value="other">Other</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="description">Description</Label>
                  <Textarea
                    id="description"
                    value={uploadForm.description}
                    onChange={(e) => setUploadForm({ ...uploadForm, description: e.target.value })}
                    placeholder="Optional file description"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="tags">Tags (comma-separated)</Label>
                  <Input
                    id="tags"
                    value={uploadForm.tags}
                    onChange={(e) => setUploadForm({ ...uploadForm, tags: e.target.value })}
                    placeholder="e.g. syllabus, semester1, cs101"
                  />
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>
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

      <Tabs defaultValue="files" className="space-y-4">
        <TabsList>
          <TabsTrigger value="files">Files</TabsTrigger>
          <TabsTrigger value="folders">Folders</TabsTrigger>
        </TabsList>

        <TabsContent value="files" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Files</CardTitle>
              <CardDescription>
                {currentFolderId ? 'Files in current folder' : 'All files'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex items-center justify-center py-16">
                  <Loader2 className="h-6 w-6 animate-spin" />
                </div>
              ) : files.length === 0 ? (
                <div className="text-center py-16">
                  <p className="text-muted-foreground">No files found</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Size</TableHead>
                      <TableHead>Category</TableHead>
                      <TableHead>Version</TableHead>
                      <TableHead>Tags</TableHead>
                      <TableHead>Uploaded</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {files.map((file) => (
                      <TableRow key={file.id}>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <File className="h-4 w-4" />
                            <span className="font-medium">{file.name}</span>
                          </div>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatFileSize(file.size)}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">{file.category}</Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className="bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400">
                            v{file.version_number}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="flex gap-1 flex-wrap">
                            {file.tags?.slice(0, 2).map((tag, idx) => (
                              <Badge key={idx} variant="outline" className="text-xs">
                                <Tag className="h-3 w-3 mr-1" />
                                {tag}
                              </Badge>
                            ))}
                            {file.tags && file.tags.length > 2 && (
                              <Badge variant="outline" className="text-xs">
                                +{file.tags.length - 2}
                              </Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {new Date(file.created_at).toLocaleDateString()}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleViewVersions(file)}
                            >
                              <Clock className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleDownload(file.id)}
                            >
                              <Download className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleDeleteFile(file.id)}
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
        </TabsContent>

        <TabsContent value="folders" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Folders</CardTitle>
              <CardDescription>
                Organize your files with folders
              </CardDescription>
            </CardHeader>
            <CardContent>
              {folders.length === 0 ? (
                <div className="text-center py-16">
                  <p className="text-muted-foreground">No folders found</p>
                </div>
              ) : (
                <div className="grid gap-4 md:grid-cols-3">
                  {folders.map((folder) => (
                    <Card
                      key={folder.id}
                      className="cursor-pointer hover:bg-accent"
                      onClick={() => handleOpenFolder(folder)}
                    >
                      <CardContent className="p-4">
                        <div className="flex items-center gap-3">
                          <Folder className="h-8 w-8 text-blue-500" />
                          <div>
                            <h3 className="font-medium">{folder.name}</h3>
                            <p className="text-sm text-muted-foreground">
                              {folder.file_count || 0} files
                            </p>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Version History Dialog */}
      <Dialog open={versionDialogOpen} onOpenChange={setVersionDialogOpen}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>Version History - {selectedFile?.name}</DialogTitle>
            <DialogDescription>
              View and manage file versions
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <Button
              onClick={() => setNewVersionDialogOpen(true)}
              size="sm"
            >
              <Upload className="mr-2 h-4 w-4" />
              Upload New Version
            </Button>

            {loadingVersions ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin" />
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Version</TableHead>
                    <TableHead>Size</TableHead>
                    <TableHead>Comment</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {versions.map((version) => (
                    <TableRow key={version.id}>
                      <TableCell>
                        <Badge>v{version.version_number}</Badge>
                      </TableCell>
                      <TableCell>{formatFileSize(version.size)}</TableCell>
                      <TableCell className="text-sm">{version.comment || '-'}</TableCell>
                      <TableCell className="text-sm">
                        {new Date(version.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell>
                        {version.is_current && (
                          <Badge className="bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400">
                            Current
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        {!version.is_current && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleSetCurrentVersion(version.id)}
                          >
                            Set as Current
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* New Version Upload Dialog */}
      <Dialog open={newVersionDialogOpen} onOpenChange={setNewVersionDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Upload New Version</DialogTitle>
            <DialogDescription>
              Upload a new version of {selectedFile?.name}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="new-version-file">File</Label>
              <Input
                id="new-version-file"
                ref={newVersionInputRef}
                type="file"
                onChange={handleUploadNewVersion}
                disabled={uploadingNewVersion}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="version-comment">Version Comment</Label>
              <Textarea
                id="version-comment"
                value={versionComment}
                onChange={(e) => setVersionComment(e.target.value)}
                placeholder="What changed in this version?"
              />
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
