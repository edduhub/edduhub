package handler

import (
	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/forum"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type ForumHandler struct {
	forumService forum.ForumService
}

func NewForumHandler(forumService forum.ForumService) *ForumHandler {
	return &ForumHandler{forumService: forumService}
}

func (h *ForumHandler) ListThreads(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	filter := models.ForumThreadFilter{
		CollegeID: collegeID,
		Limit:     limit,
		Offset:    offset,
	}

	courseIDStr := c.QueryParam("course_id")
	if courseIDStr == "" {
		courseIDStr = c.QueryParam("courseId")
	}
	if courseIDStr != "" {
		courseID, err := strconv.Atoi(courseIDStr)
		if err != nil || courseID <= 0 {
			return helpers.Error(c, "invalid course_id", 400)
		}
		filter.CourseID = &courseID
	}

	if cat := c.QueryParam("category"); cat != "" {
		cf := models.ForumCategory(cat)
		if !cf.IsValid() {
			return helpers.Error(c, "invalid category", 400)
		}
		filter.Category = &cf
	}
	if query := strings.TrimSpace(c.QueryParam("search")); query != "" {
		filter.Search = &query
	}

	threads, err := h.forumService.ListThreads(c.Request().Context(), filter)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}
	return helpers.Success(c, threads, 200)
}

func (h *ForumHandler) GetThread(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	threadID, err := strconv.Atoi(c.Param("threadID"))
	if err != nil {
		return helpers.Error(c, "invalid thread ID", 400)
	}

	thread, err := h.forumService.GetThread(c.Request().Context(), collegeID, threadID)
	if err != nil {
		return helpers.Error(c, err.Error(), 404)
	}
	return helpers.Success(c, thread, 200)
}

func (h *ForumHandler) CreateThread(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return err
	}

	var thread models.ForumThread
	if err := c.Bind(&thread); err != nil {
		return helpers.Error(c, "invalid body", 400)
	}

	thread.CollegeID = collegeID
	thread.AuthorID = userID

	if err := h.forumService.CreateThread(c.Request().Context(), &thread); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}
	return helpers.Success(c, thread, 201)
}

func (h *ForumHandler) ListReplies(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	threadID, err := strconv.Atoi(c.Param("threadID"))
	if err != nil {
		return helpers.Error(c, "invalid thread ID", 400)
	}

	replies, err := h.forumService.ListReplies(c.Request().Context(), collegeID, threadID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}
	return helpers.Success(c, replies, 200)
}

func (h *ForumHandler) CreateReply(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return err
	}
	threadID, err := strconv.Atoi(c.Param("threadID"))
	if err != nil {
		return helpers.Error(c, "invalid thread ID", 400)
	}

	var reply models.ForumReply
	if err := c.Bind(&reply); err != nil {
		return helpers.Error(c, "invalid body", 400)
	}

	reply.CollegeID = collegeID
	reply.AuthorID = userID
	reply.ThreadID = threadID

	if err := h.forumService.CreateReply(c.Request().Context(), &reply); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}
	return helpers.Success(c, reply, 201)
}
