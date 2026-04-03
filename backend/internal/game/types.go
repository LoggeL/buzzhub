package game

import (
	"context"
	"time"
)

type GameInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MinPlayers  int    `json:"minPlayers"`
	MaxPlayers  int    `json:"maxPlayers"`
	Icon        string `json:"icon"`
}

type Phase struct {
	Name          string
	Duration      time.Duration
	HostData      any
	PlayerData    map[string]any // key = playerID
	BroadcastData any            // fallback if no host
}

type PlayerEvent struct {
	PlayerID string
	Type     string
	Data     map[string]any
}

type StateUpdate struct {
	Phase      *Phase
	Scores     map[string]int
	HostUpdate any
	PlayerUpdates map[string]any
	BroadcastUpdate any
	GameOver   bool
}

type Game interface {
	Info() GameInfo
	Init(ctx context.Context, players []string) error
	Phases() []string
	NextPhase(ctx context.Context) (*Phase, error)
	HandleEvent(ctx context.Context, event PlayerEvent) (*StateUpdate, error)
	TimerExpired(ctx context.Context) (*StateUpdate, error)
	Scores() map[string]int
	Cleanup()
}

type Factory func() Game
