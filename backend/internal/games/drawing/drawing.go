package drawing

import (
	"context"
	"math/rand"
	"time"

	"github.com/logge/buzzhub/internal/game"
)

func init() {
	game.Register("drawing", func() game.Game { return &Drawing{} })
}

type Drawing struct {
	players      []string
	words        []string
	drawOrder    []int // indices into players
	currentTurn  int
	totalTurns   int
	currentWord  string
	guesses      map[string]bool // who guessed correctly
	strokes      []any           // accumulated strokes
	scores       map[string]int
	phase        string
}

func (d *Drawing) Info() game.GameInfo {
	return game.GameInfo{
		ID:          "drawing",
		Name:        "Kritzelei",
		Description: "Zeichne auf deinem Handy und lass die anderen raten!",
		MinPlayers:  3,
		MaxPlayers:  16,
		Icon:        "pencil",
	}
}

func (d *Drawing) Init(_ context.Context, players []string) error {
	d.players = players
	d.scores = make(map[string]int)
	d.totalTurns = len(players)
	if d.totalTurns > 6 {
		d.totalTurns = 6
	}
	d.words = pickWords(d.totalTurns)

	// Randomize draw order
	d.drawOrder = make([]int, len(players))
	for i := range d.drawOrder {
		d.drawOrder[i] = i
	}
	rand.Shuffle(len(d.drawOrder), func(i, j int) { d.drawOrder[i], d.drawOrder[j] = d.drawOrder[j], d.drawOrder[i] })

	d.currentTurn = -1
	d.phase = ""
	for _, p := range players {
		d.scores[p] = 0
	}
	return nil
}

func (d *Drawing) Phases() []string {
	return []string{"draw", "reveal", "scoreboard"}
}

func (d *Drawing) NextPhase(_ context.Context) (*game.Phase, error) {
	switch d.phase {
	case "":
		d.currentTurn = 0
		d.phase = "draw"
		return d.drawPhase()
	case "draw":
		d.phase = "reveal"
		return d.revealPhase()
	case "reveal":
		d.currentTurn++
		if d.currentTurn >= d.totalTurns {
			d.phase = "scoreboard"
			return d.scoreboardPhase()
		}
		d.phase = "draw"
		return d.drawPhase()
	case "scoreboard":
		return nil, nil
	}
	return nil, nil
}

func (d *Drawing) drawPhase() (*game.Phase, error) {
	d.guesses = make(map[string]bool)
	d.strokes = nil
	d.currentWord = d.words[d.currentTurn]

	drawerIdx := d.drawOrder[d.currentTurn % len(d.drawOrder)]
	drawerID := d.players[drawerIdx]

	// Drawer sees the word, others see "rate!"
	playerData := make(map[string]any)
	for _, pid := range d.players {
		if pid == drawerID {
			playerData[pid] = map[string]any{
				"role":    "drawer",
				"word":    d.currentWord,
				"turnNum": d.currentTurn + 1,
				"totalTurns": d.totalTurns,
			}
		} else {
			playerData[pid] = map[string]any{
				"role":    "guesser",
				"hint":    generateHint(d.currentWord),
				"turnNum": d.currentTurn + 1,
				"totalTurns": d.totalTurns,
			}
		}
	}

	hostData := map[string]any{
		"drawer":     drawerID,
		"hint":       generateHint(d.currentWord),
		"turnNum":    d.currentTurn + 1,
		"totalTurns": d.totalTurns,
	}

	return &game.Phase{
		Name:       "draw",
		Duration:   45 * time.Second,
		HostData:   hostData,
		PlayerData: playerData,
	}, nil
}

