// ==================== handlers/user_handler.go ====================
package handlers

import (
	"amur/models"
	"amur/service"
	"encoding/json"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GET /api/users - Barcha foydalanuvchilarni olish
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Foydalanuvchilarni olishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Foydalanuvchilar muvaffaqiyatli olindi", users)
}

// GET /api/users/stats - Foydalanuvchilar statistikasi
func (h *UserHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	count := h.userService.GetUserCount()
	stats := map[string]interface{}{
		"total_users": count,
	}

	h.sendSuccessResponse(w, "Statistika muvaffaqiyatli olindi", stats)
}

func (h *UserHandler) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string, error string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.APIResponse{
		Success: false,
		Message: message,
		Error:   error,
	}

	json.NewEncoder(w).Encode(response)
}
