package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sentul-golf-be/config"
	"sentul-golf-be/models"
	"sentul-golf-be/utils"

	"github.com/gorilla/mux"
)

// GetHoles retrieves all holes
func GetHoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cacheKey := "holes:list"
	
	// Try to get from cache first
	var holes []models.Hole
	if err := utils.CacheGet(ctx, cacheKey, &holes); err == nil {
		// Cache hit - add BASE_URL and return
		baseURL := config.GetEnv("BASE_URL", "")
		for i := range holes {
			holes[i].ImageURL = utils.PrependBaseURL(holes[i].ImageURL, baseURL)
		}
		
		utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
			"holes": holes,
		}, nil)
		return
	}
	
	// Cache miss - get from database
	db := config.GetDB()
	if err := db.Order("hole_index ASC").Find(&holes).Error; err != nil {
		utils.RespondInternalError(w)
		return
	}

	// Store in cache (without BASE_URL prepended)
	_ = utils.CacheSet(ctx, cacheKey, holes, utils.CacheTTLHolesList)

	// Add BASE_URL to all image URLs for response
	baseURL := config.GetEnv("BASE_URL", "")
	for i := range holes {
		holes[i].ImageURL = utils.PrependBaseURL(holes[i].ImageURL, baseURL)
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"holes": holes,
	}, nil)
}

// GetHole retrieves a single hole by ID
func GetHole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	ctx := r.Context()
	cacheKey := utils.BuildCacheKey("hole", id)

	// Try to get from cache first
	var hole models.Hole
	if err := utils.CacheGet(ctx, cacheKey, &hole); err == nil {
		// Cache hit - add BASE_URL and return
		baseURL := config.GetEnv("BASE_URL", "")
		hole.ImageURL = utils.PrependBaseURL(hole.ImageURL, baseURL)
		
		utils.RespondSuccess(w, http.StatusOK, hole, nil)
		return
	}

	// Cache miss - get from database
	db := config.GetDB()
	if err := db.First(&hole, "id = ?", id).Error; err != nil {
		utils.RespondNotFound(w, "Hole")
		return
	}

	// Store in cache (without BASE_URL prepended)
	_ = utils.CacheSet(ctx, cacheKey, hole, utils.CacheTTLHoleDetail)

	// Add BASE_URL to image URL for response
	baseURL := config.GetEnv("BASE_URL", "")
	hole.ImageURL = utils.PrependBaseURL(hole.ImageURL, baseURL)

	utils.RespondSuccess(w, http.StatusOK, hole, nil)
}

// CreateHole creates a new hole with image upload
func CreateHole(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with max memory of 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.RespondBadRequest(w, "Failed to parse form data")
		return
	}

	// Get form values
	name := r.FormValue("name")
	description := r.FormValue("description")
	parStr := r.FormValue("par")
	distanceStr := r.FormValue("distance")

	// Validate required fields
	fields := make(map[string]string)
	if name == "" {
		fields["name"] = "Name is required"
	}

	var par, distance int
	var err error

	if parStr == "" {
		fields["par"] = "Par is required"
	} else {
		par, err = strconv.Atoi(parStr)
		if err != nil || par <= 0 {
			fields["par"] = "Par must be a positive number"
		}
	}

	if distanceStr == "" {
		fields["distance"] = "Distance is required"
	} else {
		distance, err = strconv.Atoi(distanceStr)
		if err != nil || distance <= 0 {
			fields["distance"] = "Distance must be a positive number"
		}
	}

	if len(fields) > 0 {
		utils.RespondValidationError(w, fields)
		return
	}

	// Get the image file
	file, header, err := r.FormFile("image")
	if err != nil {
		utils.RespondBadRequest(w, "Image file is required")
		return
	}
	defer file.Close()

	// Validate and save the image
	imageResult, err := utils.SaveImage(file, header, "holes")
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "INVALID_IMAGE", err.Error(), nil)
		return
	}

	// Get max hole index
	var maxIndex int
	db := config.GetDB()
	var lastHole models.Hole
	if err := db.Order("hole_index DESC").First(&lastHole).Error; err == nil {
		maxIndex = lastHole.HoleIndex
	}

	// Create hole object
	hole := models.Hole{
		Name:        name,
		Description: description,
		Par:         par,
		Distance:    distance,
		HoleIndex:   maxIndex + 1,
		ImageURL:    imageResult.URL,
	}

	// Save to database
	if err := db.Create(&hole).Error; err != nil {
		// If database save fails, delete the uploaded image
		utils.DeleteImage(imageResult.URL)
		utils.RespondInternalError(w)
		return
	}

	// Invalidate holes list cache
	ctx := r.Context()
	_ = utils.CacheDelete(ctx, "holes:list")

	// Add BASE_URL to response
	baseURL := config.GetEnv("BASE_URL", "")
	hole.ImageURL = utils.PrependBaseURL(hole.ImageURL, baseURL)

	utils.RespondSuccess(w, http.StatusCreated, hole, nil)
}

