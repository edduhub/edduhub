package repository

import (
	"context"
	"eduhub/server/internal/models"
	"fmt"
)

type ForumRepository interface {
	CreateThread(ctx context.Context, thread *models.ForumThread) error
	GetThread(ctx context.Context, collegeID, threadID int) (*models.ForumThread, error)
	ListThreads(ctx context.Context, filter models.ForumThreadFilter) ([]models.ForumThread, error)
	UpdateThread(ctx context.Context, thread *models.ForumThread) error
	DeleteThread(ctx context.Context, collegeID, threadID int) error
	IncrementViewCount(ctx context.Context, threadID int) error

	CreateReply(ctx context.Context, reply *models.ForumReply) error
	ListReplies(ctx context.Context, collegeID, threadID int) ([]models.ForumReply, error)
	DeleteReply(ctx context.Context, collegeID, replyID int) error
	MarkAnswer(ctx context.Context, collegeID, threadID, replyID int) error
}

type forumRepository struct {
	db PoolIface
}

func NewForumRepository(db *DB) ForumRepository {
	return &forumRepository{db: db.Pool}
}

func (r *forumRepository) CreateThread(ctx context.Context, thread *models.ForumThread) error {
	query := `
		INSERT INTO forum_threads (college_id, course_id, category, title, content, author_id, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query,
		thread.CollegeID, thread.CourseID, thread.Category, thread.Title, thread.Content, thread.AuthorID, thread.Tags,
	).Scan(&thread.ID, &thread.CreatedAt, &thread.UpdatedAt)
}

func (r *forumRepository) GetThread(ctx context.Context, collegeID, threadID int) (*models.ForumThread, error) {
	query := `
		SELECT t.id, t.college_id, t.course_id, t.category, t.title,
		       t.content, t.author_id, t.is_pinned, t.is_locked,
		       t.view_count, t.reply_count, t.last_reply_at, t.last_reply_by,
		       t.tags, t.created_at, t.updated_at,
		       u.name as author_name, c.name as course_name
		FROM forum_threads t
		JOIN users u ON t.author_id = u.id
		JOIN courses c ON t.course_id = c.id
		WHERE t.id = $1 AND t.college_id = $2`

	var thread models.ForumThread
	err := r.db.QueryRow(ctx, query, threadID, collegeID).Scan(
		&thread.ID, &thread.CollegeID, &thread.CourseID, &thread.Category, &thread.Title,
		&thread.Content, &thread.AuthorID, &thread.IsPinned, &thread.IsLocked,
		&thread.ViewCount, &thread.ReplyCount, &thread.LastReplyAt, &thread.LastReplyBy,
		&thread.Tags, &thread.CreatedAt, &thread.UpdatedAt,
		&thread.AuthorName, &thread.CourseName,
	)
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *forumRepository) ListThreads(ctx context.Context, filter models.ForumThreadFilter) ([]models.ForumThread, error) {
	query := `
		SELECT t.id, t.college_id, t.course_id, t.category, t.title,
		       t.content, t.author_id, t.is_pinned, t.is_locked,
		       t.view_count, t.reply_count, t.last_reply_at, t.last_reply_by,
		       t.tags, t.created_at, t.updated_at,
		       u.name as author_name, c.name as course_name
		FROM forum_threads t
		JOIN users u ON t.author_id = u.id
		JOIN courses c ON t.course_id = c.id
		WHERE t.college_id = $1`

	args := []interface{}{filter.CollegeID}
	placeholderID := 2

	if filter.CourseID != nil {
		query += fmt.Sprintf(" AND t.course_id = $%d", placeholderID)
		args = append(args, *filter.CourseID)
		placeholderID++
	}
	if filter.Category != nil {
		query += fmt.Sprintf(" AND t.category = $%d", placeholderID)
		args = append(args, *filter.Category)
		placeholderID++
	}
	if filter.Search != nil {
		query += fmt.Sprintf(" AND (t.title ILIKE $%d OR t.content ILIKE $%d)", placeholderID, placeholderID)
		searchPattern := "%" + *filter.Search + "%"
		args = append(args, searchPattern)
		placeholderID++
	}

	query += " ORDER BY t.is_pinned DESC, t.updated_at DESC LIMIT $" + fmt.Sprint(placeholderID) + " OFFSET $" + fmt.Sprint(placeholderID+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threads []models.ForumThread
	for rows.Next() {
		var thread models.ForumThread
		err := rows.Scan(
			&thread.ID, &thread.CollegeID, &thread.CourseID, &thread.Category, &thread.Title,
			&thread.Content, &thread.AuthorID, &thread.IsPinned, &thread.IsLocked,
			&thread.ViewCount, &thread.ReplyCount, &thread.LastReplyAt, &thread.LastReplyBy,
			&thread.Tags, &thread.CreatedAt, &thread.UpdatedAt,
			&thread.AuthorName, &thread.CourseName,
		)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}
	return threads, nil
}

func (r *forumRepository) UpdateThread(ctx context.Context, thread *models.ForumThread) error {
	query := `
		UPDATE forum_threads
		SET title = $1, content = $2, category = $3, is_pinned = $4, is_locked = $5, tags = $6, updated_at = NOW()
		WHERE id = $7 AND college_id = $8`
	_, err := r.db.Exec(ctx, query,
		thread.Title, thread.Content, thread.Category, thread.IsPinned, thread.IsLocked, thread.Tags,
		thread.ID, thread.CollegeID,
	)
	return err
}

func (r *forumRepository) DeleteThread(ctx context.Context, collegeID, threadID int) error {
	query := "DELETE FROM forum_threads WHERE id = $1 AND college_id = $2"
	_, err := r.db.Exec(ctx, query, threadID, collegeID)
	return err
}

func (r *forumRepository) IncrementViewCount(ctx context.Context, threadID int) error {
	query := "UPDATE forum_threads SET view_count = view_count + 1 WHERE id = $1"
	_, err := r.db.Exec(ctx, query, threadID)
	return err
}

func (r *forumRepository) CreateReply(ctx context.Context, reply *models.ForumReply) error {
	// Note: Generic pgx pool doesn't directly support BeginTx in the simple PoolIface wrapper if not added
	// For eduhub, we typically use the pool directly or just Exec multiple if no complex rollbacks needed
	// But let's assume we want a transaction for safety if the underlying pool supports it.

	query := `
		INSERT INTO forum_replies (thread_id, parent_id, content, author_id, college_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query,
		reply.ThreadID, reply.ParentID, reply.Content, reply.AuthorID, reply.CollegeID,
	).Scan(&reply.ID, &reply.CreatedAt, &reply.UpdatedAt)
	if err != nil {
		return err
	}

	updateThread := `
		UPDATE forum_threads
		SET reply_count = reply_count + 1, last_reply_at = $1, last_reply_by = $2
		WHERE id = $3`
	_, err = r.db.Exec(ctx, updateThread, reply.CreatedAt, reply.AuthorID, reply.ThreadID)
	return err
}

