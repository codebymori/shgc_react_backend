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

// SimplifiedAuthor for response
type SimplifiedAuthor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewsResponse with simplified author (for list)
type NewsResponse struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Slug      string           `json:"slug"`
	Published bool             `json:"published"`
	ImageURL  string           `json:"image_url"`
	AuthorID  string           `json:"author_id"`
	Author    SimplifiedAuthor `json:"author"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// NewsDetailResponse with full content (for detail by ID)
type NewsDetailResponse struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Content   string           `json:"content"`
	Slug      string           `json:"slug"`
	Published bool             `json:"published"`
	ImageURL  string           `json:"image_url"`
	AuthorID  string           `json:"author_id"`
	Author    SimplifiedAuthor `json:"author"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// GetNews retrieves all news articles with pagination
func GetNews(w http.ResponseWriter, r *http.Request) {
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
	cacheKey := utils.BuildCacheKey("news", "list", "page", page, "limit", limit, "published", publishedOnly)
	type CachedNewsResponse struct {
		NewsResponse []NewsResponse `json:"news"`
		Meta         *utils.Meta    `json:"meta"`
	}
	var cached CachedNewsResponse
	if err := utils.CacheGet(ctx, cacheKey, &cached); err == nil {
		utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
			"news": cached.NewsResponse,
		}, cached.Meta)
		return
	}
	
	// Cache miss - get from database
	db := config.GetDB()
	var news []models.News
	query := db.Preload("Author")
	if publishedOnly {
		query = query.Where("published = ?", true)
	}
	
	// Count total items
	var total int64
	countQuery := db.Model(&models.News{})
	if !ok || claims.Role != string(models.RoleAdmin) {
		countQuery = countQuery.Where("published = ?", true)
	}
	countQuery.Count(&total)
	
	// Get paginated results
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&news).Error; err != nil {
		utils.RespondInternalError(w)
		return
	}
	
	// Transform to response format with simplified author
	newsResponse := make([]NewsResponse, len(news))
	for i, n := range news {
		newsResponse[i] = NewsResponse{
			ID:        n.ID,
			Title:     n.Title,
			Slug:      n.Slug,
			Published: n.Published,
			ImageURL:  n.ImageURL,
			AuthorID:  n.AuthorID,
			Author: SimplifiedAuthor{
				ID:   n.Author.ID,
				Name: n.Author.Name,
			},
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
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
	_ = utils.CacheSet(ctx, cacheKey, CachedNewsResponse{
		NewsResponse: newsResponse,
		Meta:         meta,
	}, utils.CacheTTLNewsList)

	utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"news": newsResponse,
	}, meta)
}

// GetNewsBySlug retrieves a single news article by slug
func GetNewsBySlug(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	slug := params["slug"]
	ctx := r.Context()
	cacheKey := utils.BuildCacheKey("news", "slug", slug)

	// Try cache first
	var response NewsDetailResponse
	if err := utils.CacheGet(ctx, cacheKey, &response); err == nil {
		utils.RespondSuccess(w, http.StatusOK, response, nil)
		return
	}

	// Cache miss - get from database
	db := config.GetDB()
	var news models.News
	if err := db.Preload("Author").Where("slug = ?", slug).First(&news).Error; err != nil {
		utils.RespondNotFound(w, "News")
		return
	}

	// Build response
	response = NewsDetailResponse{
		ID:        news.ID,
		Title:     news.Title,
		Content:   news.Content,
		Slug:      news.Slug,
		Published: news.Published,
		ImageURL:  news.ImageURL,
		AuthorID:  news.AuthorID,
		Author: SimplifiedAuthor{
			ID:   news.Author.ID,
			Name: news.Author.Name,
		},
		CreatedAt: news.CreatedAt,
		UpdatedAt: news.UpdatedAt,
	}

	// Cache the response
	_ = utils.CacheSet(ctx, cacheKey, response, utils.CacheTTLNewsDetail)

	utils.RespondSuccess(w, http.StatusOK, response, nil)
}

// GetNewsByID retrieves a single news article by ID
func GetNewsByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	ctx := r.Context()
	cacheKey := utils.BuildCacheKey("news", "id", id)

	// Try cache first
	var response NewsDetailResponse
	if err := utils.CacheGet(ctx, cacheKey, &response); err == nil {
		utils.RespondSuccess(w, http.StatusOK, response, nil)
		return
	}

	// Cache miss - get from database
	db := config.GetDB()
	var news models.News
	if err := db.Preload("Author").Where("id = ?", id).First(&news).Error; err != nil {
		utils.RespondNotFound(w, "News")
		return
	}

	// Build response
	response = NewsDetailResponse{
		ID:        news.ID,
		Title:     news.Title,
		Content:   news.Content,
		Slug:      news.Slug,
		Published: news.Published,
		ImageURL:  news.ImageURL,
		AuthorID:  news.AuthorID,
		Author: SimplifiedAuthor{
			ID:   news.Author.ID,
			Name: news.Author.Name,
		},
		CreatedAt: news.CreatedAt,
		UpdatedAt: news.UpdatedAt,
	}

	// Cache the response
	_ = utils.CacheSet(ctx, cacheKey, response, utils.CacheTTLNewsDetail)

	utils.RespondSuccess(w, http.StatusOK, response, nil)
}

