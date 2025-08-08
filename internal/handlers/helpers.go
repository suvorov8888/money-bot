package handlers

import "time"

// GetStartAndEndOfDay возвращает начало и конец текущего дня
func GetStartAndEndOfDay() (time.Time, time.Time) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24*time.Hour - time.Second)
	return startOfDay, endOfDay
}

// GetStartAndEndOfWeek возвращает начало и конец текущей недели
func GetStartAndEndOfWeek() (time.Time, time.Time) {
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, int(time.Monday-now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, now.Location())
	endOfWeek := startOfWeek.AddDate(0, 0, 7)
	return startOfWeek, endOfWeek
}

// GetStartAndEndOfMonth возвращает начало и конец текущего месяца
func GetStartAndEndOfMonth() (time.Time, time.Time) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	endOfMonth = time.Date(endOfMonth.Year(), endOfMonth.Month(), endOfMonth.Day(), 23, 59, 59, 0, now.Location())
	return startOfMonth, endOfMonth
}
