package jwt_auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("juda_maxfiy_kalit_hech_kim_bilmasin") // Ishlab chiqarishda buni atrof-muhit o'zgaruvchisidan oling!

// Claims strukturasi, JWT ichidagi ma'lumotlar uchun
type Claims struct {
	TelegramID int64  `json:"telegram_id"` // <--- Mana bu joyni qo'shish yoki o'zgartirish kerak!
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken berilgan foydalanuvchi Telegram ID va roli uchun yangi JWT yaratadi
// Endi UserID o'rniga telegramID int64 qabul qiladi
func GenerateToken(telegramID int64, role string) (string, error) { // <--- Imzosini o'zgartirish
	expirationTime := time.Now().Add(24 * time.Hour) // Token 24 soat amal qiladi

	claims := &Claims{
		TelegramID: telegramID, // <--- Bu yerda TelegramID ishlatiladi
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app-name",
			Subject:   fmt.Sprintf("%d", telegramID), // Subject ham Telegram ID bo'lishi mumkin
			Audience:  []string{"users"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("token yaratishda xatolik: %w", err)
	}
	return tokenString, nil
}

// ValidateToken berilgan token stringini tekshiradi va agar to'g'ri bo'lsa, claimsni qaytaradi
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("kutilmagan imzolash usuli: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("tokenni tekshirishda xatolik: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token yaroqsiz")
	}

	return claims, nil
}

func GetSecret() []byte {
	return jwtSecret
}
