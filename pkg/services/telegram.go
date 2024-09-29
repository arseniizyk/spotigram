package services

import (
	"Spotigram/config"
	"fmt"
	"log"
	"path/filepath"

	"github.com/zelenin/go-tdlib/client"
)

var TdlibClient *client.Client

func handleError(text string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", text, err)
	}
}

func AuthorizeTelegram() {
	// client authorizer
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
		ApplicationVersion:  "1.0.3",
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
}

func ChangeBio(song string) {
	result, err := TdlibClient.SetBio(&client.SetBioRequest{
		Bio: song,
	})
	handleError("ChangeBio error:", err)
	fmt.Println(result)
}
