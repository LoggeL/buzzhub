package crossword

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/logge/buzzhub/internal/game"
)

func init() {
	game.Register("crossword", func() game.Game { return &Crossword{} })
}

type WordPlacement struct {
	Word string `json:"word"`
	Row  int    `json:"row"`
	Col  int    `json:"col"`
	DRow int    `json:"dRow"` // direction: 0,1,-1
	DCol int    `json:"dCol"`
}

type FoundWord struct {
	Word     string `json:"word"`
	PlayerID string `json:"playerId"`
	Order    int    `json:"order"` // 1st, 2nd, 3rd to find
	Points   int    `json:"points"`
}

type Crossword struct {
	players    []string
	grid       [][]byte
	gridSize   int
	words      []WordPlacement
	wordSet    map[string]bool
	found      []FoundWord
	foundBy    map[string]map[string]bool // word -> set of playerIDs who found it
	findOrder  map[string]int             // word -> how many times found
	scores     map[string]int
	phase      string
	duration   time.Duration
}

func (c *Crossword) Info() game.GameInfo {
	return game.GameInfo{
		ID:          "crossword",
		Name:        "Woertersuche",
		Description: "Finde versteckte Woerter im Buchstabengitter!",
		MinPlayers:  2,
		MaxPlayers:  16,
		Icon:        "grid",
	}
}

func (c *Crossword) Init(ctx context.Context, players []string) error {
	c.players = players
	c.scores = make(map[string]int)
	c.foundBy = make(map[string]map[string]bool)
	c.findOrder = make(map[string]int)
	c.phase = ""

	// Default settings - can be overridden via game config
	c.gridSize = 10
	c.duration = 90 * time.Second

	for _, p := range players {
		c.scores[p] = 0
	}

	c.generateGrid()
	return nil
}

func (c *Crossword) Configure(settings map[string]any) {
	changed := false
	if size, ok := settings["gridSize"].(float64); ok {
		s := int(size)
		if s >= 8 && s <= 16 && s != c.gridSize {
			c.gridSize = s
			changed = true
		}
	}
	if dur, ok := settings["duration"].(float64); ok {
		d := int(dur)
		if d >= 30 && d <= 300 {
			c.duration = time.Duration(d) * time.Second
		}
	}
	if changed {
		c.foundBy = make(map[string]map[string]bool)
		c.findOrder = make(map[string]int)
		c.found = nil
		c.generateGrid()
	}
}

func (c *Crossword) Phases() []string {
	return []string{"playing", "results"}
}

func (c *Crossword) NextPhase(_ context.Context) (*game.Phase, error) {
	switch c.phase {
	case "":
		c.phase = "playing"
		return c.playingPhase()
	case "playing":
		c.phase = "results"
		return c.resultsPhase()
	case "results":
		return nil, nil
	}
	return nil, nil
}

func (c *Crossword) playingPhase() (*game.Phase, error) {
	// Convert grid to string rows
	gridStr := make([]string, len(c.grid))
	for i, row := range c.grid {
		gridStr[i] = string(row)
	}

	// Word list (without positions - players must find them)
	var wordList []string
	for _, wp := range c.words {
		wordList = append(wordList, wp.Word)
	}

	data := map[string]any{
		"grid":       gridStr,
		"gridSize":   c.gridSize,
		"wordCount":  len(c.words),
		"duration":   c.duration.Seconds(),
	}

	hostData := map[string]any{
		"grid":       gridStr,
		"gridSize":   c.gridSize,
		"words":      wordList,
		"wordCount":  len(c.words),
		"placements": c.words,
	}

	return &game.Phase{
		Name:          "playing",
		Duration:      c.duration,
		HostData:      hostData,
		BroadcastData: data,
	}, nil
}

