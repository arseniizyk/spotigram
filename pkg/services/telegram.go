package services

import (
	"Spotigram/config"
	"Spotigram/models"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/zelenin/go-tdlib/client"
)

var TdlibClient *client.Client

func handleError(text string, err error) {
	if err != nil {
		log.Printf("%v : %v\nОкно закроется автоматически через 15 секунд.", text, err)
		time.Sleep(15 * time.Second)
		os.Exit(1)
	}
}

// client authorizer
func AuthorizeTelegram() {
	authorizer := client.ClientAuthorizer()
	go client.CliInteractor(authorizer)

	authorizer.TdlibParameters <- &client.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   filepath.Join(".tdlib", "database"),
		FilesDirectory:      filepath.Join(".tdlib", "files"),
		UseFileDatabase:     false,
		UseChatInfoDatabase: false,
		UseMessageDatabase:  false,
		UseSecretChats:      false,
		ApiId:               config.Conf.TelegramAPI_id,
		ApiHash:             config.Conf.TelegramAPI_hash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Spotigram",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.5",
		// EnableStorageOptimizer: true,
		// IgnoreFileNames:        false,
	}

	_, err := client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 0,
	})
	handleError("SetLogVerbosityLevel error:", err)

	TdlibClient, err = client.NewClient(authorizer)
	handleError("NewClient error:", err)

	optionValue, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	handleError("GetOption error:", err)

	log.Printf("TDLib version: %s", optionValue.(*client.OptionValueString).Value)

	me, err := TdlibClient.GetMe()
	handleError("GetMe error:", err)

	log.Printf("Me: %s %s [%s]", me.FirstName, me.LastName, me.Usernames)
	// Запуск функции обновления трека сразу после авторизации в телеграмм
	go UpdateCurrentTrack()
	getBio(me)
}

func ChangeBio(song string) {
	result, err := TdlibClient.SetBio(&client.SetBioRequest{
		Bio: song,
	})
	handleError("ChangeBio error:", err)
	fmt.Println(result)
}

func getBio(me *client.User) {
	result, err := TdlibClient.GetUserFullInfo(&client.GetUserFullInfoRequest{
		UserId: me.Id,
	})
	handleError("Что-то пошло не так при попытке получить Bio", err)
	models.UserInstance.TelegramBio = result.Bio.Text
	log.Println("Bio:", models.UserInstance.TelegramBio)
}
