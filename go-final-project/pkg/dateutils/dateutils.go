package dateutils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("пустой запрос")
	}

	date, err := time.Parse("20060102", dstart)
	if err != nil {
		return "", fmt.Errorf("ошибка формата")
	}

	parts := strings.Split(repeat, " ")

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", fmt.Errorf("некоректный формат дней")
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval <= 0 || interval > 400 {
			return "", fmt.Errorf("некоректный формат года")
		}
		date = date.AddDate(0, 0, interval)
		for !date.After(now) {
			date = date.AddDate(0, 0, interval)
		}

	case "y":
		date = date.AddDate(1, 0, 0)
		if date.Month() == time.February && date.Day() == 28 && dstart[4:8] == "0229" {
			date = date.AddDate(0, 0, 1)
		}
		for !date.After(now) {
			date = date.AddDate(1, 0, 0)
			if date.Month() == time.February && date.Day() == 28 && dstart[4:8] == "0229" {
				date = date.AddDate(0, 0, 1)
			}
		}

	default:
		return "", fmt.Errorf("неподерживаемые форматы")
	}

	return date.Format("20060102"), nil
}
