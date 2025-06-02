package repository

import (
	"amur/models"
	"database/sql"
	"fmt"
	"log"
)

// CartRepository savatdagi elementlar uchun ma'lumotlar bazasi operatsiyalarini boshqaradi.
type CartRepository struct {
	db *sql.DB
}

// NewCartRepository CartRepository ning yangi instansini yaratadi.
func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

// CreateOrUpdateCartItem savatga mahsulot qo'shadi yoki agar mavjud bo'lsa, miqdorini yangilaydi.
func (r *CartRepository) CreateOrUpdateCartItem(item *models.CartItem) (*models.CartItem, error) {
	// UNIQUE (user_id, food_id) cheklovi tufayli INSERT OR REPLACE ishlatish mumkin
	// yoki avval tekshirib, keyin INSERT/UPDATE qilish mumkin.
	// Hozirda UPSERT (UPDATE OR INSERT) logikasi qo'llaniladi.

	// Avval mavjudligini tekshiramiz
	existingItem := &models.CartItem{}
	err := r.db.QueryRow("SELECT id, quantity FROM cart_items WHERE user_id = ? AND food_id = ?", item.UserID, item.FoodID).
		Scan(&existingItem.ID, &existingItem.Quantity)

	if err == sql.ErrNoRows {
		// Mavjud emas, yangi element qo'shamiz
		stmt, err := r.db.Prepare(`
			INSERT INTO cart_items(user_id, food_id, quantity)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			log.Printf("Cart Create prepare xatolik: %v", err)
			return nil, err
		}
		defer stmt.Close()

		result, err := stmt.Exec(item.UserID, item.FoodID, item.Quantity)
		if err != nil {
			log.Printf("Cart Create exec xatolik: %v", err)
			return nil, err
		}

		id, _ := result.LastInsertId()
		item.ID = int(id)
		log.Printf("✅ Savatga yangi mahsulot qo'shildi: UserID=%d, FoodID=%d, Quantity=%d (ID: %d)", item.UserID, item.FoodID, item.Quantity, item.ID)
		return item, nil
	} else if err != nil {
		log.Printf("Cart item mavjudligini tekshirishda xatolik: %v", err)
		return nil, err
	}

	// Mavjud, miqdorini yangilaymiz
	newQuantity := existingItem.Quantity + item.Quantity
	if item.Quantity == 0 { // Agar miqdor 0 bo'lsa, o'chirishni nazarda tutish mumkin
		return nil, fmt.Errorf("miqdor 0 dan katta bo'lishi kerak")
	}
	if item.Quantity < 0 && existingItem.Quantity+item.Quantity < 0 {
		return nil, fmt.Errorf("miqdor manfiy bo'lishi mumkin emas")
	}

	stmt, err := r.db.Prepare(`
		UPDATE cart_items SET
			quantity = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND food_id = ?
	`)
	if err != nil {
		log.Printf("Cart Update prepare xatolik: %v", err)
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newQuantity, item.UserID, item.FoodID)
	if err != nil {
		log.Printf("Cart Update exec xatolik: %v", err)
		return nil, err
	}
	item.ID = existingItem.ID   // ID ni mavjud elementdan olish
	item.Quantity = newQuantity // Yangi miqdorni qaytarish
	log.Printf("🔄 Savatdagi mahsulot miqdori yangilandi: UserID=%d, FoodID=%d, NewQuantity=%d (ID: %d)", item.UserID, item.FoodID, newQuantity, item.ID)
	return item, nil
}

// GetCartItemsByUserID foydalanuvchining savatidagi barcha elementlarni qaytaradi.
func (r *CartRepository) GetCartItemsByUserID(userID int64) ([]*models.CartItem, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, food_id, quantity, created_at, updated_at
		FROM cart_items
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		log.Printf("Savat elementlarini olishda xatolik: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []*models.CartItem
	for rows.Next() {
		item := &models.CartItem{}
		err := rows.Scan(&item.ID, &item.UserID, &item.FoodID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			log.Printf("Savat elementini skanerlashda xatolik: %v", err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// UpdateCartItemQuantity savatdagi ma'lum bir mahsulot miqdorini yangilaydi.
func (r *CartRepository) UpdateCartItemQuantity(userID int64, foodID int, quantity int) error {
	if quantity <= 0 {
		// Agar miqdor 0 yoki undan kam bo'lsa, elementni o'chirishni nazarda tutish mumkin
		return r.DeleteCartItem(userID, foodID)
	}

	stmt, err := r.db.Prepare(`
		UPDATE cart_items SET
			quantity = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND food_id = ?
	`)
	if err != nil {
		log.Printf("Cart item miqdorini yangilashda prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(quantity, userID, foodID)
	if err != nil {
		log.Printf("Cart item miqdorini yangilashda exec xatolik: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("savatdagi mahsulot topilmadi yoki yangilanmadi")
	}

	log.Printf("🔄 Savatdagi mahsulot miqdori yangilandi: UserID=%d, FoodID=%d, Quantity=%d", userID, foodID, quantity)
	return nil
}

// DeleteCartItem savatdagi ma'lum bir mahsulotni o'chiradi.
func (r *CartRepository) DeleteCartItem(userID int64, foodID int) error {
	stmt, err := r.db.Prepare("DELETE FROM cart_items WHERE user_id = ? AND food_id = ?")
	if err != nil {
		log.Printf("Cart item o'chirishda prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(userID, foodID)
	if err != nil {
		log.Printf("Cart item o'chirishda exec xatolik: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("savatdagi mahsulot topilmadi")
	}

	log.Printf("🗑️ Savatdagi mahsulot o'chirildi: UserID=%d, FoodID=%d", userID, foodID)
	return nil
}

// ClearCart foydalanuvchining savatidagi barcha elementlarni o'chiradi.
func (r *CartRepository) ClearCart(userID int64) error {
	stmt, err := r.db.Prepare("DELETE FROM cart_items WHERE user_id = ?")
	if err != nil {
		log.Printf("Savatni tozalashda prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID)
	if err != nil {
		log.Printf("Savatni tozalashda exec xatolik: %v", err)
		return err
	}

	log.Printf("🧹 Savat tozalandi: UserID=%d", userID)
	return nil
}
