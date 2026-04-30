package models

import "time"

type Reminder struct {
	ID     int
	ChatID int64
	Text   string
	Time   time.Time
}

type UserState struct {
	Step     string
	Reminder Reminder
}
