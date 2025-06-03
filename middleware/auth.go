package middleware

import (
	"context"
	"net/http"
	"strings"

	"amur/pkg/jwt_auth" // jwt_auth paketi to'g'ri yo'lni ko'rsatishini tekshiring
)

// ContextKey o'zgaruvchisi context ichiga ma'lumot saqlash uchun
type ContextKey string

const (
	TelegramIDContextKey ContextKey = "telegram_id" // <--- Yangi key
	RoleContextKey       ContextKey = "role"
)

// AuthMiddleware JWT tokenini tekshiradi va foydalanuvchi ma'lumotlarini contextga qo'shadi
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Avtorizatsiya tokeni topilmadi", http.StatusUnauthorized)
			return
		}

		// "Bearer " prefiksini olib tashlash
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		if tokenString == authHeader { // Agar "Bearer " topilmasa
			http.Error(w, "Noto'g'ri avtorizatsiya formati", http.StatusUnauthorized)
			return
		}

		claims, err := jwt_auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Yaroqsiz token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Foydalanuvchi Telegram ID va Rolini contextga qo'shish
		// claims.UserID o'rniga claims.TelegramID ni ishlatamiz
		ctx := context.WithValue(r.Context(), TelegramIDContextKey, claims.TelegramID) // <--- O'zgarish
		ctx = context.WithValue(ctx, RoleContextKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RoleMiddleware faqat ma'lum bir rolga ega foydalanuvchilarga ruxsat beradi
func RoleMiddleware(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(RoleContextKey).(string)
		if !ok {
			http.Error(w, "Foydalanuvchi roli topilmadi (AuthMiddleware avval ishlashi kerak)", http.StatusInternalServerError)
			return
		}

		if role != requiredRole {
			http.Error(w, "Sizda bu operatsiya uchun ruxsat yo'q", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
