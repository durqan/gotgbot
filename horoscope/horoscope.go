package horoscope

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var Signs = []string{
	"♈ Овен", "♉ Телец", "♊ Близнецы", "♋ Рак",
	"♌ Лев", "♍ Дева", "♎ Весы", "♏ Скорпион",
	"♐ Стрелец", "♑ Козерог", "♒ Водолей", "♓ Рыбы",
}

var signMap = map[string]string{
	"♈ Овен":     "aries",
	"♉ Телец":    "taurus",
	"♊ Близнецы": "gemini",
	"♋ Рак":      "cancer",
	"♌ Лев":      "leo",
	"♍ Дева":     "virgo",
	"♎ Весы":     "libra",
	"♏ Скорпион": "scorpio",
	"♐ Стрелец":  "sagittarius",
	"♑ Козерог":  "capricorn",
	"♒ Водолей":  "aquarius",
	"♓ Рыбы":     "pisces",
}

type HoroscopeResponse struct {
	DateRange     string `json:"date_range"`
	CurrentDate   string `json:"current_date"`
	Description   string `json:"description"`
	Compatibility string `json:"compatibility"`
	Mood          string `json:"mood"`
	Color         string `json:"color"`
	LuckyNumber   string `json:"lucky_number"`
	LuckyTime     string `json:"lucky_time"`
}

func GetDailyHoroscope(signName string) (*HoroscopeResponse, error) {
	englishSign, ok := signMap[signName]
	if !ok {
		return nil, fmt.Errorf("неизвестный знак: %s", signName)
	}

	url := fmt.Sprintf("https://aztro.sameerkumar.website?sign=%s&day=today", englishSign)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API вернул статус: %d", resp.StatusCode)
	}

	var horoscope HoroscopeResponse
	if err := json.NewDecoder(resp.Body).Decode(&horoscope); err != nil {
		return nil, err
	}

	return &horoscope, nil
}

func FormatHoroscope(signName string, h *HoroscopeResponse) string {
	return fmt.Sprintf(
		"🔮 *%s* — гороскоп на %s\n\n"+
			"📖 %s\n\n"+
			"✨ *Совместимость:* %s\n"+
			"🎭 *Настроение:* %s\n"+
			"🎨 *Цвет:* %s\n"+
			"🔢 *Счастливое число:* %s\n"+
			"⏰ *Удачное время:* %s",
		signName, h.CurrentDate, h.Description,
		h.Compatibility, h.Mood, h.Color, h.LuckyNumber, h.LuckyTime,
	)
}
