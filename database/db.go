package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq" // PostgreSQL drayveri
)

// Database structurasi ma'lumotlar bazasi bilan ishlash uchun.
type Database struct {
	db *sql.DB
}

// NewDatabase yangi Database instansiyasini yaratadi va jadvallarni sozlaydi.
func NewDatabase(connectionString string) (*Database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	// PostgreSQL ulanishini tekshirish
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("PostgreSQL ga ulanib bo'lmadi: %v", err)
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		db.Close() // Xatolik yuz bersa, DB ni yopish
		return nil, err
	}

	return database, nil
}

// createTables kerakli jadvallarni (agar mavjud bo'lmasa) yaratadi.
func (d *Database) createTables() error {
	// Users table
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		userid SERIAL PRIMARY KEY,
		telegram_id BIGINT UNIQUE NOT NULL,
		first_name TEXT NOT NULL DEFAULT 'N/A',
		username TEXT NOT NULL DEFAULT 'N/A',
		language_code TEXT NOT NULL DEFAULT 'uz',
		phone TEXT,
		password_hash TEXT, -- Yangi: Parol hash
		role TEXT NOT NULL DEFAULT 'user', -- Yangi: Foydalanuvchi roli (user, admin)
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(userTable); err != nil {
		log.Printf("Users jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'users' jadvali mavjud yoki yaratildi.")

	// Foods table
	foodTable := `
	CREATE TABLE IF NOT EXISTS foods (
		food_id SERIAL PRIMARY KEY,
		food_name TEXT NOT NULL,
		food_category TEXT NOT NULL,
		food_price DECIMAL(10,2) NOT NULL,
		food_image TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(foodTable); err != nil {
		log.Printf("Foods jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'foods' jadvali mavjud yoki yaratildi.")

	// Tables table (orders dan oldin yaratish kerak foreign key uchun)
	tableTable := `
	CREATE TABLE IF NOT EXISTS tables (
		table_id SERIAL PRIMARY KEY,
		table_name TEXT UNIQUE NOT NULL,
		qr_code_token TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := d.db.Exec(tableTable); err != nil {
		log.Printf("Tables jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'tables' jadvali mavjud yoki yaratildi.")

	// Orders table
	orderTable := `
	CREATE TABLE IF NOT EXISTS orders (
		order_id SERIAL PRIMARY KEY,
		telegram_id BIGINT NOT NULL,
		order_status TEXT NOT NULL DEFAULT 'pending',
		total_price DECIMAL(10,2) NOT NULL DEFAULT 0.0,
		order_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		delivery_type TEXT NOT NULL,
		delivery_latitude DECIMAL(10,8),
		delivery_longitude DECIMAL(11,8),
		table_id INTEGER,
		comment TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (telegram_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
		FOREIGN KEY (table_id) REFERENCES tables(table_id) ON DELETE SET NULL
	);`
	if _, err := d.db.Exec(orderTable); err != nil {
		log.Printf("Orders jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'orders' jadvali mavjud yoki yaratildi (yangi ustunlar bilan).")

	// Order Items table
	orderItemTable := `
	CREATE TABLE IF NOT EXISTS order_items (
		order_item_id SERIAL PRIMARY KEY,
		order_id INTEGER NOT NULL,
		food_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		item_price DECIMAL(10,2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE,
		FOREIGN KEY (food_id) REFERENCES foods(food_id) ON DELETE CASCADE
	);`
	if _, err := d.db.Exec(orderItemTable); err != nil {
		log.Printf("Order_items jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'order_items' jadvali mavjud yoki yaratildi.")

	// Basket Orders table
	basketOrderTable := `
	CREATE TABLE IF NOT EXISTS basket_orders (
		basket_order_id SERIAL PRIMARY KEY,
		telegram_id BIGINT NOT NULL,
		food_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(telegram_id, food_id),
		FOREIGN KEY (telegram_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
		FOREIGN KEY (food_id) REFERENCES foods(food_id) ON DELETE CASCADE
	);`

	if _, err := d.db.Exec(basketOrderTable); err != nil {
		log.Printf("Basket_orders jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'basket_orders' jadvali mavjud yoki yaratildi.")

	// Ustunlarni qo'shish yoki o'zgartirish
	if err := d.handleColumnMigrations(); err != nil {
		return err
	}

	return nil
}

// handleColumnMigrations ustunlarni qo'shish va o'zgartirish bilan ishlaydi
func (d *Database) handleColumnMigrations() error {
	// `orders` jadvali uchun ustunlarni qo'shish
	ordersColumnsToAdd := map[string]string{
		"order_status":       "TEXT NOT NULL DEFAULT 'pending'",
		"order_time":         "TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		"delivery_type":      "TEXT",
		"delivery_latitude":  "DECIMAL(10,8)",
		"delivery_longitude": "DECIMAL(11,8)",
		"table_id":           "INTEGER",
		"comment":            "TEXT",
	}

	for colName, colDef := range ordersColumnsToAdd {
		if !d.columnExists("orders", colName) {
			alterQuery := fmt.Sprintf("ALTER TABLE orders ADD COLUMN %s %s;", colName, colDef)
			_, err := d.db.Exec(alterQuery)
			if err != nil {
				log.Printf("Orders jadvaliga '%s' ustunini qo'shishda xatolik: %v", colName, err)
				return err
			}
			log.Printf("✅ 'orders' jadvaliga '%s' ustuni qo'shildi.", colName)
		} else {
			log.Printf("ℹ️ 'orders' jadvalida '%s' ustuni allaqachon mavjud.", colName)
		}
	}

	// `order_items` jadvali uchun ustunlarni qo'shish
	orderItemsColumnsToAdd := map[string]string{
		"item_price": "DECIMAL(10,2) NOT NULL DEFAULT 0.0",
	}

	for colName, colDef := range orderItemsColumnsToAdd {
		if !d.columnExists("order_items", colName) {
			alterQuery := fmt.Sprintf("ALTER TABLE order_items ADD COLUMN %s %s;", colName, colDef)
			_, err := d.db.Exec(alterQuery)
			if err != nil {
				log.Printf("Order_items jadvaliga '%s' ustunini qo'shishda xatolik: %v", colName, err)
				return err
			}
			log.Printf("✅ 'order_items' jadvaliga '%s' ustuni qo'shildi.", colName)
		} else {
			log.Printf("ℹ️ 'order_items' jadvalida '%s' ustuni allaqachon mavjud.", colName)
		}
	}

	return nil
}

// columnExists PostgreSQL da jadvalda ustun borligini tekshiradi.
func (d *Database) columnExists(tableName, columnName string) bool {
	query := `
		SELECT COUNT(*)
		FROM information_schema.columns 
		WHERE table_name = $1 AND column_name = $2;`

	var count int
	err := d.db.QueryRow(query, strings.ToLower(tableName), strings.ToLower(columnName)).Scan(&count)
	if err != nil {
		log.Printf("Ustun mavjudligini tekshirishda xatolik: %v", err)
		return false
	}

	return count > 0
}

// GetDB joriy *sql.DB instansiyasini qaytaradi.
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// Close ma'lumotlar bazasi ulanishini yopadi.
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}
