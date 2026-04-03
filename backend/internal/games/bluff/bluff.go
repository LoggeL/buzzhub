package bluff

import (
	"context"
	"math/rand"
	"time"

	"github.com/logge/buzzhub/internal/game"
)

func init() {
	game.Register("bluff", func() game.Game { return &Bluff{} })
}

type BluffQuestion struct {
	Text       string `json:"text"`
	RealAnswer string `json:"realAnswer"`
}

type Bluff struct {
	players      []string
	questions    []BluffQuestion
	currentRound int
	totalRounds  int
	fakeAnswers  map[string]string // playerID -> fake answer
	guesses      map[string]string // voterID -> guessed answer (playerID or "truth")
	scores       map[string]int
	phase        string
}

func (b *Bluff) Info() game.GameInfo {
	return game.GameInfo{
		ID:          "bluff",
		Name:        "Bluff Master",
		Description: "Erfinde falsche Antworten und taeuschen deine Mitspieler!",
		MinPlayers:  3,
		MaxPlayers:  16,
		Icon:        "mask",
	}
}

func (b *Bluff) Init(_ context.Context, players []string) error {
	b.players = players
	b.scores = make(map[string]int)
	b.totalRounds = 5
	if len(players) <= 4 {
		b.totalRounds = 4
	}
	b.questions = pickBluffQuestions(b.totalRounds)
	b.currentRound = -1
	b.phase = ""
	for _, p := range players {
		b.scores[p] = 0
	}
	return nil
}

func (b *Bluff) Phases() []string {
	return []string{"write", "guess", "reveal", "scoreboard"}
}

func (b *Bluff) NextPhase(_ context.Context) (*game.Phase, error) {
	switch b.phase {
	case "":
		b.currentRound = 0
		b.phase = "write"
		return b.writePhase()
	case "write":
		b.phase = "guess"
		return b.guessPhase()
	case "guess":
		b.phase = "reveal"
		return b.revealPhase()
	case "reveal":
		b.currentRound++
		if b.currentRound >= b.totalRounds {
			b.phase = "scoreboard"
			return b.scoreboardPhase()
		}
		b.phase = "write"
		return b.writePhase()
	case "scoreboard":
		return nil, nil
	}
	return nil, nil
}

