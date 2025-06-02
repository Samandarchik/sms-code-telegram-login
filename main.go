package main

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

// User modeli
type User struct {
	TgID         int64
	FirstName    string
	Username     string
	LanguageCode string
	PhoneNumber  string
}

func main() {
	botToken := "7609705273:AAFX60_khniloe_ExejY4VRJdxEmeP4aloQ" // ❗ Bot tokeningizni shu yerga yozing
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// SQLite bazaga ulanish
	db, err := sql.Open("sqlite3", "./user.db")
	if err != nil {
		log.Fatal("DB ulanishda xato:", err)
	}
	defer db.Close()

	createTable(db)

	log.Printf("🤖 Bot %s ishga tushdi", bot.Self.UserName)
	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		// Telefon raqami kelganda
		if update.Message.Contact != nil {
			contact := update.Message.Contact
			user := User{
				TgID:         contact.UserID,
				FirstName:    update.Message.From.FirstName,
				Username:     update.Message.From.UserName,
				LanguageCode: update.Message.From.LanguageCode,
				PhoneNumber:  contact.PhoneNumber,
			}
			if !userExists(db, user.TgID) {
				saveUser(db, user)

				// Telegram ID dan oxirgi 4 raqamni olish
				tgIDStr := fmt.Sprintf("%d", user.TgID)
				code := "0000"
				if len(tgIDStr) >= 4 {
					code = tgIDStr[len(tgIDStr)-4:]
				}

				msg := tgbotapi.NewMessage(chatID,
					fmt.Sprintf("✅ Raqamingiz saqlandi!\nSizning code:\n```%s```", code))
				msg.ParseMode = "Markdown"
				bot.Send(msg)

			} else {
				// Foydalanuvchi mavjud bo‘lsa — yana code ni yuborish
				tgIDStr := fmt.Sprintf("%d", user.TgID)
				code := "0000"
				if len(tgIDStr) >= 4 {
					code = tgIDStr[len(tgIDStr)-4:]
				}

				msg := tgbotapi.NewMessage(chatID,
					fmt.Sprintf("ℹ️ Siz allaqachon ro‘yxatdan o‘tgansiz.\nSizning code:\n```%s```", code))
				msg.ParseMode = "Markdown"
				bot.Send(msg)
			}

			continue
		}

		// /start buyrug‘i
		if update.Message.IsCommand() && update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(chatID, "👋 Salom! Iltimos, telefon raqamingizni yuboring:")
			button := tgbotapi.NewKeyboardButtonContact("📱 Telefon raqamni yuborish")
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(button),
			)
			keyboard.OneTimeKeyboard = true
			keyboard.ResizeKeyboard = true
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		}
	}
}

// Jadval yaratish
func createTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		userid INTEGER PRIMARY KEY AUTOINCREMENT,
		tg_id INTEGER UNIQUE,
		first_name TEXT,
		username TEXT,
		language_code TEXT,
		phone TEXT
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Jadval yaratishda xatolik:", err)
	}
}

// Foydalanuvchini bazaga saqlash
func saveUser(db *sql.DB, user User) {
	stmt, err := db.Prepare(`
	INSERT INTO users(tg_id, first_name, username, language_code, phone)
	VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Println("Saqlashda xatolik (prepare):", err)
		return
	}
	_, err = stmt.Exec(user.TgID, user.FirstName, user.Username, user.LanguageCode, user.PhoneNumber)
	if err != nil {
		log.Println("Saqlashda xatolik (exec):", err)
	}
}

// Foydalanuvchi mavjudligini tekshirish
func userExists(db *sql.DB, tgID int64) bool {
	row := db.QueryRow("SELECT tg_id FROM users WHERE tg_id = ?", tgID)
	var id int64
	err := row.Scan(&id)
	return err == nil
}

// Foydalanuvchilar sonini olish
func getUserCount(db *sql.DB) int {
	row := db.QueryRow("SELECT COUNT(*) FROM users")
	var count int
	err := row.Scan(&count)
	if err != nil {
		log.Println("COUNT olishda xatolik:", err)
		return 0
	}
	return count
}
