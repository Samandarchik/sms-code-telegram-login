package handlers

import (
	"amur/models"
	"amur/service"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// sendErrorResponse yordamchi funksiyasi
func (h *UserHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   message,
		"details": details,
	})
	log.Printf("Xato javobi yuborildi: Status=%d, Xabar='%s', Tafsilotlar='%s'", statusCode, message, details)
}

// sendSuccessResponse yordamchi funksiyasi
func (h *UserHandler) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
		"data":    data,
	})
	log.Printf("Muvaffaqiyatli javob yuborildi: Xabar='%s'", message)
}

// RegisterUser Telegram ID orqali foydalanuvchini ro'yxatdan o'tkazadi yoki mavjud bo'lsa qaytaradi.
// POST /api/register
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest // Bu hozirda JSON so'rovini qabul qilish uchun kerak
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "So'rov tanasini tahlil qilishda xatolik", err.Error())
		return
	}

	if req.UserID == 0 {
		h.sendErrorResponse(w, http.StatusBadRequest, "Telegram ID majburiy", "Telegram ID (user_id) nol bo'lmasligi kerak.")
		return
	}

	// models.CreateUserRequest dan models.User obyektini yaratamiz
	userToRegister := &models.User{
		TelegramID: req.UserID,
		Username:   req.Username,
		FirstName:  req.FirstName,
		// Role va boshqa maydonlar service qatlamida o'rnatilishi yoki repositoryda default qiymatlar bo'lishi mumkin.
		// Hozircha Role 'user' deb Service qatlamida o'rnatilgan.
		// LanguageCode - agar siz uni ham Requestda qabul qilmoqchi bo'lsangiz, uni ham qo'shishingiz mumkin.
	}

	// CreateUser o'rniga SaveOrUpdateUser dan foydalanish yaxshiroq,
	// chunki u foydalanuvchining mavjudligini tekshiradi va shunga qarab yaratadi yoki yangilaydi.
	// user, err := h.userService.CreateUser(userToRegister) // OLD: h.userService.CreateUser(&req)
	registeredUser, err := h.userService.SaveOrUpdateUser(userToRegister) // <-- Shu qatorni o'zgartirdik
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Foydalanuvchini ro'yxatdan o'tkazishda/yangilashda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Foydalanuvchi muvaffaqiyatli ro'yxatdan o'tkazildi/topildi", registeredUser)
}

// GetAllUsers barcha foydalanuvchilarni qaytaradi (faqat adminlar uchun bo'lishi kerak)
// GET /api/users
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers() // Bu endi []*models.User qaytaradi
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Foydalanuvchilarni olishda xatolik", err.Error())
		return
	}
	h.sendSuccessResponse(w, "Foydalanuvchilar muvaffaqiyatli olindi", users)
}

// GetUserStats foydalanuvchilar sonini qaytaradi (admin funksiyasi)
// GET /api/users/stats
func (h *UserHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	count, err := h.userService.GetUserStats()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Foydalanuvchi statistikasini olishda xatolik", err.Error())
		return
	}
	h.sendSuccessResponse(w, "Foydalanuvchilar statistikasi muvaffaqiyatli olindi", map[string]int{"total_users": count})
}

// Login foydalanuvchini tizimga kiritadi va JWT tokenini qaytaradi
// POST /api/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "So'rov tanasini tahlil qilishda xatolik", err.Error())
		return
	}

	if req.TelegramID == 0 {
		h.sendErrorResponse(w, http.StatusBadRequest, "Telegram ID majburiy", "Telegram ID (telegram_id) nol bo'lmasligi kerak.")
		return
	}

	token, err := h.userService.Login(req.TelegramID, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Agar foydalanuvchi topilmasa
			h.sendErrorResponse(w, http.StatusUnauthorized, "Autentifikatsiya muvaffaqiyatsiz", "Foydalanuvchi topilmadi. Iltimos, avval ro'yxatdan o'ting.")
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, "Tizimga kirishda xatolik", err.Error())
		}
		return
	}

	h.sendSuccessResponse(w, "Muvaffaqiyatli tizimga kirish", models.LoginResponse{Token: token})
}
