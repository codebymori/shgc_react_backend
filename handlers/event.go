package handlers

import (
	"net/http"
	"strconv"
	"time"

	"sentul-golf-be/config"
	"sentul-golf-be/middleware"
	"sentul-golf-be/models"
	"sentul-golf-be/utils"

	"github.com/gorilla/mux"
)

// EventResponse with simplified author (for list)
type EventResponse struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Slug       string           `json:"slug"`
	Published  bool             `json:"published"`
	ImageURL   string           `json:"image_url"`
	AuthorID   string           `json:"author_id"`
	Author     SimplifiedAuthor `json:"author"`
	EventStart *time.Time       `json:"event_start"`
	EventEnd   *time.Time       `json:"event_end"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// EventDetailResponse with full content (for detail by ID)
type EventDetailResponse struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Content    string           `json:"content"`
	Slug       string           `json:"slug"`
	Published  bool             `json:"published"`
	ImageURL   string           `json:"image_url"`
	AuthorID   string           `json:"author_id"`
	Author     SimplifiedAuthor `json:"author"`
	EventStart *time.Time       `json:"event_start"`
	EventEnd   *time.Time       `json:"event_end"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// GetEvents retrieves all events with pagination
func GetEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get pagination parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	
	page := 1
	limit := 10 // Default 10 items per page
	
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	offset := (page - 1) * limit

	// Filter by published status if not admin
	claims, ok := ctx.Value(middleware.UserContextKey).(*utils.Claims)
	publishedOnly := !ok || claims.Role != string(models.RoleAdmin)
	
	// Try cache first
	cacheKey := utils.BuildCacheKey("event", "list", "page", page, "limit", limit, "published", publishedOnly)
	type CachedEventResponse struct {
		EventResponse []EventResponse `json:"events"`
		Meta          *utils.Meta     `json:"meta"`
	}
	var cached CachedEventResponse
	if err := utils.CacheGet(ctx, cacheKey, &cached); err == nil {
		utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
			"events": cached.EventResponse,
		}, cached.Meta)
		return
	}
	
	// Cache miss - get from database
	db := config.GetDB()
	var events []models.Event
	query := db.Preload("Author")
	if publishedOnly {
		query = query.Where("published = ?", true)
	}
	
	// Count total items
	var total int64
	countQuery := db.Model(&models.Event{})
	if !ok || claims.Role != string(models.RoleAdmin) {
		countQuery = countQuery.Where("published = ?", true)
	}
	countQuery.Count(&total)
	
	// Get paginated results
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		utils.RespondInternalError(w)
		return
	}
	
	// Transform to response format with simplified author
	eventsResponse := make([]EventResponse, len(events))
	for i, e := range events {
		eventsResponse[i] = EventResponse{
			ID:         e.ID,
			Title:      e.Title,
			Slug:       e.Slug,
			Published:  e.Published,
			ImageURL:   e.ImageURL,
			AuthorID:   e.AuthorID,
			Author: SimplifiedAuthor{
				ID:   e.Author.ID,
				Name: e.Author.Name,
			},
			EventStart: e.EventStart,
			EventEnd:   e.EventEnd,
			CreatedAt:  e.CreatedAt,
			UpdatedAt:  e.UpdatedAt,
		}
	}
	
	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	meta := &utils.Meta{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}

	// Cache the response
	_ = utils.CacheSet(ctx, cacheKey, CachedEventResponse{
		EventResponse: eventsResponse,
		Meta:          meta,
	}, utils.CacheTTLEventsList)

	// Add BASE_URL to response
	baseURL := config.GetEnv("BASE_URL", "")
	for i := range eventsResponse {
		eventsResponse[i].ImageURL = utils.PrependBaseURL(eventsResponse[i].ImageURL, baseURL)
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"events": eventsResponse,
	}, meta)
}

