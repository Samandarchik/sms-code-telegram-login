package models

import "time"

type User struct {
	UserID       int       `json:"user_id" db:"userid"`
	TelegramID   int64     `json:"TelegramID" db:"TelegramID"`
	FirstName    string    `json:"first_name" db:"first_name"`
	Username     string    `json:"username" db:"username"`
	LanguageCode string    `json:"language_code" db:"language_code"`
	PhoneNumber  string    `json:"phone_number" db:"phone"`
	Role         string    `json:"role" db:"role"` // "user", "admin", "superadmin"
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
type LoginRequest struct {
	TelegramID int64  `json:"telegram_id"` // Foydalanuvchi Telegram ID
	Username   string `json:"username"`    // Telegram username (ixtiyoriy, tekshirish uchun)
}

// LoginResponse muvaffaqiyatli kirishdan keyin qaytariladigan javob
type LoginResponse struct {
	Token string `json:"token"` // JWT tokeni
}
type CreateUserRequest struct {
	UserID    int64  `json:"user_id"` // Telegram ID
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