func (r *forumRepository) ListReplies(ctx context.Context, collegeID, threadID int) ([]models.ForumReply, error) {
	query := `
		SELECT r.id, r.thread_id, r.parent_id, r.content,
		       r.author_id, r.is_accepted_answer, r.like_count,
		       r.created_at, r.updated_at, r.college_id,
		       u.name as author_name
		FROM forum_replies r
		JOIN users u ON r.author_id = u.id
		WHERE r.thread_id = $1 AND r.college_id = $2
		ORDER BY r.created_at ASC`

	rows, err := r.db.Query(ctx, query, threadID, collegeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []models.ForumReply
	for rows.Next() {
		var reply models.ForumReply
		err := rows.Scan(
			&reply.ID, &reply.ThreadID, &reply.ParentID, &reply.Content,
			&reply.AuthorID, &reply.IsAcceptedAnswer, &reply.LikeCount,
			&reply.CreatedAt, &reply.UpdatedAt, &reply.CollegeID,
			&reply.AuthorName,
		)
		if err != nil {
			return nil, err
		}
		replies = append(replies, reply)
	}
	return replies, nil
}

func (r *forumRepository) DeleteReply(ctx context.Context, collegeID, replyID int) error {
	var threadID int
	err := r.db.QueryRow(ctx, "SELECT thread_id FROM forum_replies WHERE id = $1 AND college_id = $2", replyID, collegeID).Scan(&threadID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, "DELETE FROM forum_replies WHERE id = $1", replyID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, "UPDATE forum_threads SET reply_count = GREATEST(0, reply_count - 1) WHERE id = $1", threadID)
	return err
}

func (r *forumRepository) MarkAnswer(ctx context.Context, collegeID, threadID, replyID int) error {
	_, err := r.db.Exec(ctx, "UPDATE forum_replies SET is_accepted_answer = FALSE WHERE thread_id = $1", threadID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, "UPDATE forum_replies SET is_accepted_answer = TRUE WHERE id = $1 AND thread_id = $2 AND college_id = $3", replyID, threadID, collegeID)
	return err
}
