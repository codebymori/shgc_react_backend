package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"sentul-golf-be/config"
	"sentul-golf-be/models"
	"sentul-golf-be/utils"
)

type RegisterRequest struct {
	Name     string      `json:"name"`
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Role     models.Role `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginData struct {
	UserID    string `json:"user_id"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp
}

// Register creates a new user
func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request payload")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		fields := make(map[string]string)
		if req.Name == "" {
			fields["name"] = "Name is required"
		}
		if req.Email == "" {
			fields["email"] = "Email is required"
		}
		if req.Password == "" {
			fields["password"] = "Password is required"
		}
		utils.RespondValidationError(w, fields)
		return
	}

	// Default role is user
	if req.Role == "" {
		req.Role = models.RoleUser
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.RespondInternalError(w)
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     req.Role,
	}

	db := config.GetDB()
	if err := db.Create(&user).Error; err != nil {
		// Check if email already exists
		if err.Error() == "duplicate key value violates unique constraint \"idx_users_email\"" || 
		   err.Error() == "UNIQUE constraint failed: users.email" {
			utils.RespondError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already registered", nil)
			return
		}
		utils.RespondInternalError(w)
		return
	}

	// Return user data without password
	userData := map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	}

	utils.RespondSuccess(w, http.StatusCreated, userData, nil)
}

// Login authenticates a user and returns a JWT token
func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondBadRequest(w, "Invalid request payload")
		return
	}

	if req.Email == "" || req.Password == "" {
		fields := make(map[string]string)
		if req.Email == "" {
			fields["email"] = "Email is required"
		}
		if req.Password == "" {
			fields["password"] = "Password is required"
		}
		utils.RespondValidationError(w, fields)
		return
	}

	db := config.GetDB()
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.RespondUnauthorized(w, "Invalid credentials")
		return
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.Password) {
		utils.RespondUnauthorized(w, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, string(user.Role), os.Getenv("JWT_SECRET"))
	if err != nil {
		utils.RespondInternalError(w)
		return
	}

	// Calculate expiry time (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	// Return user_id, token, and expiry
	loginData := LoginData{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	utils.RespondSuccess(w, http.StatusOK, loginData, nil)
}
