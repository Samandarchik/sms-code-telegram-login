package repository

import (
	"amur/models"
	"database/sql"
	"log"
	"time"
)

type BasketOrderRepository struct {
	db *sql.DB
}

func NewBasketOrderRepository(db *sql.DB) *BasketOrderRepository {
	return &BasketOrderRepository{db: db}
}

// AddToBasket savatchaga mahsulot qo'shadi yoki mavjud bo'lsa miqdorini oshiradi
func (r *BasketOrderRepository) AddToBasket(telegramID int64, foodID int) (*models.BasketOrder, error) {
	// Avval ushbu foydalanuvchi va ovqat ID'si bo'yicha mavjud buyurtmani qidiramiz
	var existingOrder models.BasketOrder
	err := r.db.QueryRow(`
        SELECT basket_order_id, telegram_id, food_id, quantity, created_at, updated_at
        FROM basket_orders
        WHERE telegram_id = ? AND food_id = ?
    `, telegramID, foodID).Scan(
		&existingOrder.BasketOrderID,
		&existingOrder.TelegramID,
		&existingOrder.FoodID,
		&existingOrder.Quantity,
		&existingOrder.CreatedAt,
		&existingOrder.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Agar mavjud bo'lmasa, yangi yozuv yaratamiz
		stmt, err := r.db.Prepare(`
            INSERT INTO basket_orders(telegram_id, food_id, quantity)
            VALUES (?, ?, 1)
        `)
		if err != nil {
			log.Printf("BasketOrder AddToBasket (insert) prepare xatolik: %v", err)
			return nil, err
		}
		defer stmt.Close()

		result, err := stmt.Exec(telegramID, foodID)
		if err != nil {
			log.Printf("BasketOrder AddToBasket (insert) exec xatolik: %v", err)
			return nil, err
		}

		id, _ := result.LastInsertId()
		newOrder := &models.BasketOrder{
			BasketOrderID: int(id),
			TelegramID:    telegramID,
			FoodID:        foodID,
			Quantity:      1,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		log.Printf("✅ Yangi savatcha buyurtmasi qo'shildi: TelegramID=%d, FoodID=%d", telegramID, foodID)
		return newOrder, nil
	} else if err != nil {
		log.Printf("BasketOrder AddToBasket (select) xatolik: %v", err)
		return nil, err
	}

	// Agar mavjud bo'lsa, miqdorini oshiramiz
	stmt, err := r.db.Prepare(`
        UPDATE basket_orders
        SET quantity = quantity + 1, updated_at = CURRENT_TIMESTAMP
        WHERE basket_order_id = ?
    `)
	if err != nil {
		log.Printf("BasketOrder AddToBasket (update) prepare xatolik: %v", err)
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(existingOrder.BasketOrderID)
	if err != nil {
		log.Printf("BasketOrder AddToBasket (update) exec xatolik: %v", err)
		return nil, err
	}

	existingOrder.Quantity++ // Miqdorni lokal obyektda ham yangilaymiz
	existingOrder.UpdatedAt = time.Now()
	log.Printf("🔄 Savatcha buyurtmasi miqdori yangilandi: TelegramID=%d, FoodID=%d, Yangi miqdor=%d", telegramID, foodID, existingOrder.Quantity)
	return &existingOrder, nil
}

// GetBasketOrdersByTelegramID berilgan Telegram ID bo'yicha savatchadagi barcha buyurtmalarni oladi
func (r *BasketOrderRepository) GetBasketOrdersByTelegramID(telegramID int64) ([]*models.BasketOrder, error) {
	rows, err := r.db.Query(`
        SELECT basket_order_id, telegram_id, food_id, quantity, created_at, updated_at
        FROM basket_orders
        WHERE telegram_id = ?
        ORDER BY created_at DESC
    `, telegramID)
	if err != nil {
		log.Printf("BasketOrder GetBasketOrdersByTelegramID xatolik: %v", err)
		return nil, err
	}
	defer rows.Close()

	var orders []*models.BasketOrder
	for rows.Next() {
		var order models.BasketOrder
		err := rows.Scan(&order.BasketOrderID, &order.TelegramID, &order.FoodID,
			&order.Quantity, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			log.Printf("BasketOrder GetBasketOrdersByTelegramID scan xatolik: %v", err)
			continue
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

// RemoveFromBasket savatchadan mahsulotni olib tashlaydi
func (r *BasketOrderRepository) RemoveFromBasket(telegramID int64, foodID int) error {
	stmt, err := r.db.Prepare("DELETE FROM basket_orders WHERE telegram_id = ? AND food_id = ?")
	if err != nil {
		log.Printf("BasketOrder RemoveFromBasket prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(telegramID, foodID)
	if err != nil {
		log.Printf("BasketOrder RemoveFromBasket exec xatolik: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows // Agar hech narsa o'chirilmasa, xato qaytaramiz
	}

	log.Printf("🗑️ Savatchadan mahsulot o'chirildi: TelegramID=%d, FoodID=%d", telegramID, foodID)
	return nil
}

// ClearBasket berilgan Telegram ID bo'yicha savatchani tozalaydi
func (r *BasketOrderRepository) ClearBasket(telegramID int64) error {
	stmt, err := r.db.Prepare("DELETE FROM basket_orders WHERE telegram_id = ?")
	if err != nil {
		log.Printf("BasketOrder ClearBasket prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(telegramID)
	if err != nil {
		log.Printf("BasketOrder ClearBasket exec xatolik: %v", err)
		return err
	}

	log.Printf("🗑️ Savatcha tozalab tashlandi: TelegramID=%d", telegramID)
	return nil
}
