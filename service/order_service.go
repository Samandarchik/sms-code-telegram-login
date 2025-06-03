package service

import (
	"amur/models"
	"amur/repository"
	"database/sql"
	"encoding/json" // Import for JSON handling
	"errors"
	"fmt"
	"io/ioutil" // Import for file reading
	"time"
)

type OrderService struct {
	orderRepo  *repository.OrderRepository
	basketRepo *repository.BasketOrderRepository
	foodRepo   *repository.FoodRepository
	tableMap   map[string]string // Add tableMap to store table_id (token) -> table_name
}

func NewOrderService(orderRepo *repository.OrderRepository, basketRepo *repository.BasketOrderRepository, foodRepo *repository.FoodRepository) *OrderService {
	// Load table data when the service is initialized
	tableMap, err := loadTableData("table.json")
	if err != nil {
		// Handle this error appropriately in a real application (e.g., log it and exit)
		fmt.Printf("Error loading table data: %v\n", err)
		// For now, we'll continue, but orders for "zalga" will fail if table data isn't loaded.
	}

	return &OrderService{
		orderRepo:  orderRepo,
		basketRepo: basketRepo,
		foodRepo:   foodRepo,
		tableMap:   tableMap,
	}
}

// loadTableData loads table names from a JSON file into a map (token -> table name)
func loadTableData(filename string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read table data file: %w", err)
	}

	var rawTableMap map[string]string
	if err := json.Unmarshal(data, &rawTableMap); err != nil {
		return nil, fmt.Errorf("could not unmarshal table data: %w", err)
	}

	// Invert the map: token -> table name
	invertedMap := make(map[string]string)
	for tableName, token := range rawTableMap {
		invertedMap[token] = tableName
	}
	return invertedMap, nil
}

// CreateOrder savatchadagi mahsulotlardan yangi buyurtma yaratadi
func (s *OrderService) CreateOrder(telegramID int64, req *models.CreateOrderRequest) (*models.OrderDetailsResponse, error) {
	// 1. Foydalanuvchining savatchasini olish
	basketItems, err := s.basketRepo.GetBasketOrdersByTelegramID(telegramID)
	if err != nil {
		return nil, fmt.Errorf("savatchani olishda xatolik: %w", err)
	}
	if len(basketItems) == 0 {
		return nil, errors.New("savatcha bo'sh, buyurtma berish mumkin emas")
	}

	var totalOrderPrice float64
	var orderItemsToCreate []*models.OrderItem // Buyurtma uchun qo'shiladigan mahsulotlar (vaqtinchalik pointerlar slice'i)

	// 2. Har bir savatcha elementi uchun mahsulot narxini olish va umumiy narxni hisoblash
	for _, item := range basketItems {
		food, err := s.foodRepo.GetByID(item.FoodID)
		if err != nil {
			// Agar ovqat topilmasa, bu buyurtmani yaratishga to'sqinlik qilishi kerak
			return nil, fmt.Errorf("FoodID %d uchun ovqat topilmadi: %w", item.FoodID, err)
		}
		itemTotalPrice := food.FoodPrice * float64(item.Quantity)
		totalOrderPrice += itemTotalPrice

		orderItemsToCreate = append(orderItemsToCreate, &models.OrderItem{
			FoodID:    item.FoodID,
			Quantity:  item.Quantity,
			ItemPrice: food.FoodPrice, // Buyurtma qilingan vaqtdagi narx
		})
	}

	// 3. Buyurtma yaratish (asosiy order ma'lumotlari)
	order := &models.Order{
		TelegramID:   telegramID,
		OrderTime:    time.Now(),
		OrderStatus:  "buyurtma qabul qilindi", // Default holat
		DeliveryType: req.DeliveryType,
		TotalPrice:   totalOrderPrice,
		Comment:      req.Comment, // Buyurtma izohi
	}

	// Handle delivery type specific logic
	switch req.DeliveryType {
	case "yetkazib berish":
		if req.DeliveryLatitude == nil || req.DeliveryLongitude == nil {
			return nil, errors.New("yetkazib berish uchun lokatsiya ma'lumotlari (latitude va longitude) majburiy")
		}
		order.DeliveryLatitude = req.DeliveryLatitude
		order.DeliveryLongitude = req.DeliveryLongitude
	case "o'zi olib ketish":
		order.DeliveryLatitude = nil
		order.DeliveryLongitude = nil
		order.TableID = nil // Ensure TableID is null for this type
	case "zalga":
		order.DeliveryLatitude = nil
		order.DeliveryLongitude = nil
		if req.TableID == nil {
			return nil, errors.New("zalga buyurtma berish uchun stol ID (QR kod tokeni) majburiy")
		}

		// Look up the table name using the provided token
		tableName, ok := s.tableMap[*req.TableID]
		if !ok {
		}
		// Set the TableID field in the Order model to the resolved table name
		tableIDStr := tableName // Convert string to *string
		order.TableID = &tableIDStr
		// You can also set the comment based on the table name if you wish
		// If req.Comment is not already set, you could do:
		// if order.Comment == nil {
		// 	comment := fmt.Sprintf("Stol: %s", tableName)
		// 	order.Comment = &comment
		// }
	default:
		return nil, fmt.Errorf("noto'g'ri yetkazib berish turi: %s", req.DeliveryType)
	}

	createdOrder, err := s.orderRepo.CreateOrder(order)
	if err != nil {
		return nil, fmt.Errorf("buyurtma yaratishda xatolik: %w", err)
	}

	// 4. Buyurtma elementlarini (order_items) qo'shish
	for _, item := range orderItemsToCreate { // `orderItemsToCreate` dan foydalanamiz
		item.OrderID = createdOrder.OrderID // Yangi yaratilgan buyurtma ID'sini bog'lash
		if err := s.orderRepo.AddOrderItem(item); err != nil {
			return nil, fmt.Errorf("buyurtma elementini qo'shishda xatolik: %w", err)
		}
	}

	// 5. Savatchani tozalash
	if err := s.basketRepo.ClearBasket(telegramID); err != nil {
		fmt.Printf("Savatchani tozalashda xatolik (Buyurtma yaratildi, lekin savatcha tozalanmadi): %v\n", err)
	}

	// 6. To'liq buyurtma ma'lumotlarini qaytarish
	// Repositorydan pointerlar slice'ini olamiz
	_, fetchedOrderItemsPointers, err := s.orderRepo.GetOrderWithItemsByID(createdOrder.OrderID)
	if err != nil {
		return nil, fmt.Errorf("yaratilgan buyurtmani olishda xatolik: %w", err)
	}

	// `[]*models.OrderItem` ni `[]models.OrderItem` ga aylantiramiz
	var finalOrderItems []models.OrderItem
	for _, itemPtr := range fetchedOrderItemsPointers {
		if itemPtr != nil {
			finalOrderItems = append(finalOrderItems, *itemPtr) // Pointerdan qiymatni olish
		}
	}

	return &models.OrderDetailsResponse{
		Order:      *createdOrder,
		OrderItems: finalOrderItems, // Endi to'g'ri tip
	}, nil
}

