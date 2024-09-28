package models

var Public = struct {
	SpotifyAccessToken        string
	SpotifyAccessTokenExpires int
	TelegramToken             string
}{}

var User = struct {
	Name                string `json:"display_name"`
	ID                  string `json:"id"`
	SpotifyAccessToken  string
	SpotifyRefreshToken string
	Last_login          string
}{}
