package handlers

import (
	// Yangi middleware paketini import qilish
	"amur/models"
	"amur/service"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type FoodHandler struct {
	foodService *service.FoodService
}

func NewFoodHandler(foodService *service.FoodService) *FoodHandler {
	return &FoodHandler{foodService: foodService}
}

// sendErrorResponse yordamchi funksiyasi xato javobini yuborish uchun
func (h *FoodHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   message,
		"details": details,
	})
	log.Printf("Xato javobi yuborildi: Status=%d, Xabar='%s', Tafsilotlar='%s'", statusCode, message, details)
}

// sendSuccessResponse yordamchi funksiyasi muvaffaqiyatli javobni yuborish uchun
func (h *FoodHandler) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
		"data":    data,
	})
	log.Printf("Muvaffaqiyatli javob yuborildi: Xabar='%s'", message)
}

// GET /api/foods - Barcha ovqatlarni olish (Ruxsat talab qilinmaydi yoki oddiy user)
func (h *FoodHandler) GetAllFoods(w http.ResponseWriter, r *http.Request) {
	foods, err := h.foodService.GetAllFoods()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Ovqatlarni olishda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Ovqatlar muvaffaqiyatli olindi", foods)
}

// GET /api/foods/{id} - Bitta ovqatni olish (Ruxsat talab qilinmaydi yoki oddiy user)
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

// POST /api/foods - Yangi ovqat qo'shish (Faqat admin uchun)
func (h *FoodHandler) CreateFood(w http.ResponseWriter, r *http.Request) {
	// Bu yerda middleware'lar yordamida avtorizatsiya allaqachon tekshirilgan bo'ladi.
	// Faqat admin roli bo'lganlar bu funksiyaga yetib keladi.
	log.Println("CreateFood so'rovi qabul qilindi.")

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Form ma'lumotlarini tahlil qilishda xatolik", err.Error())
		return
	}
	log.Println("Form ma'lumotlari muvaffaqiyatli tahlil qilindi.")

	foodName := r.FormValue("food_name")
	foodCategory := r.FormValue("food_category")
	foodPriceStr := r.FormValue("food_price")

	log.Printf("Form qiymatlari: FoodName='%s', FoodCategory='%s', FoodPriceStr='%s'", foodName, foodCategory, foodPriceStr)

	if foodName == "" || foodCategory == "" || foodPriceStr == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Majburiy maydonlar to'ldirilmagan", "food_name, food_category, food_price maydonlari majburiy.")
		return
	}

	foodPrice, err := strconv.ParseFloat(foodPriceStr, 64)
	if err != nil || foodPrice <= 0 {
		h.sendErrorResponse(w, http.StatusBadRequest, "Narx noto'g'ri formatda", "food_price musbat son bo'lishi kerak.")
		return
	}

	var foodImageURL string
	file, handler, err := r.FormFile("food_image")
	if err == nil {
		defer file.Close()
		log.Printf("Rasm fayli topildi: %s, Hajmi: %d bytes", handler.Filename, handler.Size)

		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			log.Printf("'%s' papkasi topilmadi, yaratilmoqda...", uploadDir)
			err = os.Mkdir(uploadDir, 0755)
			if err != nil {
				h.sendErrorResponse(w, http.StatusInternalServerError, "Yuklash papkasini yaratishda xatolik", err.Error())
				return
			}
			log.Printf("'%s' papkasi muvaffaqiyatli yaratildi.", uploadDir)
		} else if err != nil {
			h.sendErrorResponse(w, http.StatusInternalServerError, "Yuklash papkasi holatini tekshirishda xatolik", err.Error())
			return
		}

		ext := filepath.Ext(handler.Filename)
		newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		filePath := filepath.Join(uploadDir, newFileName)
		log.Printf("Fayl saqlash yo'li: %s", filePath)

		dst, err := os.Create(filePath)
		if err != nil {
			h.sendErrorResponse(w, http.StatusInternalServerError, "Faylni saqlashda xatolik", err.Error())
			return
		}
		defer dst.Close()
		log.Println("Bo'sh fayl yaratildi.")

		bytesCopied, err := io.Copy(dst, file)
		if err != nil {
			h.sendErrorResponse(w, http.StatusInternalServerError, "Faylni nusxalashda xatolik", err.Error())
			return
		}
		log.Printf("Faylga %d bayt nusxalandi.", bytesCopied)

		foodImageURL = "/uploads/" + newFileName
		log.Printf("Rasm URL manzili: %s", foodImageURL)

	} else if err == http.ErrMissingFile {
		log.Println("Rasm fayli yuklanmadi (ixtiyoriy).")
	} else {
		h.sendErrorResponse(w, http.StatusBadRequest, "Rasm yuklashda kutilmagan xatolik", err.Error())
		return
	}

	req := models.CreateFoodRequest{
		FoodName:     foodName,
		FoodCategory: foodCategory,
		FoodPrice:    foodPrice,
		FoodImage:    foodImageURL,
	}

	food, err := h.foodService.CreateFood(&req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "Ovqat qo'shishda xizmat xatoligi", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Ovqat muvaffaqiyatli qo'shildi", food)
	log.Println("Ovqat yaratish jarayoni yakunlandi.")
}

