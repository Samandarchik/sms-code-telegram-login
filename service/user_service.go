package service

import (
	"amur/models"
	"amur/pkg/jwt_auth"
	"amur/repository"
	"database/sql"
	"errors"
	"fmt"
	"log"
	// string to int64 conversion for code check
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// CreateUser yangi foydalanuvchi yaratadi yoki agar mavjud bo'lsa, uni qaytaradi
func (s *UserService) CreateUser(user *models.User) (*models.User, error) {
	existingUser, err := s.userRepo.GetByTgID(user.TelegramID)
	if err == nil {
		log.Printf("Foydalanuvchi allaqachon mavjud, ID: %d", user.TelegramID)
		return existingUser, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("foydalanuvchini tekshirishda xatolik: %w", err)
	}

	user.Role = "user"
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("foydalanuvchini yaratishda xatolik: %w", err)
	}

	log.Printf("Yangi foydalanuvchi yaratildi, ID: %d", user.TelegramID)
	return user, nil
}

// GetAllUsers barcha foydalanuvchilarni qaytaradi (admin funksiyasi)
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	users, err := s.userRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("foydalanuvchilarni olishda xatolik: %w", err)
	}
	return users, nil
}

// GetUserByID ID bo'yicha foydalanuvchini qaytaradi
func (s *UserService) GetUserByID(userID int64) (*models.User, error) {
	user, err := s.userRepo.GetByTgID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("foydalanuvchi topilmadi: %w", err)
		}
		return nil, fmt.Errorf("foydalanuvchini olishda xatolik: %w", err)
	}
	return user, nil
}

// GetUserStats foydalanuvchilar sonini qaytaradi
func (s *UserService) GetUserStats() (int, error) {
	count := s.userRepo.Count()
	return count, nil
}

// --- Yangilangan Login funksiyasi ---

// Login foydalanuvchini telefon raqami va kod orqali tizimga kiritadi
func (s *UserService) Login(phoneNumber, code string) (string, error) {
	// 1. Telefon raqami bo'yicha foydalanuvchini topish
	user, err := s.userRepo.GetByPhoneNumber(phoneNumber) // repository/user_repository.go da GetByPhoneNumber ni yaratish kerak!
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("telefon raqamiga ega foydalanuvchi topilmadi: %w", err)
		}
		return "", fmt.Errorf("foydalanuvchini telefon raqami bo'yicha tekshirishda xatolik: %w", err)
	}

	// 2. Foydalanuvchining Telegram ID'sidan kodni generatsiya qilish
	generatedCode := s.GenerateUserCode(user.TelegramID)

	// 3. Foydalanuvchi kiritgan kodni tekshirish
	if generatedCode != code {
		return "", errors.New("kod noto'g'ri, iltimos, tekshirib qaytadan urinib ko'ring")
	}

	// 4. Tekshiruvdan o'tdi, endi JWT tokenini yaratamiz
	token, err := jwt_auth.GenerateToken(user.TelegramID, user.Role)
	if err != nil {
		return "", fmt.Errorf("token yaratishda xatolik: %w", err)
	}

	log.Printf("Foydalanuvchi tizimga kirdi, Telegram ID: %d", user.TelegramID)
	return token, nil
}

// validateAndCleanUser foydalanuvchi ma'lumotlarini tozalaydi
func (s *UserService) validateAndCleanUser(user *models.User) {
	if user.FirstName == "" {
		user.FirstName = "N/A"
	}
	if user.Username == "" {
		user.Username = "N/A"
	}
	if user.LanguageCode == "" {
		user.LanguageCode = "uz"
	}
}

// SaveOrUpdateUser funksiyasi CreateUser va UpdateUser ni o'z ichiga oladi
func (s *UserService) SaveOrUpdateUser(user *models.User) (*models.User, error) {
	s.validateAndCleanUser(user)

	if s.userRepo.Exists(user.TelegramID) {
		if err := s.userRepo.Update(user); err != nil {
			return nil, fmt.Errorf("foydalanuvchini yangilashda xatolik: %w", err)
		}
		log.Printf("Foydalanuvchi yangilandi, ID: %d", user.TelegramID)
		return user, nil
	} else {
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("foydalanuvchini yaratishda xatolik: %w", err)
		}
		log.Printf("Yangi foydalanuvchi yaratildi, ID: %d", user.TelegramID)
		return user, nil
	}
}

// GenerateUserCode foydalanuvchi uchun kod yaratadi (service qatlamida qoldiriladi)
func (s *UserService) GenerateUserCode(tgID int64) string {
	tgIDStr := fmt.Sprintf("%d", tgID)
	if len(tgIDStr) >= 4 {
		return tgIDStr[len(tgIDStr)-4:]
	}
	return "0000" // Agar Telegram ID 4 raqamdan kam bo'lsa default
}

// GetUserCount foydalanuvchilar sonini qaytaradi
func (s *UserService) GetUserCount() (int, error) {
	count := s.userRepo.Count()
	return count, nil
}

// UserExists foydalanuvchi mavjudligini tekshiradi
func (s *UserService) UserExists(tgID int64) bool {
	return s.userRepo.Exists(tgID)
}
