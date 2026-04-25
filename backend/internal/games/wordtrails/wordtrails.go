package wordtrails

import (
	"context"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/logge/buzzhub/internal/game"
)

func init() {
	game.Register("wordtrails", func() game.Game { return &WordTrails{} })
}

type Puzzle struct {
	Letters []string `json:"letters"`
	Words   []string `json:"words"`
	Theme   string   `json:"theme"`
}

type FoundWord struct {
	Word     string `json:"word"`
	PlayerID string `json:"playerId"`
	Order    int    `json:"order"`
	Points   int    `json:"points"`
}

type WordTrails struct {
	players     []string
	puzzle      Puzzle
	wordSet     map[string]bool
	found       []FoundWord
	foundBy     map[string]map[string]bool
	findOrder   map[string]int
	globalFound []string
	scores      map[string]int
	phase       string
	duration    time.Duration
}

func (w *WordTrails) Info() game.GameInfo {
	return game.GameInfo{
		ID:          "wordtrails",
		Name:        "Buchstaben-Tease",
		Description: "Verbinde Buchstaben und fuelle alle Slots schneller als die anderen.",
		MinPlayers:  2,
		MaxPlayers:  16,
		Icon:        "tongue",
	}
}

func (w *WordTrails) Init(_ context.Context, players []string) error {
	w.players = players
	w.scores = make(map[string]int)
	w.wordSet = make(map[string]bool)
	w.foundBy = make(map[string]map[string]bool)
	w.findOrder = make(map[string]int)
	w.duration = 90 * time.Second
	w.phase = ""

	for _, p := range players {
		w.scores[p] = 0
	}

	w.puzzle = pickPuzzle()
	for _, word := range w.puzzle.Words {
		w.wordSet[word] = true
	}
	return nil
}

func (w *WordTrails) Phases() []string {
	return []string{"playing", "results"}
}

func (w *WordTrails) NextPhase(_ context.Context) (*game.Phase, error) {
	switch w.phase {
	case "":
		w.phase = "playing"
		return w.playingPhase()
	case "playing":
		w.phase = "results"
		return w.resultsPhase()
	case "results":
		return nil, nil
	}
	return nil, nil
}

func (w *WordTrails) playingPhase() (*game.Phase, error) {
	data := map[string]any{
		"modeName":  "Buchstaben-Tease",
		"theme":     w.puzzle.Theme,
		"letters":   shuffledLetters(w.puzzle.Letters),
		"slots":     w.slots(false),
		"wordCount": len(w.puzzle.Words),
	}
	hostData := map[string]any{
		"modeName":  "Buchstaben-Tease",
		"theme":     w.puzzle.Theme,
		"letters":   w.puzzle.Letters,
		"slots":     w.slots(true),
		"wordCount": len(w.puzzle.Words),
	}
	return &game.Phase{
		Name:          "playing",
		Duration:      w.duration,
		HostData:      hostData,
		BroadcastData: data,
	}, nil
}

