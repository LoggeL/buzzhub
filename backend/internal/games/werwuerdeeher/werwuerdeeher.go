package werwuerdeeher

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/logge/buzzhub/internal/game"
)

func init() {
	game.Register("werwuerdeeher", func() game.Game { return &WerWuerdeEher{} })
}

type WerWuerdeEher struct {
	players      []string
	prompts      []string
	currentRound int
	totalRounds  int
	votes        map[string]string
	scores       map[string]int
	phase        string
}

func (w *WerWuerdeEher) Info() game.GameInfo {
	return game.GameInfo{
		ID:          "werwuerdeeher",
		Name:        "Wer Wuerde Eher",
		Description: "Waehlt, wer zur schmutzigen Frage am besten passt.",
		MinPlayers:  3,
		MaxPlayers:  16,
		Icon:        "eyes",
	}
}

func (w *WerWuerdeEher) Init(_ context.Context, players []string) error {
	w.players = players
	w.scores = make(map[string]int)
	w.totalRounds = 6
	if len(players) <= 4 {
		w.totalRounds = 5
	}
	w.prompts = pickPrompts(w.totalRounds)
	w.currentRound = -1
	w.phase = ""
	for _, player := range players {
		w.scores[player] = 0
	}
	return nil
}

func (w *WerWuerdeEher) Phases() []string {
	return []string{"pick", "results", "scoreboard"}
}

func (w *WerWuerdeEher) NextPhase(_ context.Context) (*game.Phase, error) {
	switch w.phase {
	case "":
		w.currentRound = 0
		w.phase = "pick"
		return w.pickPhase()
	case "pick":
		w.phase = "results"
		return w.resultsPhase()
	case "results":
		w.currentRound++
		if w.currentRound >= w.totalRounds {
			w.phase = "scoreboard"
			return w.scoreboardPhase()
		}
		w.phase = "pick"
		return w.pickPhase()
	case "scoreboard":
		return nil, nil
	}
	return nil, nil
}

func (w *WerWuerdeEher) pickPhase() (*game.Phase, error) {
	w.votes = make(map[string]string)

	playerData := make(map[string]any)
	for _, playerID := range w.players {
		choices := make([]string, 0, len(w.players)-1)
		for _, candidateID := range w.players {
			if candidateID != playerID {
				choices = append(choices, candidateID)
			}
		}
		data := w.baseData()
		data["choices"] = choices
		playerData[playerID] = data
	}

	hostData := w.baseData()
	hostData["choices"] = w.players
	return &game.Phase{
		Name:       "pick",
		Duration:   25 * time.Second,
		HostData:   hostData,
		PlayerData: playerData,
	}, nil
}

func (w *WerWuerdeEher) resultsPhase() (*game.Phase, error) {
	voteCounts := make(map[string]int)
	for _, votedFor := range w.votes {
		voteCounts[votedFor]++
	}

	winners := topVoted(voteCounts)
	roundScores := make(map[string]int)
	for playerID, count := range voteCounts {
		points := count * 100
		roundScores[playerID] += points
		w.scores[playerID] += points
	}
	for voterID, votedFor := range w.votes {
		if contains(winners, votedFor) {
			roundScores[voterID] += 50
			w.scores[voterID] += 50
		}
	}

	data := w.baseData()
	data["votes"] = w.votes
	data["voteCounts"] = voteCounts
	data["winners"] = winners
	data["roundScores"] = roundScores
	data["scores"] = w.scores
	return &game.Phase{
		Name:          "results",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (w *WerWuerdeEher) scoreboardPhase() (*game.Phase, error) {
	data := map[string]any{
		"modeName": "Wer Wuerde Eher",
		"scores":   w.scores,
		"final":    true,
	}
	return &game.Phase{
		Name:          "scoreboard",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (w *WerWuerdeEher) HandleEvent(_ context.Context, event game.PlayerEvent) (*game.StateUpdate, error) {
	if event.Type != "vote" || w.phase != "pick" {
		return nil, nil
	}

	votedFor, _ := event.Data["playerId"].(string)
	if votedFor == "" || votedFor == event.PlayerID || !contains(w.players, votedFor) {
		return nil, nil
	}
	w.votes[event.PlayerID] = votedFor

	return &game.StateUpdate{
		BroadcastUpdate: map[string]any{
			"votedCount":   len(w.votes),
			"totalPlayers": len(w.players),
		},
		PhaseComplete: len(w.votes) >= len(w.players),
	}, nil
}

func (w *WerWuerdeEher) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	return nil, nil
}

func (w *WerWuerdeEher) Scores() map[string]int {
	return w.scores
}

func (w *WerWuerdeEher) Cleanup() {}

func (w *WerWuerdeEher) baseData() map[string]any {
	return map[string]any{
		"modeName":    "Wer Wuerde Eher",
		"prompt":      w.prompts[w.currentRound],
		"roundNum":    w.currentRound + 1,
		"totalRounds": w.totalRounds,
	}
}

func pickPrompts(n int) []string {
	shuffled := make([]string, len(promptBank))
	copy(shuffled, promptBank)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	if n > len(shuffled) {
		n = len(shuffled)
	}
	return shuffled[:n]
}

func topVoted(voteCounts map[string]int) []string {
	highest := 0
	for _, count := range voteCounts {
		if count > highest {
			highest = count
		}
	}
	if highest == 0 {
		return nil
	}
	winners := make([]string, 0)
	for playerID, count := range voteCounts {
		if count == highest {
			winners = append(winners, playerID)
		}
	}
	sort.Strings(winners)
	return winners
}

func contains(items []string, needle string) bool {
	for _, item := range items {
		if item == needle {
			return true
		}
	}
	return false
}

var promptBank = []string{
	"Wer wuerde eher eine Party verlassen und erst am naechsten Tag Bescheid sagen?",
	"Wer wuerde eher im Gruppenchat nur mit Sprachnachrichten antworten?",
	"Wer wuerde eher bei Karaoke ploetzlich ernsthaft performen?",
	"Wer wuerde eher beim ersten Date aus Versehen eine Lebenskrise teilen?",
	"Wer wuerde eher einen fremden Hund auf einer Party adoptieren wollen?",
	"Wer wuerde eher eine Entschuldigung als PowerPoint vorbereiten?",
	"Wer wuerde eher wegen eines Snacks einen Streit anfangen?",
	"Wer wuerde eher im Urlaub den kompletten Plan verlieren und trotzdem fuehren wollen?",
	"Wer wuerde eher bei einem Escape Room den Notausgang suchen?",
	"Wer wuerde eher nach einem Film sofort eine viel zu tiefe Analyse starten?",
	"Wer wuerde eher aus Versehen einen Screenshot in die falsche Gruppe schicken?",
	"Wer wuerde eher beim Kochen behaupten, das sei alles Absicht?",
	"Wer wuerde eher einen Abend mit 'nur ein Drink' komplett eskalieren lassen?",
	"Wer wuerde eher eine Nachricht tippen, loeschen und dann gar nichts sagen?",
	"Wer wuerde eher bei einem Brettspiel die Regeln persoenlich nehmen?",
	"Wer wuerde eher ein Outfit nach dem Wetterbericht ignorieren?",
}
