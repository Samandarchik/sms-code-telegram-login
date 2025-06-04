package routes

import (
	"amur/handlers"
	"amur/pkg/jwt_auth" // jwt_auth paketini import qilamiz
	"net/http"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// SetupRoutes funksiyasi barcha API marshrutlarini sozlaydi
func SetupRoutes(foodHandler *handlers.FoodHandler, userHandler *handlers.UserHandler, basketOrderHandler *handlers.BasketOrderHandler, orderHandler *handlers.OrderHandler) http.Handler {
	r := mux.NewRouter()

	// API prefix
	api := r.PathPrefix("/api").Subrouter()

	// --- Public (autentifikatsiya talab qilinmaydigan) marshrutlar ---
	// Bu endpointlarga har kim token bo'lmasa ham murojaat qila oladi.
	api.HandleFunc("/register", userHandler.RegisterUser).Methods("POST") // Ro'yxatdan o'tish
	api.HandleFunc("/login", userHandler.Login).Methods("POST")           // Tizimga kirish

	// Health check (server holatini tekshirish uchun)
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK", "message": "Server ishga tushdi."}`))
	}).Methods("GET")

	// Statik fayllarni (yuklangan rasmlarni) taqdim etish uchun marshrut
	// Bunga ham autentifikatsiya kerak emas, chunki bu ommaviy kontent.
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// --- Autentifikatsiya talab qilinadigan marshrutlar uchun Subrouter ---
	// Bu subrouterga 'AuthMiddleware' qo'llaniladi.
	// Barcha marshrutlar ushbu subrouter orqali o'tadi va token tekshiriladi.
	authRequired := api.PathPrefix("/").Subrouter()
	authRequired.Use(jwt_auth.AuthMiddleware) // Bu yerda AuthMiddleware qo'llaniladi

	// --- AuthMiddleware orqali himoyalangan marshrutlar ---

	// Food routes
	// Bular endi himoyalangan. Faqat to'g'ri JWT tokeni bilan kirish mumkin.
	authRequired.HandleFunc("/foods", foodHandler.GetAllFoods).Methods("GET")
	authRequired.HandleFunc("/foods", foodHandler.CreateFood).Methods("POST")
	authRequired.HandleFunc("/foods/{id:[0-9]+}", foodHandler.GetFoodByID).Methods("GET")
	authRequired.HandleFunc("/foods/{id:[0-9]+}", foodHandler.UpdateFood).Methods("PUT")
	authRequired.HandleFunc("/foods/{id:[0-9]+}", foodHandler.DeleteFood).Methods("DELETE")
	authRequired.HandleFunc("/foods/category/{category}", foodHandler.GetFoodsByCategory).Methods("GET")
	authRequired.HandleFunc("/foods/stats", foodHandler.GetFoodStats).Methods("GET")

	// User routes
	// Foydalanuvchilar ro'yxati va statistikasi ham himoyalangan.
	// Keyinchalik, RoleMiddleware yordamida faqat administratorlarga ruxsat berishingiz mumkin.
	authRequired.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	authRequired.HandleFunc("/users/stats", userHandler.GetUserStats).Methods("GET")

	// Basket Order Routes
	// Endi `telegramID` URLdan emas, JWT tokendan olinadi.
	authRequired.HandleFunc("/basket-order", basketOrderHandler.AddToBasket).Methods("POST")
	authRequired.HandleFunc("/basket-order", basketOrderHandler.GetBasketOrders).Methods("GET")
	authRequired.HandleFunc("/basket-order/{foodID:[0-9]+}", basketOrderHandler.RemoveFromBasket).Methods("DELETE")
	authRequired.HandleFunc("/basket-order", basketOrderHandler.ClearBasket).Methods("DELETE")

	// Order Routes
	// Buyurtmalar yaratish va ko'rish uchun ham `telegramID` tokendan olinadi.
	authRequired.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	authRequired.HandleFunc("/orders", orderHandler.GetUserOrders).Methods("GET")
	authRequired.HandleFunc("/orders/{orderID:[0-9]+}", orderHandler.GetOrderDetails).Methods("GET")
	authRequired.HandleFunc("/orders/{orderID:[0-9]+}/status", orderHandler.UpdateOrderStatus).Methods("PUT") // Admin roli bilan himoyalash kerak
	authRequired.HandleFunc("/orders/stats", orderHandler.GetOrderStats).Methods("GET")                       // Admin roli bilan himoyalash kerak

	// Admin-only routes
	// Bu marshrutlar AuthMiddleware orqali himoyalangan. Rol bo'yicha cheklovlarni qo'shishni unutmang.
	authRequired.HandleFunc("/admin/orders/{orderID:[0-9]+}", orderHandler.DeleteOrderAdmin).Methods("DELETE")

	// --- CORS middleware ---
	// Bu barcha so'rovlar uchun CORS sozlamalarini o'rnatadi.
	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins([]string{"*"}), // Diqqat: Productionda faqat kerakli originlarni ko'rsating!
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
	)(r)

	return corsHandler
}
