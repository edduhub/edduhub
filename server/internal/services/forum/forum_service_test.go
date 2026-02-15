package forum

import (
	"context"
	"eduhub/server/internal/models"
	"testing"
)

type forumRepositoryStub struct {
	createThreadFn func(ctx context.Context, thread *models.ForumThread) error
	createReplyFn  func(ctx context.Context, reply *models.ForumReply) error
}

func (s *forumRepositoryStub) CreateThread(ctx context.Context, thread *models.ForumThread) error {
	if s.createThreadFn != nil {
		return s.createThreadFn(ctx, thread)
	}
	return nil
}

func (s *forumRepositoryStub) GetThread(context.Context, int, int) (*models.ForumThread, error) {
	return nil, nil
}

func (s *forumRepositoryStub) ListThreads(context.Context, models.ForumThreadFilter) ([]models.ForumThread, error) {
	return nil, nil
}

func (s *forumRepositoryStub) UpdateThread(context.Context, *models.ForumThread) error {
	return nil
}

func (s *forumRepositoryStub) DeleteThread(context.Context, int, int) error {
	return nil
}

func (s *forumRepositoryStub) IncrementViewCount(context.Context, int) error {
	return nil
}

func (s *forumRepositoryStub) CreateReply(ctx context.Context, reply *models.ForumReply) error {
	if s.createReplyFn != nil {
		return s.createReplyFn(ctx, reply)
	}
	return nil
}

func (s *forumRepositoryStub) GetReply(context.Context, int, int) (*models.ForumReply, error) {
	return nil, nil
}

func (s *forumRepositoryStub) ListReplies(context.Context, int, int) ([]models.ForumReply, error) {
	return nil, nil
}

func (s *forumRepositoryStub) DeleteReply(context.Context, int, int) error {
	return nil
}

func (s *forumRepositoryStub) MarkAnswer(context.Context, int, int, int) error {
	return nil
}

func TestCreateThreadValidatesAndSanitizes(t *testing.T) {
	var captured *models.ForumThread
	repo := &forumRepositoryStub{
		createThreadFn: func(_ context.Context, thread *models.ForumThread) error {
			copied := *thread
			captured = &copied
			return nil
		},
	}
	service := NewForumService(repo)

	thread := &models.ForumThread{
		CourseID: 42,
		Title:    "  Thread Title  ",
		Content:  "  Thread Content  ",
		Tags:     []string{"  React ", "react", "API"},
	}
	err := service.CreateThread(context.Background(), thread)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if captured == nil {
		t.Fatalf("expected repository call to capture thread")
	}
	if captured.Category != models.CategoryGeneral {
		t.Fatalf("expected default category %q, got %q", models.CategoryGeneral, captured.Category)
	}
	if captured.Title != "Thread Title" {
		t.Fatalf("expected title to be trimmed, got %q", captured.Title)
	}
	if captured.Content != "Thread Content" {
		t.Fatalf("expected content to be trimmed, got %q", captured.Content)
	}
	if len(captured.Tags) != 2 || captured.Tags[0] != "React" || captured.Tags[1] != "API" {
		t.Fatalf("unexpected tags after sanitize: %#v", captured.Tags)
	}
}

func TestCreateThreadRejectsInvalidInput(t *testing.T) {
	service := NewForumService(&forumRepositoryStub{})

	cases := []struct {
		name   string
		thread models.ForumThread
	}{
		{
			name: "missing course",
			thread: models.ForumThread{
				Title:   "Thread",
				Content: "Body",
			},
		},
		{
			name: "invalid category",
			thread: models.ForumThread{
				CourseID: 1,
				Title:    "Thread",
				Content:  "Body",
				Category: "invalid",
			},
		},
		{
			name: "missing content",
			thread: models.ForumThread{
				CourseID: 1,
				Title:    "Thread",
				Content:  "   ",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.CreateThread(context.Background(), &tc.thread)
			if err == nil {
				t.Fatalf("expected validation error for %s", tc.name)
			}
		})
	}
}

func TestCreateReplyRejectsWhitespaceOnlyContent(t *testing.T) {
	service := NewForumService(&forumRepositoryStub{})
	reply := &models.ForumReply{Content: "   "}

	if err := service.CreateReply(context.Background(), reply); err == nil {
		t.Fatalf("expected error for empty reply content")
	}
}