func (w *WordTrails) resultsPhase() (*game.Phase, error) {
	data := map[string]any{
		"modeName": "Buchstaben-Tease",
		"theme":    w.puzzle.Theme,
		"letters":  w.puzzle.Letters,
		"slots":    w.slots(true),
		"found":    w.found,
		"scores":   w.scores,
		"final":    true,
	}
	return &game.Phase{
		Name:          "results",
		Duration:      10 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (w *WordTrails) HandleEvent(_ context.Context, event game.PlayerEvent) (*game.StateUpdate, error) {
	if w.phase != "playing" || event.Type != "guess" {
		return nil, nil
	}

	raw, _ := event.Data["word"].(string)
	guess := strings.ToUpper(strings.TrimSpace(raw))
	if guess == "" || !w.wordSet[guess] || !canBuild(guess, w.puzzle.Letters) {
		return &game.StateUpdate{
			PlayerUpdates: map[string]any{
				event.PlayerID: map[string]any{"wrong": true},
			},
		}, nil
	}

	if w.foundBy[guess] == nil {
		w.foundBy[guess] = make(map[string]bool)
	}
	if w.foundBy[guess][event.PlayerID] {
		return nil, nil
	}

	w.foundBy[guess][event.PlayerID] = true
	w.findOrder[guess]++
	order := w.findOrder[guess]
	if order == 1 {
		w.globalFound = append(w.globalFound, guess)
	}

	basePoints := len([]rune(guess)) * 25
	points := basePoints
	switch order {
	case 1:
		points = basePoints + 50
	case 2:
		points = basePoints * 60 / 100
	case 3:
		points = basePoints * 35 / 100
	default:
		points = basePoints * 15 / 100
	}
	w.scores[event.PlayerID] += points

	found := FoundWord{
		Word:     guess,
		PlayerID: event.PlayerID,
		Order:    order,
		Points:   points,
	}
	w.found = append(w.found, found)

	allFound := len(w.globalFound) >= len(w.puzzle.Words)
	return &game.StateUpdate{
		BroadcastUpdate: map[string]any{
			"found":      found,
			"slots":      w.slots(false),
			"totalFound": len(w.globalFound),
			"totalWords": len(w.puzzle.Words),
		},
		PlayerUpdates: map[string]any{
			event.PlayerID: map[string]any{
				"correct": guess,
				"points":  points,
				"order":   order,
			},
		},
		PhaseComplete: allFound,
	}, nil
}

func (w *WordTrails) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	return nil, nil
}

func (w *WordTrails) Scores() map[string]int {
	return w.scores
}

func (w *WordTrails) Cleanup() {}

func (w *WordTrails) slots(reveal bool) []map[string]any {
	words := make([]string, len(w.puzzle.Words))
	copy(words, w.puzzle.Words)
	sort.Slice(words, func(i, j int) bool {
		if len(words[i]) == len(words[j]) {
			return words[i] < words[j]
		}
		return len(words[i]) < len(words[j])
	})

	slots := make([]map[string]any, 0, len(words))
	for _, word := range words {
		found := w.findOrder[word] > 0
		text := strings.Repeat("_", len([]rune(word)))
		if reveal || found {
			text = word
		}
		slots = append(slots, map[string]any{
			"length": len([]rune(word)),
			"word":   text,
			"found":  found,
		})
	}
	return slots
}

func canBuild(word string, letters []string) bool {
	counts := make(map[rune]int)
	for _, l := range letters {
		for _, r := range l {
			counts[r]++
		}
	}
	for _, r := range word {
		counts[r]--
		if counts[r] < 0 {
			return false
		}
	}
	return true
}

func shuffledLetters(letters []string) []string {
	out := make([]string, len(letters))
	copy(out, letters)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func pickPuzzle() Puzzle {
	puzzles := make([]Puzzle, len(puzzleBank))
	copy(puzzles, puzzleBank)
	return puzzles[rand.Intn(len(puzzles))]
}

var puzzleBank = []Puzzle{
	{
		Theme:   "Nachtleben",
		Letters: []string{"B", "A", "R", "T", "A", "N", "Z"},
		Words:   []string{"BAR", "ART", "TANZ", "RAN", "ZART", "ARZT"},
	},
	{
		Theme:   "Chaos",
		Letters: []string{"P", "A", "N", "I", "K", "E", "N"},
		Words:   []string{"PANIK", "PEN", "KNEIP", "KANN", "PIE", "PIN"},
	},
	{
		Theme:   "Snack",
		Letters: []string{"P", "I", "Z", "Z", "A", "T", "E"},
		Words:   []string{"PIZZA", "ZEIT", "TAP", "PIE", "ZITAT", "TAT"},
	},
	{
		Theme:   "Party",
		Letters: []string{"F", "E", "I", "E", "R", "N", "D"},
		Words:   []string{"FEIER", "FEIN", "REIF", "EIER", "DEIN", "RIND"},
	},
	{
		Theme:   "Drama",
		Letters: []string{"S", "T", "R", "E", "I", "T", "A"},
		Words:   []string{"STREIT", "RAST", "TIER", "EIS", "SEIT", "TEAR"},
	},
	{
		Theme:   "WG",
		Letters: []string{"K", "U", "E", "C", "H", "E", "N"},
		Words:   []string{"KUECHE", "KUCHEN", "ECKE", "HEU", "EHE", "NECK"},
	},
}
