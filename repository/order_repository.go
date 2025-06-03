// repository/order_repository.go
package repository

import (
	"amur/models"
	"database/sql"
	"log"
	"time"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder buyurtmani ma'lumotlar bazasiga qo'shadi va uning ID'sini qaytaradi
func (r *OrderRepository) CreateOrder(order *models.Order) (*models.Order, error) {
	stmt, err := r.db.Prepare(`
        INSERT INTO orders(telegram_id, order_time, order_status, delivery_type, total_price, delivery_latitude, delivery_longitude, comment)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING order_id
    `)
	if err != nil {
		log.Printf("Order CreateOrder prepare xatolik: %v", err)
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(
		order.TelegramID,
		order.OrderTime,
		order.OrderStatus,
		order.DeliveryType,
		order.TotalPrice,
		order.DeliveryLatitude,
		order.DeliveryLongitude,
		order.Comment,
	).Scan(&order.OrderID)
	if err != nil {
		log.Printf("Order CreateOrder exec xatolik: %v", err)
		return nil, err
	}

	// Set timestamps in Go since they're not in DB
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	log.Printf("✅ Yangi buyurtma yaratildi: OrderID=%d, TelegramID=%d", order.OrderID, order.TelegramID)
	return order, nil
}

// AddOrderItem buyurtma elementini (mahsulotni) ma'lumotlar bazasiga qo'shadi
func (r *OrderRepository) AddOrderItem(item *models.OrderItem) error {
	stmt, err := r.db.Prepare(`
        INSERT INTO order_items(order_id, food_id, quantity, item_price)
        VALUES ($1, $2, $3, $4)
    `)
	if err != nil {
		log.Printf("Order AddOrderItem prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(item.OrderID, item.FoodID, item.Quantity, item.ItemPrice)
	if err != nil {
		log.Printf("Order AddOrderItem exec xatolik: %v", err)
		return err
	}
	log.Printf("✅ Buyurtma elementi qo'shildi: OrderID=%d, FoodID=%d, Quantity=%d", item.OrderID, item.FoodID, item.Quantity)
	return nil
}

// GetOrderWithItemsByID buyurtmani uning elementlari bilan birga oladi
func (r *OrderRepository) GetOrderWithItemsByID(orderID int) (*models.Order, []*models.OrderItem, error) {
	var order models.Order
	err := r.db.QueryRow(`
        SELECT order_id, telegram_id, order_time, order_status, delivery_type, total_price, delivery_latitude, delivery_longitude, comment
        FROM orders
        WHERE order_id = $1
    `, orderID).Scan(
		&order.OrderID,
		&order.TelegramID,
		&order.OrderTime,
		&order.OrderStatus,
		&order.DeliveryType,
		&order.TotalPrice,
		&order.DeliveryLatitude,
		&order.DeliveryLongitude,
		&order.Comment,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, sql.ErrNoRows // Buyurtma topilmadi
		}
		log.Printf("Order GetOrderWithItemsByID (order) xatolik: %v", err)
		return nil, nil, err
	}

	// Set default timestamps
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	rows, err := r.db.Query(`
        SELECT order_item_id, order_id, food_id, quantity, item_price
        FROM order_items
        WHERE order_id = $1
    `, orderID)
	if err != nil {
		log.Printf("Order GetOrderWithItemsByID (items) xatolik: %v", err)
		return nil, nil, err
	}
	defer rows.Close()

	var orderItems []*models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(
			&item.OrderItemID,
			&item.OrderID,
			&item.FoodID,
			&item.Quantity,
			&item.ItemPrice,
		)
		if err != nil {
			log.Printf("Order GetOrderWithItemsByID (item scan) xatolik: %v", err)
			continue
		}
		// Set default timestamps for items too
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()
		orderItems = append(orderItems, &item)
	}

	return &order, orderItems, nil
}

// UpdateOrderStatus buyurtma holatini yangilaydi
func (r *OrderRepository) UpdateOrderStatus(orderID int, status string) error {
	stmt, err := r.db.Prepare(`
        UPDATE orders
        SET order_status = $1
        WHERE order_id = $2
    `)
	if err != nil {
		log.Printf("Order UpdateOrderStatus prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(status, orderID)
	if err != nil {
		log.Printf("Order UpdateOrderStatus exec xatolik: %v", err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows // Agar buyurtma topilmasa
	}
	log.Printf("🔄 Buyurtma holati yangilandi: OrderID=%d, Yangi holat='%s'", orderID, status)
	return nil
}

// GetUserOrders berilgan Telegram ID bo'yicha foydalanuvchining barcha buyurtmalarini oladi
func (r *OrderRepository) GetUserOrders(telegramID int64) ([]*models.Order, error) {
	rows, err := r.db.Query(`
        SELECT order_id, telegram_id, order_time, order_status, delivery_type, total_price, delivery_latitude, delivery_longitude, comment
        FROM orders
        WHERE telegram_id = $1
        ORDER BY order_time DESC
    `, telegramID)
	if err != nil {
		log.Printf("Order GetUserOrders query xatolik: %v", err)
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.OrderID,
			&order.TelegramID,
			&order.OrderTime,
			&order.OrderStatus,
			&order.DeliveryType,
			&order.TotalPrice,
			&order.DeliveryLatitude,
			&order.DeliveryLongitude,
			&order.Comment,
		)
		if err != nil {
			log.Printf("Order GetUserOrders scan xatolik: %v", err)
			continue
		}
		// Set default timestamps
		order.CreatedAt = time.Now()
		order.UpdatedAt = time.Now()
		orders = append(orders, &order)
	}
	return orders, nil
}

// GetOrderStats buyurtma statistikasini oladi
func (r *OrderRepository) GetOrderStats() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
	if err != nil {
		log.Printf("Order GetOrderStats xatolik: %v", err)
		return 0, err
	}
	return count, nil
}

// DeleteOrder buyurtmani ID bo'yicha o'chiradi (FAQAT ADMIN UCHUN)
func (r *OrderRepository) DeleteOrder(orderID int) error {
	stmt, err := r.db.Prepare("DELETE FROM orders WHERE order_id = $1")
	if err != nil {
		log.Printf("Order DeleteOrder prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(orderID)
	if err != nil {
		log.Printf("Order DeleteOrder exec xatolik: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows // Agar buyurtma topilmadi
	}

	log.Printf("🗑️ Buyurtma o'chirildi: OrderID=%d", orderID)
	return nil
}
