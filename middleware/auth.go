package middleware

import (
	"context"
	"log" // Log uchun import qiling
	"net/http"
	"strings"

	"amur/pkg/jwt_auth" // jwt_auth paketi to'g'ri yo'lni ko'rsatishini tekshiring
)

// ContextKey o'zgaruvchisi context ichiga ma'lumot saqlash uchun
type ContextKey string

const (
	TelegramIDContextKey ContextKey = "telegram_id"
	RoleContextKey       ContextKey = "role"
)

// AuthMiddleware JWT tokenini tekshiradi va foydalanuvchi ma'lumotlarini contextga qo'shadi
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("AuthMiddleware: Avtorizatsiya tokeni topilmadi")
			http.Error(w, "Avtorizatsiya tokeni topilmadi", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		if tokenString == authHeader { // Agar "Bearer " topilmasa
			log.Println("AuthMiddleware: Noto'g'ri avtorizatsiya formati, 'Bearer ' prefiksi yo'q")
			http.Error(w, "Noto'g'ri avtorizatsiya formati", http.StatusUnauthorized)
			return
		}

		// ValidateToken endi to'g'ridan-to'g'ri jwt.MapClaims qaytarishi kerak.
		claims, err := jwt_auth.ValidateToken(tokenString)
		if err != nil {
			log.Printf("AuthMiddleware: Token validatsiya xatoligi: %v", err)
			http.Error(w, "Yaroqsiz token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// --- Asosiy o'zgarishlar shu yerda ---
		// jwt.MapClaims dan qiymatlarni olish
		telegramIDFloat, ok := claims["telegram_id"].(float64) // JWT claimlarida raqamlar float64 bo'lishi mumkin
		if !ok {
			log.Println("AuthMiddleware: 'telegram_id' claim topilmadi yoki turi noto'g'ri")
			http.Error(w, "Foydalanuvchi Telegram IDsi topilmadi", http.StatusUnauthorized)
			return
		}
		telegramID := int64(telegramIDFloat) // int64 ga o'girish

		role, ok := claims["role"].(string)
		if !ok {
			log.Println("AuthMiddleware: 'role' claim topilmadi yoki turi noto'g'ri")
			// Rol majburiy bo'lmasa, bu xatoni qaytarmasligingiz mumkin, balki default qiymat berasiz.
			// Bu misolda, topilmasa ham davom etamiz, lekin loglaymiz.
			role = "" // Default bo'sh string
		}

		// Foydalanuvchi Telegram ID va Rolini contextga qo'shish
		ctx := context.WithValue(r.Context(), TelegramIDContextKey, telegramID)
		ctx = context.WithValue(ctx, RoleContextKey, role) // role string bo'ladi
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RoleMiddleware faqat ma'lum bir rolga ega foydalanuvchilarga ruxsat beradi
func RoleMiddleware(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(RoleContextKey).(string)
		if !ok {
			log.Println("RoleMiddleware: Foydalanuvchi roli kontekstda topilmadi")
			http.Error(w, "Foydalanuvchi roli topilmadi (AuthMiddleware avval ishlashi kerak)", http.StatusInternalServerError)
			return
		}

		if role != requiredRole {
			log.Printf("RoleMiddleware: Ruxsat berilmagan harakat. Talab qilingan rol: %s, Foydalanuvchi roli: %s", requiredRole, role)
			http.Error(w, "Sizda bu operatsiya uchun ruxsat yo'q", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
