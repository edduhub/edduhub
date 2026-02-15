import {
  normalizeForumThread,
  normalizeForumReply,
  ForumThreadApi,
  ForumReplyApi,
} from '@/lib/forum';

describe('normalizeForumThread', () => {
  it('normalizes snake_case to camelCase', () => {
    const raw: ForumThreadApi = {
      id: 1,
      course_id: 2,
      course_name: 'Math 101',
      title: 'Test Thread',
      content: 'Test content',
      author_id: 3,
      author_name: 'John Doe',
      is_pinned: true,
      is_locked: false,
      view_count: 100,
      reply_count: 5,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
    };

    const result = normalizeForumThread(raw);

    expect(result.id).toBe(1);
    expect(result.courseId).toBe(2);
    expect(result.courseName).toBe('Math 101');
    expect(result.title).toBe('Test Thread');
    expect(result.authorId).toBe(3);
    expect(result.authorName).toBe('John Doe');
    expect(result.isPinned).toBe(true);
    expect(result.viewCount).toBe(100);
    expect(result.replyCount).toBe(5);
  });

  it('uses default values for missing fields', () => {
    const raw: ForumThreadApi = {};

    const result = normalizeForumThread(raw);

    expect(result.id).toBe(0);
    expect(result.courseId).toBe(0);
    expect(result.title).toBe('');
    expect(result.content).toBe('');
    expect(result.authorName).toBe('Unknown');
    expect(result.isPinned).toBe(false);
    expect(result.viewCount).toBe(0);
    expect(result.replyCount).toBe(0);
    expect(result.tags).toEqual([]);
    expect(result.category).toBe('general');
  });

  it('handles optional fields', () => {
    const raw: ForumThreadApi = {
      id: 1,
      author_avatar: 'avatar.png',
      tags: ['help', 'urgent'],
      last_reply_at: '2024-01-03T00:00:00Z',
      last_reply_by: 5,
    };

    const result = normalizeForumThread(raw);

    expect(result.authorAvatar).toBe('avatar.png');
    expect(result.tags).toEqual(['help', 'urgent']);
    expect(result.lastReplyAt).toBe('2024-01-03T00:00:00Z');
    expect(result.lastReplyBy).toBe(5);
  });

  it('handles tags as non-array', () => {
    const raw: ForumThreadApi = {
      tags: 'not-an-array' as unknown as string[],
    };

    const result = normalizeForumThread(raw);
    expect(result.tags).toEqual([]);
  });

  it('normalizes valid categories', () => {
    expect(normalizeForumThread({ category: 'general' } as ForumThreadApi).category).toBe('general');
    expect(normalizeForumThread({ category: 'academic' } as ForumThreadApi).category).toBe('academic');
    expect(normalizeForumThread({ category: 'assignment' } as ForumThreadApi).category).toBe('assignment');
    expect(normalizeForumThread({ category: 'question' } as ForumThreadApi).category).toBe('question');
    expect(normalizeForumThread({ category: 'announcement' } as ForumThreadApi).category).toBe('announcement');
  });

  it('defaults invalid categories to general', () => {
    expect(normalizeForumThread({ category: 'invalid' } as ForumThreadApi).category).toBe('general');
    expect(normalizeForumThread({ category: undefined } as ForumThreadApi).category).toBe('general');
  });
});

describe('normalizeForumReply', () => {
  it('normalizes snake_case to camelCase', () => {
    const raw: ForumReplyApi = {
      id: 1,
      thread_id: 2,
      parent_id: 3,
      content: 'Reply content',
      author_id: 4,
      author_name: 'Jane Doe',
      is_accepted_answer: true,
      like_count: 10,
      has_liked: false,
      created_at: '2024-01-01T00:00:00Z',
    };

    const result = normalizeForumReply(raw);

    expect(result.id).toBe(1);
    expect(result.threadId).toBe(2);
    expect(result.parentId).toBe(3);
    expect(result.content).toBe('Reply content');
    expect(result.authorId).toBe(4);
    expect(result.authorName).toBe('Jane Doe');
    expect(result.isAcceptedAnswer).toBe(true);
    expect(result.likeCount).toBe(10);
    expect(result.hasLiked).toBe(false);
  });

  it('uses default values for missing fields', () => {
    const raw: ForumReplyApi = {};

    const result = normalizeForumReply(raw);

    expect(result.id).toBe(0);
    expect(result.threadId).toBe(0);
    expect(result.content).toBe('');
    expect(result.authorName).toBe('Unknown');
    expect(result.isAcceptedAnswer).toBe(false);
    expect(result.likeCount).toBe(0);
    expect(result.hasLiked).toBe(false);
  });

  it('handles optional fields', () => {
    const raw: ForumReplyApi = {
      author_avatar: 'avatar.png',
    };

    const result = normalizeForumReply(raw);
    expect(result.authorAvatar).toBe('avatar.png');
  });
});
