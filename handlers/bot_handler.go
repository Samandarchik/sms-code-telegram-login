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
	chatID := update.Message.Chat.ID

	user := h.extractUserFromContact(update)

	// SaveOrUpdateUser endi 2 ta qiymat qaytaradi
	savedUser, err := h.userService.SaveOrUpdateUser(user) // <-- Shu qatorni o'zgartirdik
	if err != nil {
		log.Printf("Foydalanuvchini saqlashda yoki yangilashda xatolik: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ma'lumotlarni saqlashda xatolik yuz berdi. Qaytadan urinib ko'ring.")
		h.bot.Send(msg)
		return
	}

	// GenerateUserCode, GetUserCount, UserExists funksiyalari UserService ichida bo'lishi kerak.
	// Agar service/user_service.go ichida bu funksiyalar to'g'ridan-to'g'ri
	// repository.UserRepository metodlarini chaqirsa ham,
	// BotHandler ularni UserService orqali chaqirishi kerak.
	// Men service/user_service.go da bu funksiyalarni yana qayta kiritib beraman.

	code := h.userService.GenerateUserCode(savedUser.TelegramID) // <-- savedUser.TelegramID ishlatildi
	userCount, err := h.userService.GetUserCount()               // <-- GetUserCount endi error qaytaradi
	if err != nil {
		log.Printf("Foydalanuvchilar sonini olishda xatolik: %v", err)
		// Xato bo'lsa ham ishni davom ettirish yoki xato xabarini yuborish mumkin
		userCount = 0 // Default qiymat
	}

	var messageText string
	// UserExists endi GetUserByID ni chaqirishi kerak, yoki UserService ichida bo'lishi kerak.
	// UserService ichida UserExists funksiyasi allaqachon bor.
	if h.userService.UserExists(savedUser.TelegramID) { // <-- savedUser.TelegramID ishlatildi
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
	count, err := h.userService.GetUserCount() // <-- GetUserCount endi error qaytaradi
	if err != nil {
		log.Printf("Foydalanuvchilar statistikasini olishda xatolik: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Statistikalarni olishda xatolik yuz berdi.")
		h.bot.Send(msg)
		return
	}
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📊 Jami ro'yxatdan o'tgan foydalanuvchilar: %d", count))
	h.bot.Send(msg)
}

func (h *BotHandler) extractUserFromContact(update tgbotapi.Update) *models.User {
	contact := update.Message.Contact
	from := update.Message.From

	var telegramID int64 // userID o'rniga telegramID ishlatamiz
	if contact.UserID != 0 {
		telegramID = contact.UserID
	} else {
		telegramID = from.ID
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
		TelegramID:   telegramID, // <-- o'zgartirildi
		FirstName:    firstName,
		Username:     username,
		LanguageCode: languageCode,
		PhoneNumber:  contact.PhoneNumber,
		// LastName:     from.LastName, // Agar kerak bo'lsa
	}
}
