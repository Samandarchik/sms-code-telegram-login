// ==================== service/user_service.go ====================
package service

import (
	"amur/models"
	"amur/repository"
	"fmt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(user *models.User) error {
	s.validateAndCleanUser(user)

	if s.userRepo.Exists(user.TgID) {
		return fmt.Errorf("foydalanuvchi allaqachon mavjud")
	}

	return s.userRepo.Create(user)
}

func (s *UserService) UpdateUser(user *models.User) error {
	s.validateAndCleanUser(user)
	return s.userRepo.Update(user)
}

func (s *UserService) GetUser(tgID int64) (*models.User, error) {
	return s.userRepo.GetByTgID(tgID)
}

func (s *UserService) UserExists(tgID int64) bool {
	return s.userRepo.Exists(tgID)
}

func (s *UserService) GetUserCount() int {
	return s.userRepo.Count()
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.userRepo.GetAll()
}

func (s *UserService) SaveOrUpdateUser(user *models.User) error {
	if s.UserExists(user.TgID) {
		return s.UpdateUser(user)
	}
	return s.CreateUser(user)
}

func (s *UserService) GenerateUserCode(tgID int64) string {
	tgIDStr := fmt.Sprintf("%d", tgID)
	if len(tgIDStr) >= 4 {
		return tgIDStr[len(tgIDStr)-4:]
	}
	return "0000"
}

func (s *UserService) validateAndCleanUser(user *models.User) {
	if user.FirstName == "" {
		user.FirstName = "N/A"
	}
	if user.Username == "" {
		user.Username = "N/A"
	}
	if user.LanguageCode == "" {
		user.LanguageCode = "uz"
	}
}
