-- Migration: Create forum tables
-- This enables discussion forum feature

CREATE TABLE IF NOT EXISTS forum_threads (
    id SERIAL PRIMARY KEY,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    course_id INTEGER REFERENCES courses(course_id) ON DELETE SET NULL,
    author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    author_name VARCHAR(100),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    is_pinned BOOLEAN DEFAULT FALSE,
    is_locked BOOLEAN DEFAULT FALSE,
    view_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    last_reply_at TIMESTAMP,
    last_reply_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS forum_replies (
    id SERIAL PRIMARY KEY,
    thread_id INTEGER NOT NULL REFERENCES forum_threads(id) ON DELETE CASCADE,
    author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    author_name VARCHAR(100),
    content TEXT NOT NULL,
    parent_reply_id INTEGER REFERENCES forum_replies(id) ON DELETE CASCADE,
    is_solution BOOLEAN DEFAULT FALSE,
    upvotes INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS forum_thread_views (
    id SERIAL PRIMARY KEY,
    thread_id INTEGER NOT NULL REFERENCES forum_threads(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(thread_id, user_id)
);

-- Indexes
CREATE INDEX idx_threads_college ON forum_threads(college_id);
CREATE INDEX idx_threads_course ON forum_threads(course_id);
CREATE INDEX idx_threads_author ON forum_threads(author_id);
CREATE INDEX idx_threads_pinned ON forum_threads(is_pinned, created_at);
CREATE INDEX idx_threads_last_reply ON forum_threads(last_reply_at);
CREATE INDEX idx_replies_thread ON forum_replies(thread_id);
CREATE INDEX idx_replies_author ON forum_replies(author_id);
CREATE INDEX idx_replies_parent ON forum_replies(parent_reply_id);
CREATE INDEX idx_thread_views_thread ON forum_thread_views(thread_id);
CREATE INDEX idx_thread_views_user ON forum_thread_views(user_id);

-- Triggers
CREATE TRIGGER update_threads_updated_at 
    BEFORE UPDATE ON forum_threads 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_replies_updated_at 
    BEFORE UPDATE ON forum_replies 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
