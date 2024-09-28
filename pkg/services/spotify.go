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
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/pkg/browser"
)

var staterand string

var FormattedSong string

var credentials = fmt.Sprintf("%s:%s", config.SpotifyClientID, config.SpotifyClientSecret)
var encodedCredentials = base64.StdEncoding.EncodeToString([]byte(credentials))

var tokenResponse = struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}{}

var userTokenResponse = struct {
	Access_token  string `json:"access_token"`
	Token_type    string `json:"token_type"`
	Scope         string `json:"scope"`
	Expires_in    int    `json:"expires_in"`
	Refresh_token string `json:"refresh_token"`
}{}

var trackResponse = struct {
	Playing bool `json:"is_playing"`
	Item    struct {
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
		Name string `json:"name"`
	} `json:"item"`
}{}

func GetPublicAccessToken() {
	// Параметры запроса
	requestParams := []byte(fmt.Sprintf("grant_type=client_credentials&client_id=%v&client_secret=%v", config.SpotifyClientID, config.SpotifyClientSecret))

	// Формирование запроса
	request, err := http.NewRequest("POST", config.SpotifyTokenURL, bytes.NewBuffer(requestParams))
	if err != nil {
		panic(err)
	}

	// Добавление заголовка запросу
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")

	// Отправка запроса и получение ответа
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	// Проверка статуса ответа
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Error: status code %d", response.StatusCode)
	}

	json.NewDecoder(response.Body).Decode(&tokenResponse)

	models.Public.SpotifyAccessToken = tokenResponse.Access_token
	models.Public.SpotifyAccessTokenExpires = tokenResponse.Expires_in
	// fmt.Println("Public token:", models.Public.SpotifyAccessToken)
}

func redirectToHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.RedirectURL, http.StatusFound)
}

func GetUserAccessToken() {

	staterand = generateRandomString(16)
	scope := "user-read-private user-read-currently-playing user-read-playback-state"

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", config.SpotifyClientID)
	params.Add("scope", scope)
	params.Add("redirect_uri", config.SpotifyRedirectURI)
	params.Add("state", staterand)

	redirectURL := fmt.Sprintf("https://accounts.spotify.com/authorize?%s", params.Encode())
	// fmt.Println(redirectURL)
	browser.OpenURL(redirectURL)
	// http.Redirect(w, r, redirectURL, http.StatusFound)
}

func RefreshUserAccessToken() {
	params := url.Values{}

	params.Add("grant_type", "refresh_token")
	params.Add("refresh_token", models.User.SpotifyRefreshToken)

	req, err := http.NewRequest("POST", config.SpotifyTokenURL, bytes.NewBuffer([]byte(params.Encode())))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+encodedCredentials)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&userTokenResponse)
	models.User.SpotifyAccessToken = userTokenResponse.Access_token
	models.User.SpotifyRefreshToken = userTokenResponse.Refresh_token
}

// генерация рандомной строки для прохождения авторизации API spotify
func generateRandomString(length int) string {
	possible := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	possibleLength := len(possible)
	result := make([]byte, length)

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err) // обработка ошибки при генерации случайных чисел
	}

	for i := 0; i < length; i++ {
		result[i] = possible[randomBytes[i]%byte(possibleLength)]
	}

	return string(result)
}

// Обработчик ответа авторизации от Spotify
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	state := queryParams.Get("state")
	urlErr := queryParams.Get("error")
	if urlErr != "" {
		redirectToHome(w, r)
		return
	}

	if state != staterand {
		redirectToHome(w, r)
		return
	}

	params := url.Values{}
	params.Add("code", code)
	params.Add("redirect_uri", config.SpotifyRedirectURI)
	params.Add("grant_type", "authorization_code")

	request, err := http.NewRequest("POST", config.SpotifyTokenURL, bytes.NewBuffer([]byte(params.Encode())))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic "+encodedCredentials)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		http.Error(w, "Failed to execute request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&userTokenResponse)
	models.User.SpotifyAccessToken = userTokenResponse.Access_token
	models.User.SpotifyRefreshToken = userTokenResponse.Refresh_token
	// fmt.Println("Callbackhandler отработал успешно")
	GetUsername()
	database.WriteDatabase()
	// fmt.Println("Current access_token: ", userTokenResponse.Access_token)
	// fmt.Println("Current refresh_token: ", userTokenResponse.Refresh_token)
	browser.OpenURL("http://localhost:8888/close")
	http.Redirect(w, r, "/close", http.StatusSeeOther)
	// redirectToHome(w, r)
}

// Получение имени пользователя для последующей записи в базу данных
func GetUsername() {
	endpoint := "https://api.spotify.com/v1/me"

	req, err := http.NewRequest("GET", endpoint, nil)
	handleError("Request error: ", err)
	req.Header.Add("Authorization", "Bearer "+models.User.SpotifyAccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	handleError("Response error: ", err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	handleError("Read body error: ", err)

	err = json.Unmarshal(body, &models.User)
	handleError("JSON Unmarshal error: ", err)

}

func IsTokenExpired() {
	data, err := database.DB.Query("SELECT spotify_refresh_token, last_login FROM users WHERE spotify_id = $1", models.User.ID)
	handleError("Error in db.exec", err)
	fmt.Println(data)
}

func GetCurrentlyPlayingTrackHandler() {
	endpoint := "https://api.spotify.com/v1/me/player/currently-playing?market=BY"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		panic("Wrong request")
	}
	req.Header.Add("Authorization", "Bearer "+models.User.SpotifyAccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("Wrong response")
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&trackResponse)

	var formattedArtists string
	lengthArtists := len(trackResponse.Item.Artists)
	// if lengthArtists > 2 {
	// 	lengthArtists = 2
	// }
	for i := 0; i < lengthArtists; i++ {
		formattedArtists += trackResponse.Item.Artists[i].Name
		if i < lengthArtists-1 { // Добавляем запятую, только если это не последний артист
			formattedArtists += ", "
		}
	}
	// fmt.Println(trackResponse)

	// + playing
	FormattedSong = fmt.Sprintf("%v - %v", formattedArtists, trackResponse.Item.Name)
	for len(FormattedSong) > 70 && lengthArtists > 1 {
		lengthArtists--
		formattedArtists = ""

		for i := 0; i < lengthArtists; i++ {
			formattedArtists += trackResponse.Item.Artists[i].Name
			if i < lengthArtists-1 {
				formattedArtists += ", "
			}
		}
		// Обновляем строку с песней
		FormattedSong = fmt.Sprintf("%v - %v", formattedArtists, trackResponse.Item.Name)
	}
	fmt.Println(FormattedSong)
}
