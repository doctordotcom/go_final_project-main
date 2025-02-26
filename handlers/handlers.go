package handlers

import (
	"database/sql"
	"encoding/json"
	"m/models"
	"m/rules"
	"net/http"
	"time"
)

// Глобальная переменная для базы данных
var db *sql.DB

// GetNextDateHandler - обработчик для "/api/nextdate", получение следующей даты
func GetNextDateHandler(w http.ResponseWriter, r *http.Request) {
	// проверка метода Get
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
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

// PostTaskHandler - обработчик для добавления задачи
func PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода Post
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	// Глобальная переменная для структуры Task из пакета models
	var task models.Task
	// Десериализация JSON-запроса в структуру Task
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
	// Получение текущей даты
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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Ошибка кодирования JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetTaskHandler - обработчик для получения списка задач
func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода GET
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	// Получение параметра фильтрации по заголовку, если он есть
	titleFilter := r.URL.Query().Get("title")

	// Определение максимального количества задач, которое будет возвращаться
	const maxTasks = 50
	limit := maxTasks

	// Создание SQL-запроса для получения задач
	sqlQuery := `SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`
	// sqlQuery := `SELECT date, title, comment, repeat FROM scheduler WHERE title LIKE ? LIMIT ?`
	// фильтрация с помощью SQL LIKE
	filter := "%" + titleFilter + "%"

	// Выполнение запроса к базе данных
	rows, err := db.Query(sqlQuery, filter, limit)
	if err != nil {
		http.Error(w, "Ошибка выполнения запроса к базе данных: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создание слайса для хранения задач
	var tasks []models.Task

	// Итерация строк результата запроса
	for rows.Next() {
		var task models.Task
		// Считывание значений в структуру Task
		if err := rows.Scan(&task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			http.Error(w, "Ошибка считывания данных задачи: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Добавление задач в слайс
		tasks = append(tasks, task)
	}

	// Проверка на наличие ошибок в процессе итерации строк
	if err := rows.Err(); err != nil {
		http.Error(w, "Ошибка при переборе задач: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Установка заголовка Content-Type для ответа
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Сериализация слайса задач в JSON и отправка его в ответе
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, "Ошибка кодирования JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
