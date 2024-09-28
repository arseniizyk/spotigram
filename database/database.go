package database

import (
	"Spotigram/config"
	"Spotigram/models"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB
var err error

// Инициализация базы данных
func InitDatabase() {
	err := os.Mkdir("database", 0755)
	if err != nil {
		fmt.Println("Директория database уже существует")
	}
	DB, err = sql.Open("sqlite3", "database/tokens.db")
	if err != nil {
		fmt.Println("Ошибка при открытии базы данных: ", err)
	}

	if err = DB.Ping(); err != nil {
		fmt.Println("Не удалось проинициализировать базу данных: ", err)
	}

	fmt.Println("База данных успешно проинициализирована")

	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		username TEXT,
		spotify_id TEXT UNIQUE,
		spotify_refresh_token TEXT,
		last_login TIMESTAMP
		);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		fmt.Println("Ошибка при создании таблицы:", err)
	}

	fmt.Println("Таблица 'users' успешно открыта")
}

// Чтение данных из базы
func ReadDatabase() error {
	row := DB.QueryRow("SELECT username, spotify_id, spotify_refresh_token, last_login FROM users")

	err = row.Scan(&models.User.Name, &models.User.ID, &models.User.SpotifyRefreshToken, &models.User.Last_login)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Данные не найдены, авторизуйтесь")
			config.IsFirstStart = true
			return nil
		}
		return err
	}
	fmt.Println("Данные найдены: ", models.User)
	return nil
}

// Запись токенов в базу данных
func WriteDatabase() error {
	query := `INSERT INTO users (username, spotify_id, spotify_refresh_token, last_login)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`

	_, err := DB.Exec(query, models.User.Name, models.User.ID, models.User.SpotifyRefreshToken)
	if err != nil {
		log.Fatal("Что-то пошло не так при попытке записи в базу данных")
	}

	return nil
}

func UpdateDatabase() error {
	_, err := DB.Exec("UPDATE users SET spotify_refresh_token = $1, last_login = CURRENT_TIMESTAMP WHERE spotify_id = $2", models.User.SpotifyRefreshToken, models.User.ID)
	if err != nil {
		fmt.Println("Ошибка при обновлении базы данных: ", err)
		return nil
	}
	fmt.Println("База данных успешно обновлена")
	return nil
}

func CloseDatabase() error {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			log.Printf("Ошибка при закрытии базы данных: %v", err)
			return err
		}
	}
	return nil
}
