package services

import (
	"Spotigram/config"
	"Spotigram/database"
	"Spotigram/models"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/browser"
)

var staterand = generateRandomString(16)

var credentials = fmt.Sprintf("%s:%s", config.Conf.SpotifyClientID, config.Conf.SpotifyClientSecret)
var encodedCredentials = base64.StdEncoding.EncodeToString([]byte(credentials))

// var publicInstance models.Public
var tokenResponse models.TokenResponse
var trackResponse models.TrackResponse

// генерация рандомной строки для прохождения авторизации Spotify
func generateRandomString(length int) string {
	possible := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	possibleLength := len(possible)
	result := make([]byte, length)

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		handleError("Ошибка генерации чисел", err) // обработка ошибки при генерации случайных чисел
	}

	for i := 0; i < length; i++ {
		result[i] = possible[randomBytes[i]%byte(possibleLength)]
	}

	return string(result)
}

// перенаправление на http://localhost:8888/
func redirectToHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.Conf.RedirectURL, http.StatusConflict)
}

/* получение публичного API токена
func GetPublicAccessToken() {
	// Параметры запроса
	params := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {config.Conf.SpotifyClientID},
		"client_secret": {config.Conf.SpotifyClientSecret},
	}

	// Отправка запроса
	resp, err := http.PostForm(config.Conf.SpotifyTokenURL, params)
	handleError("Ошибка при отправке POST запроса для получения публичного токена:", err)

	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ошибка при получении публичного токена: %d", resp.StatusCode)
	}

	// Временная структура для декодирования JSON
	tokenResponse := struct {
		AccessToken string `json:"access_token"`
	}{}

	// Декодирование
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		log.Fatalf("Ошибка при декодировании JSON: %v", err)
	}

	PublicInstance.SpotifyAccessToken = tokenResponse.AccessToken
}
	No need
*/

// получение пользовательского API токена
func GetUserAccessToken() {
	scope := "user-read-private user-read-currently-playing user-read-playback-state"

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {config.Conf.SpotifyClientID},
		"scope":         {scope},
		"redirect_uri":  {config.Conf.SpotifyRedirectURI},
		"state":         {staterand},
	}

	redirectURL := fmt.Sprintf("https://accounts.spotify.com/authorize?%s", params.Encode())
	// переадресация на авторизацию в spotify
	browser.OpenURL(redirectURL)
}

// обновление пользовательского API токена
func RefreshUserAccessToken() {
	params := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {models.UserInstance.SpotifyRefreshToken},
	}

	req, err := http.NewRequest("POST", config.Conf.SpotifyTokenURL, bytes.NewBufferString(params.Encode()))
	handleError("Ошибка при формировании POST запроса", err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+encodedCredentials)

	client := &http.Client{}
	resp, err := client.Do(req)
	handleError("Ошибка при отправке POST запроса", err)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка при обновлении пользовательского токена: %d", resp.StatusCode)
		log.Printf("Окно закроется через 15 секунд")
		time.Sleep(15 * time.Second)
		os.Exit(1)
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	handleError("Ошибка при декодировании JSON", err)

	models.UserInstance.SpotifyAccessToken = tokenResponse.Access_token
	models.UserInstance.SpotifyRefreshToken = tokenResponse.Refresh_token
}

// Обработчик ответа авторизации от Spotify
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	code, state, urlErr := params.Get("code"), params.Get("state"), params.Get("error")
	if urlErr != "" || state != staterand {
		log.Printf("Ошибка при получении данных от Spotify %d", urlErr)
		redirectToHome(w, r)
		return
	}

	params = url.Values{
		"code":         {code},
		"redirect_uri": {config.Conf.SpotifyRedirectURI},
		"grant_type":   {"authorization_code"},
	}

	req, err := http.NewRequest("POST", config.Conf.SpotifyTokenURL, bytes.NewBufferString(params.Encode()))
	handleError("Ошибка при формировании запроса", err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+encodedCredentials)

	client := &http.Client{}
	resp, err := client.Do(req)
	handleError("Ошибка при отправке запроса", err)
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	handleError("Ошибка при декодировании JSON", err)

	models.UserInstance.SpotifyAccessToken = tokenResponse.Access_token
	models.UserInstance.SpotifyRefreshToken = tokenResponse.Refresh_token

	GetUsername()
	database.WriteDatabase()

	browser.OpenURL("http://localhost:8888/close")
	http.Redirect(w, r, "/close", http.StatusSeeOther)
}

// Получение имени пользователя для последующей записи в базу данных
func GetUsername() {
	req, err := http.NewRequest("GET", models.Endpoint, nil)
	handleError("Request error: ", err)
	req.Header.Add("Authorization", "Bearer "+models.UserInstance.SpotifyAccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	handleError("Response error: ", err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ошибка при получении имени пользователя: %d", resp.StatusCode)
		log.Printf("Окно закроется через 15 секунд")
		time.Sleep(15 * time.Second)
		os.Exit(1)
	}

	err = json.NewDecoder(resp.Body).Decode(&models.UserInstance)
	handleError("Ошибка при декодировании JSON", err)
}

func GetCurrentlyPlayingTrackHandler() string {
	endpoint := fmt.Sprintf("%v/player/currently-playing", models.Endpoint)

	req, err := http.NewRequest("GET", endpoint, nil)
	handleError("Ошибка при создании GET запроса", err)
	req.Header.Add("Authorization", "Bearer "+models.UserInstance.SpotifyAccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	handleError("Ошибка при отправке GET запроса", err)
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&trackResponse)
	handleError("Ошибка при декодировании JSON", err)

	var formattedArtists string
	lengthArtists := len(trackResponse.Item.Artists)

	for i := 0; i < lengthArtists; i++ {
		formattedArtists += trackResponse.Item.Artists[i].Name
		if i < lengthArtists-1 {
			formattedArtists += ", "
		}
	}

	formattedSong := fmt.Sprintf("%v - %v", formattedArtists, trackResponse.Item.Name)
	for len(formattedSong) > 70 && lengthArtists > 1 {
		lengthArtists--
		formattedArtists = ""

		for i := 0; i < lengthArtists; i++ {
			formattedArtists += trackResponse.Item.Artists[i].Name
			if i < lengthArtists-1 {
				formattedArtists += ", "
			}
		}

		formattedSong = fmt.Sprintf("%v - %v", formattedArtists, trackResponse.Item.Name)
	}
	return formattedSong
}
