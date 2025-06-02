package main

import (
	"database/sql"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID           int64
	FirstName    string
	Username     string
	LanguageCode string
	PhoneNumber  string
}

func main() {
	botToken := "7609705273:AAFX60_khniloe_ExejY4VRJdxEmeP4aloQ"
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// SQLite DB ulanish
	db, err := sql.Open("sqlite3", "./user.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable(db)

	bot.Debug = true
	log.Println("Bot ishga tushdi...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		// Telefon raqami yuborilgan bo‘lsa
		if update.Message.Contact != nil {
			contact := update.Message.Contact
			user := User{
				ID:           contact.UserID,
				FirstName:    update.Message.From.FirstName,
				Username:     update.Message.From.UserName,
				LanguageCode: update.Message.From.LanguageCode,
				PhoneNumber:  contact.PhoneNumber,
			}

			if !userExists(db, user.ID) {
				saveUser(db, user)
				msg := tgbotapi.NewMessage(chatID, "Raqamingiz saqlandi. Rahmat!")
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Siz allaqachon ro‘yxatdan o‘tgansiz.")
				bot.Send(msg)
			}
			continue
		}

		// /start komandasi
		if update.Message.IsCommand() && update.Message.Command() == "start" {
			// Telefon raqamini so‘rash
			msg := tgbotapi.NewMessage(chatID, "Iltimos, telefon raqamingizni yuboring:")
			button := tgbotapi.NewKeyboardButtonContact("📱 Raqamni yuborish")
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

func createTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		first_name TEXT,
		username TEXT,
		language_code TEXT,
		phone TEXT
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Jadval yaratishda xatolik:", err)
	}
}

func saveUser(db *sql.DB, user User) {
	stmt, err := db.Prepare(`INSERT INTO users(id, first_name, username, language_code, phone) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		log.Println("Insert prepare error:", err)
		return
	}
	_, err = stmt.Exec(user.ID, user.FirstName, user.Username, user.LanguageCode, user.PhoneNumber)
	if err != nil {
		log.Println("Insert error:", err)
	}
}

func userExists(db *sql.DB, userID int64) bool {
	row := db.QueryRow("SELECT id FROM users WHERE id = ?", userID)
	var id int64
	err := row.Scan(&id)
	return err == nil
}
