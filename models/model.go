package models

var Endpoint = "https://api.spotify.com/v1/me"

// type Public struct {
// 	SpotifyAccessToken string
// }

type User struct {
	Name                string `json:"display_name"`
	ID                  string `json:"id"`
	SpotifyAccessToken  string
	SpotifyRefreshToken string
	LastLogin           string
}

type TokenResponse struct {
	Access_token  string `json:"access_token"`
	Scope         string `json:"scope"`
	Refresh_token string `json:"refresh_token"`
}

type TrackResponse struct {
	Playing bool `json:"is_playing"`
	Item    struct {
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
		Name string `json:"name"`
	} `json:"item"`
}

var UserInstance User
