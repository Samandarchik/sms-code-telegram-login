// service/basket_order_service.go
package service

import (
	"amur/models"
	"amur/repository"
	"database/sql"
	"errors"
	"fmt"
)

type BasketOrderService struct {
	basketRepo *repository.BasketOrderRepository
	foodRepo   *repository.FoodRepository // Oziq-ovqat ma'lumotlarini olish uchun
}

func NewBasketOrderService(basketRepo *repository.BasketOrderRepository, foodRepo *repository.FoodRepository) *BasketOrderService {
	return &BasketOrderService{
		basketRepo: basketRepo,
		foodRepo:   foodRepo,
	}
}

// AddToBasket savatchaga mahsulot qo'shadi yoki miqdorini yangilaydi
func (s *BasketOrderService) AddToBasket(telegramID int64, req *models.AddToBasketRequest) (*models.BasketOrder, error) {
	// Mahsulot mavjudligini tekshirish
	food, err := s.foodRepo.GetByID(req.FoodID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("id=%d bo'lgan ovqat topilmadi", req.FoodID)
		}
		return nil, fmt.Errorf("ovqatni olishda xatolik: %w", err)
	}

	// Savatchaga qo'shish yoki miqdorini oshirish
	order, err := s.basketRepo.AddToBasket(telegramID, food.FoodID)
	if err != nil {
		return nil, fmt.Errorf("savatchaga qo'shishda xatolik: %w", err)
	}
	return order, nil
}

// GetBasketOrders savatchadagi barcha buyurtmalarni Food ma'lumotlari bilan qaytaradi
func (s *BasketOrderService) GetBasketOrders(telegramID int64) ([]map[string]interface{}, error) {
	basketItems, err := s.basketRepo.GetBasketOrdersByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}

	var detailedBasket []map[string]interface{}
	for _, item := range basketItems {
		food, err := s.foodRepo.GetByID(item.FoodID)
		if err != nil {
			// Agar ovqat topilmasa, uni o'tkazib yuboramiz yoki xato qaytaramiz
			// Hozircha logga yozamiz va o'tkazib yuboramiz
			// Real ilovada bu yerdagi xatolarni yaxshiroq boshqarish kerak
			fmt.Printf("FoodID %d uchun ovqat topilmadi, savatchadan o'tkazib yuborilmoqda: %v\n", item.FoodID, err)
			continue
		}

		detailedItem := map[string]interface{}{
			"basket_order_id": item.BasketOrderID,
			"tg_id":           item.TelegramID,
			"food_id":         item.FoodID,
			"quantity":        item.Quantity,
			"food_name":       food.FoodName,
			"food_category":   food.FoodCategory,
			"food_price":      food.FoodPrice,
			"food_image":      food.FoodImage,
			"total_price":     float64(item.Quantity) * food.FoodPrice,
			"created_at":      item.CreatedAt,
			"updated_at":      item.UpdatedAt,
		}
		detailedBasket = append(detailedBasket, detailedItem)
	}

	return detailedBasket, nil
}

// RemoveFromBasket savatchadan mahsulotni olib tashlaydi
func (s *BasketOrderService) RemoveFromBasket(telegramID int64, foodID int) error {
	err := s.basketRepo.RemoveFromBasket(telegramID, foodID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("tg_id=%d va food_id=%d bo'lgan savatcha elementi topilmadi", telegramID, foodID)
		}
		return fmt.Errorf("savatchadan o'chirishda xatolik: %w", err)
	}
	return nil
}

// ClearBasket savatchani tozalaydi
func (s *BasketOrderService) ClearBasket(telegramID int64) error {
	err := s.basketRepo.ClearBasket(telegramID)
	if err != nil {
		return fmt.Errorf("savatchani tozalashda xatolik: %w", err)
	}
	return nil
}