// UpdateHole updates a hole
// Supports partial updates - only send fields you want to change
// Always use multipart/form-data for all updates (with or without image)
func UpdateHole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	db := config.GetDB()
	var hole models.Hole
	if err := db.First(&hole, "id = ?", id).Error; err != nil {
		utils.RespondNotFound(w, "Hole")
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
	if name := r.FormValue("name"); name != "" {
		hole.Name = name
		updated["name"] = true
	}
	if description := r.FormValue("description"); description != "" {
		hole.Description = description
		updated["description"] = true
	}

	// Update numeric fields if provided
	if parStr := r.FormValue("par"); parStr != "" {
		par, err := strconv.Atoi(parStr)
		if err != nil || par <= 0 {
			utils.RespondValidationError(w, map[string]string{
				"par": "Par must be a positive number",
			})
			return
		}
		hole.Par = par
		updated["par"] = true
	}

	if distanceStr := r.FormValue("distance"); distanceStr != "" {
		distance, err := strconv.Atoi(distanceStr)
		if err != nil || distance <= 0 {
			utils.RespondValidationError(w, map[string]string{
				"distance": "Distance must be a positive number",
			})
			return
		}
		hole.Distance = distance
		updated["distance"] = true
	}

	// Handle image operations
	deleteImage := r.FormValue("delete_image") == "true"
	file, header, err := r.FormFile("image")
	hasNewImage := err == nil

	if deleteImage {
		// Delete image without uploading new one
		oldImageURL := hole.ImageURL
		hole.ImageURL = ""
		updated["image_deleted"] = true

		// Delete old image file
		if oldImageURL != "" {
			utils.DeleteImage(oldImageURL)
		}
	} else if hasNewImage {
		// New image uploaded
		defer file.Close()
		
		// Validate and save the new image
		imageResult, err := utils.SaveImage(file, header, "holes")
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "INVALID_IMAGE", err.Error(), nil)
			return
		}

		// Store old image URL for deletion
		oldImageURL := hole.ImageURL
		hole.ImageURL = imageResult.URL
		updated["image_updated"] = true

		// Delete old image after new one is saved
		if oldImageURL != "" {
			utils.DeleteImage(oldImageURL)
		}
	}

	// Save to database if any field was updated
	if len(updated) > 0 {
		if err := db.Save(&hole).Error; err != nil {
			utils.RespondInternalError(w)
			return
		}

		// Invalidate caches
		ctx := r.Context()
		_ = utils.CacheDelete(ctx, "holes:list")
		_ = utils.CacheDelete(ctx, utils.BuildCacheKey("hole", id))
	}

	// Add BASE_URL to response
	baseURL := config.GetEnv("BASE_URL", "")
	hole.ImageURL = utils.PrependBaseURL(hole.ImageURL, baseURL)

	utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"message":        "Hole updated successfully",
		"updated_fields": updated,
		"data":           hole,
	}, nil)
}

// DeleteHole soft deletes a hole and removes its image
func DeleteHole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	db := config.GetDB()

	// Get the hole first to retrieve image URL
	var hole models.Hole
	if err := db.First(&hole, "id = ?", id).Error; err != nil {
		utils.RespondNotFound(w, "Hole")
		return
	}

	// Delete from database
	if err := db.Delete(&hole, "id = ?", id).Error; err != nil {
		utils.RespondInternalError(w)
		return
	}

	// Delete the image file
	utils.DeleteImage(hole.ImageURL)

	// Invalidate caches
	ctx := r.Context()
	_ = utils.CacheDelete(ctx, "holes:list")
	_ = utils.CacheDelete(ctx, utils.BuildCacheKey("hole", id))

	utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Hole deleted successfully",
	}, nil)
}

// ReorderHolesRequest represents the request body for reordering holes
type ReorderHolesRequest struct {
	HoleIDs []string `json:"hole_ids"`
}

// ReorderHoles updates the sequence index of holes based on the provided ID list
func ReorderHoles(w http.ResponseWriter, r *http.Request) {
	var req ReorderHolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request body")
		return
	}

	if len(req.HoleIDs) == 0 {
		utils.RespondBadRequest(w, "hole_ids cannot be empty")
		return
	}

	db := config.GetDB()
	
	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		utils.RespondInternalError(w)
		return
	}

	// Update each hole's index based on its position in the list
	for i, id := range req.HoleIDs {
		// Index starts from 1
		newIndex := i + 1
		
		if err := tx.Model(&models.Hole{}).Where("id = ?", id).Update("hole_index", newIndex).Error; err != nil {
			tx.Rollback()
			utils.RespondInternalError(w)
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		utils.RespondInternalError(w)
		return
	}

	// Invalidate holes list cache (order changed)
	ctx := r.Context()
	_ = utils.CacheDelete(ctx, "holes:list")

	utils.RespondSuccess(w, http.StatusOK, map[string]string{
		"message": "Holes reordered successfully",
	}, nil)
}
