package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dataSourceName string) (*Database, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Printf("✅ Ma'lumotlar bazasiga ulanish muvaffaqiyatli: %s", dataSourceName)

	database := &Database{db: db}
	database.InitDB() // Jadval yaratishni chaqirish
	return database, nil
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}

func (d *Database) Close() {
	if d.db != nil {
		d.db.Close()
		log.Println("🔌 Ma'lumotlar bazasi aloqasi uzildi.")
	}
}

// InitDB barcha kerakli jadvallarni yaratadi
func (d *Database) InitDB() {
	// users jadvali
	usersTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tg_id INTEGER UNIQUE NOT NULL,
		first_name TEXT,
		username TEXT,
		language_code TEXT,
		phone_number TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// foods jadvali
	foodsTableQuery := `
	CREATE TABLE IF NOT EXISTS foods (
		food_id INTEGER PRIMARY KEY AUTOINCREMENT,
		food_name TEXT NOT NULL,
		food_category TEXT NOT NULL,
		food_price REAL NOT NULL,
		food_image TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// cart_items jadvali (Yangi)
	cartItemsTableQuery := `
	CREATE TABLE IF NOT EXISTS cart_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		food_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, food_id), -- Bir foydalanuvchi bir xil mahsulotni faqat bir marta qo'sha oladi (miqdorini yangilaydi)
		FOREIGN KEY (user_id) REFERENCES users(tg_id) ON DELETE CASCADE,
		FOREIGN KEY (food_id) REFERENCES foods(food_id) ON DELETE CASCADE
	);`

	_, err := d.db.Exec(usersTableQuery)
	if err != nil {
		log.Fatalf("Users jadvalini yaratishda xatolik: %v", err)
	}
	log.Println("✅ Users jadvali tekshirildi/yaratildi.")

	_, err = d.db.Exec(foodsTableQuery)
	if err != nil {
		log.Fatalf("Foods jadvalini yaratishda xatolik: %v", err)
	}
	log.Println("✅ Foods jadvali tekshirildi/yaratildi.")

	_, err = d.db.Exec(cartItemsTableQuery)
	if err != nil {
		log.Fatalf("Cart_items jadvalini yaratishda xatolik: %v", err)
	}
	log.Println("✅ Cart_items jadvali tekshirildi/yaratildi.")
}
