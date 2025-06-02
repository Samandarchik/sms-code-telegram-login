package service

import (
	"amur/models"
	"amur/repository"
	"fmt"
	"strings"
)

type FoodService struct {
	foodRepo *repository.FoodRepository
}

func NewFoodService(foodRepo *repository.FoodRepository) *FoodService {
	return &FoodService{foodRepo: foodRepo}
}

func (s *FoodService) CreateFood(req *models.CreateFoodRequest) (*models.Food, error) {
	if err := s.validateCreateFoodRequest(req); err != nil {
		return nil, err
	}

	food := &models.Food{
		FoodName:     strings.TrimSpace(req.FoodName),
		FoodCategory: strings.TrimSpace(req.FoodCategory),
		FoodPrice:    req.FoodPrice,
		FoodImage:    strings.TrimSpace(req.FoodImage), // Rasm URL manzilini qabul qiladi
	}

	err := s.foodRepo.Create(food)
	if err != nil {
		return nil, err
	}

	return food, nil
}

func (s *FoodService) GetAllFoods() ([]*models.Food, error) {
	return s.foodRepo.GetAll()
}

func (s *FoodService) GetFoodByID(id int) (*models.Food, error) {
	if id <= 0 {
		return nil, fmt.Errorf("noto'g'ri food ID")
	}
	return s.foodRepo.GetByID(id)
}

func (s *FoodService) UpdateFood(id int, req *models.UpdateFoodRequest) (*models.Food, error) {
	if id <= 0 {
		return nil, fmt.Errorf("noto'g'ri food ID")
	}

	existingFood, err := s.foodRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("ovqat topilmadi")
	}

	updatedFood := &models.Food{
		FoodID:       existingFood.FoodID,
		FoodName:     existingFood.FoodName,
		FoodCategory: existingFood.FoodCategory,
		FoodPrice:    existingFood.FoodPrice,
		FoodImage:    existingFood.FoodImage,
	}

	if req.FoodName != "" {
		updatedFood.FoodName = strings.TrimSpace(req.FoodName)
	}
	if req.FoodCategory != "" {
		updatedFood.FoodCategory = strings.TrimSpace(req.FoodCategory)
	}
	if req.FoodPrice > 0 {
		updatedFood.FoodPrice = req.FoodPrice
	}
	if req.FoodImage != "" { // Agar yangi rasm URL manzili kelsa
		updatedFood.FoodImage = strings.TrimSpace(req.FoodImage)
	}

	err = s.foodRepo.Update(id, updatedFood)
	if err != nil {
		return nil, err
	}

	return s.foodRepo.GetByID(id)
}

func (s *FoodService) DeleteFood(id int) error {
	if id <= 0 {
		return fmt.Errorf("noto'g'ri food ID")
	}

	_, err := s.foodRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("ovqat topilmadi")
	}

	return s.foodRepo.Delete(id)
}

func (s *FoodService) GetFoodsByCategory(category string) ([]*models.Food, error) {
	if category == "" {
		return nil, fmt.Errorf("kategoriya nomi bo'sh bo'lishi mumkin emas")
	}
	return s.foodRepo.GetByCategory(strings.TrimSpace(category))
}

func (s *FoodService) GetFoodCount() int {
	return s.foodRepo.Count()
}

func (s *FoodService) validateCreateFoodRequest(req *models.CreateFoodRequest) error {
	if strings.TrimSpace(req.FoodName) == "" {
		return fmt.Errorf("ovqat nomi bo'sh bo'lishi mumkin emas")
	}
	if strings.TrimSpace(req.FoodCategory) == "" {
		return fmt.Errorf("kategoriya nomi bo'sh bo'lishi mumkin emas")
	}
	if req.FoodPrice <= 0 {
		return fmt.Errorf("narx 0 dan katta bo'lishi kerak")
	}
	return nil
}
