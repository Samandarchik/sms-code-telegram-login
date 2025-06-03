package routes

import (
	"amur/handlers"
	"net/http"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// SetupRoutes funksiyasi barcha API marshrutlarini sozlaydi
func SetupRoutes(foodHandler *handlers.FoodHandler, userHandler *handlers.UserHandler, basketOrderHandler *handlers.BasketOrderHandler, orderHandler *handlers.OrderHandler) http.Handler {
	r := mux.NewRouter()

	// API prefix
	api := r.PathPrefix("/api").Subrouter()

	// --- Yangi User auth routes ---
	api.HandleFunc("/register", userHandler.RegisterUser).Methods("POST") // Ro'yxatdan o'tish
	api.HandleFunc("/login", userHandler.Login).Methods("POST")           // Tizimga kirish

	// Food routes (unchanged)
	api.HandleFunc("/foods", foodHandler.GetAllFoods).Methods("GET")
	api.HandleFunc("/foods", foodHandler.CreateFood).Methods("POST")
	api.HandleFunc("/foods/{id:[0-9]+}", foodHandler.GetFoodByID).Methods("GET")
	api.HandleFunc("/foods/{id:[0-9]+}", foodHandler.UpdateFood).Methods("PUT")
	api.HandleFunc("/foods/{id:[0-9]+}", foodHandler.DeleteFood).Methods("DELETE")
	api.HandleFunc("/foods/category/{category}", foodHandler.GetFoodsByCategory).Methods("GET")
	api.HandleFunc("/foods/stats", foodHandler.GetFoodStats).Methods("GET")

	// User routes (unchanged, lekin "register" va "login"ni qo'shdik)
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")        // Bunga AuthMiddleware va RoleMiddleware kerak bo'ladi!
	api.HandleFunc("/users/stats", userHandler.GetUserStats).Methods("GET") // Bunga AuthMiddleware va RoleMiddleware kerak bo'ladi!

	// Basket Order Routes (unchanged)
	// Eslatma: Hozirgi holatda telegramID URLda, lekin middlewaredan olish yaxshiroq
	// api.HandleFunc("/{telegramID:[0-9]+}/basket-order", basketOrderHandler.AddToBasket).Methods("POST")
	// api.HandleFunc("/{telegramID:[0-9]+}/basket-order", basketOrderHandler.GetBasketOrders).Methods("GET")
	// api.HandleFunc("/{telegramID:[0-9]+}/basket-order/{foodID:[0-9]+}", basketOrderHandler.RemoveFromBasket).Methods("DELETE")
	// api.HandleFunc("/{telegramID:[0-9]+}/basket-order", basketOrderHandler.ClearBasket).Methods("DELETE")
	// Endi middlewaredan telegramID olish uchun marshrutlarni yangilashimiz kerak:
	api.HandleFunc("/basket-order", basketOrderHandler.AddToBasket).Methods("POST")
	api.HandleFunc("/basket-order", basketOrderHandler.GetBasketOrders).Methods("GET")
	api.HandleFunc("/basket-order/{foodID:[0-9]+}", basketOrderHandler.RemoveFromBasket).Methods("DELETE")
	api.HandleFunc("/basket-order", basketOrderHandler.ClearBasket).Methods("DELETE")

	// Order Routes (unchanged, faqat order_handler.go o'zgaradi)
	// Eslatma: Hozirgi holatda telegramID URLda, lekin middlewaredan olish yaxshiroq
	// api.HandleFunc("/{telegramID:[0-9]+}/orders", orderHandler.CreateOrder).Methods("POST")
	// api.HandleFunc("/{telegramID:[0-9]+}/orders", orderHandler.GetUserOrders).Methods("GET")
	// Endi middlewaredan telegramID olish uchun marshrutlarni yangilashimiz kerak:
	api.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	api.HandleFunc("/orders", orderHandler.GetUserOrders).Methods("GET")

	api.HandleFunc("/orders/{orderID:[0-9]+}", orderHandler.GetOrderDetails).Methods("GET")
	api.HandleFunc("/orders/{orderID:[0-9]+}/status", orderHandler.UpdateOrderStatus).Methods("PUT") // Admin roli bilan himoyalash kerak
	api.HandleFunc("/orders/stats", orderHandler.GetOrderStats).Methods("GET")                       // Admin roli bilan himoyalash kerak

	// Admin-only routes (unchanged, faqat order_handler.go o'zgaradi)
	api.HandleFunc("/admin/orders/{orderID:[0-9]+}", orderHandler.DeleteOrderAdmin).Methods("DELETE") // Admin roli bilan himoyalash kerak

	// Health check (unchanged)
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK", "message": "Server is running"}`))
	}).Methods("GET")

	// Statik fayllarni (yuklangan rasmlarni) taqdim etish uchun marshrut (unchanged)
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// CORS middleware (unchanged)
	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}),
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
	)(r)

	return corsHandler
}
