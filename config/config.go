package config

var IsFirstStart bool
var IsAuthorized bool = false

type Config struct {
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyTokenURL     string
	RedirectURL         string
	SpotifyRedirectURI  string
	TelegramAPI_id      int32
	TelegramAPI_hash    string
}