func (b *Bluff) writePhase() (*game.Phase, error) {
	b.fakeAnswers = make(map[string]string)
	b.guesses = make(map[string]string)

	data := map[string]any{
		"question":    b.questions[b.currentRound].Text,
		"roundNum":    b.currentRound + 1,
		"totalRounds": b.totalRounds,
	}

	return &game.Phase{
		Name:          "write",
		Duration:      30 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (b *Bluff) guessPhase() (*game.Phase, error) {
	// Mix real answer with fake answers
	type answerOption struct {
		ID     string `json:"id"` // playerID or "truth"
		Answer string `json:"answer"`
	}

	var options []answerOption
	options = append(options, answerOption{
		ID:     "truth",
		Answer: b.questions[b.currentRound].RealAnswer,
	})
	for pid, fake := range b.fakeAnswers {
		options = append(options, answerOption{ID: pid, Answer: fake})
	}

	// Shuffle
	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	// Each player sees all options except their own fake answer
	playerData := make(map[string]any)
	for _, pid := range b.players {
		var filtered []answerOption
		for _, opt := range options {
			if opt.ID != pid {
				filtered = append(filtered, opt)
			}
		}
		playerData[pid] = map[string]any{
			"question": b.questions[b.currentRound].Text,
			"options":  filtered,
		}
	}

	hostData := map[string]any{
		"question": b.questions[b.currentRound].Text,
		"options":  options,
	}

	return &game.Phase{
		Name:       "guess",
		Duration:   20 * time.Second,
		HostData:   hostData,
		PlayerData: playerData,
	}, nil
}

func (b *Bluff) revealPhase() (*game.Phase, error) {
	roundScores := make(map[string]int)

	for voterID, guessedID := range b.guesses {
		if guessedID == "truth" {
			// Found the truth: +200 points
			roundScores[voterID] += 200
		} else {
			// Fell for a bluff: bluffer gets +100
			roundScores[guessedID] += 100
		}
	}

	for pid, points := range roundScores {
		b.scores[pid] += points
	}

	data := map[string]any{
		"question":    b.questions[b.currentRound].Text,
		"realAnswer":  b.questions[b.currentRound].RealAnswer,
		"fakeAnswers": b.fakeAnswers,
		"guesses":     b.guesses,
		"roundScores": roundScores,
		"scores":      b.scores,
	}

	return &game.Phase{
		Name:          "reveal",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (b *Bluff) scoreboardPhase() (*game.Phase, error) {
	data := map[string]any{
		"scores": b.scores,
		"final":  true,
	}
	return &game.Phase{
		Name:          "scoreboard",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (b *Bluff) HandleEvent(_ context.Context, event game.PlayerEvent) (*game.StateUpdate, error) {
	switch event.Type {
	case "submit":
		if b.phase != "write" {
			return nil, nil
		}
		answer, _ := event.Data["answer"].(string)
		if answer == "" {
			return nil, nil
		}
		b.fakeAnswers[event.PlayerID] = answer
		return &game.StateUpdate{
			BroadcastUpdate: map[string]any{
				"submittedCount": len(b.fakeAnswers),
				"totalPlayers":   len(b.players),
			},
			PhaseComplete: len(b.fakeAnswers) >= len(b.players),
		}, nil

	case "guess":
		if b.phase != "guess" {
			return nil, nil
		}
		guessedID, _ := event.Data["answerId"].(string)
		if guessedID == "" {
			return nil, nil
		}
		b.guesses[event.PlayerID] = guessedID
		return &game.StateUpdate{
			BroadcastUpdate: map[string]any{
				"guessedCount": len(b.guesses),
				"totalPlayers": len(b.players),
			},
			PhaseComplete: len(b.guesses) >= len(b.players),
		}, nil
	}
	return nil, nil
}

func (b *Bluff) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	return nil, nil
}

func (b *Bluff) Scores() map[string]int {
	return b.scores
}

func (b *Bluff) Cleanup() {}

var bluffQuestionBank = []BluffQuestion{
	{Text: "Was bedeutet das Wort 'Petrichor'?", RealAnswer: "Der Geruch von Regen auf trockener Erde"},
	{Text: "Wie lang ist der laengste Fluss der Welt in km?", RealAnswer: "6.650 km (Nil)"},
	{Text: "Was ist das Nationaltier von Schottland?", RealAnswer: "Das Einhorn"},
	{Text: "Wie viele Herzen hat ein Oktopus?", RealAnswer: "Drei"},
	{Text: "Was war der urspruengliche Name von Google?", RealAnswer: "BackRub"},
	{Text: "Welches Land hat die meisten Zeitzonen?", RealAnswer: "Frankreich (12 Zeitzonen)"},
	{Text: "Woraus besteht der Saturn-Ring hauptsaechlich?", RealAnswer: "Eis und Gestein"},
	{Text: "Wie heisst die Angst vor langen Woertern?", RealAnswer: "Hippopotomonstrosesquippedaliophobie"},
	{Text: "Welches Tier schlaeft mit einem offenen Auge?", RealAnswer: "Der Delfin"},
	{Text: "Was ist das seltenste natuerliche Element auf der Erde?", RealAnswer: "Astat"},
	{Text: "Wie schnell waechst ein Fingernagel pro Monat?", RealAnswer: "Etwa 3-4 Millimeter"},
	{Text: "Was ist die aelteste bekannte Sportart?", RealAnswer: "Ringen"},
}

func pickBluffQuestions(n int) []BluffQuestion {
	shuffled := make([]BluffQuestion, len(bluffQuestionBank))
	copy(shuffled, bluffQuestionBank)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	if n > len(shuffled) {
		n = len(shuffled)
	}
	return shuffled[:n]
}