// GetEventBySlug retrieves a single event by slug
func GetEventBySlug(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	slug := params["slug"]
	ctx := r.Context()
	cacheKey := utils.BuildCacheKey("event", "slug", slug)

	// Try cache first
	var response EventDetailResponse
	if err := utils.CacheGet(ctx, cacheKey, &response); err == nil {
		utils.RespondSuccess(w, http.StatusOK, response, nil)
		return
	}

	// Cache miss - get from database
	db := config.GetDB()
	var event models.Event
	if err := db.Preload("Author").Where("slug = ?", slug).First(&event).Error; err != nil {
		utils.RespondNotFound(w, "Event")
		return
	}

	// Build response
	response = EventDetailResponse{
		ID:         event.ID,
		Title:      event.Title,
		Content:    event.Content,
		Slug:       event.Slug,
		Published:  event.Published,
		ImageURL:   event.ImageURL,
		AuthorID:   event.AuthorID,
		Author: SimplifiedAuthor{
			ID:   event.Author.ID,
			Name: event.Author.Name,
		},
		EventStart: event.EventStart,
		EventEnd:   event.EventEnd,
		CreatedAt:  event.CreatedAt,
		UpdatedAt:  event.UpdatedAt,
	}

	// Cache the response
	_ = utils.CacheSet(ctx, cacheKey, response, utils.CacheTTLEventDetail)

	// Add BASE_URL to response
	baseURL := config.GetEnv("BASE_URL", "")
	response.ImageURL = utils.PrependBaseURL(response.ImageURL, baseURL)

	utils.RespondSuccess(w, http.StatusOK, response, nil)
}

// GetEventByID retrieves a single event by ID
func GetEventByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	ctx := r.Context()
	cacheKey := utils.BuildCacheKey("event", "id", id)

	// Try cache first
	var response EventDetailResponse
	if err := utils.CacheGet(ctx, cacheKey, &response); err == nil {
		utils.RespondSuccess(w, http.StatusOK, response, nil)
		return
	}

	// Cache miss - get from database
	db := config.GetDB()
	var event models.Event
	if err := db.Preload("Author").Where("id = ?", id).First(&event).Error; err != nil {
		utils.RespondNotFound(w, "Event")
		return
	}

	// Build response
	response = EventDetailResponse{
		ID:         event.ID,
		Title:      event.Title,
		Content:    event.Content,
		Slug:       event.Slug,
		Published:  event.Published,
		ImageURL:   event.ImageURL,
		AuthorID:   event.AuthorID,
		Author: SimplifiedAuthor{
			ID:   event.Author.ID,
			Name: event.Author.Name,
		},
		EventStart: event.EventStart,
		EventEnd:   event.EventEnd,
		CreatedAt:  event.CreatedAt,
		UpdatedAt:  event.UpdatedAt,
	}

	// Cache the response
	_ = utils.CacheSet(ctx, cacheKey, response, utils.CacheTTLEventDetail)

	// Add BASE_URL to response
	baseURL := config.GetEnv("BASE_URL", "")
	response.ImageURL = utils.PrependBaseURL(response.ImageURL, baseURL)

	utils.RespondSuccess(w, http.StatusOK, response, nil)
}

