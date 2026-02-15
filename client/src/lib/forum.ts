import type { ForumCategory, ForumReply, ForumThread } from './types';

export type ForumThreadApi = {
  id?: number;
  courseId?: number;
  course_id?: number;
  courseName?: string;
  course_name?: string;
  category?: ForumCategory | string;
  title?: string;
  content?: string;
  authorId?: number;
  author_id?: number;
  authorName?: string;
  author_name?: string;
  authorAvatar?: string;
  author_avatar?: string;
  isPinned?: boolean;
  is_pinned?: boolean;
  isLocked?: boolean;
  is_locked?: boolean;
  viewCount?: number;
  view_count?: number;
  replyCount?: number;
  reply_count?: number;
  lastReplyAt?: string;
  last_reply_at?: string;
  lastReplyBy?: number;
  last_reply_by?: number;
  tags?: string[];
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
  collegeId?: number;
  college_id?: number;
};

export type ForumReplyApi = {
  id?: number;
  threadId?: number;
  thread_id?: number;
  parentId?: number;
  parent_id?: number;
  content?: string;
  authorId?: number;
  author_id?: number;
  authorName?: string;
  author_name?: string;
  authorAvatar?: string;
  author_avatar?: string;
  isAcceptedAnswer?: boolean;
  is_accepted_answer?: boolean;
  likeCount?: number;
  like_count?: number;
  hasLiked?: boolean;
  has_liked?: boolean;
  createdAt?: string;
  created_at?: string;
  updatedAt?: string;
  updated_at?: string;
  collegeId?: number;
  college_id?: number;
};

function normalizeCategory(raw: string | ForumCategory | undefined): ForumCategory {
  if (
    raw === 'general' ||
    raw === 'academic' ||
    raw === 'assignment' ||
    raw === 'question' ||
    raw === 'announcement'
  ) {
    return raw;
  }
  return 'general';
}

export function normalizeForumThread(raw: ForumThreadApi): ForumThread {
  const createdAt = raw.createdAt ?? raw.created_at ?? new Date().toISOString();
  const updatedAt = raw.updatedAt ?? raw.updated_at ?? createdAt;
  return {
    id: raw.id ?? 0,
    courseId: raw.courseId ?? raw.course_id ?? 0,
    courseName: raw.courseName ?? raw.course_name,
    category: normalizeCategory(raw.category),
    title: raw.title ?? '',
    content: raw.content ?? '',
    authorId: raw.authorId ?? raw.author_id ?? 0,
    authorName: raw.authorName ?? raw.author_name ?? 'Unknown',
    authorAvatar: raw.authorAvatar ?? raw.author_avatar,
    isPinned: raw.isPinned ?? raw.is_pinned ?? false,
    isLocked: raw.isLocked ?? raw.is_locked ?? false,
    viewCount: raw.viewCount ?? raw.view_count ?? 0,
    replyCount: raw.replyCount ?? raw.reply_count ?? 0,
    lastReplyAt: raw.lastReplyAt ?? raw.last_reply_at,
    lastReplyBy: raw.lastReplyBy ?? raw.last_reply_by,
    tags: Array.isArray(raw.tags) ? raw.tags : [],
    createdAt,
    updatedAt,
    collegeId: raw.collegeId ?? raw.college_id ?? 0,
  };
}

export function normalizeForumReply(raw: ForumReplyApi): ForumReply {
  const createdAt = raw.createdAt ?? raw.created_at ?? new Date().toISOString();
  const updatedAt = raw.updatedAt ?? raw.updated_at ?? createdAt;
  return {
    id: raw.id ?? 0,
    threadId: raw.threadId ?? raw.thread_id ?? 0,
    parentId: raw.parentId ?? raw.parent_id,
    content: raw.content ?? '',
    authorId: raw.authorId ?? raw.author_id ?? 0,
    authorName: raw.authorName ?? raw.author_name ?? 'Unknown',
    authorAvatar: raw.authorAvatar ?? raw.author_avatar,
    isAcceptedAnswer: raw.isAcceptedAnswer ?? raw.is_accepted_answer ?? false,
    likeCount: raw.likeCount ?? raw.like_count ?? 0,
    hasLiked: raw.hasLiked ?? raw.has_liked ?? false,
    createdAt,
    updatedAt,
    collegeId: raw.collegeId ?? raw.college_id ?? 0,
  };
}
