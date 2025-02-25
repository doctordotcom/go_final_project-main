package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open() (*sql.DB, error) {
	// os.Getwd возвращает путь к исполняемому файлу программы (вместо Executable, чтобы не было ошибки пути)
	appPath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения рабочей директории: %w", err)
	}
	// filepath.Join объединяет несколько путей в один.
	// filepath.Dir возвращает путь к директории в которой находится файл.
	dbFile := filepath.Join(appPath, "scheduler.db")
	// Проверка, существует ли файл базы данных
	// os.Stat получает информацию о файле.
	// os. IsNotExist проверяет существование файла
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Println("Файл базы данных не существует.")
		return nil, nil
	}
	// Открытие базы данных
	_, err = sql.Open("sqlite", dbFile)
	// db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Println("Ошибка при открытии базы данных:", err)
		return nil, err
	}
	log.Println("База данных открыта успешно.")
	return nil, err
	// return db, nil
}
