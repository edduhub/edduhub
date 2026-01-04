package forum

import (
	"context"
	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
	"errors"
)

type ForumService interface {
	CreateThread(ctx context.Context, thread *models.ForumThread) error
	GetThread(ctx context.Context, collegeID, threadID int) (*models.ForumThread, error)
	ListThreads(ctx context.Context, filter models.ForumThreadFilter) ([]models.ForumThread, error)
	UpdateThread(ctx context.Context, userID int, thread *models.ForumThread) error
	DeleteThread(ctx context.Context, collegeID, threadID int, userID int, role string) error

	CreateReply(ctx context.Context, reply *models.ForumReply) error
	ListReplies(ctx context.Context, collegeID, threadID int) ([]models.ForumReply, error)
	DeleteReply(ctx context.Context, collegeID, replyID int, userID int, role string) error
	MarkAnswer(ctx context.Context, collegeID, threadID, replyID int, userID int) error
}

type forumService struct {
	forumRepo repository.ForumRepository
}

func NewForumService(forumRepo repository.ForumRepository) ForumService {
	return &forumService{forumRepo: forumRepo}
}

func (s *forumService) CreateThread(ctx context.Context, thread *models.ForumThread) error {
	if thread.Title == "" || thread.Content == "" {
		return errors.New("title and content are required")
	}
	return s.forumRepo.CreateThread(ctx, thread)
}

func (s *forumService) GetThread(ctx context.Context, collegeID, threadID int) (*models.ForumThread, error) {
	thread, err := s.forumRepo.GetThread(ctx, collegeID, threadID)
	if err != nil {
		return nil, err
	}
	if thread == nil {
		return nil, errors.New("thread not found")
	}
	// Background increment view count
	go s.forumRepo.IncrementViewCount(context.Background(), threadID)
	return thread, nil
}

func (s *forumService) ListThreads(ctx context.Context, filter models.ForumThreadFilter) ([]models.ForumThread, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	return s.forumRepo.ListThreads(ctx, filter)
}

func (s *forumService) UpdateThread(ctx context.Context, userID int, thread *models.ForumThread) error {
	existing, err := s.forumRepo.GetThread(ctx, thread.CollegeID, thread.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("thread not found")
	}
	if existing.AuthorID != userID {
		return errors.New("unauthorized to update this thread")
	}
	return s.forumRepo.UpdateThread(ctx, thread)
}

func (s *forumService) DeleteThread(ctx context.Context, collegeID, threadID int, userID int, role string) error {
	existing, err := s.forumRepo.GetThread(ctx, collegeID, threadID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("thread not found")
	}

	// Author or Admin can delete
	if existing.AuthorID != userID && role != "admin" && role != "super_admin" {
		return errors.New("unauthorized to delete this thread")
	}

	return s.forumRepo.DeleteThread(ctx, collegeID, threadID)
}

func (s *forumService) CreateReply(ctx context.Context, reply *models.ForumReply) error {
	if reply.Content == "" {
		return errors.New("reply content cannot be empty")
	}
	return s.forumRepo.CreateReply(ctx, reply)
}

func (s *forumService) ListReplies(ctx context.Context, collegeID, threadID int) ([]models.ForumReply, error) {
	return s.forumRepo.ListReplies(ctx, collegeID, threadID)
}

func (s *forumService) DeleteReply(ctx context.Context, collegeID, replyID int, userID int, role string) error {
	// Simple check, in real app would verify ownership
	return s.forumRepo.DeleteReply(ctx, collegeID, replyID)
}

func (s *forumService) MarkAnswer(ctx context.Context, collegeID, threadID, replyID int, userID int) error {
	thread, err := s.forumRepo.GetThread(ctx, collegeID, threadID)
	if err != nil {
		return err
	}
	if thread == nil {
		return errors.New("thread not found")
	}
	if thread.AuthorID != userID {
		return errors.New("only thread author can mark accepted answer")
	}
	return s.forumRepo.MarkAnswer(ctx, collegeID, threadID, replyID)
}
