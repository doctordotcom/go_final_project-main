package rules

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату для задачи согласно заданным правилам.
func NextDate(now time.Time, date string, repeat string) (string, error) {

	// Проверка пустой строки в repeat
	if repeat == "" {
		return "", errors.New("правило повторения не указано")
	}

	// Проверка и парсинг исходной даты
	startDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", errors.New("некорректный формат даты: " + date)
	}

	// Определение следующей даты на основе правила повторения
	var nextDate time.Time
	repeatSlice := strings.Split(repeat, " ")
	if len(repeatSlice) < 1 || len(repeatSlice) > 2 {
		return "", errors.New("неверный формат правила повторения: " + repeat)
	}
	switch repeatSlice[0] {
	case "d":
		if len(repeatSlice) != 1 {
			days, err := strconv.Atoi(repeatSlice[1])
			if err != nil {
				return "", errors.New("некорректный формат")
			}
			if days < 1 || days > 400 {
				return "", errors.New("число дней должно быть от 1 до 400")
			}
			nextDate = startDate.AddDate(0, 0, days)
			for nextDate.Before(now) {
				nextDate = nextDate.AddDate(0, 0, days)
			}
		}

	case "y":
		nextDate = startDate.AddDate(1, 0, 0)
		for nextDate.Before(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}

	default:
		return "", errors.New("неподдерживаемый формат правила повторения: " + repeat)
	}

	// Проверка, больше ли следующая дата указанного времени
	if nextDate.Before(now) {
		return "", errors.New("следующая дата должна быть позже текущей даты")
	}
	return nextDate.Format("20060102"), nil
}

// // Обработка правила повторения
// var newDate time.Time
// var err error
// switch task.Repeat {
// case "":
// 	// Задача будет удалена после выполнения
// 	newDate, err = time.Parse("20060102", task.Date)

// case "y":
// 	// Ежегодное выполнение
// 	newDate, err = time.Parse("20060102", task.Date)
// 	newDate = newDate.AddDate(1, 0, 0)

// default:
// 	if len(task.Repeat) > 2 && task.Repeat[:2] == "d " {
// 		days, err := strconv.Atoi(task.Repeat[2:])
// 		if err != nil || days > 400 {
// 			http.Error(w, "Invalid repeat value", http.StatusBadRequest)
// 			return
// 		}
// 		newDate, err = time.Parse("20060102", task.Date)
// 		newDate = newDate.AddDate(0, 0, days)
// 	} else {
// 		http.Error(w, "Invalid repeat rule", http.StatusBadRequest)
// 		return
// 	}
// }

// if err != nil {
// 	http.Error(w, "Error parsing date: "+err.Error(), http.StatusInternalServerError)
// 	return
// }