// CreateEvent creates a new event with image upload
func CreateEvent(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with max memory of 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondBadRequest(w, "Failed to parse form data")
		return
	}

	// Get form values
	title := r.FormValue("title")
	content := r.FormValue("content")
	// Sanitize HTML content to prevent XSS
	content = utils.SanitizeHTML(content)
	slug := r.FormValue("slug")
	published := r.FormValue("published") == "true"
	eventStartStr := r.FormValue("event_start")
	eventEndStr := r.FormValue("event_end")

	// Validate required fields
	fields := make(map[string]string)
	if title == "" {
		fields["title"] = "Title is required"
	}
	if content == "" {
		fields["content"] = "Content is required"
	}
	
	if len(fields) > 0 {
		utils.RespondValidationError(w, fields)
		return
	}

	// Auto-generate slug from title if not provided
	if slug == "" {
		slug = utils.GenerateSlug(title)
	}

	// Parse event start date if provided
	var eventStart *time.Time
	if eventStartStr != "" {
		parsedDate, err := time.Parse(time.RFC3339, eventStartStr)
		if err != nil {
			// Try alternative format (date only, will use 00:00:00)
			parsedDate, err = time.Parse("2006-01-02", eventStartStr)
			if err != nil {
				utils.RespondError(w, http.StatusBadRequest, "INVALID_DATE", "Invalid event_start format. Use RFC3339 (2006-01-02T15:04:05Z07:00) or YYYY-MM-DD", nil)
				return
			}
		}
		eventStart = &parsedDate
	}

	// Parse event end date if provided
	var eventEnd *time.Time
	if eventEndStr != "" {
		parsedDate, err := time.Parse(time.RFC3339, eventEndStr)
		if err != nil {
			// Try alternative format (date only, will use 00:00:00)
			parsedDate, err = time.Parse("2006-01-02", eventEndStr)
			if err != nil {
				utils.RespondError(w, http.StatusBadRequest, "INVALID_DATE", "Invalid event_end format. Use RFC3339 (2006-01-02T15:04:05Z07:00) or YYYY-MM-DD", nil)
				return
			}
		}
		eventEnd = &parsedDate
	}

	// Get the image file (optional)
	var imageURL string
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Validate and save the image
		imageResult, err := utils.SaveImage(file, header, "events")
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "INVALID_IMAGE", err.Error(), nil)
			return
		}
		imageURL = imageResult.URL
	}

	// Get author ID from token
	claims, _ := r.Context().Value(middleware.UserContextKey).(*utils.Claims)

	// Check if slug already exists
	db := config.GetDB()
	var existingEvent models.Event
	if err := db.Where("slug = ?", slug).First(&existingEvent).Error; err == nil {
		// Slug already exists
		if imageURL != "" {
			utils.DeleteImage(imageURL) // Clean up uploaded image
		}
		utils.RespondValidationError(w, map[string]string{
			"slug": "Slug already exists. Please use a different slug.",
		})
		return
	}

	// Create event object
	event := models.Event{
		Title:      title,
		Content:    content,
		Excerpt:    utils.MakeExcerpt(content, 160),
		Slug:       slug,
		Published:  published,
		ImageURL:   imageURL,
		AuthorID:   claims.UserID,
		EventStart: eventStart,
		EventEnd:   eventEnd,
	}

	// Save to database
	if err := db.Create(&event).Error; err != nil {
		// If database save fails, delete the uploaded image
		if imageURL != "" {
			utils.DeleteImage(imageURL)
		}
		utils.RespondInternalError(w)
		return
	}

	// Invalidate all event list caches
	ctx := r.Context()
	_ = utils.CacheDeletePattern(ctx, "event:list:*")

	utils.RespondSuccess(w, http.StatusCreated, map[string]interface{}{
		"id": event.ID,
	}, nil)
}

