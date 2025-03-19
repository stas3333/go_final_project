package main

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	parseDate, err := time.Parse(DateFormat, date)
	if err != nil {
		return "", err
	}
	if repeat == "" {
		return "", errors.New("правило repeat отсутствует")
	} else if strings.Contains(repeat, "d ") {
		days, err := strconv.Atoi(strings.TrimPrefix(repeat, "d "))
		if err != nil {
			return "", err
		}
		if days > 400 {
			return "", errors.New("максимальное количество дней не должно превышать 400")
		}
		newDate := parseDate.AddDate(0, 0, days)
		for newDate.Before(now) {
			newDate = newDate.AddDate(0, 0, days)
		}
		return newDate.Format(DateFormat), nil

	} else if repeat == "y" {
		newDate := parseDate.AddDate(1, 0, 0)
		for newDate.Before(now) {
			newDate = newDate.AddDate(1, 0, 0)
		}
		return newDate.Format(DateFormat), nil
	} else {
		return "", errors.New("неверный формат repeat")
	}
}
