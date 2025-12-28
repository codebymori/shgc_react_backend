package handlers

import (
	"encoding/json"
	"net/http"

	"sentul-golf-be/config"
	"sentul-golf-be/middleware"
	"sentul-golf-be/models"
	"sentul-golf-be/utils"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// GetCurrentUser retrieves current authenticated user info
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user info from context (set by AuthMiddleware)
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.Claims)
	if !ok {
		utils.RespondUnauthorized(w, "Unauthorized")
		return
	}

	db := config.GetDB()
	var user models.User
	if err := db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		utils.RespondNotFound(w, "User")
		return
	}

	// Return only necessary fields
	userInfo := map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	}

	utils.RespondSuccess(w, http.StatusOK, userInfo, nil)
}

// GetUsers retrieves all users (admin only)
func GetUsers(w http.ResponseWriter, r *http.Request) {
	db := config.GetDB()
	var users []models.User

	if err := db.Find(&users).Error; err != nil {
		utils.RespondInternalError(w)
		return
	}

	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]interface{}{
"users": users,
}, nil)
}

// GetUser retrieves a single user by ID
func GetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	db := config.GetDB()
	var user models.User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		utils.RespondNotFound(w, "User")
		return
	}

	user.Password = ""
	utils.RespondSuccess(w, http.StatusOK, user, nil)
}

// UpdateUser updates a user (admin can update any, user can update self)
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	// Get current user from context
	claims, _ := r.Context().Value("user").(*utils.Claims)

	// Check if user can update (admin or self)
	if claims.Role != string(models.RoleAdmin) && claims.UserID != id {
		utils.RespondForbidden(w, "You don't have permission to update this user")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.RespondBadRequest(w, "Invalid request payload")
		return
	}

	// Don't allow updating role unless admin
if claims.Role != string(models.RoleAdmin) {
delete(updates, "role")
}

// Hash password if it's being updated
	if password, ok := updates["password"].(string); ok {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			utils.RespondInternalError(w)
			return
		}
		updates["password"] = string(hashedPassword)
	}

	db := config.GetDB()
	var user models.User
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		utils.RespondNotFound(w, "User")
		return
	}

	if err := db.Model(&user).Updates(updates).Error; err != nil {
		utils.RespondInternalError(w)
		return
	}

	user.Password = ""
	utils.RespondSuccess(w, http.StatusOK, user, nil)
}

// DeleteUser soft deletes a user (admin only)
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	db := config.GetDB()
	if err := db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		utils.RespondInternalError(w)
		return
	}

	utils.RespondSuccess(w, http.StatusOK, map[string]string{
"message": "User deleted successfully",
}, nil)
}