// UpdateEvent updates an event
// Supports partial updates - only send fields you want to change
// Always use multipart/form-data for all updates (with or without image)
func UpdateEvent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	db := config.GetDB()
	var event models.Event
	if err := db.Preload("Author").First(&event, "id = ?", id).Error; err != nil {
		utils.RespondNotFound(w, "Event")
		return
	}

	// Parse multipart form data (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondBadRequest(w, "Failed to parse form data")
		return
	}

	// Track what was updated for response
	updated := make(map[string]bool)

	// Update text fields if provided (only fields that are sent)
	if title := r.FormValue("title"); title != "" {
		event.Title = title
		updated["title"] = true
	}
	if content := r.FormValue("content"); content != "" {
		// Sanitize HTML content to prevent XSS
		event.Content = utils.SanitizeHTML(content)
		// Regenerate excerpt from sanitized content
		event.Excerpt = utils.MakeExcerpt(event.Content, 160)
		updated["content"] = true
		updated["excerpt"] = true
	}
	if slug := r.FormValue("slug"); slug != "" {
		event.Slug = slug
		updated["slug"] = true
	}
	if published := r.FormValue("published"); published != "" {
		event.Published = published == "true"
		updated["published"] = true
	}
	if eventStartStr := r.FormValue("event_start"); eventStartStr != "" {
		parsedDate, err := time.Parse(time.RFC3339, eventStartStr)
		if err != nil {
			// Try alternative format
			parsedDate, err = time.Parse("2006-01-02", eventStartStr)
			if err != nil {
				utils.RespondError(w, http.StatusBadRequest, "INVALID_DATE", "Invalid event_start format. Use RFC3339 or YYYY-MM-DD", nil)
				return
			}
		}
		event.EventStart = &parsedDate
		updated["event_start"] = true
	}
	if eventEndStr := r.FormValue("event_end"); eventEndStr != "" {
		parsedDate, err := time.Parse(time.RFC3339, eventEndStr)
		if err != nil {
			// Try alternative format
			parsedDate, err = time.Parse("2006-01-02", eventEndStr)
			if err != nil {
				utils.RespondError(w, http.StatusBadRequest, "INVALID_DATE", "Invalid event_end format. Use RFC3339 or YYYY-MM-DD", nil)
				return
			}
		}
		event.EventEnd = &parsedDate
		updated["event_end"] = true
	}

	// Handle image operations
	deleteImage := r.FormValue("delete_image") == "true"
	file, header, err := r.FormFile("image")
	hasNewImage := err == nil

	if deleteImage {
		// Delete image without uploading new one
		oldImageURL := event.ImageURL
		event.ImageURL = ""
		updated["image_deleted"] = true

		// Delete old image file
		if oldImageURL != "" {
			utils.DeleteImage(oldImageURL)
		}
	} else if hasNewImage {
		// New image uploaded
		defer file.Close()
		
		// Validate and save the new image
		imageResult, err := utils.SaveImage(file, header, "events")
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "INVALID_IMAGE", err.Error(), nil)
			return
		}

		// Store old image URL for deletion
		oldImageURL := event.ImageURL
		event.ImageURL = imageResult.URL
		updated["image_updated"] = true

		// Delete old image after new one is saved
		if oldImageURL != "" {
			utils.DeleteImage(oldImageURL)
		}
	}

	// Save to database if any field was updated
	oldSlug := event.Slug
	if len(updated) > 0 {
		if err := db.Save(&event).Error; err != nil {
			utils.RespondInternalError(w)
			return
		}

		// Invalidate caches
		ctx := r.Context()
		_ = utils.CacheDeletePattern(ctx, "event:list:*")
		_ = utils.CacheDelete(ctx, utils.BuildCacheKey("event", "id", id))
		_ = utils.CacheDelete(ctx, utils.BuildCacheKey("event", "slug", oldSlug))
		if updated["slug"] && event.Slug != oldSlug {
			_ = utils.CacheDelete(ctx, utils.BuildCacheKey("event", "slug", event.Slug))
		}
	}

	// Prepare response with updated event data
	response := EventDetailResponse{
		ID:         event.ID,
		Title:      event.Title,
		Content:    event.Content,
		Slug:       event.Slug,
		Published:  event.Published,
		ImageURL:   event.ImageURL,
		AuthorID:   event.AuthorID,
		Author: SimplifiedAuthor{
			ID:   event.Author.ID,
			Name: event.Author.Name,
		},
		EventStart: event.EventStart,
		EventEnd:   event.EventEnd,
		CreatedAt:  event.CreatedAt,
		UpdatedAt:  event.UpdatedAt,
	}

	// Add BASE_URL to response
	baseURL := config.GetEnv("BASE_URL", "")
	response.ImageURL = utils.PrependBaseURL(response.ImageURL, baseURL)

	utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Event updated successfully",
		"updated_fields": updated,
		"data":    response,
	}, nil)
}

// DeleteEvent soft deletes an event and removes its image
func DeleteEvent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	db := config.GetDB()
	
	// Get the event first to retrieve image URL
	var event models.Event
	if err := db.First(&event, "id = ?", id).Error; err != nil {
		utils.RespondNotFound(w, "Event")
		return
	}

	// Delete from database
	if err := db.Delete(&event, "id = ?", id).Error; err != nil {
		utils.RespondInternalError(w)
		return
	}

	// Delete the image file
	utils.DeleteImage(event.ImageURL)

	// Invalidate caches
	ctx := r.Context()
	_ = utils.CacheDeletePattern(ctx, "event:list:*")
	_ = utils.CacheDelete(ctx, utils.BuildCacheKey("event", "id", id))
	_ = utils.CacheDelete(ctx, utils.BuildCacheKey("event", "slug", event.Slug))

	utils.RespondSuccess(w, http.StatusOK, map[string]string{
		"message": "Event deleted successfully",
	}, nil)
}
