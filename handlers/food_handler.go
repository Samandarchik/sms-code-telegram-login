// ==================== handlers/food_handler.go ====================
package handlers

import (
	"amur/models"
	"amur/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type FoodHandler struct {
	foodService *service.FoodService
}

func NewFoodHandler(foodService *service.FoodService) *FoodHandler {
	return &FoodHandler{foodService: foodService}
}

// GET /api/foods - Barcha ovqatlarni olish
func (h *FoodHandler) GetAllFoods(w http.ResponseWriter, r *http.Request) {
	foods, err := h.foodService.GetAllFoods()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Ovqatlarni olishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Ovqatlar muvaffaqiyatli olindi", foods)
}

// GET /api/foods/{id} - Bitta ovqatni olish
func (h *FoodHandler) GetFoodByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri ID", err.Error())
		return
	}

	food, err := h.foodService.GetFoodByID(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Ovqat topilmadi", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Ovqat muvaffaqiyatli topildi", food)
}

// POST /api/foods - Yangi ovqat qo'shish
func (h *FoodHandler) CreateFood(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "JSON formatida xatolik", err.Error())
		return
	}

	food, err := h.foodService.CreateFood(&req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Ovqat qo'shishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Ovqat muvaffaqiyatli qo'shildi", food)
}

// PUT /api/foods/{id} - Ovqatni yangilash
func (h *FoodHandler) UpdateFood(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri ID", err.Error())
		return
	}

	var req models.UpdateFoodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "JSON formatida xatolik", err.Error())
		return
	}

	food, err := h.foodService.UpdateFood(id, &req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Ovqatni yangilashda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Ovqat muvaffaqiyatli yangilandi", food)
}

// DELETE /api/foods/{id} - Ovqatni o'chirish
func (h *FoodHandler) DeleteFood(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri ID", err.Error())
		return
	}

	err = h.foodService.DeleteFood(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Ovqatni o'chirishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Ovqat muvaffaqiyatli o'chirildi", nil)
}
func (h *FoodHandler) GetFoodsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]

	foods, err := h.foodService.GetFoodsByCategory(category)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Kategoriya bo'yicha ovqatlarni olishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Kategoriya bo'yicha ovqatlar muvaffaqiyatli olindi", foods)
}

// GET /api/foods/stats - Ovqatlar statistikasi
func (h *FoodHandler) GetFoodStats(w http.ResponseWriter, r *http.Request) {
	count := h.foodService.GetFoodCount()
	stats := map[string]interface{}{
		"total_foods": count,
	}

	h.sendSuccessResponse(w, "Statistika muvaffaqiyatli olindi", stats)
}

// Helper functions
func (h *FoodHandler) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *FoodHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string, error string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.APIResponse{
		Success: false,
		Message: message,
		Error:   error,
	}

	json.NewEncoder(w).Encode(response)
}
