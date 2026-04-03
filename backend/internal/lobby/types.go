package lobby

import "time"

type Lobby struct {
	Code       string    `json:"code"`
	HostID     string    `json:"hostId"`
	GameID     string    `json:"gameId"`
	Status     string    `json:"status"` // "waiting", "playing", "finished"
	MaxPlayers int       `json:"maxPlayers"`
	Players    []Player  `json:"players"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Player struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
	Score     int    `json:"score"`
}
