package services

import (
	"Spotigram/database"
	"fmt"
	"time"
)

// Обновление access_token каждые 3600 секунд
func UpdateAPI() {
	ticker := time.NewTicker(3600 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// GetPublicAccessToken()
			RefreshUserAccessToken()
			database.UpdateDatabase()
		}
	}
}

func UpdateCurrentTrack() {
	timing := 15
	ticker := time.NewTicker(time.Duration(timing) * time.Second)
	defer ticker.Stop()
	var prevTrack string

	for {
		select {
		case <-ticker.C:
			currentTrack := GetCurrentlyPlayingTrackHandler()
			if prevTrack != currentTrack {
				prevTrack = currentTrack
				fmt.Println(currentTrack)
				ChangeBio(currentTrack)
			} else {
				fmt.Println("Track wasn't changed.")
			}
		}
	}
}
