package handler

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/audit"
	"eduhub/server/internal/services/profile"
	"eduhub/server/internal/services/storage"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ProfileHandler struct {
	profileService profile.ProfileService
	auditService   audit.AuditService
	storageService storage.StorageService
}

func NewProfileHandler(profileService profile.ProfileService, auditService audit.AuditService, storageService storage.StorageService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		auditService:   auditService,
		storageService: storageService,
	}
}

// GetUserProfile retrieves the current user's profile
func (h *ProfileHandler) GetUserProfile(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	profileData, err := h.profileService.GetProfileByUserID(c.Request().Context(), userID)
	if err != nil {
		return helpers.Error(c, "profile not found", 404)
	}

	return helpers.Success(c, profileData, 200)
}

// UploadProfileImage handles profile picture upload
func (h *ProfileHandler) UploadProfileImage(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "college ID required", 401)
	}

	// Get file from form
	file, err := c.FormFile("image")
	if err != nil {
		return helpers.Error(c, "image file is required", 400)
	}

	// Validate file size (5MB limit for images)
	if file.Size > 5*1024*1024 {
		return helpers.Error(c, "image size exceeds 5MB limit", 400)
	}

	// Validate file type
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedTypes[ext] {
		return helpers.Error(c, "file type not allowed. Only JPG, PNG, GIF are supported", 400)
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return helpers.Error(c, "failed to open file", 500)
	}
	defer src.Close()

	// Generate unique filename
	uniqueFilename := fmt.Sprintf("%s_%s_profile%s", uuid.New().String(), strconv.Itoa(userID), ext)

	// Build storage path: college_id/profiles/user_id/filename
	objectKey := fmt.Sprintf("%d/profiles/%d/%s", collegeID, userID, uniqueFilename)

	// Upload to storage
	fileURL, err := h.storageService.UploadFile(c.Request().Context(), objectKey, src, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	// Update profile with new image URL
	updateReq := &models.UpdateProfileRequest{
		ProfileImage: &fileURL,
	}

	// Get current profile for history tracking
	currentProfile, err := h.profileService.GetProfileByUserID(c.Request().Context(), userID)
	if err != nil {
		return helpers.Error(c, "failed to get current profile", 500)
	}

	err = h.profileService.UpdateProfile(c.Request().Context(), userID, updateReq)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	// Log profile image change
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	profileHistory := &models.ProfileHistory{
		ProfileID: currentProfile.ID,
		UserID:    userID,
		Action:    "UPLOAD_IMAGE",
		Field:     "profile_image",
		OldValue:  currentProfile.ProfileImage,
		NewValue:  fileURL,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	if err := h.profileService.CreateProfileHistory(c.Request().Context(), profileHistory); err != nil {
		c.Logger().Error("Failed to log profile history:", err)
	}

	// Log audit event
	auditLog := &models.AuditLog{
		CollegeID:  collegeID,
		UserID:     userID,
		Action:     "UPLOAD",
		EntityType: "profile_image",
		EntityID:   currentProfile.ID,
		Changes:    map[string]interface{}{"filename": file.Filename, "size": file.Size},
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	if err := h.auditService.LogAction(c.Request().Context(), auditLog); err != nil {
		c.Logger().Error("Failed to log audit event:", err)
	}

	return helpers.Success(c, map[string]interface{}{
		"url":      fileURL,
		"filename": file.Filename,
		"size":     file.Size,
	}, 201)
}

// GetProfileHistory retrieves the change history for a user's profile
func (h *ProfileHandler) GetProfileHistory(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	// Get current profile to verify ownership
	profile, err := h.profileService.GetProfileByUserID(c.Request().Context(), userID)
	if err != nil {
		return helpers.Error(c, "profile not found", 404)
	}

	// Parse pagination parameters
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := 50 // default limit
	offset := 0 // default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	history, err := h.profileService.GetProfileHistory(c.Request().Context(), profile.ID, limit, offset)
	if err != nil {
		return helpers.Error(c, "failed to retrieve profile history", 500)
	}

	return helpers.Success(c, map[string]interface{}{
		"history": history,
		"limit":   limit,
		"offset":  offset,
	}, 200)
}

// Helper methods

func (h *ProfileHandler) logProfileChanges(ctx context.Context, currentProfile *models.Profile, req *models.UpdateProfileRequest, userID, collegeID int, ipAddress, userAgent string) {
	now := time.Now()

	if req.Bio != nil && *req.Bio != currentProfile.Bio {
		history := &models.ProfileHistory{
			ProfileID: currentProfile.ID,
			UserID:    userID,
			Action:    "UPDATE",
			Field:     "bio",
			OldValue:  currentProfile.Bio,
			NewValue:  *req.Bio,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			CreatedAt: now,
		}
		if err := h.profileService.CreateProfileHistory(ctx, history); err != nil {
			// Log error but continue
		}
	}

	if req.PhoneNumber != nil && *req.PhoneNumber != currentProfile.PhoneNumber {
		history := &models.ProfileHistory{
			ProfileID: currentProfile.ID,
			UserID:    userID,
			Action:    "UPDATE",
			Field:     "phone_number",
			OldValue:  currentProfile.PhoneNumber,
			NewValue:  *req.PhoneNumber,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			CreatedAt: now,
		}
		if err := h.profileService.CreateProfileHistory(ctx, history); err != nil {
			// Log error but continue
		}
	}

	if req.Address != nil && *req.Address != currentProfile.Address {
		history := &models.ProfileHistory{
			ProfileID: currentProfile.ID,
			UserID:    userID,
			Action:    "UPDATE",
			Field:     "address",
			OldValue:  currentProfile.Address,
			NewValue:  *req.Address,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			CreatedAt: now,
		}
		if err := h.profileService.CreateProfileHistory(ctx, history); err != nil {
			// Log error but continue
		}
	}

	if req.DateOfBirth != nil && (!currentProfile.DateOfBirth.IsZero() && !req.DateOfBirth.Equal(currentProfile.DateOfBirth)) {
		history := &models.ProfileHistory{
			ProfileID: currentProfile.ID,
			UserID:    userID,
			Action:    "UPDATE",
			Field:     "date_of_birth",
			OldValue:  currentProfile.DateOfBirth.Format("2006-01-02"),
			NewValue:  req.DateOfBirth.Format("2006-01-02"),
			IPAddress: ipAddress,
			UserAgent: userAgent,
			CreatedAt: now,
		}
		if err := h.profileService.CreateProfileHistory(ctx, history); err != nil {
			// Log error but continue
		}
	}
}

func getUpdatedFields(req *models.UpdateProfileRequest) []string {
	var fields []string
	if req.UserID != nil {
		fields = append(fields, "user_id")
	}
	if req.CollegeID != nil {
		fields = append(fields, "college_id")
	}
	if req.Bio != nil {
		fields = append(fields, "bio")
	}
	if req.ProfileImage != nil {
		fields = append(fields, "profile_image")
	}
	if req.PhoneNumber != nil {
		fields = append(fields, "phone_number")
	}
	if req.Address != nil {
		fields = append(fields, "address")
	}
	if req.DateOfBirth != nil {
		fields = append(fields, "date_of_birth")
	}
	if req.Preferences != nil {
		fields = append(fields, "preferences")
	}
	if req.SocialLinks != nil {
		fields = append(fields, "social_links")
	}
	return fields
}

// UpdateUserProfile updates the current user's profile with enhanced validation and history tracking
func (h *ProfileHandler) UpdateUserProfile(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "college ID required", 401)
	}

	var req models.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	// Validate request fields
	if err := c.Validate(&req); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	// Get current profile for comparison
	currentProfile, err := h.profileService.GetProfileByUserID(c.Request().Context(), userID)
	if err != nil {
		return helpers.Error(c, "failed to get current profile", 500)
	}

	// Update profile
	err = h.profileService.UpdateProfile(c.Request().Context(), userID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	// Log changes to audit trail
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	// Track changes for each field
	h.logProfileChanges(c.Request().Context(), currentProfile, &req, userID, collegeID, ipAddress, userAgent)

	// Log audit event
	auditLog := &models.AuditLog{
		CollegeID: collegeID,
		UserID:    userID,
		Action:    "UPDATE",
		EntityType: "profile",
		EntityID:  currentProfile.ID,
		Changes:   map[string]interface{}{"updated_fields": getUpdatedFields(&req)},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := h.auditService.LogAction(c.Request().Context(), auditLog); err != nil {
		// Log error but don't fail the request
		c.Logger().Error("Failed to log audit event:", err)
	}

	return helpers.Success(c, "Profile updated successfully", 200)
}

// GetProfile retrieves a specific user's profile (admin only)
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	profileIDStr := c.Param("profileID")
	profileID, err := strconv.Atoi(profileIDStr)
	if err != nil {
		return helpers.Error(c, "invalid profile ID", 400)
	}

	profileData, err := h.profileService.GetProfileByID(c.Request().Context(), profileID)
	if err != nil {
		return helpers.Error(c, "profile not found", 404)
	}

	return helpers.Success(c, profileData, 200)
}
