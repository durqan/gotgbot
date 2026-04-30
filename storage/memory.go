package storage

import (
	"tgbot/models"
	"time"
)

var (
	Reminders   = make(map[int64][]models.Reminder)
	UserStates  = make(map[int64]*models.UserState)
	LastMessage = make(map[int64]time.Time)
	NextID      = 1
)

func AddReminder(chatID int64, reminder models.Reminder) int {
	reminder.ID = NextID
	Reminders[chatID] = append(Reminders[chatID], reminder)
	NextID++
	return reminder.ID
}

func GetReminders(chatID int64) []models.Reminder {
	return Reminders[chatID]
}

func DeleteReminder(chatID int64, id int) bool {
	userReminders := Reminders[chatID]
	for i, r := range userReminders {
		if r.ID == id {
			Reminders[chatID] = append(userReminders[:i], userReminders[i+1:]...)
			return true
		}
	}
	return false
}

func GetUserState(chatID int64) *models.UserState {
	return UserStates[chatID]
}

func SetUserState(chatID int64, state *models.UserState) {
	UserStates[chatID] = state
}

func DeleteUserState(chatID int64) {
	delete(UserStates, chatID)
}

func UpdateLastMessage(chatID int64) {
	LastMessage[chatID] = time.Now()
}

func CanSendMessage(chatID int64) bool {
	if last, ok := LastMessage[chatID]; ok && time.Since(last) < time.Second {
		return false
	}
	return true
}
