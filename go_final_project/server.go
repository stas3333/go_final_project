package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTask(w, r)
	case http.MethodGet:
		getTaskId(w, r)
	case http.MethodPut:
		updateTask(w, r)
	case http.MethodDelete:
		deleteTaskId(w, r)
	default:
		JSONError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

// Функция getNextDate. Реализуем обработчик для NextDate()
func getNextDate(w http.ResponseWriter, r *http.Request) {
	now, err := time.Parse(DateFormat, r.FormValue("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")
	nextDate, err := NextDate(now, date, repeat)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(nextDate))

	if err != nil {
		http.Error(w, fmt.Errorf("writing tasks data error: %w", err).Error(), http.StatusBadRequest)
	}
}

// Функция addTask. Добавляем задачу
func addTask(w http.ResponseWriter, r *http.Request) {
	var task Task

	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		JSONError(w, "Неудалось открыть базу данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Получаем текущую дату
	today := time.Now().Format(DateFormat)

	// Разрешаем только POST-запросы
	if r.Method != http.MethodPost {
		JSONError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Декодируем JSON-запрос в структуру Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		JSONError(w, "Ошибка декодирования JSON в структуру Task", http.StatusBadRequest)
		return
	}

	if task.Date == "" {
		task.Date = today
	} else if _, err := time.Parse(DateFormat, task.Date); err != nil {
		JSONError(w, "Дата представлена в формате, отличном от 20060102", http.StatusBadRequest)
		return
	}

	// Проверка: Вызываем ошибку если Title пустой
	if task.Title == "" {
		JSONError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}
	// Проверяем правило повторения на пустоту, если дата в прошлом, возвращаем сегодняшний день
	if task.Date < today {
		if task.Repeat == "" {
			task.Date = today
		}
	}

	// Проверяем правило повторения, если дата в прошлом, пытаемся вычислить новую дату
	if task.Date < today {
		newDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			JSONError(w, "Правило повторения указано в неправильном формате", http.StatusBadRequest)
			return
		}
		task.Date = newDate
	}

	//Получаем ID по созданной задачи
	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		JSONError(w, "Неудалось выполнить запрос в базу данных", http.StatusInternalServerError)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		JSONError(w, "Ошибка получения задачи по ID", http.StatusInternalServerError)
		return
	}

	// Отправляем JSON-ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": fmt.Sprintf("%d", id)})

}

// Функция getListTasks. Получаем список ближайших задач
func getListTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []Task
	var task Task

	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		JSONError(w, "Ошибка открытия базы данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Получаем текущую дату
	today := time.Now().Format(DateFormat)

	// Разрешаем только GET-запросы
	if r.Method != http.MethodGet {
		JSONError(w, "Метод GET не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Создаём запрос в базу данных
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= 10 ORDER BY date LIMIT 50`
	rows, err := db.Query(query, today)
	if err != nil {
		JSONError(w, "Ошибка запроса к базе данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Обрабатываем строки результата запроса
	for rows.Next() {

		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			JSONError(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []Task{}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

// Функция getTaskId. Возвратит все параметры задачи по ID
func getTaskId(w http.ResponseWriter, r *http.Request) {
	var task Task

	// Открываем соединение с базой данных
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		JSONError(w, "Ошибка открытия базы данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Получаем параметр id из GET запроса
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		JSONError(w, "Не указан ID задачи", http.StatusBadRequest)
		return
	}

	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	row := db.QueryRow(query, taskID)
	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "Задача не найдена", http.StatusNotFound)
		} else {
			JSONError(w, "Ошибка получения данных", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// Функция updateTask. Обновляем параметры задач
func updateTask(w http.ResponseWriter, r *http.Request) {
	var task Task

	// Открываем БД
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		JSONError(w, "Ошибка открытия базы данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Получаем текущую дату
	today := time.Now().Format(DateFormat)

	// Декодируем JSON-запрос
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		JSONError(w, "Ошибка декодирования JSON в структуру Task", http.StatusBadRequest)
		return
	}

	// Проверка: Вызываем ошибку если ID пустой
	if task.ID == "" {
		JSONError(w, "Не указан идентификатор задачи", http.StatusBadRequest)
		return
	}

	// Проверка: Вызываем ошибку если Title пустой
	if task.Title == "" {
		JSONError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	// Проверка: Вызываем ошибку если формат даты не соответствует "20060102"
	if _, err := time.Parse(DateFormat, task.Date); err != nil {
		JSONError(w, "Неверный формат даты", http.StatusBadRequest)
		return
	}

	// Проверяем правило повторения на пустоту, если дата в прошлом возвращаем сегодняшний день
	if task.Date < today {
		if task.Repeat == "" {
			task.Date = today
		}
	}

	// Проверяем правило повторения, если дата в прошлом, пытаемся вычислить новую дату
	if task.Date < today {
		newDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			JSONError(w, "Правило повторения указано в неправильном формате", http.StatusBadRequest)
			return
		}
		task.Date = newDate
	}

	// Обновляем запись
	query := "UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?"
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		JSONError(w, "Неудалось выполнить Update запрос в базу данных", http.StatusInternalServerError)
		return
	}

	// Проверка на наличие обновлений
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		JSONError(w, "Ошибка проверки обновления задачи", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		JSONError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func doneTask(w http.ResponseWriter, r *http.Request) {
	var task Task

	// Открываем соединение с БД
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		JSONError(w, "Ошибка открытия базы данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Разрешаем только POST-запросы
	if r.Method != http.MethodPost {
		JSONError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметр id из POST запроса и если пусто, возвращаем ошибку
	tasksId := r.URL.Query().Get("id")
	if tasksId == "" {
		JSONError(w, "Не указан ID задачи", http.StatusBadRequest)
		return
	}

	// Получаем данные задач по ID
	query := "SELECT id, date, repeat FROM scheduler WHERE id = ?"
	row := db.QueryRow(query, tasksId)
	err = row.Scan(&task.ID, &task.Date, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			JSONError(w, "Задача не найдена", http.StatusNotFound)
		} else {
			JSONError(w, "Ошибка получения данных", http.StatusInternalServerError)
		}
		return
	}

	// Если у задачи нет правила повторения - удаляем её по ID
	if task.Repeat == "" {
		_, err := db.Exec("DELETE FROM scheduler WHERE id = ?", tasksId)
		if err != nil {
			JSONError(w, "Ошибка удаления задачи", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
		return
	}

	// Вычисляем новую дату
	newDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		JSONError(w, "Правило повторения указано в неправильном формате", http.StatusBadRequest)
		return
	}

	// Обновляем дату выполнения задачи
	_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", newDate, tasksId)
	if err != nil {
		JSONError(w, "Ошибка обновления задачи", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{}`))
}

// Функция для удаления задач
func deleteTaskId(w http.ResponseWriter, r *http.Request) {
	// Открываем БД
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		JSONError(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Получаем идентификатор задачи
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		JSONError(w, "Не указан идентификатор задачи", http.StatusBadRequest)
		return
	}

	// Удаляем задачу
	query := "DELETE FROM scheduler WHERE id = ?"
	res, err := db.Exec(query, taskID)
	if err != nil {
		JSONError(w, "Ошибка удаления задачи", http.StatusInternalServerError)
		return
	}

	// Проверяем, была ли удалена хотя бы одна строка
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		JSONError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func JSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
