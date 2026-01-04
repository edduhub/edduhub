package models

import "time"

type ForumCategory string

const (
	CategoryGeneral      ForumCategory = "general"
	CategoryAcademic     ForumCategory = "academic"
	CategoryAssignment   ForumCategory = "assignment"
	CategoryQuestion     ForumCategory = "question"
	CategoryAnnouncement ForumCategory = "announcement"
)

type ForumThread struct {
	ID          int           `json:"id" db:"id"`
	CollegeID   int           `json:"college_id" db:"college_id"`
	CourseID    int           `json:"course_id" db:"course_id"`
	Category    ForumCategory `json:"category" db:"category"`
	Title       string        `json:"title" db:"title"`
	Content     string        `json:"content" db:"content"`
	AuthorID    int           `json:"author_id" db:"author_id"`
	IsPinned    bool          `json:"is_pinned" db:"is_pinned"`
	IsLocked    bool          `json:"is_locked" db:"is_locked"`
	ViewCount   int           `json:"view_count" db:"view_count"`
	ReplyCount  int           `json:"reply_count" db:"reply_count"`
	LastReplyAt *time.Time    `json:"last_reply_at" db:"last_reply_at"`
	LastReplyBy *int          `json:"last_reply_by" db:"last_reply_by"`
	Tags        []string      `json:"tags" db:"tags"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`

	// Derived fields
	AuthorName   string `json:"author_name" db:"author_name"`
	AuthorAvatar string `json:"author_avatar" db:"author_avatar"`
	CourseName   string `json:"course_name" db:"course_name"`
}

type ForumReply struct {
	ID               int       `json:"id" db:"id"`
	ThreadID         int       `json:"thread_id" db:"thread_id"`
	ParentID         *int      `json:"parent_id" db:"parent_id"`
	Content          string    `json:"content" db:"content"`
	AuthorID         int       `json:"author_id" db:"author_id"`
	IsAcceptedAnswer bool      `json:"is_accepted_answer" db:"is_accepted_answer"`
	LikeCount        int       `json:"like_count" db:"like_count"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	CollegeID        int       `json:"college_id" db:"college_id"`

	// Derived fields
	AuthorName   string `json:"author_name" db:"author_name"`
	AuthorAvatar string `json:"author_avatar" db:"author_avatar"`
	HasLiked     bool   `json:"has_liked" db:"-"`
}

type ForumThreadFilter struct {
	CollegeID  int
	CourseID   *int
	Category   *ForumCategory
	Tag        *string
	AuthorID   *int
	Search     *string
	PinnedOnly *bool
	Limit      int
	Offset     int
}
