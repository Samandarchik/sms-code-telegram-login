package service

import (
	"amur/models"
	"amur/pkg/jwt_auth" // JWT yaratish uchun paket
	"amur/repository"   // Yangi repository paketini import qilish
	"database/sql"      // Faqat GetUserByIDda sql.ErrNoRows uchun kerak
	"errors"
	"fmt"
	"log"
)

type UserService struct {
	userRepo *repository.UserRepository // Endi *sql.DB o'rniga UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// CreateUser yangi foydalanuvchi yaratadi yoki agar mavjud bo'lsa, uni qaytaradi
// Bu funksiya models.CreateUserRequest o'rniga to'g'ridan-to'g'ri models.User qabul qiladi
func (s *UserService) CreateUser(user *models.User) (*models.User, error) {
	// Avval foydalanuvchi mavjudligini tekshiramiz
	existingUser, err := s.userRepo.GetByTgID(user.TelegramID)
	if err == nil {
		log.Printf("Foydalanuvchi allaqachon mavjud, ID: %d", user.TelegramID)
		return existingUser, nil // Foydalanuvchi allaqachon mavjud bo'lsa, uni qaytaramiz
	}
	if !errors.Is(err, sql.ErrNoRows) {
		// Boshqa turdagi xato bo'lsa, uni qaytaramiz
		return nil, fmt.Errorf("foydalanuvchini tekshirishda xatolik: %w", err)
	}

	// Foydalanuvchi mavjud emas, yangisini yaratamiz
	// Default rol 'user' qilib belgilash
	user.Role = "user"
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("foydalanuvchini yaratishda xatolik: %w", err)
	}

	log.Printf("Yangi foydalanuvchi yaratildi, ID: %d", user.TelegramID)
	return user, nil
}

// GetAllUsers barcha foydalanuvchilarni qaytaradi (admin funksiyasi)
func (s *UserService) GetAllUsers() ([]*models.User, error) { // []*models.User qaytaradi
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
	count := s.userRepo.Count() // Repositorydagi Count funksiyasi int qaytaradi
	return count, nil
}

// --- Yangi Login funksiyasi ---

// Login foydalanuvchini Telegram ID orqali tizimga kiritadi va JWT tokenini qaytaradi.
// Agar foydalanuvchi topilmasa, xato qaytaradi.
func (s *UserService) Login(telegramID int64, username string) (string, error) {
	user, err := s.userRepo.GetByTgID(telegramID) // userRepo.GetByTgID dan foydalanamiz
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("foydalanuvchi topilmadi, iltimos ro'yxatdan o'ting: %w", err)
		}
		return "", fmt.Errorf("foydalanuvchini tekshirishda xatolik: %w", err)
	}

	// Agar username berilgan bo'lsa, uni ham tekshirish mumkin (ixtiyoriy)
	// Bu yerda `models.User` dagi `Username` maydoni bo'sh bo'lsa,
	// `validateAndCleanUser` uni "N/A" ga o'zgartirishi mumkinligini hisobga oling.
	// Shuning uchun bu tekshiruvni o'chirib qo'yish afzalroq yoki yanada murakkabroq tekshiruv qo'yish kerak.
	// if username != "" && user.Username != username {
	//     return "", errors.New("noto'g'ri foydalanuvchi nomi")
	// }

	// Foydalanuvchi topildi, endi JWT tokenini yaratamiz
	// user.TelegramID user.UserID o'rniga ishlatiladi, chunki models.User endi TelegramID maydoniga ega.
	token, err := jwt_auth.GenerateToken(user.TelegramID, user.Role)
	if err != nil {
		return "", fmt.Errorf("token yaratishda xatolik: %w", err)
	}

	log.Printf("Foydalanuvchi tizimga kirdi, ID: %d", user.TelegramID)
	return token, nil
}

// service/user_service.go fayli ichida

// ... (yuqoridagi mavjud kod) ...

// SaveOrUpdateUser funksiyasi CreateUser va UpdateUser ni o'z ichiga oladi
// Bu funksiyani RegisterUser handlerida ishlatish oqilona bo'ladi.
func (s *UserService) SaveOrUpdateUser(user *models.User) (*models.User, error) { // <-- Endi *models.User ham qaytaradi
	s.validateAndCleanUser(user) // Ma'lumotlarni saqlashdan oldin tozalash

	if s.userRepo.Exists(user.TelegramID) {
		if err := s.userRepo.Update(user); err != nil {
			return nil, fmt.Errorf("foydalanuvchini yangilashda xatolik: %w", err)
		}
		log.Printf("Foydalanuvchi yangilandi, ID: %d", user.TelegramID)
		return user, nil // Yangilangan foydalanuvchini qaytarish
	} else {
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("foydalanuvchini yaratishda xatolik: %w", err)
		}
		log.Printf("Yangi foydalanuvchi yaratildi, ID: %d", user.TelegramID)
		return user, nil // Yaratilgan foydalanuvchini qaytarish
	}
}

// GenerateUserCode foydalanuvchi uchun kod yaratadi
func (s *UserService) GenerateUserCode(tgID int64) string {
	tgIDStr := fmt.Sprintf("%d", tgID)
	if len(tgIDStr) >= 4 {
		return tgIDStr[len(tgIDStr)-4:]
	}
	return "0000"
}

// GetUserCount foydalanuvchilar sonini qaytaradi
func (s *UserService) GetUserCount() (int, error) { // <-- error qaytaradi
	count := s.userRepo.Count()
	// Agar repository Count() dan xato qaytarish imkoniyati bo'lsa, uni shu yerda tekshiring.
	// Hozircha repositorydagi Count int qaytargani uchun error bo'lmaydi.
	return count, nil
}

// UserExists foydalanuvchi mavjudligini tekshiradi
func (s *UserService) UserExists(tgID int64) bool {
	return s.userRepo.Exists(tgID)
}

// validateAndCleanUser foydalanuvchi ma'lumotlarini tozalaydi
// Bu funksiyani CreateUser va UpdateUser ichidan chaqirish kerak
func (s *UserService) validateAndCleanUser(user *models.User) {
	if user.FirstName == "" {
		user.FirstName = "N/A"
	}
	if user.Username == "" {
		user.Username = "N/A"
	}
	if user.LanguageCode == "" {
		user.LanguageCode = "uz" // Default til
	}
}
