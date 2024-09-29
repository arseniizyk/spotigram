package handlers

import (
	"Spotigram/config"
	"Spotigram/database"
	"Spotigram/pkg/services"
	"fmt"
	"log"
	"net/http"
)

func Handlers() {
	// База данных
	database.InitDatabase()
	database.ReadDatabase()

	http.HandleFunc("/close", closeHandler)

	// Проверка, есть ли данные в базы данных
	if config.IsFirstStart == true {
		http.HandleFunc("/callback", services.CallbackHandler)
		services.GetUserAccessToken()
		config.IsFirstStart = false
	} else {
		services.RefreshUserAccessToken()
	}

	// Фоновые горутины
	// go services.GetPublicAccessToken()
	go services.UpdateAPI()

	// Запуск сервера
	fmt.Println("Server is running on http://localhost:8888")
	go services.AuthorizeTelegram()
	log.Fatal(http.ListenAndServe(":8888", nil))

}

func closeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
	<html>
	<body>
		<script>
			window.close();
		</script>
		<h1>Закрытие вкладки...</h1>
	</body>
	</html>
	`))
}
