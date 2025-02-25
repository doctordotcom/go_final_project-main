package handlers

import (
	"database/sql"
	"encoding/json"
	"m/models"
	"m/rules"
	"net/http"
	"strings"
	"time"
)

// // функция для обработки ошибки
// func sendErrorResponse(w http.ResponseWriter, message string, status int) {
// 	http.Error(w, message, status)
// }

// Глобальная переменная для базы данных
var db *sql.DB

// getNextDateHandler - обработчик для "/api/nextdate", получение следующей даты
func GetNextDateHandler(w http.ResponseWriter, r *http.Request) {
	// проверка метода Get
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	// // проверка формата json
	// if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
	// 	log.Printf("запрос не содержит json")
	// 	sendErrorResponse(w, "запрос не содержит json", http.StatusUnsupportedMediaType)
	// 	return
	// }
	// Получение параметров запроса
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")
	// Проверка пустых параметров
	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, "Параметры now, date и repeat обязательны", http.StatusBadRequest)
		return
	}
	// Парсинг nowStr
	Now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "Ошибка парсинга параметра now", http.StatusBadRequest)
		return
	}
	// Парсинг dateStr
	Date, err := time.Parse("20060102", dateStr)
	if err != nil {
		http.Error(w, "Ошибка парсинга параметра date", http.StatusBadRequest)
		return
	}
	// Получение следующей даты
	NextDate, err := rules.NextDate(Now, Date.Format("20060102"), repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Вывод
	w.Write([]byte(NextDate))
}

// postNextTaskHandler - обработчик для добавления задачи
func PostNextTaskHandler(w http.ResponseWriter, r *http.Request) {
	// проверка метода Post
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	// // Проверка формата json
	// if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
	// 	http.Error(w, "Запрос не содержит JSON", http.StatusUnsupportedMediaType)
	// 	return
	// }
	// глобальная переменная для структуры Task из пакета models
	var task models.Task
	// Проверка, не пустое ли r.Body
	if r.Body == nil {
		http.Error(w, "Тело запроса не может быть пустым", http.StatusBadRequest)
		return
	}
	// Декодировка JSON-запроса в структуру Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		defer r.Body.Close()
		return
	}
	// Проверка обязательного поля Title
	if task.Title == "" {
		http.Error(w, "поле Title не может быть пустым", http.StatusBadRequest)
		return
	}
	// Получаем текущую дату
	today := time.Now().Format("20060102")
	todayDate, err := time.Parse("20060102", today)
	if err != nil {
		http.Error(w, "Ошибка парсинга текущей даты", http.StatusInternalServerError)
		return
	}
	var taskDate time.Time
	// Проверка даты и установка
	if task.Date == "" {
		// Если дата не указана, используем сегодняшнюю
		task.Date = today
	} else {
		// Проверка формата 20060102
		var err error
		taskDate, err = time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, "Неправильный формат даты, должно быть YYYYMMDD", http.StatusBadRequest)
			return
		}
		// Если дата меньше сегодня
		if taskDate.Before(time.Now()) {
			// Если правило повторения пустое, берётся сегодняшняя дата
			if task.Repeat == "" {
				task.Date = today
			} else {
				// Проверка правила повторения
				if strings.Contains(task.Repeat, "wm") {
					http.Error(w, "Лишь правила d (день) и y (год) допустимы", http.StatusBadRequest)
					return
				}
				// Функция NextDate вычисляет следующую дату с учетом repeat
				nextDateStr, err := rules.NextDate(todayDate, task.Date, task.Repeat)
				if err != nil {
					http.Error(w, "Ошибка вычисления следующей даты: "+err.Error(), http.StatusBadRequest)
					return
				}
				task.Date = nextDateStr
			}
		}
	}
	// Добавление новой задачи в таблицу базы данных
	insertTaskSQL := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	_, err = db.Exec(insertTaskSQL, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, "Ошибка добавления задачи: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Возвращение ответа с кодом 201 (Created)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Ошибка кодирования JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// json.NewEncoder(w).Encode(task)
}
