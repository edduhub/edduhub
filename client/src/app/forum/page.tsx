"use client";

import { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  MessageSquare,
  Search,
  Filter,
  Pin,
  Lock,
  TrendingUp,
  Clock,
  User,
  Tag
} from 'lucide-react';
import type { ForumThread, ForumCategory } from '@/lib/types';
import { api, endpoints } from '@/lib/api-client';

export default function ForumPage() {
  const searchParams = useSearchParams();
  const courseId = searchParams.get('courseId');

  const [threads, setThreads] = useState<ForumThread[]>([]);
  const [filteredThreads, setFilteredThreads] = useState<ForumThread[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<ForumCategory | 'all'>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);

  const fetchThreads = useCallback(async () => {
    try {
      setIsLoading(true);
      // Build query params
      let url = endpoints.forum.threads;
      const params = new URLSearchParams();
      if (selectedCategory !== 'all') {
        params.append('category', selectedCategory);
      }
      if (params.toString()) {
        url += `?${params.toString()}`;
      }
      const data = await api.get<ForumThread[]>(url);
      setThreads(data || []);
    } catch (error) {
      console.error('Failed to fetch threads:', error);
      setThreads([]);
    } finally {
      setIsLoading(false);
    }
  }, [selectedCategory]);

  const filterThreads = useCallback(() => {
    let filtered = threads;

    // Filter by category
    if (selectedCategory !== 'all') {
      filtered = filtered.filter(t => t.category === selectedCategory);
    }

    // Filter by search query
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(t =>
        t.title.toLowerCase().includes(query) ||
        t.content.toLowerCase().includes(query) ||
        t.tags.some(tag => tag.toLowerCase().includes(query))
      );
    }

    setFilteredThreads(filtered);
  }, [threads, searchQuery, selectedCategory]);

  useEffect(() => {
    fetchThreads();
  }, [fetchThreads]);

  useEffect(() => {
    filterThreads();
  }, [filterThreads]);


  const categoryColors: Record<ForumCategory, string> = {
    general: 'bg-blue-100 text-blue-800 hover:bg-blue-200',
    academic: 'bg-green-100 text-green-800 hover:bg-green-200',
    assignment: 'bg-purple-100 text-purple-800 hover:bg-purple-200',
    question: 'bg-yellow-100 text-yellow-800 hover:bg-yellow-200',
    announcement: 'bg-red-100 text-red-800 hover:bg-red-200',
  };

  const categoryLabels: Record<ForumCategory, string> = {
    general: 'General',
    academic: 'Academic',
    assignment: 'Assignment',
    question: 'Question',
    announcement: 'Announcement',
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-muted-foreground">Loading discussions...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-muted/10">
      <div className="container mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">
              <MessageSquare className="w-8 h-8" />
              Discussion Forum
            </h1>
            <p className="text-muted-foreground mt-1">
              {courseId ? 'Course Discussions' : 'College-wide Discussions'}
            </p>
          </div>
          <Button onClick={() => setIsCreateDialogOpen(true)}>
            <MessageSquare className="w-4 h-4 mr-2" />
            New Thread
          </Button>
        </div>

        {/* Search and Filters */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex gap-4 flex-wrap">
              <div className="flex-1 min-w-[200px]">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input
                    placeholder="Search discussions..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10"
                  />
                </div>
              </div>

              <Tabs value={selectedCategory} onValueChange={(v) => setSelectedCategory(v as ForumCategory | 'all')}>
                <TabsList>
                  <TabsTrigger value="all">All</TabsTrigger>
                  <TabsTrigger value="general">General</TabsTrigger>
                  <TabsTrigger value="academic">Academic</TabsTrigger>
                  <TabsTrigger value="assignment">Assignment</TabsTrigger>
                  <TabsTrigger value="question">Questions</TabsTrigger>
                  <TabsTrigger value="announcement">Announcements</TabsTrigger>
                </TabsList>
              </Tabs>
            </div>
          </CardContent>
        </Card>

        {/* Sort Options */}
        <div className="flex items-center gap-4">
          <Button variant="outline" size="sm">
            <Clock className="w-4 h-4 mr-2" />
            Latest
          </Button>
          <Button variant="outline" size="sm">
            <TrendingUp className="w-4 h-4 mr-2" />
            Popular
          </Button>
          <Button variant="outline" size="sm">
            <Filter className="w-4 h-4 mr-2" />
            Unanswered
          </Button>
        </div>

        {/* Thread List */}
        <div className="space-y-4">
          {filteredThreads.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <MessageSquare className="w-16 h-16 mx-auto text-muted-foreground mb-4" />
                <h3 className="text-xl font-semibold mb-2">No discussions found</h3>
                <p className="text-muted-foreground">
                  {searchQuery
                    ? 'Try adjusting your search or filters'
                    : 'Be the first to start a discussion!'}
                </p>
              </CardContent>
            </Card>
          ) : (
            filteredThreads.map((thread) => (
              <ForumThreadCard key={thread.id} thread={thread} categoryColors={categoryColors} categoryLabels={categoryLabels} />
            ))
          )}
        </div>

        {/* Create Thread Dialog */}
        {isCreateDialogOpen && (
          <CreateThreadDialog
            courseId={courseId}
            onClose={() => setIsCreateDialogOpen(false)}
            onSuccess={fetchThreads}
          />
        )}
      </div>
    </div>
  );
}

