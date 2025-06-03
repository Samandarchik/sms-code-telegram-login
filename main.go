// main.go
package main

import (
	"amur/config"
	"amur/database"
	"amur/handlers"
	"amur/repository"
	"amur/routes"
	"amur/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Konfiguratsiyani yuklash
	cfg := config.LoadConfig()

	// PostgreSQL ma'lumotlar bazasini sozlash
	connStr := "host=" + cfg.DatabaseHost +
		" port=" + cfg.DatabasePort +
		" user=" + cfg.DatabaseUser +
		" password=" + cfg.DatabasePassword +
		" dbname=" + cfg.DatabaseName +
		" sslmode=disable"
	db, err := database.NewDatabase(connStr)
	if err != nil {
		log.Fatalf("PostgreSQL ma'lumotlar bazasiga ulanishda xatolik: %v", err)
	}
	defer db.Close()
	log.Println("✅ PostgreSQL ma'lumotlarlar bazasiga muvaffaqiyatli ulanildi.")

	// Repository'larni yaratish
	userRepo := repository.NewUserRepository(db.GetDB())
	foodRepo := repository.NewFoodRepository(db.GetDB())
	basketOrderRepo := repository.NewBasketOrderRepository(db.GetDB())
	orderRepo := repository.NewOrderRepository(db.GetDB())

	// Service'larni yaratish
	userService := service.NewUserService(userRepo)
	foodService := service.NewFoodService(foodRepo)
	basketOrderService := service.NewBasketOrderService(basketOrderRepo, foodRepo)
	orderService := service.NewOrderService(orderRepo, basketOrderRepo, foodRepo)

	// Handler'larni yaratish
	userHandler := handlers.NewUserHandler(userService)
	foodHandler := handlers.NewFoodHandler(foodService)
	basketOrderHandler := handlers.NewBasketOrderHandler(basketOrderService)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Telegram botni sozlash
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Botni yaratishda xatolik: %v", err)
	}

	bot.Debug = false
	log.Printf("🤖 Bot @%s sifatida ishga tushdi", bot.Self.UserName)

	botHandler := handlers.NewBotHandler(bot, userService)

	// HTTP serverni sozlash
	router := routes.SetupRoutes(foodHandler, userHandler, basketOrderHandler, orderHandler)
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Goroutine'da HTTP serverni ishga tushirish
	go func() {
		log.Printf("🌐 HTTP server %s portda ishga tushdi", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server xatoligi: %v", err)
		}
	}()

	// Telegram bot update'larini olish
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Graceful shutdown uchun
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Asosiy loop (Telegram bot update'larini qayta ishlash)
	go func() {
		for update := range updates {
			if update.Message != nil {
				chatID := update.Message.Chat.ID

				switch {
				case update.Message.IsCommand():
					switch update.Message.Command() {
					case "start":
						botHandler.HandleStart(chatID)
					case "stats":
						botHandler.HandleStats(chatID)
					default:
						msg := tgbotapi.NewMessage(chatID, "❓ Noma'lum buyruq. /start bosing.")
						bot.Send(msg)
					}

				case update.Message.Contact != nil:
					botHandler.HandleContact(update)

				default:
					msg := tgbotapi.NewMessage(chatID, "📱 Iltimos, telefon raqamingizni yuboring yoki /start bosing.")
					bot.Send(msg)
				}
			}
		}
	}()

	log.Println("✅ Bot va API server ishga tushdi. To'xtatish uchun Ctrl+C bosing.")

	// Graceful shutdown
	<-quit
	log.Println("🛑 Server to'xtatilmoqda...")

	// HTTP serverni to'xtatish
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server to'xtatishda xatolik: %v", err)
	}

	// Botni to'xtatish
	bot.StopReceivingUpdates()

	log.Println("✅ Server muvaffaqiyatli to'xtatildi.")
}
