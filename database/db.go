// ==================== database/db.go ====================
package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) createTables() error {
	// Users table
	userTable := `
    CREATE TABLE IF NOT EXISTS users (
        userid INTEGER PRIMARY KEY AUTOINCREMENT,
        tg_id INTEGER UNIQUE NOT NULL,
        first_name TEXT NOT NULL DEFAULT 'N/A',
        username TEXT NOT NULL DEFAULT 'N/A',
        language_code TEXT NOT NULL DEFAULT 'uz',
        phone TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	if _, err := d.db.Exec(userTable); err != nil {
		log.Printf("Users jadval yaratishda xatolik: %v", err)
		return err
	}

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
		log.Printf("Foods jadval yaratishda xatolik: %v", err)
		return err
	}

	return nil
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}

func (d *Database) Close() error {
	return d.db.Close()
}