// GetOrderDetails buyurtma va uning elementlarini qaytaradi
func (s *OrderService) GetOrderDetails(orderID int) (*models.OrderDetailsResponse, error) {
	order, orderItemsPointers, err := s.orderRepo.GetOrderWithItemsByID(orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("id=%d bo'lgan buyurtma topilmadi", orderID)
		}
		return nil, fmt.Errorf("buyurtma ma'lumotlarini olishda xatolik: %w", err)
	}

	// `[]*models.OrderItem` ni `[]models.OrderItem` ga aylantiramiz
	var orderItems []models.OrderItem
	for _, itemPtr := range orderItemsPointers {
		if itemPtr != nil {
			orderItems = append(orderItems, *itemPtr) // Pointerdan qiymatni olish
		}
	}

	return &models.OrderDetailsResponse{
		Order:      *order,
		OrderItems: orderItems, // Endi to'g'ri tip
	}, nil
}

// GetUserOrdersWithDetails foydalanuvchining barcha buyurtmalarini va ularning elementlarini oladi
func (s *OrderService) GetUserOrdersWithDetails(telegramID int64) ([]models.OrderDetailsResponse, error) {
	orders, err := s.orderRepo.GetUserOrders(telegramID)
	if err != nil {
		return nil, fmt.Errorf("foydalanuvchi buyurtmalarini olishda xatolik: %w", err)
	}

	var allOrderDetails []models.OrderDetailsResponse
	for _, order := range orders {
		_, orderItemsPointers, err := s.orderRepo.GetOrderWithItemsByID(order.OrderID)
		if err != nil {
			fmt.Printf("Order ID %d uchun buyurtma elementlarini olishda xatolik: %v\n", order.OrderID, err)
			continue
		}

		// `[]*models.OrderItem` ni `[]models.OrderItem` ga aylantiramiz
		var orderItems []models.OrderItem
		for _, itemPtr := range orderItemsPointers {
			if itemPtr != nil {
				orderItems = append(orderItems, *itemPtr) // Pointerdan qiymatni olish
			}
		}

		allOrderDetails = append(allOrderDetails, models.OrderDetailsResponse{
			Order:      *order,
			OrderItems: orderItems, // Endi to'g'ri tip
		})
	}
	return allOrderDetails, nil
}

// UpdateOrderStatus buyurtma holatini yangilaydi
func (s *OrderService) UpdateOrderStatus(orderID int, newStatus string) error {
	validStatuses := map[string]bool{
		"buyurtma qabul qilindi":  true,
		"buyurtma tayyorlanmoqda": true,
		"buyurtma tayyor":         true,
		"buyurtma yetkazildi":     true,
		"bekor qilindi":           true,
	}
	if !validStatuses[newStatus] {
		return fmt.Errorf("noto'g'ri buyurtma holati: %s", newStatus)
	}

	err := s.orderRepo.UpdateOrderStatus(orderID, newStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("id=%d bo'lgan buyurtma topilmadi", orderID)
		}
		return fmt.Errorf("buyurtma holatini yangilashda xatolik: %w", err)
	}
	return nil
}

// GetOrderStats buyurtma statistikasini oladi
func (s *OrderService) GetOrderStats() (int, error) {
	count, err := s.orderRepo.GetOrderStats()
	if err != nil {
		return 0, fmt.Errorf("buyurtma statistikasini olishda xatolik: %w", err)
	}
	return count, nil
}

// DeleteOrderAdmin buyurtmani o'chiradi (FAQAT ADMIN UCHUN)
func (s *OrderService) DeleteOrderAdmin(orderID int) error {
	err := s.orderRepo.DeleteOrder(orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("id=%d bo'lgan buyurtma topilmadi", orderID)
		}
		return fmt.Errorf("buyurtmani o'chirishda xatolik: %w", err)
	}
	return nil
}
