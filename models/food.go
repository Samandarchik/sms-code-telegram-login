package models

import "time"

type Food struct {
	FoodID       int       `json:"food_id" db:"food_id"`
	FoodName     string    `json:"food_name" db:"food_name"`
	FoodCategory string    `json:"food_category" db:"food_category"`
	FoodPrice    float64   `json:"food_price" db:"food_price"`
	FoodImage    string    `json:"food_image" db:"food_image"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type CreateFoodRequest struct {
	FoodName     string  `json:"food_name" validate:"required"`
	FoodCategory string  `json:"food_category" validate:"required"`
	FoodPrice    float64 `json:"food_price" validate:"required,gt=0"`
	FoodImage    string  `json:"food_image"`
}

type UpdateFoodRequest struct {
	FoodName     string  `json:"food_name"`
	FoodCategory string  `json:"food_category"`
	FoodPrice    float64 `json:"food_price"`
	FoodImage    string  `json:"food_image"`
}
