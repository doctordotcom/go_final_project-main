package scheduler

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Build() {

	// os.Getwd возвращает путь к исполняемому файлу программы (альтернатива os.Executable)
	appPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// filepath.Join объединяет несколько путей в один.
	// filepath.Dir возвращающает путь к директории в которой находится файл.
	// os.Stat получает информацию о файле.
	dbFile := filepath.Join(appPath, "scheduler.db")
	_, err = os.Stat(dbFile)
	// если файла нет, код установит значение install в true, что позволит создать базу данных.
	var install bool
	if err != nil {
		install = true
	}
	// если install равен true, после открытия БД требуется выполнить
	// sql-запрос с CREATE TABLE и CREATE INDEX.
	if install {
		// os.Create создаёт файл по определённому пути, который передаётся в качестве параметра.
		file, err := os.Create(dbFile)
		if err != nil {
			log.Fatal("Не удалось создать файл по указанному пути", err)
		}
		// defer file.Close - отложенное закрытие file после завершения работы с ним.
		defer file.Close()
		// проверка, существует ли sql файл перед его прочтением
		_, err = os.Stat("scheduler.sql")
		if os.IsNotExist(err) {
			log.Fatal("Файл scheduler.sql не найден", err)
		}
		// os.ReadFile читает файл, где указано как строить таблицу.
		n, err := os.ReadFile("scheduler.sql")
		if err != nil {
			log.Fatal("Не удалось прочесть файл scheduler.sql", err)
		}
		// sql.Open подключает к базе данных.
		DB, err := sql.Open("sqlite", dbFile)
		if err != nil {
			log.Fatal("Не удалось подключиться к базе данных", err)
			return
		}
		defer DB.Close()
		// перевод n в нужный тип string с помощью m.
		// DB.Exec(m) принимает sql запрос и возвращает значение sql.Result.
		m := string(n)
		_, err = DB.Exec(m)
		if err != nil {
			log.Fatal("Ошибка приёма или возврата значения sql запроса", err)
			return
		}
	}
}
