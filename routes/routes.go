package routes

import (
	"amur/handlers"
	"net/http"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func SetupRoutes(foodHandler *handlers.FoodHandler, userHandler *handlers.UserHandler, cartHandler *handlers.CartHandler) http.Handler {
	r := mux.NewRouter()

	// API prefix
	api := r.PathPrefix("/api").Subrouter()

	// Food routes
	api.HandleFunc("/foods", foodHandler.GetAllFoods).Methods("GET")
	api.HandleFunc("/foods", foodHandler.CreateFood).Methods("POST")
	api.HandleFunc("/foods/{id:[0-9]+}", foodHandler.GetFoodByID).Methods("GET")
	api.HandleFunc("/foods/{id:[0-9]+}", foodHandler.UpdateFood).Methods("PUT")
	api.HandleFunc("/foods/{id:[0-9]+}", foodHandler.DeleteFood).Methods("DELETE")
	api.HandleFunc("/foods/category/{category}", foodHandler.GetFoodsByCategory).Methods("GET")
	api.HandleFunc("/foods/stats", foodHandler.GetFoodStats).Methods("GET")

	// User routes
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	api.HandleFunc("/users/stats", userHandler.GetUserStats).Methods("GET")

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK", "message": "Server is running"}`))
	}).Methods("GET")

	// Statik fayllarni (yuklangan rasmlarni) taqdim etish uchun marshrut
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// CORS middleware
	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
	)(r)

	return corsHandler
}
