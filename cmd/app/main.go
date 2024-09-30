package main

import (
	"Spotigram/database"
	"Spotigram/pkg/handlers"
)

func main() {
	handlers.Handlers()

	defer database.CloseDatabase()

}