// CreateNews creates a new news article with image upload
func CreateNews(w http.ResponseWriter, r *http.Request) {
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

	// Get the image file (optional)
	var imageURL string
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Validate and save the image
		imageResult, err := utils.SaveImage(file, header, "news")
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
	var existingNews models.News
	if err := db.Where("slug = ?", slug).First(&existingNews).Error; err == nil {
		// Slug already exists
		if imageURL != "" {
			utils.DeleteImage(imageURL) // Clean up uploaded image
		}
		utils.RespondValidationError(w, map[string]string{
			"slug": "Slug already exists. Please use a different slug.",
		})
		return
	}

	// Create news object
	news := models.News{
		Title:     title,
		Content:   content,
		Excerpt:   utils.MakeExcerpt(content, 160),
		Slug:      slug,
		Published: published,
		ImageURL:  imageURL,
		AuthorID:  claims.UserID,
	}

	// Save to database
	if err := db.Create(&news).Error; err != nil {
		// If database save fails, delete the uploaded image
		if imageURL != "" {
			utils.DeleteImage(imageURL)
		}
		utils.RespondInternalError(w)
		return
	}

	// Invalidate all news list caches
	ctx := r.Context()
	_ = utils.CacheDeletePattern(ctx, "news:list:*")

	utils.RespondSuccess(w, http.StatusCreated, map[string]interface{}{
		"id": news.ID,
	}, nil)
}

// UpdateNews updates a news article
// Supports partial updates - only send fields you want to change
// Always use multipart/form-data for all updates (with or without image)
func UpdateNews(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	db := config.GetDB()
	var news models.News
	if err := db.Preload("Author").First(&news, "id = ?", id).Error; err != nil {
		utils.RespondNotFound(w, "News")
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
		news.Title = title
		updated["title"] = true
	}
	if content := r.FormValue("content"); content != "" {
		// Sanitize HTML content to prevent XSS
		news.Content = utils.SanitizeHTML(content)
		// Regenerate excerpt from sanitized content
		news.Excerpt = utils.MakeExcerpt(news.Content, 160)
		updated["content"] = true
		updated["excerpt"] = true
	}
	if slug := r.FormValue("slug"); slug != "" {
		news.Slug = slug
		updated["slug"] = true
	}
	if published := r.FormValue("published"); published != "" {
		news.Published = published == "true"
		updated["published"] = true
	}

	// Handle image operations
	deleteImage := r.FormValue("delete_image") == "true"
	file, header, err := r.FormFile("image")
	hasNewImage := err == nil

	if deleteImage {
		// Delete image without uploading new one
		oldImageURL := news.ImageURL
		news.ImageURL = ""
		updated["image_deleted"] = true

		// Delete old image file
		if oldImageURL != "" {
			utils.DeleteImage(oldImageURL)
		}
	} else if hasNewImage {
		// New image uploaded
		defer file.Close()
		
		// Validate and save the new image
		imageResult, err := utils.SaveImage(file, header, "news")
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "INVALID_IMAGE", err.Error(), nil)
			return
		}

		// Store old image URL for deletion
		oldImageURL := news.ImageURL
		news.ImageURL = imageResult.URL
		updated["image_updated"] = true

		// Delete old image after new one is saved
		if oldImageURL != "" {
			utils.DeleteImage(oldImageURL)
		}
	}

	// Save to database if any field was updated
	oldSlug := news.Slug
	if len(updated) > 0 {
		if err := db.Save(&news).Error; err != nil {
			utils.RespondInternalError(w)
			return
		}

		// Invalidate caches
		ctx := r.Context()
		_ = utils.CacheDeletePattern(ctx, "news:list:*")
		_ = utils.CacheDelete(ctx, utils.BuildCacheKey("news", "id", id))
		_ = utils.CacheDelete(ctx, utils.BuildCacheKey("news", "slug", oldSlug))
		if updated["slug"] && news.Slug != oldSlug {
			_ = utils.CacheDelete(ctx, utils.BuildCacheKey("news", "slug", news.Slug))
		}
	}

	// Prepare response with updated news data
	response := NewsDetailResponse{
		ID:        news.ID,
		Title:     news.Title,
		Content:   news.Content,
		Slug:      news.Slug,
		Published: news.Published,
		ImageURL:  news.ImageURL,
		AuthorID:  news.AuthorID,
		Author: SimplifiedAuthor{
			ID:   news.Author.ID,
			Name: news.Author.Name,
		},
		CreatedAt: news.CreatedAt,
		UpdatedAt: news.UpdatedAt,
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "News updated successfully",
		"updated_fields": updated,
		"data":    response,
	}, nil)
}

// DeleteNews soft deletes a news article and removes its image
func DeleteNews(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	db := config.GetDB()
	
	// Get the news first to retrieve image URL
	var news models.News
	if err := db.First(&news, "id = ?", id).Error; err != nil {
		utils.RespondNotFound(w, "News")
		return
	}

	// Delete from database
	if err := db.Delete(&news, "id = ?", id).Error; err != nil {
		utils.RespondInternalError(w)
		return
	}

	// Delete the image file
	utils.DeleteImage(news.ImageURL)

	// Invalidate caches
	ctx := r.Context()
	_ = utils.CacheDeletePattern(ctx, "news:list:*")
	_ = utils.CacheDelete(ctx, utils.BuildCacheKey("news", "id", id))
	_ = utils.CacheDelete(ctx, utils.BuildCacheKey("news", "slug", news.Slug))

	utils.RespondSuccess(w, http.StatusOK, map[string]string{
		"message": "News deleted successfully",
	}, nil)
}
