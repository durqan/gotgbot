package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"tgbot/commands"
	"tgbot/horoscope"
	"tgbot/models"
	"tgbot/storage"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		handleCallback(bot, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	log.Printf("Пришло сообщение от %d: %s", chatID, update.Message.Text)

	if !storage.CanSendMessage(chatID) {
		log.Printf("Антиспам: пропущено сообщение от %d", chatID)
		return
	}
	storage.UpdateLastMessage(chatID)

	if len(update.Message.Text) > 500 {
		msg := tgbotapi.NewMessage(chatID, "Слишком длинное сообщение (макс 500 символов)")
		bot.Send(msg)
		return
	}

	var response string

	if state := storage.GetUserState(chatID); state != nil {
		response = handleDialog(bot, chatID, state, update.Message.Text)
	} else if update.Message.IsCommand() {
		response = handleCommand(bot, chatID, update.Message)
	} else {
		response = "Используй команды: /help для списка команд"
	}

	log.Printf("Ответ для %d: %s", chatID, response)
	msg := tgbotapi.NewMessage(chatID, response)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки для %d: %s", chatID, err)
	}
}

func handleCommand(bot *tgbotapi.BotAPI, chatID int64, message *tgbotapi.Message) string {
	switch message.Command() {
	case "start":
		sendStartMenu(bot, chatID)
		return ""
	case "help":
		sendHelpMenu(bot, chatID)
		return ""
	case "remind":
		return commands.HandleRemind(bot, chatID, message.Text)
	case "reminders":
		return commands.GetRemindersList(chatID)
	case "horoscope":
		return handleHoroscopeList(bot, chatID)
	case "sethoroscope":
		sendSignKeyboard(bot, chatID)
		return ""
	case "myhoroscope":
		return commands.HandleMyHoroscope(chatID)
	case "cancel":
		words := strings.Split(message.Text, " ")
		if len(words) < 2 {
			return "Формат: /cancel ID\nПример: /cancel 1"
		}
		id, err := strconv.Atoi(words[1])
		if err != nil {
			return "ID должен быть числом"
		}
		return commands.HandleCancel(chatID, id)
	default:
		return "Неизвестная команда. Используй /help"
	}
}

func handleDialog(bot *tgbotapi.BotAPI, chatID int64, state *models.UserState, text string) string {
	switch state.Step {
	case "waiting_text":
		state.Reminder.Text = text
		state.Step = "waiting_minutes"
		return "⏰ Через сколько минут напомнить?\n(Отправь число)"
	case "waiting_minutes":
		minutes, err := strconv.Atoi(text)
		if err != nil || minutes <= 0 {
			return "❌ Пожалуйста, отправь положительное число (например: 5)"
		}

		reminder := models.Reminder{
			ChatID: chatID,
			Text:   state.Reminder.Text,
			Time:   time.Now().Add(time.Duration(minutes) * time.Minute),
		}

		id := storage.AddReminder(chatID, reminder)
		storage.DeleteUserState(chatID)

		go func() {
			time.Sleep(time.Duration(minutes) * time.Minute)
			msg := tgbotapi.NewMessage(chatID, "🔔 НАПОМИНАНИЕ: "+reminder.Text)
			bot.Send(msg)
		}()

		return fmt.Sprintf("✅ Напоминание \"%s\" установлено на %d минут (ID: %d)", reminder.Text, minutes, id)
	}
	return "Что-то пошло не так"
}

func sendStartMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Привет! Я бот напоминалок")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить напоминание", "add"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Мои напоминания", "list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Отменить напоминание", "cancel"),
			tgbotapi.NewInlineKeyboardButtonData("🔮 Гороскоп", "horoscope"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "help"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func sendHelpMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "📋 Доступные команды:\n/start - главное меню\n/remind 5 текст - быстрая команда\n/reminders - список напоминаний\n/cancel ID - отменить\n/horoscope - гороскоп\n/sethoroscope - выбрать знак\n/myhoroscope - мой гороскоп")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить", "add"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Список", "list"),
			tgbotapi.NewInlineKeyboardButtonData("🔮 Гороскоп", "horoscope"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func sendSignKeyboard(bot *tgbotapi.BotAPI, chatID int64) {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(horoscope.Signs); i += 3 {
		row := []tgbotapi.InlineKeyboardButton{}
		for j := i; j < i+3 && j < len(horoscope.Signs); j++ {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(horoscope.Signs[j], "sign_"+horoscope.Signs[j]))
		}
		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, "🔮 Выбери свой знак зодиака:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func handleHoroscopeList(bot *tgbotapi.BotAPI, chatID int64) string {
	sendSignKeyboard(bot, chatID)
	return "🔮 Выбери знак зодиака, чтобы узнать гороскоп:"
}

func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	data := query.Data

	var response string

	switch data {
	case "add":
		storage.SetUserState(chatID, &models.UserState{Step: "waiting_text"})
		response = "📝 Напиши текст напоминания"
	case "list":
		response = commands.GetRemindersList(chatID)
	case "cancel":
		response = "❌ Отправь ID напоминания, которое нужно отменить\n(Например: 1)\n\nСписок ID можно посмотреть через кнопку \"📋 Мои напоминания\""
	case "help":
		response = "📋 Доступные команды:\n• /start - главное меню\n• /remind 5 текст - быстрая команда\n• /reminders - список напоминаний\n• /cancel ID - отменить\n• /horoscope - гороскоп\n• /sethoroscope - выбрать знак\n• /myhoroscope - мой гороскоп"
	case "horoscope":
		response = handleHoroscopeList(bot, chatID)
		bot.Send(tgbotapi.NewCallback(query.ID, ""))
		msg := tgbotapi.NewMessage(chatID, response)
		bot.Send(msg)
		return
	default:
		if strings.HasPrefix(data, "sign_") {
			sign := strings.TrimPrefix(data, "sign_")
			response = commands.HandleHoroscope(chatID, sign)
		} else if id, err := strconv.Atoi(data); err == nil {
			response = commands.HandleCancel(chatID, id)
		} else {
			response = "Неизвестная команда"
		}
	}

	bot.Send(tgbotapi.NewCallback(query.ID, ""))
	msg := tgbotapi.NewMessage(chatID, response)
	bot.Send(msg)
}
