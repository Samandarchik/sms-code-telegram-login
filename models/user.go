package models

import "time"

type User struct {
	UserID       int       `json:"user_id" db:"userid"`
	TelegramID   int64     `json:"TelegramID" db:"TelegramID"`
	FirstName    string    `json:"first_name" db:"first_name"`
	Username     string    `json:"username" db:"username"`
	LanguageCode string    `json:"language_code" db:"language_code"`
	PhoneNumber  string    `json:"phone_number" db:"phone"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
