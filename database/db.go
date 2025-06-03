package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings" // strings paketini import qildim

	_ "github.com/mattn/go-sqlite3" // SQLite3 drayveri
)

// Database structurasi ma'lumotlar bazasi bilan ishlash uchun.
type Database struct {
	db *sql.DB
}

// NewDatabase yangi Database instansiyasini yaratadi va jadvallarni sozlaydi.
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
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
        userid INTEGER PRIMARY KEY AUTOINCREMENT,
        telegram_id INTEGER UNIQUE NOT NULL,
        first_name TEXT NOT NULL DEFAULT 'N/A',
        username TEXT NOT NULL DEFAULT 'N/A',
        language_code TEXT NOT NULL DEFAULT 'uz',
        phone TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	if _, err := d.db.Exec(userTable); err != nil {
		log.Printf("Users jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'users' jadvali mavjud yoki yaratildi.")

	// Foods table
	foodTable := `
    CREATE TABLE IF NOT EXISTS foods (
        food_id INTEGER PRIMARY KEY AUTOINCREMENT,
        food_name TEXT NOT NULL,
        food_category TEXT NOT NULL,
        food_price REAL NOT NULL,
        food_image TEXT DEFAULT '',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	if _, err := d.db.Exec(foodTable); err != nil {
		log.Printf("Foods jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'foods' jadvali mavjud yoki yaratildi.")

	// Orders table
	orderTable := `
    CREATE TABLE IF NOT EXISTS orders (
        order_id INTEGER PRIMARY KEY AUTOINCREMENT,
        telegram_id INTEGER NOT NULL,
        order_status TEXT NOT NULL DEFAULT 'pending',
        total_price REAL NOT NULL DEFAULT 0.0,
        order_time DATETIME DEFAULT CURRENT_TIMESTAMP,
        delivery_type TEXT NOT NULL,
        delivery_latitude REAL,
        delivery_longitude REAL,
        table_id INTEGER,
        comment TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (telegram_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
        FOREIGN KEY (table_id) REFERENCES tables(table_id) ON DELETE SET NULL
    );`
	if _, err := d.db.Exec(orderTable); err != nil {
		log.Printf("Orders jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'orders' jadvali mavjud yoki yaratildi (yangi ustunlar bilan).")

	// Order Items table
	// YANGILANGAN: `price_at_order` -> `item_price` ga o'zgartirildi
	orderItemTable := `
    CREATE TABLE IF NOT EXISTS order_items (
        order_item_id INTEGER PRIMARY KEY AUTOINCREMENT,
        order_id INTEGER NOT NULL,
        food_id INTEGER NOT NULL,
        quantity INTEGER NOT NULL,
        item_price REAL NOT NULL, -- 'price_at_order' -> 'item_price' ga o'zgartirildi
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
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
        basket_order_id INTEGER PRIMARY KEY AUTOINCREMENT,
        telegram_id INTEGER NOT NULL,
        food_id INTEGER NOT NULL,
        quantity INTEGER NOT NULL DEFAULT 1,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(telegram_id, food_id),
        FOREIGN KEY (telegram_id) REFERENCES users(telegram_id) ON DELETE CASCADE,
        FOREIGN KEY (food_id) REFERENCES foods(food_id) ON DELETE CASCADE
    );`

	if _, err := d.db.Exec(basketOrderTable); err != nil {
		log.Printf("Basket_orders jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'basket_orders' jadvali mavjud yoki yaratildi.")

	// Tables table
	tableTable := `
    CREATE TABLE IF NOT EXISTS tables (
        table_id INTEGER PRIMARY KEY AUTOINCREMENT,
        table_name TEXT UNIQUE NOT NULL,
        qr_code_token TEXT UNIQUE NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	if _, err := d.db.Exec(tableTable); err != nil {
		log.Printf("Tables jadvalini yaratishda xatolik: %v", err)
		return err
	}
	log.Println("✅ 'tables' jadvali mavjud yoki yaratildi.")

	// `orders` jadvali uchun ustunlarni qo'shish yoki nomini o'zgartirish
	// `order_status` ustunini qo'shish yoki nomini o'zgartirish logikasi
	// Avval 'status' ustuni borligini tekshiramiz va 'order_status' yo'qligini tekshiramiz
	if d.columnExists("orders", "status") && !d.columnExists("orders", "order_status") {
		renameQuery := "ALTER TABLE orders RENAME COLUMN status TO order_status;"
		_, err := d.db.Exec(renameQuery)
		if err != nil {
			log.Printf("Orders jadvalidagi 'status' ustunini 'order_status' ga o'zgartirishda xatolik: %v", err)
			// Agar xatolik jiddiy bo'lsa, bu yerda return err; qilinishi mumkin.
		} else {
			log.Println("✅ 'orders' jadvalidagi 'status' ustuni 'order_status' ga o'zgartirildi.")
		}
	}

	// Orders jadvaliga yangi ustunlar qo'shish.
	ordersColumnsToAdd := map[string]string{
		"order_status":       "TEXT NOT NULL DEFAULT 'pending'",
		"order_time":         "DATETIME DEFAULT CURRENT_TIMESTAMP",
		"delivery_type":      "TEXT",
		"delivery_latitude":  "REAL",
		"delivery_longitude": "REAL",
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
			log.Printf("ℹ️  'orders' jadvalida '%s' ustuni allaqachon mavjud.", colName)
		}
	}

	// `order_items` jadvali uchun ustunlarni qo'shish yoki nomini o'zgartirish
	// Agar "price_at_order" ustuni mavjud bo'lsa va "item_price" ustuni yo'q bo'lsa, nomini o'zgartiramiz.
	if d.columnExists("order_items", "price_at_order") && !d.columnExists("order_items", "item_price") {
		renameQuery := "ALTER TABLE order_items RENAME COLUMN price_at_order TO item_price;"
		_, err := d.db.Exec(renameQuery)
		if err != nil {
			log.Printf("Order_items jadvalidagi 'price_at_order' ustunini 'item_price' ga o'zgartirishda xatolik: %v", err)
			// Agar xatolik jiddiy bo'lsa, bu yerda return err; qilinishi mumkin.
		} else {
			log.Println("✅ 'order_items' jadvalidagi 'price_at_order' ustuni 'item_price' ga o'zgartirildi.")
		}
	}

	// order_items jadvaliga yangi ustunlar qo'shish (asosan, agar kelajakda yana bo'lsa)
	// Hozirda faqat item_price mavjud va yuqorida uni yaratish yoki nomini o'zgartirish bilan hal qildik.
	// Agar yana qo'shimcha ustunlar kerak bo'lsa, shu yerga qo'shiladi.
	orderItemsColumnsToAdd := map[string]string{
		"item_price": "REAL NOT NULL", // Faqat ushbu misolda item_price ni alohida qo'shish.
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
			log.Printf("ℹ️  'order_items' jadvalida '%s' ustuni allaqachon mavjud.", colName)
		}
	}

	return nil
}

// columnExists jadvalda ustun borligini tekshiradi.
func (d *Database) columnExists(tableName, columnName string) bool {
	query := fmt.Sprintf("PRAGMA table_info(%s);", tableName)
	rows, err := d.db.Query(query)
	if err != nil {
		log.Printf("Ustun mavjudligini tekshirishda xatolik: %v", err)
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			log.Printf("Ustun ma'lumotlarini skanlashda xatolik: %v", err)
			continue
		}
		if strings.EqualFold(name, columnName) {
			return true
		}
	}
	return false
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
