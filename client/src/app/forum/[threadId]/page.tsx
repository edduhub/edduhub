"use client";

import { useCallback, useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { api, endpoints } from '@/lib/api-client';
import { normalizeForumReply, normalizeForumThread, type ForumReplyApi, type ForumThreadApi } from '@/lib/forum';
import type { ForumCategory, ForumReply, ForumThread } from '@/lib/types';
import { logger } from '@/lib/logger';
import { ArrowLeft, CheckCircle, Clock, Lock, MessageSquare, Pin, Send, Tag, User } from 'lucide-react';

const categoryColors: Record<ForumCategory, string> = {
  general: 'bg-blue-100 text-blue-800',
  academic: 'bg-green-100 text-green-800',
  assignment: 'bg-purple-100 text-purple-800',
  question: 'bg-yellow-100 text-yellow-800',
  announcement: 'bg-red-100 text-red-800',
};

const categoryLabels: Record<ForumCategory, string> = {
  general: 'General',
  academic: 'Academic',
  assignment: 'Assignment',
  question: 'Question',
  announcement: 'Announcement',
};

export default function ForumThreadPage() {
  const params = useParams<{ threadId: string }>();
  const router = useRouter();
  const threadID = Number.parseInt(params.threadId, 10);

  const [thread, setThread] = useState<ForumThread | null>(null);
  const [replies, setReplies] = useState<ForumReply[]>([]);
  const [replyContent, setReplyContent] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadThread = useCallback(async () => {
    if (!Number.isFinite(threadID) || threadID <= 0) {
      setError('Invalid discussion thread ID.');
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      const [threadData, repliesData] = await Promise.all([
        api.get<ForumThreadApi>(endpoints.forum.thread(threadID)),
        api.get<ForumReplyApi[]>(endpoints.forum.replies(threadID)),
      ]);

      setThread(normalizeForumThread(threadData || {}));
      const normalizedReplies = Array.isArray(repliesData)
        ? repliesData.map(normalizeForumReply)
        : [];
      setReplies(normalizedReplies);
    } catch (err) {
      logger.error('Failed to load forum thread:', err as Error);
      setError('Failed to load this discussion. Please try again.');
    } finally {
      setIsLoading(false);
    }
  }, [threadID]);

  useEffect(() => {
    loadThread();
  }, [loadThread]);

  const handleCreateReply = async () => {
    const content = replyContent.trim();
    if (!content) {
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);
      await api.post(endpoints.forum.createReply(threadID), { content });
      setReplyContent('');
      await loadThread();
    } catch (err) {
      logger.error('Failed to create forum reply:', err as Error);
      setError('Failed to post reply. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const goBack = () => {
    if (thread?.courseId) {
      router.push(`/forum?courseId=${thread.courseId}`);
      return;
    }
    router.push('/forum');
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto" />
          <p className="mt-4 text-muted-foreground">Loading discussion...</p>
        </div>
      </div>
    );
  }

  if (!thread) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardContent className="py-12 text-center space-y-4">
            <p className="text-destructive">{error ?? 'Thread not found.'}</p>
            <Button variant="outline" onClick={() => router.push('/forum')}>
              Back to Forum
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-muted/10">
      <div className="container mx-auto p-6 space-y-6">
        <div className="flex items-center justify-between gap-4">
          <Button variant="outline" onClick={goBack}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Forum
          </Button>
          <div className="text-sm text-muted-foreground">
            {thread.courseName || (thread.courseId ? `Course ${thread.courseId}` : 'Discussion')}
          </div>
        </div>

        <Card>
          <CardHeader className="space-y-3">
            <div className="flex items-center gap-2 flex-wrap">
              <CardTitle className="text-2xl">{thread.title}</CardTitle>
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

            <div className="flex items-center gap-4 text-sm text-muted-foreground flex-wrap">
              <span className="flex items-center gap-1">
                <User className="w-4 h-4" />
                {thread.authorName}
              </span>
              <span className="flex items-center gap-1">
                <Clock className="w-4 h-4" />
                {new Date(thread.createdAt).toLocaleString()}
              </span>
              <span className="flex items-center gap-1">
                <MessageSquare className="w-4 h-4" />
                {thread.replyCount} replies
              </span>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-base whitespace-pre-wrap">{thread.content}</p>
            {thread.tags.length > 0 && (
              <div className="flex items-center gap-2 flex-wrap">
                {thread.tags.map((tag) => (
                  <Badge key={tag} variant="outline" className="flex items-center gap-1">
                    <Tag className="w-3 h-3" />
                    {tag}
                  </Badge>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <div className="space-y-4">
          <h2 className="text-xl font-semibold">Replies</h2>
          {replies.length === 0 ? (
            <Card>
              <CardContent className="py-8 text-center text-muted-foreground">
                No replies yet. Start the conversation.
              </CardContent>
            </Card>
          ) : (
            replies.map((reply) => (
              <Card key={reply.id}>
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <User className="w-4 h-4" />
                      <span>{reply.authorName}</span>
                      <span>•</span>
                      <Clock className="w-4 h-4" />
                      <span>{new Date(reply.createdAt).toLocaleString()}</span>
                    </div>
                    {reply.isAcceptedAnswer && (
                      <Badge className="bg-green-100 text-green-800 flex items-center gap-1">
                        <CheckCircle className="w-3 h-3" />
                        Accepted
                      </Badge>
                    )}
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="whitespace-pre-wrap">{reply.content}</p>
                </CardContent>
              </Card>
            ))
          )}
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Post a Reply</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <textarea
              value={replyContent}
              onChange={(event) => setReplyContent(event.target.value)}
              placeholder={thread.isLocked ? 'This thread is locked.' : 'Write your reply...'}
              className="flex min-h-[140px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              disabled={thread.isLocked || isSubmitting}
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
            <div className="flex justify-end">
              <Button
                onClick={handleCreateReply}
                disabled={thread.isLocked || isSubmitting || !replyContent.trim()}
              >
                <Send className="w-4 h-4 mr-2" />
                {isSubmitting ? 'Posting...' : 'Post Reply'}
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
