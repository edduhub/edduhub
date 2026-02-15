-- Migration: Create forum tables
-- This enables discussion forum feature

CREATE TABLE IF NOT EXISTS forum_threads (
    id SERIAL PRIMARY KEY,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    view_count INTEGER NOT NULL DEFAULT 0,
    reply_count INTEGER NOT NULL DEFAULT 0,
    last_reply_at TIMESTAMP,
    last_reply_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    tags TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_forum_threads_category CHECK (category IN ('general', 'academic', 'assignment', 'question', 'announcement'))
);

CREATE TABLE IF NOT EXISTS forum_replies (
    id SERIAL PRIMARY KEY,
    thread_id INTEGER NOT NULL REFERENCES forum_threads(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES forum_replies(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_accepted_answer BOOLEAN NOT NULL DEFAULT FALSE,
    like_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS forum_thread_views (
    id SERIAL PRIMARY KEY,
    thread_id INTEGER NOT NULL REFERENCES forum_threads(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(thread_id, user_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_threads_college ON forum_threads(college_id);
CREATE INDEX IF NOT EXISTS idx_threads_course ON forum_threads(course_id);
CREATE INDEX IF NOT EXISTS idx_threads_author ON forum_threads(author_id);
CREATE INDEX IF NOT EXISTS idx_threads_pinned ON forum_threads(is_pinned, created_at);
CREATE INDEX IF NOT EXISTS idx_threads_last_reply ON forum_threads(last_reply_at);
CREATE INDEX IF NOT EXISTS idx_replies_thread ON forum_replies(thread_id);
CREATE INDEX IF NOT EXISTS idx_replies_author ON forum_replies(author_id);
CREATE INDEX IF NOT EXISTS idx_replies_parent ON forum_replies(parent_id);
CREATE INDEX IF NOT EXISTS idx_replies_college ON forum_replies(college_id);
CREATE INDEX IF NOT EXISTS idx_thread_views_thread ON forum_thread_views(thread_id);
CREATE INDEX IF NOT EXISTS idx_thread_views_user ON forum_thread_views(user_id);

-- Triggers
DROP TRIGGER IF EXISTS update_threads_updated_at ON forum_threads;
CREATE TRIGGER update_threads_updated_at
    BEFORE UPDATE ON forum_threads
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_replies_updated_at ON forum_replies;
CREATE TRIGGER update_replies_updated_at
    BEFORE UPDATE ON forum_replies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
