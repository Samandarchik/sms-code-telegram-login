// ==================== repository/user_repository.go ====================
package repository

import (
	"amur/models"
	"database/sql"
	"log"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	stmt, err := r.db.Prepare(`
        INSERT INTO users(telegram_id, first_name, username, language_code, phone)
        VALUES (?, ?, ?, ?, ?)
    `)
	if err != nil {
		log.Printf("Create prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.TelegramID, user.FirstName, user.Username, user.LanguageCode, user.PhoneNumber)
	if err != nil {
		log.Printf("Create exec xatolik: %v", err)
		return err
	}

	log.Printf("✅ Yangi foydalanuvchi saqlandi: %s (ID: %d)", user.FirstName, user.TelegramID)
	return nil
}

func (r *UserRepository) Update(user *models.User) error {
	stmt, err := r.db.Prepare(`
        UPDATE users SET 
            first_name = ?, 
            username = ?, 
            language_code = ?, 
            phone = ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE telegram_id = ?
    `)
	if err != nil {
		log.Printf("Update prepare xatolik: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.FirstName, user.Username, user.LanguageCode, user.PhoneNumber, user.TelegramID)
	if err != nil {
		log.Printf("Update exec xatolik: %v", err)
		return err
	}

	log.Printf("🔄 Foydalanuvchi yangilandi: %s (ID: %d)", user.FirstName, user.TelegramID)
	return nil
}

func (r *UserRepository) GetByTgID(tgID int64) (*models.User, error) {
	row := r.db.QueryRow(`
        SELECT userid, telegram_id, first_name, username, language_code, phone, created_at, updated_at
        FROM users WHERE telegram_id = ?
    `, tgID)

	var user models.User
	err := row.Scan(&user.UserID, &user.TelegramID, &user.FirstName, &user.Username,
		&user.LanguageCode, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Exists(tgID int64) bool {
	row := r.db.QueryRow("SELECT telegram_id FROM users WHERE telegram_id = ?", tgID)
	var id int64
	err := row.Scan(&id)
	return err == nil
}

func (r *UserRepository) Count() int {
	row := r.db.QueryRow("SELECT COUNT(*) FROM users")
	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Printf("Count xatolik: %v", err)
		return 0
	}
	return count
}

func (r *UserRepository) GetAll() ([]*models.User, error) {
	rows, err := r.db.Query(`
        SELECT userid, telegram_id, first_name, username, language_code, phone, created_at, updated_at
        FROM users ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.UserID, &user.TelegramID, &user.FirstName, &user.Username,
			&user.LanguageCode, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("GetAll scan xatolik: %v", err)
			continue
		}
		users = append(users, &user)
	}

	return users, nil
}
