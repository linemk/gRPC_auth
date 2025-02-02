package main

import (
	"errors"                                                  // пакет для работы с ошибками
	"flag"                                                    // пакет для обработки аргументов командной строки
	"fmt"                                                     // пакет для форматированного вывода
	"github.com/golang-migrate/migrate/v4"                    // пакет для управления миграциями
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3" // подключение поддержки SQLite
	_ "github.com/golang-migrate/migrate/v4/source/file"      // подключение источника миграций из файлов
)

func main() {
	var storagePath, migrationPath, migrationTable string
	// парсинг аргумента для пути к хранилищу
	flag.StringVar(&storagePath, "storage-path", "", "storage path")
	// парсинг аргумента для пути к миграциям
	flag.StringVar(&migrationPath, "migrations-path", "", "migration path")
	// парсинг аргумента для имени таблицы миграций
	flag.StringVar(&migrationTable, "migrations-table", "migrations", "name of migrations table")
	// разбор аргументов командной строки
	flag.Parse()

	// проверка, что передан путь к хранилищу
	if storagePath == "" {
		panic("storage-path is required") // выброс ошибки, если аргумент не задан
	}
	// проверка, что передан путь к миграциям
	if migrationPath == "" {
		panic("migration-path is required") // выброс ошибки, если аргумент не задан
	}

	// создание объекта для управления миграциями
	m, err := migrate.New(
		"file://"+migrationPath, // источник миграций (файловая система)
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationTable)) // подключение к базе SQLite и таблице миграций
	if err != nil {
		panic(err) // выброс ошибки, если объект не удалось создать
	}

	// применение миграций
	if err := m.Up(); err != nil {
		// обработка случая, когда нет изменений для применения
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply") // вывод сообщения об отсутствии миграций
			return
		}
		panic(err) // выброс ошибки в случае других проблем
	}
	fmt.Println("migrations applied") // сообщение об успешном применении миграций
}
