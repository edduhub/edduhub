-- Align forum schema with repository/model expectations for existing databases.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'forum_threads' AND column_name = 'course_id'
    ) AND EXISTS (
        SELECT 1 FROM forum_threads WHERE course_id IS NULL
    ) THEN
        -- Keep legacy NULLs to avoid migration failure; code paths should create course-scoped threads.
        NULL;
    ELSIF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'forum_threads' AND column_name = 'course_id'
    ) THEN
        ALTER TABLE forum_threads ALTER COLUMN course_id SET NOT NULL;
    END IF;
END $$;

ALTER TABLE forum_threads ADD COLUMN IF NOT EXISTS category VARCHAR(50) NOT NULL DEFAULT 'general';
ALTER TABLE forum_threads ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[];

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_forum_threads_category'
    ) THEN
        ALTER TABLE forum_threads
            ADD CONSTRAINT chk_forum_threads_category
            CHECK (category IN ('general', 'academic', 'assignment', 'question', 'announcement'));
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'forum_threads_course_id_fkey') THEN
        ALTER TABLE forum_threads DROP CONSTRAINT forum_threads_course_id_fkey;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_forum_threads_course_id'
    ) THEN
        ALTER TABLE forum_threads
            ADD CONSTRAINT fk_forum_threads_course_id
            FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'forum_replies' AND column_name = 'parent_reply_id'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'forum_replies' AND column_name = 'parent_id'
    ) THEN
        ALTER TABLE forum_replies RENAME COLUMN parent_reply_id TO parent_id;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'forum_replies' AND column_name = 'is_solution'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'forum_replies' AND column_name = 'is_accepted_answer'
    ) THEN
        ALTER TABLE forum_replies RENAME COLUMN is_solution TO is_accepted_answer;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'forum_replies' AND column_name = 'upvotes'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'forum_replies' AND column_name = 'like_count'
    ) THEN
        ALTER TABLE forum_replies RENAME COLUMN upvotes TO like_count;
    END IF;
END $$;

ALTER TABLE forum_replies ADD COLUMN IF NOT EXISTS parent_id INTEGER REFERENCES forum_replies(id) ON DELETE CASCADE;
ALTER TABLE forum_replies ADD COLUMN IF NOT EXISTS is_accepted_answer BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE forum_replies ADD COLUMN IF NOT EXISTS like_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE forum_replies ADD COLUMN IF NOT EXISTS college_id INTEGER REFERENCES colleges(id) ON DELETE CASCADE;

UPDATE forum_replies r
SET college_id = t.college_id
FROM forum_threads t
WHERE r.thread_id = t.id
  AND r.college_id IS NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'forum_replies' AND column_name = 'college_id'
    ) AND NOT EXISTS (
        SELECT 1 FROM forum_replies WHERE college_id IS NULL
    ) THEN
        ALTER TABLE forum_replies ALTER COLUMN college_id SET NOT NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_replies_college ON forum_replies(college_id);
