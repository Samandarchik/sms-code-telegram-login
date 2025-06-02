package service

import (
	"amur/models"
	"amur/repository"
	"fmt"
)

// CartService savat bilan bog'liq biznes logikasini boshqaradi.
type CartService struct {
	cartRepo *repository.CartRepository
	foodRepo *repository.FoodRepository // Mahsulot ma'lumotlarini olish uchun
	userRepo *repository.UserRepository // Foydalanuvchi mavjudligini tekshirish uchun
}

// NewCartService CartService ning yangi instansini yaratadi.
func NewCartService(cartRepo *repository.CartRepository, foodRepo *repository.FoodRepository, userRepo *repository.UserRepository) *CartService {
	return &CartService{cartRepo: cartRepo, foodRepo: foodRepo, userRepo: userRepo}
}

// AddToCart savatga mahsulot qo'shadi yoki mavjud bo'lsa miqdorini yangilaydi.
func (s *CartService) AddToCart(userID int64, foodID int, quantity int) (*models.CartItem, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("noto'g'ri foydalanuvchi ID")
	}
	if foodID <= 0 {
		return nil, fmt.Errorf("noto'g'ri mahsulot ID")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("miqdor 0 dan katta bo'lishi kerak")
	}

	// Foydalanuvchi mavjudligini tekshirish (ixtiyoriy, lekin yaxshi amaliyot)
	if !s.userRepo.Exists(userID) {
		return nil, fmt.Errorf("foydalanuvchi topilmadi")
	}

	// Mahsulot mavjudligini tekshirish
	food, err := s.foodRepo.GetByID(foodID)
	if err != nil || food == nil {
		return nil, fmt.Errorf("mahsulot topilmadi")
	}

	item := &models.CartItem{
		UserID:   userID,
		FoodID:   foodID,
		Quantity: quantity,
	}

	return s.cartRepo.CreateOrUpdateCartItem(item)
}

// GetCart foydalanuvchining savatidagi barcha elementlarni to'liq ma'lumotlari bilan qaytaradi.
func (s *CartService) GetCart(userID int64) ([]*models.CartItemResponse, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("noto'g'ri foydalanuvchi ID")
	}

	// Foydalanuvchi mavjudligini tekshirish
	if !s.userRepo.Exists(userID) {
		return nil, fmt.Errorf("foydalanuvchi topilmadi")
	}

	cartItems, err := s.cartRepo.GetCartItemsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var cartResponses []*models.CartItemResponse
	for _, item := range cartItems {
		food, err := s.foodRepo.GetByID(item.FoodID)
		if err != nil {
			// Agar mahsulot topilmasa, logga yozib, o'tkazib yuborishimiz mumkin
			// yoki xato qaytarishimiz mumkin. Hozircha logga yozamiz.
			fmt.Printf("Savatdagi mahsulot (FoodID: %d) topilmadi: %v\n", item.FoodID, err)
			continue
		}
		cartResponses = append(cartResponses, &models.CartItemResponse{
			ID:        item.ID,
			UserID:    item.UserID,
			Food:      food,
			Quantity:  item.Quantity,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return cartResponses, nil
}

// UpdateCartItemQuantity savatdagi ma'lum bir mahsulot miqdorini yangilaydi.
func (s *CartService) UpdateCartItemQuantity(userID int64, foodID int, quantity int) error {
	if userID <= 0 {
		return fmt.Errorf("noto'g'ri foydalanuvchi ID")
	}
	if foodID <= 0 {
		return fmt.Errorf("noto'g'ri mahsulot ID")
	}
	if quantity < 0 { // Miqdor manfiy bo'lishi mumkin emas
		return fmt.Errorf("miqdor manfiy bo'lishi mumkin emas")
	}

	// Foydalanuvchi mavjudligini tekshirish
	if !s.userRepo.Exists(userID) {
		return fmt.Errorf("foydalanuvchi topilmadi")
	}

	// Agar miqdor 0 bo'lsa, elementni o'chiramiz
	if quantity == 0 {
		return s.cartRepo.DeleteCartItem(userID, foodID)
	}

	return s.cartRepo.UpdateCartItemQuantity(userID, foodID, quantity)
}

// RemoveFromCart savatdan ma'lum bir mahsulotni o'chiradi.
func (s *CartService) RemoveFromCart(userID int64, foodID int) error {
	if userID <= 0 {
		return fmt.Errorf("noto'g'ri foydalanuvchi ID")
	}
	if foodID <= 0 {
		return nil // Noto'g'ri foodID bo'lsa, hech narsa qilmaymiz
	}

	// Foydalanuvchi mavjudligini tekshirish
	if !s.userRepo.Exists(userID) {
		return fmt.Errorf("foydalanuvchi topilmadi")
	}

	return s.cartRepo.DeleteCartItem(userID, foodID)
}

// ClearUserCart foydalanuvchining savatidagi barcha elementlarni o'chiradi.
func (s *CartService) ClearUserCart(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("noto'g'ri foydalanuvchi ID")
	}

	// Foydalanuvchi mavjudligini tekshirish
	if !s.userRepo.Exists(userID) {
		return fmt.Errorf("foydalanuvchi topilmadi")
	}

	return s.cartRepo.ClearCart(userID)
}
