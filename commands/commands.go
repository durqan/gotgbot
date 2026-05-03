package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"tgbot/models"
	"tgbot/storage"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleRemind(bot *tgbotapi.BotAPI, chatID int64, text string) string {
	words := strings.Split(text, " ")
	if len(words) < 3 {
		return "Формат: /remind 5 текст напоминания\nПример: /remind 10 купить хлеб"
	}

	minutes, err := strconv.Atoi(words[1])
	if err != nil || minutes <= 0 {
		return "Укажите число минут (пример: /remind 5 купить хлеб)"
	}

	reminderText := strings.Join(words[2:], " ")
	id := storage.AddReminder(chatID, models.Reminder{
		ChatID: chatID,
		Text:   reminderText,
		Time:   time.Now().Add(time.Duration(minutes) * time.Minute),
	})

	go func() {
		time.Sleep(time.Duration(minutes) * time.Minute)
		msg := tgbotapi.NewMessage(chatID, "🔔 НАПОМИНАНИЕ: "+reminderText)
		bot.Send(msg)

	}()

	return fmt.Sprintf("✅ Напоминание установлено!\n📝 %s\n🆔 ID: %d\n⏰ Через %d минут", reminderText, id, minutes)
}

func GetRemindersList(chatID int64) string {
	userReminders := storage.GetReminders(chatID)
	if len(userReminders) == 0 {
		return "📭 У тебя нет активных напоминаний"
	}

	var result strings.Builder
	result.WriteString("📋 Твои напоминания:\n\n")
	for _, r := range userReminders {
		result.WriteString(fmt.Sprintf("🆔 *%d*\n   📝 %s\n   ⏰ %s\n\n",
			r.ID, r.Text, r.Time.Format("15:04 02.01.2006")))
	}
	return result.String()
}

func HandleCancel(chatID int64, id int) string {
	if storage.DeleteReminder(chatID, id) {
		return fmt.Sprintf("✅ Напоминание #%d отменено", id)
	}
	return "❌ Напоминание с таким ID не найдено.\nИспользуй /reminders для просмотра"
}
