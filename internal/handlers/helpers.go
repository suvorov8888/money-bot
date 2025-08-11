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
	// Корректно вычисляем смещение до понедельника
	offset := int(time.Monday - now.Weekday())
	if offset > 0 { // Если сегодня воскресенье (в Go это 0), то смещение будет +1, что неверно.
		offset = -6 // Для воскресенья нужно отнять 6 дней, чтобы получить прошлый понедельник.
	}

	startOfWeek := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, offset)
	endOfWeek := startOfWeek.AddDate(0, 0, 7).Add(-time.Nanosecond) // Конец недели - это следующее воскресенье 23:59:59...
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
