package main

import (
	"Spotigram/database"
	"Spotigram/pkg/handlers"
)

func main() {
	handlers.Handlers()

	defer database.CloseDatabase()

	// TODO: Добавить сохранение предыдущего статуса и его восстановление перед завершением
}
