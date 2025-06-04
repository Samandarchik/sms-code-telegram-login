package repository

import (
	"amur/models"
	"database/sql"
	"errors" // errors paketini import qilish
	"fmt"
	"log"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	// 'role' maydonini ham INSERT so'roviga qo'shamiz
	query := `
        INSERT INTO users(telegram_id, first_name, username, language_code, phone, role)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING userid, created_at, updated_at
    `
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Printf("UserRepository.Create: prepare statement xatolik: %v", err)
		return fmt.Errorf("foydalanuvchi yaratish uchun SQL so'rovini tayyorlashda xatolik: %w", err)
	}
	defer stmt.Close()

	// user.UserID va vaqt maydonlarini RETURNING qismi orqali to'ldiramiz
	err = stmt.QueryRow(user.TelegramID, user.FirstName, user.Username, user.LanguageCode, user.PhoneNumber, user.Role).
		Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Printf("UserRepository.Create: exec statement xatolik: %v", err)
		return fmt.Errorf("foydalanuvchini bazaga saqlashda xatolik: %w", err)
	}

	log.Printf("✅ Yangi foydalanuvchi saqlandi: %s (ID: %d)", user.FirstName, user.TelegramID)
	return nil
}

func (r *UserRepository) Update(user *models.User) error {
	// 'role' maydonini ham UPDATE so'roviga qo'shamiz
	query := `
        UPDATE users SET
            first_name = $1,
            username = $2,
            language_code = $3,
            phone = $4,
            role = $5,
            updated_at = CURRENT_TIMESTAMP
        WHERE telegram_id = $6
        RETURNING updated_at
    `
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Printf("UserRepository.Update: prepare statement xatolik: %v", err)
		return fmt.Errorf("foydalanuvchini yangilash uchun SQL so'rovini tayyorlashda xatolik: %w", err)
	}
	defer stmt.Close()

	// updated_at ni qaytarib olamiz
	err = stmt.QueryRow(user.FirstName, user.Username, user.LanguageCode, user.PhoneNumber, user.Role, user.TelegramID).
		Scan(&user.UpdatedAt)
	if err != nil {
		log.Printf("UserRepository.Update: exec statement xatolik: %v", err)
		return fmt.Errorf("foydalanuvchini bazada yangilashda xatolik: %w", err)
	}

	log.Printf("🔄 Foydalanuvchi yangilandi: %s (ID: %d)", user.FirstName, user.TelegramID)
	return nil
}

func (r *UserRepository) GetByTgID(tgID int64) (*models.User, error) {
	query := `
        SELECT userid, telegram_id, first_name, username, language_code, phone, role, created_at, updated_at
        FROM users WHERE telegram_id = $1
    `
	row := r.db.QueryRow(query, tgID)

	var user models.User
	err := row.Scan(&user.UserID, &user.TelegramID, &user.FirstName, &user.Username,
		&user.LanguageCode, &user.PhoneNumber, &user.Role, &user.CreatedAt, &user.UpdatedAt) // 'role' ni ham scan qilamiz
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("foydalanuvchi topilmadi (Telegram ID: %d): %w", tgID, err)
		}
		log.Printf("UserRepository.GetByTgID: scan xatolik (Telegram ID: %d): %v", tgID, err)
		return nil, fmt.Errorf("foydalanuvchini Telegram ID bo'yicha olishda xatolik: %w", err)
	}

	return &user, nil
}

// --- YANGI: GetByPhoneNumber funksiyasi ---
func (r *UserRepository) GetByPhoneNumber(phoneNumber string) (*models.User, error) {
	query := `
        SELECT userid, telegram_id, first_name, username, language_code, phone, role, created_at, updated_at
        FROM users WHERE phone = $1
    `
	row := r.db.QueryRow(query, phoneNumber)

	var user models.User
	err := row.Scan(&user.UserID, &user.TelegramID, &user.FirstName, &user.Username,
		&user.LanguageCode, &user.PhoneNumber, &user.Role, &user.CreatedAt, &user.UpdatedAt) // 'role' ni ham scan qilamiz
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("telefon raqamiga ega foydalanuvchi topilmadi (%s): %w", phoneNumber, err)
		}
		log.Printf("UserRepository.GetByPhoneNumber: scan xatolik (telefon raqami: %s): %v", phoneNumber, err)
		return nil, fmt.Errorf("foydalanuvchini telefon raqami bo'yicha olishda xatolik: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) Exists(tgID int64) bool {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE telegram_id = $1)` // 'telegram_id' ni ishlatamiz
	var exists bool
	err := r.db.QueryRow(query, tgID).Scan(&exists)
	if err != nil {
		log.Printf("UserRepository.Exists: foydalanuvchi mavjudligini tekshirishda xatolik: %v", err)
		return false
	}
	return exists
}

func (r *UserRepository) Count() int {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Printf("UserRepository.Count: foydalanuvchilar sonini olishda xatolik: %v", err)
		return 0
	}
	return count
}

func (r *UserRepository) GetAll() ([]*models.User, error) {
	query := `
        SELECT userid, telegram_id, first_name, username, language_code, phone, role, created_at, updated_at
        FROM users ORDER BY created_at DESC
    `
	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("UserRepository.GetAll: query xatolik: %v", err)
		return nil, fmt.Errorf("barcha foydalanuvchilarni olishda xatolik: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.UserID, &user.TelegramID, &user.FirstName, &user.Username,
			&user.LanguageCode, &user.PhoneNumber, &user.Role, &user.CreatedAt, &user.UpdatedAt) // 'role' ni ham scan qilamiz
		if err != nil {
			log.Printf("UserRepository.GetAll: scan xatolik: %v", err)
			return nil, fmt.Errorf("foydalanuvchini qatoridan o'qishda xatolik: %w", err) // Xatolikni qaytaramiz
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("UserRepository.GetAll: rows iterator xatolik: %v", err)
		return nil, fmt.Errorf("foydalanuvchilarni olish paytida iteratsiyada xatolik: %w", err)
	}

	return users, nil
}
