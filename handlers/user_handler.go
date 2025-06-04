package handlers

import (
	"amur/models"
	"amur/service"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	// "strconv" // Endi kerak emas
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
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "So'rov tanasini tahlil qilishda xatolik", err.Error())
		return
	}

	if req.UserID == 0 {
		h.sendErrorResponse(w, http.StatusBadRequest, "Telegram ID majburiy", "Telegram ID (user_id) nol bo'lmasligi kerak.")
		return
	}

	userToRegister := &models.User{
		TelegramID: req.UserID,
		Username:   req.Username,
		FirstName:  req.FirstName,
		// PhoneNumber maydoni CreateUserRequestda yo'q, agar kerak bo'lsa uni ham qo'shish kerak.
		// Hozirda Phone kontakt yuborishda bot orqali o'rnatiladi.
	}

	registeredUser, err := h.userService.SaveOrUpdateUser(userToRegister)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Foydalanuvchini ro'yxatdan o'tkazishda/yangilashda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Foydalanuvchi muvaffaqiyatli ro'yxatdan o'tkazildi/topildi", registeredUser)
}

// GetAllUsers barcha foydalanuvchilarni qaytaradi (faqat adminlar uchun bo'lishi kerak)
// GET /api/users
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
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

// --- Yangilangan Login Handler funksiyasi ---

// Login foydalanuvchini telefon raqami va kod orqali tizimga kiritadi
// POST /api/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest // Yangi LoginRequestni ishlatamiz
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "So'rov tanasini tahlil qilishda xatolik", err.Error())
		return
	}

	if req.PhoneNumber == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Telefon raqami majburiy", "Telefon raqami bo'sh bo'lmasligi kerak.")
		return
	}
	if req.Code == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Kod majburiy", "Kod bo'sh bo'lmasligi kerak.")
		return
	}
	if len(req.Code) != 4 { // Kod uzunligini tekshirish
		h.sendErrorResponse(w, http.StatusBadRequest, "Kod noto'g'ri", "Kod 4 raqamdan iborat bo'lishi kerak.")
		return
	}

	// UserService dagi yangi Login funksiyasini chaqiramiz
	token, err := h.userService.Login(req.PhoneNumber, req.Code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Agar foydalanuvchi topilmasa
			h.sendErrorResponse(w, http.StatusUnauthorized, "Autentifikatsiya muvaffaqiyatsiz", "Foydalanuvchi topilmadi. Iltimos, avval ro'yxatdan o'ting.")
		} else if err.Error() == "kod noto'g'ri, iltimos, tekshirib qaytadan urinib ko'ring" {
			h.sendErrorResponse(w, http.StatusUnauthorized, "Kod noto'g'ri", err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, "Tizimga kirishda xatolik", err.Error())
		}
		return
	}

	h.sendSuccessResponse(w, "Muvaffaqiyatli tizimga kirish", models.LoginResponse{Token: token})
}
