package voting

import (
	"context"
	"math/rand"
	"time"

	"github.com/logge/buzzhub/internal/game"
)

func init() {
	game.Register("voting", func() game.Game { return &Voting{} })
}

type Voting struct {
	players      []string
	prompts      []string
	currentRound int
	totalRounds  int
	submissions  map[string]string // playerID -> answer
	votes        map[string]string // voterID -> votedForPlayerID
	scores       map[string]int
	phase        string
}

func (v *Voting) Info() game.GameInfo {
	return game.GameInfo{
		ID:          "voting",
		Name:        "Abstimmung",
		Description: "Schreibe die witzigste Antwort und sammle Stimmen!",
		MinPlayers:  3,
		MaxPlayers:  16,
		Icon:        "trophy",
	}
}

func (v *Voting) Init(_ context.Context, players []string) error {
	v.players = players
	v.scores = make(map[string]int)
	v.totalRounds = 5
	if len(players) <= 4 {
		v.totalRounds = 4
	}
	v.prompts = pickPrompts(v.totalRounds)
	v.currentRound = -1
	v.phase = ""
	for _, p := range players {
		v.scores[p] = 0
	}
	return nil
}

func (v *Voting) Phases() []string {
	return []string{"prompt", "vote", "results", "scoreboard"}
}

func (v *Voting) NextPhase(_ context.Context) (*game.Phase, error) {
	switch v.phase {
	case "":
		v.currentRound = 0
		v.phase = "prompt"
		return v.promptPhase()
	case "prompt":
		v.phase = "vote"
		return v.votePhase()
	case "vote":
		v.phase = "results"
		return v.resultsPhase()
	case "results":
		v.currentRound++
		if v.currentRound >= v.totalRounds {
			v.phase = "scoreboard"
			return v.scoreboardPhase()
		}
		v.phase = "prompt"
		return v.promptPhase()
	case "scoreboard":
		return nil, nil
	}
	return nil, nil
}

func (v *Voting) promptPhase() (*game.Phase, error) {
	v.submissions = make(map[string]string)
	v.votes = make(map[string]string)

	data := map[string]any{
		"prompt":     v.prompts[v.currentRound],
		"roundNum":   v.currentRound + 1,
		"totalRounds": v.totalRounds,
	}

	return &game.Phase{
		Name:          "prompt",
		Duration:      30 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (v *Voting) votePhase() (*game.Phase, error) {
	// Build anonymous submission list
	type submission struct {
		PlayerID string `json:"playerId"`
		Answer   string `json:"answer"`
	}

	var subs []submission
	for pid, answer := range v.submissions {
		subs = append(subs, submission{PlayerID: pid, Answer: answer})
	}

	// Each player sees all answers except their own
	playerData := make(map[string]any)
	for _, pid := range v.players {
		var filtered []submission
		for _, s := range subs {
			if s.PlayerID != pid {
				filtered = append(filtered, s)
			}
		}
		playerData[pid] = map[string]any{
			"prompt":      v.prompts[v.currentRound],
			"submissions": filtered,
		}
	}

	hostData := map[string]any{
		"prompt":      v.prompts[v.currentRound],
		"submissions": subs,
	}

	return &game.Phase{
		Name:       "vote",
		Duration:   20 * time.Second,
		HostData:   hostData,
		PlayerData: playerData,
	}, nil
}

func (v *Voting) resultsPhase() (*game.Phase, error) {
	// Count votes
	voteCounts := make(map[string]int)
	for _, votedFor := range v.votes {
		voteCounts[votedFor]++
	}

	// Award points: 100 per vote
	roundScores := make(map[string]int)
	for pid, count := range voteCounts {
		points := count * 100
		roundScores[pid] = points
		v.scores[pid] += points
	}

	data := map[string]any{
		"prompt":      v.prompts[v.currentRound],
		"submissions": v.submissions,
		"votes":       v.votes,
		"voteCounts":  voteCounts,
		"roundScores": roundScores,
		"scores":      v.scores,
	}

	return &game.Phase{
		Name:          "results",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (v *Voting) scoreboardPhase() (*game.Phase, error) {
	data := map[string]any{
		"scores": v.scores,
		"final":  true,
	}
	return &game.Phase{
		Name:          "scoreboard",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (v *Voting) HandleEvent(_ context.Context, event game.PlayerEvent) (*game.StateUpdate, error) {
	switch event.Type {
	case "submit":
		if v.phase != "prompt" {
			return nil, nil
		}
		answer, _ := event.Data["answer"].(string)
		if answer == "" {
			return nil, nil
		}
		v.submissions[event.PlayerID] = answer
		return &game.StateUpdate{
			BroadcastUpdate: map[string]any{
				"submittedCount": len(v.submissions),
				"totalPlayers":   len(v.players),
			},
		}, nil

	case "vote":
		if v.phase != "vote" {
			return nil, nil
		}
		votedFor, _ := event.Data["playerId"].(string)
		if votedFor == "" || votedFor == event.PlayerID {
			return nil, nil
		}
		v.votes[event.PlayerID] = votedFor
		return &game.StateUpdate{
			BroadcastUpdate: map[string]any{
				"votedCount":   len(v.votes),
				"totalPlayers": len(v.players),
			},
		}, nil
	}
	return nil, nil
}

func (v *Voting) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	return nil, nil
}

func (v *Voting) Scores() map[string]int {
	return v.scores
}

func (v *Voting) Cleanup() {}

var promptBank = []string{
	"Was wuerde ein Ausserirdischer als erstes ueber Menschen sagen?",
	"Der schlechteste Ratschlag, den man geben kann:",
	"Was steht auf dem Schild vor dem Paradies?",
	"Die schlechteste Superkraft aller Zeiten:",
	"Warum sind Flamingos rosa?",
	"Was wuerde dein Kuechlschrank sagen, wenn er reden koennte?",
	"Der wahre Grund, warum Dinosaurier ausgestorben sind:",
	"Was denkt dein Hund wirklich ueber dich?",
	"Die ungeschriebene Regel im Buero:",
	"Was wuerde auf deinem Grabstein stehen?",
	"Die schlimmste Ausrede fuer Zu-spaet-Kommen:",
	"Was wuerde Wikipedia ueber dich schreiben?",
	"Der schlechteste Name fuer ein Restaurant:",
	"Was wuerde ein Zeitreisender aus dem Jahr 1800 am meisten schockieren?",
	"Die ehrliche Antwort auf 'Wie geht's dir?':",
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