function ForumThreadCard({
  thread,
  categoryColors,
  categoryLabels
}: {
  thread: ForumThread
  categoryColors: Record<ForumCategory, string>
  categoryLabels: Record<ForumCategory, string>
}) {
  return (
    <Card className="hover:shadow-md transition-shadow cursor-pointer">
      <CardHeader>
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1 space-y-2">
            <div className="flex items-center gap-2 flex-wrap">
              <h3 className="text-lg font-semibold">{thread.title}</h3>
              {thread.isPinned && (
                <Badge variant="default" className="flex items-center gap-1">
                  <Pin className="w-3 h-3" />
                  Pinned
                </Badge>
              )}
              {thread.isLocked && (
                <Badge variant="secondary" className="flex items-center gap-1">
                  <Lock className="w-3 h-3" />
                  Locked
                </Badge>
              )}
              <Badge className={categoryColors[thread.category]}>
                {categoryLabels[thread.category]}
              </Badge>
            </div>

            <p className="text-sm text-muted-foreground line-clamp-2">
              {thread.content}
            </p>

            <div className="flex items-center gap-4 text-sm text-muted-foreground flex-wrap">
              <span className="flex items-center gap-1">
                <User className="w-4 h-4" />
                {thread.authorName}
              </span>
              <span className="flex items-center gap-1">
                <Clock className="w-4 h-4" />
                {new Date(thread.createdAt).toLocaleDateString()}
              </span>
              <span className="flex items-center gap-1">
                <MessageSquare className="w-4 h-4" />
                {thread.replyCount} replies
              </span>
              <span className="flex items-center gap-1">
                <TrendingUp className="w-4 h-4" />
                {thread.viewCount} views
              </span>
            </div>

            <div className="flex items-center gap-2 flex-wrap">
              {thread.tags.map((tag) => (
                <Badge key={tag} variant="outline" className="flex items-center gap-1">
                  <Tag className="w-3 h-3" />
                  {tag}
                </Badge>
              ))}
            </div>
          </div>
        </div>
      </CardHeader>
    </Card>
  );
}

function CreateThreadDialog({
  courseId,
  onClose,
  onSuccess
}: {
  courseId: string | null
  onClose: () => void
  onSuccess: () => void
}) {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [category, setCategory] = useState<ForumCategory>('general');
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleAddTag = () => {
    if (tagInput.trim() && !tags.includes(tagInput.trim())) {
      setTags([...tags, tagInput.trim()]);
      setTagInput('');
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setTags(tags.filter(tag => tag !== tagToRemove));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      await api.post(endpoints.forum.createThread, {
        courseId: parseInt(courseId || '0'),
        title,
        content,
        category,
        tags,
      });

      onSuccess();
      onClose();
      setTitle('');
      setContent('');
      setTags([]);
    } catch (error) {
      console.error('Failed to create thread:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
      <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Create New Discussion</CardTitle>
            <Button variant="ghost" size="sm" onClick={onClose}>
              ×
            </Button>
          </div>
          <CardDescription>
            Start a new discussion thread for your course
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label htmlFor="title" className="text-sm font-medium">
                Title
              </label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="What's your discussion about?"
                required
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="category" className="text-sm font-medium">
                Category
              </label>
              <select
                id="category"
                value={category}
                onChange={(e) => setCategory(e.target.value as ForumCategory)}
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="general">General</option>
                <option value="academic">Academic</option>
                <option value="assignment">Assignment</option>
                <option value="question">Question</option>
                <option value="announcement">Announcement</option>
              </select>
            </div>

            <div className="space-y-2">
              <label htmlFor="content" className="text-sm font-medium">
                Content
              </label>
              <textarea
                id="content"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="Describe your discussion in detail..."
                className="flex min-h-[200px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                required
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="tags" className="text-sm font-medium">
                Tags
              </label>
              <div className="flex gap-2">
                <Input
                  id="tags"
                  value={tagInput}
                  onChange={(e) => setTagInput(e.target.value)}
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleAddTag();
                    }
                  }}
                  placeholder="Add tags (press Enter)"
                />
                <Button type="button" onClick={handleAddTag}>
                  Add
                </Button>
              </div>
              <div className="flex gap-2 flex-wrap mt-2">
                {tags.map((tag) => (
                  <Badge key={tag} variant="secondary" className="flex items-center gap-1">
                    {tag}
                    <button
                      type="button"
                      onClick={() => handleRemoveTag(tag)}
                      className="ml-1 hover:text-destructive"
                    >
                      ×
                    </button>
                  </Badge>
                ))}
              </div>
            </div>

            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? 'Creating...' : 'Create Discussion'}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
