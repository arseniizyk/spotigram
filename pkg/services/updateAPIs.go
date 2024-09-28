package services

import (
	"Spotigram/database"
	"time"
)

// Обновление access_token каждые 3600 секунд
func UpdateAPI() {
	ticker := time.NewTicker(3600 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			GetPublicAccessToken()
			RefreshUserAccessToken()
			database.UpdateDatabase()
		}
	}
}

func UpdateCurrentTrack() {
	timing := 60
	ticker := time.NewTicker(time.Duration(timing) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			GetCurrentlyPlayingTrackHandler()
			ChangeBio()
			// if !trackResponse.Playing {
			// 	timing = 360
			// }
		}
	}
}
