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

// CartHandler savat bilan bog'liq HTTP so'rovlarini boshqaradi.
type CartHandler struct {
	cartService *service.CartService
}

// NewCartHandler CartHandler ning yangi instansini yaratadi.
func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

// sendErrorResponse yordamchi funksiyasi xato javobini yuborish uchun
func (h *CartHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Message: message,
		Error:   details,
	})
	log.Printf("CartHandler xato javobi yuborildi: Status=%d, Xabar='%s', Tafsilotlar='%s'", statusCode, message, details)
}

// sendSuccessResponse yordamchi funksiyasi muvaffaqiyatli javobni yuborish uchun
func (h *CartHandler) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
	log.Printf("CartHandler muvaffaqiyatli javob yuborildi: Xabar='%s'", message)
}

// AddToCart POST /api/users/{id}/cart - Savatga mahsulot qo'shadi yoki miqdorini yangilaydi.
func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri foydalanuvchi ID formati", err.Error())
		return
	}

	var req models.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "JSON formatida xatolik", err.Error())
		return
	}

	cartItem, err := h.cartService.AddToCart(userID, req.FoodID, req.Quantity)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Savatga mahsulot qo'shishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Mahsulot savatga muvaffaqiyatli qo'shildi/yangilandi", cartItem)
}

// GetCart GET /api/users/{id}/cart - Foydalanuvchining savatidagi barcha elementlarni qaytaradi.
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri foydalanuvchi ID formati", err.Error())
		return
	}

	cartItems, err := h.cartService.GetCart(userID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Savatni olishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Savat muvaffaqiyatli olindi", cartItems)
}

// UpdateCartItem PUT /api/users/{user_id}/cart/{food_id} - Savatdagi ma'lum bir mahsulot miqdorini yangilaydi.
func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["user_id"]
	foodIDStr := vars["food_id"]

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri foydalanuvchi ID formati", err.Error())
		return
	}
	foodID, err := strconv.Atoi(foodIDStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri mahsulot ID formati", err.Error())
		return
	}

	var req models.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "JSON formatida xatolik", err.Error())
		return
	}

	err = h.cartService.UpdateCartItemQuantity(userID, foodID, req.Quantity)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Savatdagi mahsulot miqdorini yangilashda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Savatdagi mahsulot miqdori muvaffaqiyatli yangilandi", nil)
}

// RemoveFromCart DELETE /api/users/{user_id}/cart/{food_id} - Savatdan ma'lum bir mahsulotni o'chiradi.
func (h *CartHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["user_id"]
	foodIDStr := vars["food_id"]

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri foydalanuvchi ID formati", err.Error())
		return
	}
	foodID, err := strconv.Atoi(foodIDStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri mahsulot ID formati", err.Error())
		return
	}

	err = h.cartService.RemoveFromCart(userID, foodID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Savatdan mahsulotni o'chirishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Mahsulot savatdan muvaffaqiyatli o'chirildi", nil)
}

// ClearCart DELETE /api/users/{id}/cart - Foydalanuvchining savatidagi barcha elementlarni o'chiradi.
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri foydalanuvchi ID formati", err.Error())
		return
	}

	err = h.cartService.ClearUserCart(userID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Savatni tozalashda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Savat muvaffaqiyatli tozalandi", nil)
}
