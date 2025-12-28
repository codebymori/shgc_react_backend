package routes

import (
	"net/http"
	"sentul-golf-be/handlers"
	"sentul-golf-be/middleware"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply CORS middleware globally
	router.Use(middleware.CORSMiddleware)
	
	// Handle all OPTIONS requests globally before route matching
	router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Static file serving for uploads
	router.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))),
	)

	// Public routes
	api := router.PathPrefix("/api").Subrouter()
	
	// Auth routes - only login is public
	api.HandleFunc("/auth/login", handlers.Login).Methods("POST")

	// Public posts endpoint - can filter by type (news or event)
	api.HandleFunc("/posts", handlers.GetPosts).Methods("GET")
	
	// Public single post by slug or ID
	api.HandleFunc("/news/{id:[0-9a-z]+}", handlers.GetNewsByID).Methods("GET")
	api.HandleFunc("/news/slug/{slug}", handlers.GetNewsBySlug).Methods("GET")
	api.HandleFunc("/events/{id:[0-9a-z]+}", handlers.GetEventByID).Methods("GET")
	api.HandleFunc("/events/slug/{slug}", handlers.GetEventBySlug).Methods("GET")

	// Public holes
	api.HandleFunc("/holes", handlers.GetHoles).Methods("GET")
	api.HandleFunc("/holes/{id}", handlers.GetHole).Methods("GET")

	// Protected routes - require authentication
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// Get current user info (authenticated users)
	protected.HandleFunc("/users/me", handlers.GetCurrentUser).Methods("GET")

	// Admin-only routes - user management
	adminUsers := protected.PathPrefix("/users").Subrouter()
	adminUsers.Use(middleware.RequireAdmin)
	adminUsers.HandleFunc("", handlers.Register).Methods("POST")
	adminUsers.HandleFunc("", handlers.GetUsers).Methods("GET")
	adminUsers.HandleFunc("/{id}", handlers.GetUser).Methods("GET")
	adminUsers.HandleFunc("/{id}", handlers.UpdateUser).Methods("PUT")
	adminUsers.HandleFunc("/{id}", handlers.DeleteUser).Methods("DELETE")

	// Admin-only routes - news management (including GET all news)
	adminNews := protected.PathPrefix("/news").Subrouter()
	adminNews.Use(middleware.RequireAdmin)
	adminNews.HandleFunc("", handlers.GetNews).Methods("GET")
	adminNews.HandleFunc("", handlers.CreateNews).Methods("POST")
	adminNews.HandleFunc("/{id}", handlers.UpdateNews).Methods("PUT")
	adminNews.HandleFunc("/{id}", handlers.DeleteNews).Methods("DELETE")

	// Admin-only routes - events management (including GET all events)
	adminEvents := protected.PathPrefix("/events").Subrouter()
	adminEvents.Use(middleware.RequireAdmin)
	adminEvents.HandleFunc("", handlers.GetEvents).Methods("GET")
	adminEvents.HandleFunc("", handlers.CreateEvent).Methods("POST")
	adminEvents.HandleFunc("/{id}", handlers.UpdateEvent).Methods("PUT")
	adminEvents.HandleFunc("/{id}", handlers.DeleteEvent).Methods("DELETE")

	// Admin-only routes - holes management
	adminHoles := protected.PathPrefix("/admin/holes").Subrouter()
	adminHoles.Use(middleware.RequireAdmin)
	adminHoles.HandleFunc("", handlers.CreateHole).Methods("POST")
	adminHoles.HandleFunc("/reorder", handlers.ReorderHoles).Methods("PUT")
	adminHoles.HandleFunc("/{id}", handlers.UpdateHole).Methods("PUT")
	adminHoles.HandleFunc("/{id}", handlers.DeleteHole).Methods("DELETE")

	return router
}

