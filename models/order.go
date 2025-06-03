// models/order.go
package models

import "time"

// BasketOrder savatchadagi bitta mahsulotni ifodalaydi (unchanged)
type BasketOrder struct {
	BasketOrderID int       `json:"basket_order_id" db:"basket_order_id"`
	TelegramID    int64     `json:"telegram_id" db:"user_tg_id"` // db tagi to'g'rilandi
	FoodID        int       `json:"food_id" db:"food_id"`
	Quantity      int       `json:"quantity" db:"quantity"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// AddToBasketRequest savatchaga mahsulot qo'shish uchun so'rov formati (unchanged)
type AddToBasketRequest struct {
	FoodID int `json:"food_id" validate:"required,gt=0"`
}

// Order buyurtmaning asosiy ma'lumotlarini ifodalaydi
type Order struct {
	OrderID           int       `json:"order_id" db:"order_id"`
	TelegramID        int64     `json:"telegram_id" db:"telegram_id"`
	OrderTime         time.Time `json:"order_time" db:"order_time"`
	OrderStatus       string    `json:"order_status" db:"order_status"`
	DeliveryType      string    `json:"delivery_type" db:"delivery_type"`
	TotalPrice        float64   `json:"total_price" db:"total_price"`
	DeliveryLatitude  *float64  `json:"delivery_latitude,omitempty" db:"delivery_latitude"`
	DeliveryLongitude *float64  `json:"delivery_longitude,omitempty" db:"delivery_longitude"`
	Comment           *string   `json:"comment,omitempty" db:"comment"`
	TableID           *string   `json:"table_id,omitempty"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// OrderItem buyurtmadagi har bir alohida mahsulotni ifodalaydi (unchanged)
type OrderItem struct {
	OrderItemID int       `json:"order_item_id" db:"order_item_id"`
	OrderID     int       `json:"order_id" db:"order_id"`
	FoodID      int       `json:"food_id" db:"food_id"`
	Quantity    int       `json:"quantity" db:"quantity"`
	ItemPrice   float64   `json:"item_price" db:"item_price"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateOrderRequest buyurtma yaratish uchun keladigan so'rov formati
type CreateOrderRequest struct {
	TelegramID        int64    `json:"telegram_id" validate:"required"`
	DeliveryType      string   `json:"delivery_type" validate:"required,oneof='yetkazib berish' 'o''zi olib ketish' 'zalga'"`
	DeliveryLatitude  *float64 `json:"delivery_latitude,omitempty"`
	DeliveryLongitude *float64 `json:"delivery_longitude,omitempty"`
	Comment           *string  `json:"comment,omitempty"`
	TableID           *string  `json:"table_id,omitempty"` // NEW: For "zalga" orders QR code token
}

// OrderDetailsResponse buyurtma va uning ichidagi mahsulotlar bilan birgalikda to'liq javob (unchanged)
type OrderDetailsResponse struct {
	Order      Order       `json:"order"`
	OrderItems []OrderItem `json:"order_items"`
}
