package handlers

import (
	"amur/middleware" // Yangi middleware paketini import qilish
	"amur/models"
	"amur/service"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// sendErrorResponse yordamchi funksiyasi xato javobini yuborish uchun
func (h *OrderHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   message,
		"details": details,
	})
	// log.Printf("Xato javobi yuborildi: Status=%d, Xabar='%s', Tafsilotlar='%s'", statusCode, message, details)
}

// sendSuccessResponse yordamchi funksiyasi muvaffaqiyatli javobni yuborish uchun
func (h *OrderHandler) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
		"data":    data,
	})
	// log.Printf("Muvaffaqiyatli javob yuborildi: Xabar='%s'", message)
}

// getTelegramIDFromContext yordamchi funksiya
func (h *OrderHandler) getTelegramIDFromContext(w http.ResponseWriter, r *http.Request) (int64, bool) {
	// OLD: userID, ok := r.Context().Value(middleware.UserContextKey).(int)
	// YANGI: Contextdan Telegram IDni to'g'ridan-to'g'ri int64 sifatida olish
	telegramID, ok := r.Context().Value(middleware.TelegramIDContextKey).(int64) // <-- Shu qatorni o'zgartiring!
	if !ok {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Foydalanuvchi Telegram IDsi kontekstda topilmadi", "Autentifikatsiya xatoligi. AuthMiddleware to'g'ri ishlamagan bo'lishi mumkin.")
		return 0, false
	}
	// Agar oldingi kodda userID ni int dan int64 ga o'girish shart bo'lgan bo'lsa,
	// endi biz to'g'ridan-to'g'ri int64 qabul qilganimiz uchun bu qator endi kerak emas:
	// telegramID := int64(userID)
	return telegramID, true
}

// Qolgan funksiyalar (CreateOrder, GetOrderDetails, GetUserOrders, UpdateOrderStatus, GetOrderStats, DeleteOrderAdmin) o'zgarishsiz qoladi.
// Chunki ularda faqat getTelegramIDFromContext chaqiriladi va qaytarilgan telegramID ishlatiladi.

// CreateOrder savatchadagi mahsulotlardan buyurtma yaratish
// POST /api/orders (telegramID endi URLda emas)
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	telegramID, ok := h.getTelegramIDFromContext(w, r)
	if !ok {
		return
	}

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "So'rov tanasini tahlil qilishda xatolik", err.Error())
		return
	}

	if req.DeliveryType != "yetkazib berish" && req.DeliveryType != "o'zi olib ketish" && req.DeliveryType != "zalga" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri yetkazib berish turi", "Qabul qilinadigan turlar: 'yetkazib berish', 'o''zi olib ketish', 'zalga'")
		return
	}
	if req.DeliveryType == "yetkazib berish" && (req.DeliveryLatitude == nil || req.DeliveryLongitude == nil) {
		h.sendErrorResponse(w, http.StatusBadRequest, "Yetkazib berish uchun lokatsiya ma'lumotlari majburiy", "latitude va longitude kiritilishi kerak")
		return
	}

	order, err := h.orderService.CreateOrder(telegramID, &req) // 'order' deb o'zgartirdim, avvalgi kodda 'orderDetails' edi.
	if err != nil {
		if errors.Is(err, errors.New("savatcha bo'sh, buyurtma berish mumkin emas")) {
			h.sendErrorResponse(w, http.StatusBadRequest, "Buyurtma yaratishda xatolik", err.Error())
		} else if errors.Is(err, errors.New("yetkazib berish uchun lokatsiya ma'lumotlari (latitude va longitude) majburiy")) {
			h.sendErrorResponse(w, http.StatusBadRequest, "Buyurtma yaratishda xatolik", err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, "Buyurtma yaratishda xatolik", err.Error())
		}
		return
	}

	h.sendSuccessResponse(w, "Buyurtma muvaffaqiyatli yaratildi", order)
}

// GetOrderDetails buyurtma ma'lumotlarini (elementlari bilan birga) olish
// GET /api/orders/{orderID}
func (h *OrderHandler) GetOrderDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["orderID"]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri Buyurtma ID", err.Error())
		return
	}

	orderDetails, err := h.orderService.GetOrderDetails(orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.sendErrorResponse(w, http.StatusNotFound, "Buyurtma topilmadi", err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, "Buyurtma ma'lumotlarini olishda xatolik", err.Error())
		}
		return
	}

	h.sendSuccessResponse(w, "Buyurtma ma'lumotlari muvaffaqiyatli olindi", orderDetails)
}

// GetUserOrders foydalanuvchining barcha buyurtmalarini (elementlari bilan birga) olish
// GET /api/orders (telegramID endi URLda emas)
func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	telegramID, ok := h.getTelegramIDFromContext(w, r)
	if !ok {
		return
	}

	allOrderDetails, err := h.orderService.GetUserOrdersWithDetails(telegramID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Foydalanuvchi buyurtmalarini olishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Foydalanuvchi buyurtmalari muvaffaqiyatli olindi", allOrderDetails)
}

// UpdateOrderStatus buyurtma holatini yangilash (Faqat admin uchun bo'lishi kerak)
// PUT /api/orders/{orderID}/status
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	// Bu yerda ham admin roli tekshirilishi kerak.
	vars := mux.Vars(r)
	orderIDStr := vars["orderID"]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri Buyurtma ID", err.Error())
		return
	}

	var requestBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "So'rov tanasini tahlil qilishda xatolik", err.Error())
		return
	}

	newStatus, ok := requestBody["status"]
	if !ok || newStatus == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Status maydoni majburiy", "JSON tanasida 'status' maydoni bo'lishi kerak.")
		return
	}

	err = h.orderService.UpdateOrderStatus(orderID, newStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.sendErrorResponse(w, http.StatusNotFound, "Buyurtma topilmadi", err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, "Buyurtma holatini yangilashda xatolik", err.Error())
		}
		return
	}

	h.sendSuccessResponse(w, "Buyurtma holati muvaffaqiyatli yangilandi", nil)
}

// GetOrderStats buyurtma statistikasini olish (Faqat admin uchun)
// GET /api/orders/stats
func (h *OrderHandler) GetOrderStats(w http.ResponseWriter, r *http.Request) {
	// Bu yerda ham admin roli tekshirilishi kerak.
	count, err := h.orderService.GetOrderStats()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Buyurtma statistikasini olishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Buyurtma statistikasi muvaffaqiyatli olindi", map[string]int{"total_orders": count})
}

// DeleteOrderAdmin buyurtmani o'chirish (FAQAT ADMIN UCHUN)
// DELETE /api/admin/orders/{orderID}
func (h *OrderHandler) DeleteOrderAdmin(w http.ResponseWriter, r *http.Request) {
	// Bu funksiya uchun avtorizatsiya AuthMiddleware va RoleMiddleware orqali amalga oshiriladi.
	// Handler ichida qayta tekshirish shart emas, chunki agar middleware ishlamagan bo'lsa, bu yerga yetib kelmaydi.

	vars := mux.Vars(r)
	orderIDStr := vars["orderID"]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri Buyurtma ID", err.Error())
		return
	}

	err = h.orderService.DeleteOrderAdmin(orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.sendErrorResponse(w, http.StatusNotFound, "Buyurtma topilmadi", err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, "Buyurtmani o'chirishda xatolik", err.Error())
		}
		return
	}

	h.sendSuccessResponse(w, "Buyurtma muvaffaqiyatli o'chirildi", nil)
}