func (d *Drawing) revealPhase() (*game.Phase, error) {
	data := map[string]any{
		"word":    d.currentWord,
		"guesses": d.guesses,
		"scores":  d.scores,
	}

	return &game.Phase{
		Name:          "reveal",
		Duration:      5 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (d *Drawing) scoreboardPhase() (*game.Phase, error) {
	data := map[string]any{
		"scores": d.scores,
		"final":  true,
	}
	return &game.Phase{
		Name:          "scoreboard",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (d *Drawing) HandleEvent(_ context.Context, event game.PlayerEvent) (*game.StateUpdate, error) {
	if d.phase != "draw" {
		return nil, nil
	}

	drawerIdx := d.drawOrder[d.currentTurn % len(d.drawOrder)]
	drawerID := d.players[drawerIdx]

	switch event.Type {
	case "draw":
		// Only drawer can draw
		if event.PlayerID != drawerID {
			return nil, nil
		}
		stroke := event.Data["stroke"]
		d.strokes = append(d.strokes, stroke)

		// Broadcast stroke to all others (host + guessers)
		return &game.StateUpdate{
			HostUpdate: map[string]any{"stroke": stroke},
			BroadcastUpdate: map[string]any{"stroke": stroke},
		}, nil

	case "guess":
		if event.PlayerID == drawerID {
			return nil, nil
		}
		if d.guesses[event.PlayerID] {
			return nil, nil // already guessed
		}

		guess, _ := event.Data["guess"].(string)
		if matchGuess(guess, d.currentWord) {
			d.guesses[event.PlayerID] = true
			// Guesser gets points, drawer also gets points
			remaining := len(d.players) - 1 - len(d.guesses)
			guesserPoints := 100 + remaining*20
			d.scores[event.PlayerID] += guesserPoints
			d.scores[drawerID] += 50

			allGuessed := len(d.guesses) >= len(d.players)-1
			return &game.StateUpdate{
				BroadcastUpdate: map[string]any{
					"correctGuess":  event.PlayerID,
					"guessedCount":  len(d.guesses),
					"totalGuessers": len(d.players) - 1,
				},
				PlayerUpdates: map[string]any{
					event.PlayerID: map[string]any{"correct": true, "points": guesserPoints},
				},
				PhaseComplete: allGuessed,
			}, nil
		}

		// Wrong guess - broadcast it as a chat message
		return &game.StateUpdate{
			BroadcastUpdate: map[string]any{
				"chatGuess": map[string]any{
					"playerId": event.PlayerID,
					"guess":    guess,
				},
			},
		}, nil
	}
	return nil, nil
}

func (d *Drawing) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	return nil, nil
}

func (d *Drawing) Scores() map[string]int {
	return d.scores
}

func (d *Drawing) Cleanup() {}

func generateHint(word string) string {
	runes := []rune(word)
	hint := make([]rune, len(runes))
	for i, r := range runes {
		if r == ' ' {
			hint[i] = ' '
		} else {
			hint[i] = '_'
		}
	}
	return string(hint)
}

func matchGuess(guess, word string) bool {
	if len(guess) == 0 {
		return false
	}
	// Simple case-insensitive comparison
	g := []rune(guess)
	w := []rune(word)
	if len(g) != len(w) {
		return false
	}
	for i := range g {
		a, b := g[i], w[i]
		if a >= 'A' && a <= 'Z' {
			a += 32
		}
		if b >= 'A' && b <= 'Z' {
			b += 32
		}
		// Handle umlauts
		if a == 0xE4 || a == 0xC4 { a = 0xE4 }
		if b == 0xE4 || b == 0xC4 { b = 0xE4 }
		if a == 0xF6 || a == 0xD6 { a = 0xF6 }
		if b == 0xF6 || b == 0xD6 { b = 0xF6 }
		if a == 0xFC || a == 0xDC { a = 0xFC }
		if b == 0xFC || b == 0xDC { b = 0xFC }
		if a != b {
			return false
		}
	}
	return true
}

var wordBank = []string{
	"Hund", "Katze", "Haus", "Baum", "Auto",
	"Sonne", "Mond", "Stern", "Blume", "Fisch",
	"Schuh", "Brille", "Stuhl", "Tisch", "Lampe",
	"Pizza", "Kuchen", "Eis", "Apfel", "Banane",
	"Gitarre", "Trompete", "Klavier", "Trommel",
	"Fahrrad", "Flugzeug", "Schiff", "Rakete",
	"Drache", "Einhorn", "Roboter", "Pirat",
	"Vulkan", "Regenbogen", "Gewitter", "Schneemann",
	"Krokodil", "Pinguin", "Elefant", "Giraffe",
	"Schloss", "Leuchtturm", "Bruecke", "Zelt",
	"Fussball", "Basketball", "Tennis", "Schwimmen",
}

func pickWords(n int) []string {
	shuffled := make([]string, len(wordBank))
	copy(shuffled, wordBank)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	if n > len(shuffled) {
		n = len(shuffled)
	}
	return shuffled[:n]
}
