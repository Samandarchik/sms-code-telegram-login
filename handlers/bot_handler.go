// ==================== handlers/bot_handler.go ====================
package handlers

import (
	"amur/models"
	"amur/service"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot         *tgbotapi.BotAPI
	userService *service.UserService
}

func NewBotHandler(bot *tgbotapi.BotAPI, userService *service.UserService) *BotHandler {
	return &BotHandler{
		bot:         bot,
		userService: userService,
	}
}

func (h *BotHandler) HandleStart(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "👋 Salom! Iltimos, telefon raqamingizni yuboring:")
	button := tgbotapi.NewKeyboardButtonContact("📱 Telefon raqamni yuborish")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(button),
	)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *BotHandler) HandleContact(update tgbotapi.Update) {
	// contact := update.Message.Contact
	chatID := update.Message.Chat.ID

	user := h.extractUserFromContact(update)

	err := h.userService.SaveOrUpdateUser(user)
	if err != nil {
		log.Printf("Foydalanuvchini saqlashda xatolik: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ma'lumotlarni saqlashda xatolik yuz berdi. Qaytadan urinib ko'ring.")
		h.bot.Send(msg)
		return
	}

	code := h.userService.GenerateUserCode(user.TelegramID)
	userCount := h.userService.GetUserCount()

	var messageText string
	if h.userService.UserExists(user.TelegramID) {
		messageText = fmt.Sprintf("ℹ️ Ma'lumotlaringiz yangilandi.\nSizning code:\n```%s```", code)
	} else {
		messageText = fmt.Sprintf("✅ Raqamingiz saqlandi!\nSizning code:\n```%s```\n\n📊 Jami foydalanuvchilar: %d",
			code, userCount)
	}

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

func (h *BotHandler) HandleStats(chatID int64) {
	count := h.userService.GetUserCount()
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📊 Jami ro'yxatdan o'tgan foydalanuvchilar: %d", count))
	h.bot.Send(msg)
}

func (h *BotHandler) extractUserFromContact(update tgbotapi.Update) *models.User {
	contact := update.Message.Contact
	from := update.Message.From

	var userID int64
	if contact.UserID != 0 {
		userID = contact.UserID
	} else {
		userID = from.ID
	}

	var firstName string
	if from.FirstName != "" {
		firstName = from.FirstName
	} else if contact.FirstName != "" {
		firstName = contact.FirstName
	} else {
		firstName = "N/A"
	}

	var username string
	if from.UserName != "" {
		username = from.UserName
	} else {
		username = "N/A"
	}

	var languageCode string
	if from.LanguageCode != "" {
		languageCode = from.LanguageCode
	} else {
		languageCode = "uz"
	}

	return &models.User{
		TelegramID:   userID,
		FirstName:    firstName,
		Username:     username,
		LanguageCode: languageCode,
		PhoneNumber:  contact.PhoneNumber,
	}
}
