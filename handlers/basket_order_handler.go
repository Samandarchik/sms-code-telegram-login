// handlers/basket_order_handler.go
package handlers

import (
	"amur/models"
	"amur/service"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type BasketOrderHandler struct {
	basketService *service.BasketOrderService
}

func NewBasketOrderHandler(basketService *service.BasketOrderService) *BasketOrderHandler {
	return &BasketOrderHandler{basketService: basketService}
}

// sendErrorResponse yordamchi funksiyasi xato javobini yuborish uchun
func (h *BasketOrderHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   message,
		"details": details,
	})
	log.Printf("Xato javobi yuborildi: Status=%d, Xabar='%s', Tafsilotlar='%s'", statusCode, message, details)
}

// sendSuccessResponse yordamchi funksiyasi muvaffaqiyatli javobni yuborish uchun
func (h *BasketOrderHandler) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
		"data":    data,
	})
	log.Printf("Muvaffaqiyatli javob yuborildi: Xabar='%s'", message)
}

// AddToBasket savatchaga mahsulot qo'shish
// POST /api/{telegramID}/basket-order
func (h *BasketOrderHandler) AddToBasket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	telegramIDStr := vars["telegramID"]
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri Telegram ID", err.Error())
		return
	}

	var req models.AddToBasketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "So'rov tanasini tahlil qilishda xatolik", err.Error())
		return
	}

	if req.FoodID <= 0 {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri Food ID", "Food ID musbat son bo'lishi kerak.")
		return
	}

	order, err := h.basketService.AddToBasket(telegramID, &req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Savatchaga mahsulot qo'shishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Mahsulot savatchaga muvaffaqiyatli qo'shildi/yangilandi", order)
}

// GetBasketOrders savatchadagi barcha mahsulotlarni olish
// GET /api/{telegramID}/basket-order
func (h *BasketOrderHandler) GetBasketOrders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	telegramIDStr := vars["telegramID"]
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri Telegram ID", err.Error())
		return
	}

	orders, err := h.basketService.GetBasketOrders(telegramID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Savatchani olishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Savatcha muvaffaqiyatli olindi", orders)
}

// RemoveFromBasket savatchadan mahsulotni o'chirish
// DELETE /api/{telegramID}/basket-order/{foodID}
func (h *BasketOrderHandler) RemoveFromBasket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	telegramIDStr := vars["telegramID"]
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri Telegram ID", err.Error())
		return
	}

	foodIDStr := vars["foodID"]
	foodID, err := strconv.Atoi(foodIDStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri Food ID", err.Error())
		return
	}

	err = h.basketService.RemoveFromBasket(telegramID, foodID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Savatchadan mahsulotni o'chirishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Mahsulot savatchadan muvaffaqiyatli o'chirildi", nil)
}

// ClearBasket savatchani tozalash
// DELETE /api/{telegramID}/basket-order
func (h *BasketOrderHandler) ClearBasket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	telegramIDStr := vars["telegramID"]
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri Telegram ID", err.Error())
		return
	}

	err = h.basketService.ClearBasket(telegramID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Savatchani tozalashda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Savatcha muvaffaqiyatli tozalandi", nil)
}
