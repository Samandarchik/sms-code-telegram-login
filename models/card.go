package models

import "time"

// CartItem savatdagi bitta mahsulotni ifodalaydi
type CartItem struct {
	ID        int       `json:"id"`
	UserID    int64     `json:"user_id"` // Telegram ID
	FoodID    int       `json:"food_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartItemResponse savatdagi mahsulotni to'liq ma'lumotlari bilan ifodalaydi (Food ma'lumotlari bilan birga)
type CartItemResponse struct {
	ID        int       `json:"id"`
	UserID    int64     `json:"user_id"`
	Food      *Food     `json:"food"` // Food modelining to'liq ma'lumotlari
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AddToCartRequest savatga mahsulot qo'shish uchun keladigan so'rov tanasi
type AddToCartRequest struct {
	FoodID   int `json:"food_id" validate:"required,gt=0"`
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

// UpdateCartItemRequest savatdagi mahsulot miqdorini yangilash uchun keladigan so'rov tanasi
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}
