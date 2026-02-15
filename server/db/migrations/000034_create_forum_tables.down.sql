DROP TRIGGER IF EXISTS update_replies_updated_at ON forum_replies;
DROP TRIGGER IF EXISTS update_threads_updated_at ON forum_threads;

DROP TABLE IF EXISTS forum_thread_views;
DROP TABLE IF EXISTS forum_replies;
DROP TABLE IF EXISTS forum_threads;