// PUT /api/foods/{id} - Ovqatni yangilash (Faqat admin uchun)
func (h *FoodHandler) UpdateFood(w http.ResponseWriter, r *http.Request) {
	// Bu yerda middleware'lar yordamida avtorizatsiya allaqachon tekshirilgan bo'ladi.
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri ID", err.Error())
		return
	}

	contentType := r.Header.Get("Content-Type")
	var req models.UpdateFoodRequest

	if strings.Contains(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			h.sendErrorResponse(w, http.StatusBadRequest, "Form ma'lumotlarini tahlil qilishda xatolik", err.Error())
			return
		}

		req.FoodName = r.FormValue("food_name")
		req.FoodCategory = r.FormValue("food_category")
		if priceStr := r.FormValue("food_price"); priceStr != "" {
			req.FoodPrice, err = strconv.ParseFloat(priceStr, 64)
			if err != nil {
				h.sendErrorResponse(w, http.StatusBadRequest, "Narx noto'g'ri formatda", err.Error())
				return
			}
		}

		var foodImageURL string
		file, handler, err := r.FormFile("food_image")
		if err == nil {
			defer file.Close()

			uploadDir := "./uploads"
			if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
				os.Mkdir(uploadDir, 0755)
			}

			ext := filepath.Ext(handler.Filename)
			newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			filePath := filepath.Join(uploadDir, newFileName)

			dst, err := os.Create(filePath)
			if err != nil {
				h.sendErrorResponse(w, http.StatusInternalServerError, "Faylni saqlashda xatolik", err.Error())
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				h.sendErrorResponse(w, http.StatusInternalServerError, "Faylni nusxalashda xatolik", err.Error())
				return
			}
			foodImageURL = "/uploads/" + newFileName
		} else if err != http.ErrMissingFile {
			h.sendErrorResponse(w, http.StatusBadRequest, "Rasm yuklashda kutilmagan xatolik", err.Error())
			return
		}
		req.FoodImage = foodImageURL

	} else if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendErrorResponse(w, http.StatusBadRequest, "JSON formatida xatolik", err.Error())
			return
		}
	} else {
		h.sendErrorResponse(w, http.StatusUnsupportedMediaType, "Qo'llab-quvvatlanmaydigan Content-Type", "application/json yoki multipart/form-data kutilmoqda.")
		return
	}

	food, err := h.foodService.UpdateFood(id, &req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Ovqatni yangilashda xatolik", err.Error())
		return
	}

	h.sendSuccessResponse(w, "Ovqat muvaffaqiyatli yangilandi", food)
}

// DELETE /api/foods/{id} - Ovqatni o'chirish (Faqat admin uchun)
func (h *FoodHandler) DeleteFood(w http.ResponseWriter, r *http.Request) {
	// Bu yerda middleware'lar yordamida avtorizatsiya allaqachon tekshirilgan bo'ladi.
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Noto'g'ri ID", err.Error())
		return
	}

	foodToDelete, err := h.foodService.GetFoodByID(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "O'chiriladigan ovqat topilmadi", err.Error())
		return
	}

	err = h.foodService.DeleteFood(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Ovqatni o'chirishda xatolik", err.Error())
		return
	}

	if foodToDelete.FoodImage != "" {
		imagePath := "." + foodToDelete.FoodImage
		if _, err := os.Stat(imagePath); err == nil {
			err := os.Remove(imagePath)
			if err != nil {
				log.Printf("Rasm faylini o'chirishda xatolik: %s - %v", imagePath, err)
			} else {
				log.Printf("Rasm fayli muvaffaqiyatli o'chirildi: %s", imagePath)
			}
		} else if !os.IsNotExist(err) {
			log.Printf("Rasm fayli holatini tekshirishda xatolik: %s - %v", imagePath, err)
		}
	}

	h.sendSuccessResponse(w, "Ovqat muvaffaqiyatli o'chirildi", nil)
}

// GET /api/foods/category/{category} - Kategoriya bo'yicha ovqatlarni olish (Ruxsat talab qilinmaydi yoki oddiy user)
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

// GET /api/foods/stats - Ovqatlar statistikasi (Faqat admin uchun)
func (h *FoodHandler) GetFoodStats(w http.ResponseWriter, r *http.Request) {
	// Bu yerda middleware'lar yordamida avtorizatsiya allaqachon tekshirilgan bo'ladi.
	count := h.foodService.GetFoodCount()
	stats := map[string]interface{}{
		"total_foods": count,
	}

	h.sendSuccessResponse(w, "Statistika muvaffaqiyatli olindi", stats)
}
