package handlers

import (
	"net/http"
	"strconv"
	"time"

	"sentul-golf-be/config"
	"sentul-golf-be/models"
	"sentul-golf-be/utils"
)

// PostResponse represents unified news and events for public access
type PostResponse struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"` // "NEWS" or "EVENT"
	Title      string           `json:"title"`
	Excerpt    string           `json:"excerpt"` // Plain text excerpt instead of full content
	Slug       string           `json:"slug"`
	Published  bool             `json:"published"`
	ImageURL   string           `json:"image_url"`
	AuthorID   string           `json:"author_id"`
	Author     SimplifiedAuthor `json:"author"`
	EventStart *time.Time       `json:"event_start,omitempty"` // Only for events
	EventEnd   *time.Time       `json:"event_end,omitempty"`   // Only for events
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// GetPosts retrieves news and/or events based on optional type query parameter
// If no type specified, returns both news and events sorted by newest first
// If type=news, returns only news
// If type=event, returns only events
func GetPosts(w http.ResponseWriter, r *http.Request) {
	db := config.GetDB()
	
	// Get type parameter (optional)
	typeParam := r.URL.Query().Get("type")
	
	// Validate type if provided
	if typeParam != "" && typeParam != "news" && typeParam != "event" {
		utils.RespondError(w, http.StatusBadRequest, "INVALID_TYPE", "Type must be 'news' or 'event'", nil)
		return
	}
	
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
	
	var posts []PostResponse
	var total int64
	
	if typeParam == "news" {
		// Get only news (published)
		var news []models.News
		newsQuery := db.Preload("Author").Where("published = ?", true)
		
		// Count total
		db.Model(&models.News{}).Where("published = ?", true).Count(&total)
		
		// Get paginated results
		if err := newsQuery.Order("created_at DESC").Limit(limit).Offset(offset).Find(&news).Error; err != nil {
			utils.RespondInternalError(w)
			return
		}
		
		// Transform to posts
		posts = make([]PostResponse, len(news))
		for i, n := range news {
			posts[i] = PostResponse{
				ID:        n.ID,
				Type:      "NEWS",
				Title:     n.Title,
				Excerpt:   n.Excerpt,
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
	} else if typeParam == "event" {
		// Get only events (published)
		var events []models.Event
		eventsQuery := db.Preload("Author").Where("published = ?", true)
		
		// Count total
		db.Model(&models.Event{}).Where("published = ?", true).Count(&total)
		
		// Get paginated results
		if err := eventsQuery.Order("created_at DESC").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
			utils.RespondInternalError(w)
			return
		}
		
		// Transform to posts
		posts = make([]PostResponse, len(events))
		for i, e := range events {
			posts[i] = PostResponse{
				ID:        e.ID,
				Type:      "EVENT",
				Title:     e.Title,
				Excerpt:   e.Excerpt,
				Slug:      e.Slug,
				Published: e.Published,
				ImageURL:  e.ImageURL,
				AuthorID:  e.AuthorID,
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
	} else {
		// Get both news and events (published), sorted by newest
		var news []models.News
		var events []models.Event
		
		// Get all news
		if err := db.Preload("Author").Where("published = ?", true).Order("created_at DESC").Find(&news).Error; err != nil {
			utils.RespondInternalError(w)
			return
		}
		
		// Get all events
		if err := db.Preload("Author").Where("published = ?", true).Order("created_at DESC").Find(&events).Error; err != nil {
			utils.RespondInternalError(w)
			return
		}
		
		// Combine into posts
		var allPosts []PostResponse
		
		// Add news
		for _, n := range news {
			allPosts = append(allPosts, PostResponse{
				ID:        n.ID,
				Type:      "NEWS",
				Title:     n.Title,
				Excerpt:   n.Excerpt,
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
			})
		}
		
		// Add events
		for _, e := range events {
			allPosts = append(allPosts, PostResponse{
				ID:         e.ID,
				Type:       "EVENT",
				Title:      e.Title,
				Excerpt:    e.Excerpt,
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
			})
		}
		
		// Sort by created_at DESC (newest first)
		for i := 0; i < len(allPosts)-1; i++ {
			for j := i + 1; j < len(allPosts); j++ {
				if allPosts[i].CreatedAt.Before(allPosts[j].CreatedAt) {
					allPosts[i], allPosts[j] = allPosts[j], allPosts[i]
				}
			}
		}
		
		// Calculate total
		total = int64(len(allPosts))
		
		// Apply pagination
		start := offset
		end := offset + limit
		
		if start > len(allPosts) {
			start = len(allPosts)
		}
		if end > len(allPosts) {
			end = len(allPosts)
		}
		
		posts = allPosts[start:end]
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
	
	utils.RespondSuccess(w, http.StatusOK, posts, meta)
}
