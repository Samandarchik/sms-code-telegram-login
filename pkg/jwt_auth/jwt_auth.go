package jwt_auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5" // jwt.v5 ni ishlatayotgan bo'lsangiz
)

// JWT_SECRET o'zgaruvchisi. Environment variable dan olinishi kerak.
// Muhim: Bu joyga kuchli va real secret key yozing!
// Productionda buni environment variable dan oling.
var JWT_SECRET = []byte("your_super_secret_jwt_key")

// GenerateToken foydalanuvchi ID va roli uchun JWT token yaratadi
func GenerateToken(telegramID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"telegram_id": telegramID, // Claim nomi
		"role":        role,
		"exp":         time.Now().Add(time.Hour * 24).Unix(), // Token 24 soat amal qiladi
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(JWT_SECRET)
	if err != nil {
		return "", fmt.Errorf("token imzolashta xatolik: %w", err)
	}
	return signedToken, nil
}
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("kutilmagan imzolash usuli: %v", token.Header["alg"])
		}
		return JWT_SECRET, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token tahlil qilishda xatolik: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token yaroqsiz")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("token claims MapClaims emas")
	}

	return claims, nil
}

// ContextKey kontekst uchun maxsus kalit turi
type contextKey string

const ContextTelegramIDKey contextKey = "telegramID" // Kontekstga qo'shish uchun kalit

// AuthMiddleware JWT tokenini tekshiradi va foydalanuvchi ma'lumotlarini kontekstga qo'shadi
// Bu funksiya katta 'A' bilan boshlanishi muhim, chunki u eksport qilinadi.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("AuthMiddleware: Authorization header topilmadi")
			http.Error(w, "Autentifikatsiya xatoligi: Authorization header topilmadi", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		if tokenString == authHeader { // "Bearer " prefiksi topilmagan bo'lsa
			log.Println("AuthMiddleware: Token Bearer prefiksi bilan berilmagan")
			http.Error(w, "Autentifikatsiya xatoligi: Token Bearer prefiksi bilan berilmagan", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("kutilmagan imzolash usuli: %v", token.Header["alg"])
			}
			return JWT_SECRET, nil
		})

		if err != nil {
			log.Printf("AuthMiddleware: Token tahlil qilishda xatolik: %v", err)
			http.Error(w, fmt.Sprintf("Autentifikatsiya xatoligi: Yaroqsiz token (%v)", err), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			log.Println("AuthMiddleware: Token yaroqsiz")
			http.Error(w, "Autentifikatsiya xatoligi: Token yaroqsiz", http.StatusUnauthorized)
			return
		}

		// Claims'dan ma'lumotlarni olish
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("AuthMiddleware: Token claims MapClaims emas")
			http.Error(w, "Autentifikatsiya xatoligi: Token claims noto'g'ri formatda", http.StatusUnauthorized)
			return
		}

		telegramIDFloat, ok := claims["telegram_id"].(float64) // JWT'dan olingan raqamlar float64 bo'lishi mumkin
		if !ok {
			log.Println("AuthMiddleware: JWT claimsda 'telegram_id' topilmadi yoki turi noto'g'ri")
			http.Error(w, "Autentifikatsiya xatoligi: Foydalanuvchi IDsi topilmadi", http.StatusUnauthorized)
			return
		}
		telegramID := int64(telegramIDFloat) // int64 ga o'girish

		// Kontekstga foydalanuvchi ID va roli qo'shish
		ctx := context.WithValue(r.Context(), ContextTelegramIDKey, telegramID)
		// Agar rol ham kerak bo'lsa va uni kontekstga qo'shmoqchi bo'lsangiz:
		// if roleStr, ok := claims["role"].(string); ok {
		//     ctx = context.WithValue(ctx, "role", roleStr)
		// }

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
