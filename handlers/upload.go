package handlers

import (
	"net/http"
	"strings"

	"sentul-golf-be/config"
	"sentul-golf-be/utils"
)

// UploadContentImage handles inline image uploads for rich text editor content.
// Accepts: multipart/form-data with field "image"
// Returns: { "url": "http://..." } — URL siap dipakai di <img src="...">
func UploadContentImage(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 5MB)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		utils.RespondBadRequest(w, "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		utils.RespondBadRequest(w, "Image file is required")
		return
	}
	defer file.Close()

	// Validate & save to ./uploads/content/
	imageResult, err := utils.SaveImage(file, header, "content")
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "INVALID_IMAGE", err.Error(), nil)
		return
	}

	// Build full URL
	baseURL := config.GetEnv("BASE_URL", "")
	fullURL := utils.PrependBaseURL(imageResult.URL, baseURL)

	utils.RespondSuccess(w, http.StatusOK, map[string]string{
		"url": fullURL,
	}, nil)
}

// DeleteSingleContentImage deletes one specific content image by its URL.
// Called by the frontend in real-time when a user removes an image from the rich text editor.
// Security: only /uploads/content/ files can be deleted via this endpoint.
func DeleteSingleContentImage(w http.ResponseWriter, r *http.Request) {
	imageURL := r.URL.Query().Get("url")
	if imageURL == "" {
		utils.RespondBadRequest(w, "url query parameter is required")
		return
	}

	// Security guard: only allow deleting from /uploads/content/
	// This prevents the endpoint from being used to delete thumbnails or other files
	if !strings.Contains(imageURL, "/uploads/content/") {
		utils.RespondError(w, http.StatusForbidden, "FORBIDDEN",
			"Only /uploads/content/ images can be deleted via this endpoint", nil)
		return
	}

	// Extract local path: "http://localhost:8000/uploads/content/abc.jpg" → "/uploads/content/abc.jpg"
	idx := strings.Index(imageURL, "/uploads/content/")
	localPath := imageURL[idx:]

	// Silently ignore if file doesn't exist (user may have refreshed/re-edited)
	_ = utils.DeleteImage(localPath)

	utils.RespondSuccess(w, http.StatusOK, map[string]string{
		"message": "Image deleted",
	}, nil)
}