func (c *Crossword) resultsPhase() (*game.Phase, error) {
	var wordList []string
	for _, wp := range c.words {
		wordList = append(wordList, wp.Word)
	}

	data := map[string]any{
		"scores":     c.scores,
		"found":      c.found,
		"words":      wordList,
		"placements": c.words,
		"final":      true,
	}
	return &game.Phase{
		Name:          "results",
		Duration:      10 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (c *Crossword) HandleEvent(_ context.Context, event game.PlayerEvent) (*game.StateUpdate, error) {
	if c.phase != "playing" || event.Type != "guess" {
		return nil, nil
	}

	guess, _ := event.Data["word"].(string)
	guess = strings.ToUpper(strings.TrimSpace(guess))
	if guess == "" {
		return nil, nil
	}

	// Check if word exists in puzzle
	if !c.wordSet[guess] {
		return &game.StateUpdate{
			PlayerUpdates: map[string]any{
				event.PlayerID: map[string]any{"wrong": guess},
			},
		}, nil
	}

	// Check if this player already found this word
	if c.foundBy[guess] == nil {
		c.foundBy[guess] = make(map[string]bool)
	}
	if c.foundBy[guess][event.PlayerID] {
		return nil, nil // already found by this player
	}

	c.foundBy[guess][event.PlayerID] = true
	c.findOrder[guess]++
	order := c.findOrder[guess]

	// Points: base = length * 20, decreasing for later finders
	basePoints := len(guess) * 20
	var points int
	switch order {
	case 1:
		points = basePoints
	case 2:
		points = basePoints * 60 / 100
	case 3:
		points = basePoints * 35 / 100
	default:
		points = basePoints * 15 / 100
	}

	c.scores[event.PlayerID] += points

	fw := FoundWord{
		Word:     guess,
		PlayerID: event.PlayerID,
		Order:    order,
		Points:   points,
	}
	c.found = append(c.found, fw)

	// Check if all words found by all players
	allFound := true
	for _, wp := range c.words {
		if c.findOrder[wp.Word] == 0 {
			allFound = false
			break
		}
	}

	return &game.StateUpdate{
		BroadcastUpdate: map[string]any{
			"found": fw,
			"totalFound": len(c.findOrder),
			"totalWords": len(c.words),
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

func (c *Crossword) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	return nil, nil
}

func (c *Crossword) Scores() map[string]int {
	return c.scores
}

func (c *Crossword) Cleanup() {}

// Grid generation

var wordPool = []string{
	// 3-4 letters
	"HAI", "EIS", "RAD", "HUT", "OHR", "ARM", "BUS", "ZUG", "TAG", "RAT",
	"BAUM", "HAUS", "MOND", "AUTO", "BOOT", "FELS", "TURM", "WAND", "FEST", "RING",
	"HUND", "WOLF", "BIER", "TOPF", "GELD", "TIER", "LAND", "BERG", "SAND", "GOLD",
	// 5-6 letters
	"SONNE", "STERN", "FEUER", "WASSER", "BLUME", "VOGEL", "KATZE", "MAUER",
	"STURM", "NACHT", "KERZE", "KRONE", "PERLE", "FUCHS", "DRACHE", "SCHATZ",
	"LICHT", "WOLKE", "BLITZ", "STEIN", "PFERD", "KRAFT", "TRAUM", "PIZZA",
	"APFEL", "INSEL", "BRAUT", "REGEN", "KNALL", "FISCH",
	// 7+ letters
	"RAKETE", "DONNER", "PIRAT", "GEWITTER", "ROBOTER",
	"SCHLANGE", "EINHORN", "DIAMANT", "TORNADO", "GEISTER",
	"VULKAN", "OZEAN", "PANZER", "MAMMUT", "KUCHEN",
	"LEUCHTTURM", "ABENTEUER", "SCHMETTERLING",
}

func (c *Crossword) generateGrid() {
	c.grid = make([][]byte, c.gridSize)
	for i := range c.grid {
		c.grid[i] = make([]byte, c.gridSize)
	}

	// Directions: horizontal, vertical, diagonal
	dirs := [][2]int{{0, 1}, {1, 0}, {1, 1}, {0, -1}, {-1, 0}, {-1, -1}, {1, -1}, {-1, 1}}

	// Sort words by length descending for better placement
	shuffled := make([]string, len(wordPool))
	copy(shuffled, wordPool)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	// Filter words that fit in grid
	var candidates []string
	for _, w := range shuffled {
		if len(w) <= c.gridSize {
			candidates = append(candidates, w)
		}
	}

	c.words = nil
	c.wordSet = make(map[string]bool)

	// Target word count based on grid size
	targetWords := c.gridSize + 2
	if targetWords > 20 {
		targetWords = 20
	}

	for _, word := range candidates {
		if len(c.words) >= targetWords {
			break
		}
		if c.tryPlace(word, dirs) {
			c.wordSet[word] = true
		}
	}

	// Fill empty cells with random letters
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := range c.grid {
		for j := range c.grid[i] {
			if c.grid[i][j] == 0 {
				c.grid[i][j] = letters[rand.Intn(len(letters))]
			}
		}
	}
}

func (c *Crossword) tryPlace(word string, dirs [][2]int) bool {
	upper := strings.ToUpper(word)
	wLen := len(upper)

	// Shuffle starting positions and directions
	type attempt struct {
		row, col, dr, dc int
	}
	var attempts []attempt
	for _, d := range dirs {
		for r := 0; r < c.gridSize; r++ {
			for col := 0; col < c.gridSize; col++ {
				attempts = append(attempts, attempt{r, col, d[0], d[1]})
			}
		}
	}
	rand.Shuffle(len(attempts), func(i, j int) { attempts[i], attempts[j] = attempts[j], attempts[i] })

	for _, a := range attempts {
		// Check if word fits
		endR := a.row + a.dr*(wLen-1)
		endC := a.col + a.dc*(wLen-1)
		if endR < 0 || endR >= c.gridSize || endC < 0 || endC >= c.gridSize {
			continue
		}

		ok := true
		for k := 0; k < wLen; k++ {
			r := a.row + a.dr*k
			cl := a.col + a.dc*k
			cell := c.grid[r][cl]
			if cell != 0 && cell != upper[k] {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}

		// Place word
		for k := 0; k < wLen; k++ {
			r := a.row + a.dr*k
			cl := a.col + a.dc*k
			c.grid[r][cl] = upper[k]
		}

		c.words = append(c.words, WordPlacement{
			Word: upper,
			Row:  a.row,
			Col:  a.col,
			DRow: a.dr,
			DCol: a.dc,
		})
		return true
	}
	return false
}
