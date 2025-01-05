package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/go-ping/ping"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	VPNClientIP             string `json:"vpn_client_ip"`
	LogDir                  string `json:"log_dir"`
	PingInterval            int    `json:"ping_interval"`
	LogRetentionDays        int    `json:"log_retention_days"`
	TelegramBotToken        string `json:"telegram_bot_token"`
	PingTimeout             int    `json:"ping_timeout"`
	AutoPingIntervalMinutes int    `json:"auto_ping_interval_minutes"`
	EnableAutoPing          bool   `json:"enable_auto_ping"`
}

var (
	config       Config
	lastPingTime time.Time
	configPath   = "config.json"
	chatIDs      sync.Map // Хранение ID чатов для рассылки уведомлений
)

//go:embed assets/icon.ico
var iconData []byte

func main() {
	// Загрузка конфигурации
	loadConfig()

	// Создание логов
	setupLogs()

	// Запуск Telegram-бота
	go startTelegramBot()

	// Запуск трея
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTooltip("VPN Статус")
	quitItem := systray.AddMenuItem("Выйти", "Выход из программы")

	go func() {
		<-quitItem.ClickedCh
		systray.Quit()
	}()
}

func onExit() {}

func startTelegramBot() {
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Fatalf("Ошибка запуска Telegram-бота: %v", err)
	}

	bot.Debug = false
	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Ошибка получения обновлений Telegram: %v", err)
	}

	if config.EnableAutoPing {
		go startAutoPing(bot)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		chatIDs.Store(chatID, true) // Сохраняем ID чата для рассылки

		switch strings.ToLower(update.Message.Text) {
		case "/status":
			handleStatusCommandAsync(bot, chatID)
		case "/enable_autoping":
			config.EnableAutoPing = true
			saveConfig()
			msg := tgbotapi.NewMessage(chatID, "Автоматический пинг включен.")
			bot.Send(msg)
		case "/disable_autoping":
			config.EnableAutoPing = false
			saveConfig()
			msg := tgbotapi.NewMessage(chatID, "Автоматический пинг отключен.")
			bot.Send(msg)
		}
	}
}

func handleStatusCommandAsync(bot *tgbotapi.BotAPI, chatID int64) {
	if time.Since(lastPingTime) < time.Duration(config.PingInterval)*time.Second {
		waitTime := config.PingInterval - int(time.Since(lastPingTime).Seconds())
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Подождите %d секунд перед следующей проверкой.", waitTime))
		bot.Send(msg)
		return
	}

	lastPingTime = time.Now()
	msg := tgbotapi.NewMessage(chatID, "Проверяю статус, подождите...")
	bot.Send(msg)

	response := handleStatusCommand()
	resultMsg := tgbotapi.NewMessage(chatID, response)
	bot.Send(resultMsg)
}

func handleStatusCommand() string {
	if isClientOnline(config.VPNClientIP) {
		lastPingTime = time.Now()
		logStatus("Клиент в сети.")
		return "Клиент в сети."
	}

	logStatus("Клиент не в сети.")
	return "Клиент не в сети."
}

func isClientOnline(ip string) bool {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		log.Printf("Ошибка создания пингера: %v", err)
		return false
	}

	pinger.Count = 3
	pinger.Timeout = time.Duration(config.PingTimeout) * time.Second
	pinger.SetPrivileged(true)

	err = pinger.Run()
	if err != nil {
		log.Printf("Ошибка выполнения пинга для IP %s: %v", ip, err)
		return false
	}

	stats := pinger.Statistics()
	log.Printf("Пинг: отправлено %d, получено %d, потеряно %d, среднее время %.2f ms",
		stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss, stats.AvgRtt.Milliseconds())

	return stats.PacketsRecv > 0
}

func startAutoPing(bot *tgbotapi.BotAPI) {
	ticker := time.NewTicker(time.Duration(config.AutoPingIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		if config.EnableAutoPing {
			response := handleStatusCommand()
			chatIDs.Range(func(key, value any) bool {
				chatID := key.(int64)
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Автоматический статус: %s", response))
				bot.Send(msg)
				return true
			})
		}
	}
}

func setupLogs() {
	if _, err := os.Stat(config.LogDir); os.IsNotExist(err) {
		os.Mkdir(config.LogDir, 0755)
	}

	go func() {
		for {
			time.Sleep(24 * time.Hour)
			rotateLogs()
		}
	}()
}

func logStatus(status string) {
	logFile := filepath.Join(config.LogDir, fmt.Sprintf("%s.log", time.Now().Format("02-01-2006")))

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Ошибка записи в лог: %v", err)
		return
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Println(status)
}

func rotateLogs() {
	files, err := os.ReadDir(config.LogDir)
	if err != nil {
		log.Printf("Ошибка чтения логов: %v", err)
		return
	}

	now := time.Now()
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(config.LogDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Ошибка получения информации о файле: %v", err)
			continue
		}

		if now.Sub(info.ModTime()) > time.Duration(config.LogRetentionDays)*24*time.Hour {
			os.Remove(filePath)
			log.Printf("Удален устаревший лог: %s", file.Name())
		}
	}
}

func loadConfig() {
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Конфигурационный файл не найден. Создаю новый с настройками по умолчанию...")
			createDefaultConfig()
			log.Println("Конфигурационный файл создан. Загрузка конфигурации...")
			loadConfig()
			return
		} else {
			log.Fatalf("Ошибка открытия конфигурационного файла: %v", err)
		}
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Ошибка декодирования конфигурационного файла: %v", err)
	}
}

func createDefaultConfig() {
	config = Config{
		VPNClientIP:             "10.9.0.2",
		LogDir:                  "logs",
		PingInterval:            10,
		LogRetentionDays:        5,
		TelegramBotToken:        "ВАШ_ТЕЛЕГРАМ_ТОКЕН",
		PingTimeout:             10,
		AutoPingIntervalMinutes: 30,
		EnableAutoPing:          true,
	}

	file, err := os.Create(configPath)
	if err != nil {
		log.Fatalf("Ошибка создания конфигурационного файла: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&config); err != nil {
		log.Fatalf("Ошибка записи конфигурационного файла: %v", err)
	}
}

func saveConfig() {
	file, err := os.Create(configPath)
	if err != nil {
		log.Printf("Ошибка сохранения конфигурационного файла: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&config); err != nil {
		log.Printf("Ошибка записи конфигурационного файла: %v", err)
	}
}
