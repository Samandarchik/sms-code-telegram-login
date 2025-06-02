package models

import "time"

type User struct {
	UserID       int       `json:"id"` // Database ID
	TgID         int64     `json:"tg_id"`
	FirstName    string    `json:"first_name"`
	Username     string    `json:"username"`
	LanguageCode string    `json:"language_code"`
	PhoneNumber  string    `json:"phone_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpdateUserCardRequest foydalanuvchi kartasini yangilash uchun
type UpdateUserCardRequest struct {
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	PhoneNumber  string `json:"phone_number"`
}
