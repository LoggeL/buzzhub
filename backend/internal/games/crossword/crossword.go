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
	Word string   `json:"word"`
	Path [][2]int `json:"path"` // list of [row, col]
}

type FoundWord struct {
	Word     string   `json:"word"`
	PlayerID string   `json:"playerId"`
	Order    int      `json:"order"`
	Points   int      `json:"points"`
	Path     [][2]int `json:"path"`
}

type Crossword struct {
	players      []string
	grid         [][]byte
	gridSize     int
	words        []WordPlacement
	wordSet      map[string]bool
	placementMap map[string]*WordPlacement // word -> placement
	found        []FoundWord
	foundBy      map[string]map[string]bool // word -> set of playerIDs
	findOrder    map[string]int             // word -> how many times found
	globalFound  []string                   // words found by anyone (ordered)
	scores       map[string]int
	phase        string
	duration     time.Duration
}

func (c *Crossword) Info() game.GameInfo {
	return game.GameInfo{
		ID:          "crossword",
		Name:        "Versteckte Woerter",
		Description: "Finde die geheimen Woerter im Buchstabengitter.",
		MinPlayers:  2,
		MaxPlayers:  16,
		Icon:        "splash",
	}
}

func (c *Crossword) Init(ctx context.Context, players []string) error {
	c.players = players
	c.scores = make(map[string]int)
	c.foundBy = make(map[string]map[string]bool)
	c.findOrder = make(map[string]int)
	c.placementMap = make(map[string]*WordPlacement)
	c.phase = ""
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
		c.placementMap = make(map[string]*WordPlacement)
		c.found = nil
		c.globalFound = nil
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
	gridStr := make([]string, len(c.grid))
	for i, row := range c.grid {
		gridStr[i] = string(row)
	}

	var wordList []string
	for _, wp := range c.words {
		wordList = append(wordList, wp.Word)
	}

	data := map[string]any{
		"grid":      gridStr,
		"gridSize":  c.gridSize,
		"wordCount": len(c.words),
		"words":     wordList,
		"duration":  c.duration.Seconds(),
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

	// Parse path from event
	pathRaw, ok := event.Data["path"].([]any)
	if !ok || len(pathRaw) < 3 {
		return nil, nil
	}

	path := make([][2]int, len(pathRaw))
	for i, p := range pathRaw {
		pair, ok := p.([]any)
		if !ok || len(pair) != 2 {
			return nil, nil
		}
		r, ok1 := pair[0].(float64)
		cl, ok2 := pair[1].(float64)
		if !ok1 || !ok2 {
			return nil, nil
		}
		ri, ci := int(r), int(cl)
		if ri < 0 || ri >= c.gridSize || ci < 0 || ci >= c.gridSize {
			return nil, nil
		}
		path[i] = [2]int{ri, ci}
	}

	// Validate adjacency and no duplicates
	seen := make(map[[2]int]bool)
	for i, pos := range path {
		if seen[pos] {
			return nil, nil
		}
		seen[pos] = true
		if i > 0 {
			prev := path[i-1]
			dr := abs(pos[0] - prev[0])
			dc := abs(pos[1] - prev[1])
			if dr > 1 || dc > 1 || (dr == 0 && dc == 0) {
				return nil, nil
			}
		}
	}

	// Find matching placement
	placement := c.findPlacement(path)
	if placement == nil {
		return &game.StateUpdate{
			PlayerUpdates: map[string]any{
				event.PlayerID: map[string]any{"wrong": true},
			},
		}, nil
	}

	guess := placement.Word

	// Check if this player already found this word
	if c.foundBy[guess] == nil {
		c.foundBy[guess] = make(map[string]bool)
	}
	if c.foundBy[guess][event.PlayerID] {
		return nil, nil
	}

	c.foundBy[guess][event.PlayerID] = true
	c.findOrder[guess]++
	order := c.findOrder[guess]

	if order == 1 {
		c.globalFound = append(c.globalFound, guess)
	}

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
		Path:     placement.Path,
	}
	c.found = append(c.found, fw)

	// Check if all words found by at least one player
	allFound := len(c.globalFound) >= len(c.words)

	return &game.StateUpdate{
		BroadcastUpdate: map[string]any{
			"found":      fw,
			"totalFound": len(c.globalFound),
			"totalWords": len(c.words),
		},
		PlayerUpdates: map[string]any{
			event.PlayerID: map[string]any{
				"correct": guess,
				"points":  points,
				"order":   order,
				"path":    placement.Path,
			},
		},
		PhaseComplete: allFound,
	}, nil
}

func (c *Crossword) findPlacement(path [][2]int) *WordPlacement {
	for i, wp := range c.words {
		if pathsMatch(wp.Path, path) {
			return &c.words[i]
		}
	}
	return nil
}

func pathsMatch(a, b [][2]int) bool {
	if len(a) != len(b) {
		return false
	}
	// Forward
	fwd := true
	for i := range a {
		if a[i] != b[i] {
			fwd = false
			break
		}
	}
	if fwd {
		return true
	}
	// Reverse
	for i := range a {
		if a[i] != b[len(b)-1-i] {
			return false
		}
	}
	return true
}

func (c *Crossword) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	return nil, nil
}

func (c *Crossword) Scores() map[string]int {
	return c.scores
}

func (c *Crossword) Cleanup() {}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

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

	shuffled := make([]string, len(wordPool))
	copy(shuffled, wordPool)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	var candidates []string
	for _, w := range shuffled {
		if len(w) <= c.gridSize {
			candidates = append(candidates, w)
		}
	}

	c.words = nil
	c.wordSet = make(map[string]bool)
	c.placementMap = make(map[string]*WordPlacement)

	targetWords := c.gridSize + 2
	if targetWords > 20 {
		targetWords = 20
	}

	for _, word := range candidates {
		if len(c.words) >= targetWords {
			break
		}
		if c.tryPlace(word) {
			c.wordSet[word] = true
			c.placementMap[word] = &c.words[len(c.words)-1]
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

func (c *Crossword) tryPlace(word string) bool {
	upper := strings.ToUpper(word)
	wLen := len(upper)

	// Generate all possible paths (straight + L-shaped)
	type pathOpt [][2]int
	var options []pathOpt

	// Straight paths in 8 directions
	dirs := [][2]int{{0, 1}, {1, 0}, {1, 1}, {0, -1}, {-1, 0}, {-1, -1}, {1, -1}, {-1, 1}}
	for _, d := range dirs {
		for r := 0; r < c.gridSize; r++ {
			for col := 0; col < c.gridSize; col++ {
				path := make([][2]int, 0, wLen)
				valid := true
				for k := 0; k < wLen; k++ {
					nr, nc := r+d[0]*k, col+d[1]*k
					if nr < 0 || nr >= c.gridSize || nc < 0 || nc >= c.gridSize {
						valid = false
						break
					}
					path = append(path, [2]int{nr, nc})
				}
				if valid {
					options = append(options, path)
				}
			}
		}
	}

	// L-shaped paths (cardinal directions only, min 4 letters)
	if wLen >= 4 {
		cardinals := [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
		for _, d1 := range cardinals {
			for _, d2 := range cardinals {
				// Must be perpendicular (dot product = 0) and not same/opposite
				if d1[0]*d2[0]+d1[1]*d2[1] != 0 {
					continue
				}
				for bendAt := 2; bendAt <= wLen-2; bendAt++ {
					for r := 0; r < c.gridSize; r++ {
						for col := 0; col < c.gridSize; col++ {
							path := make([][2]int, 0, wLen)
							valid := true
							for k := 0; k < wLen; k++ {
								var nr, nc int
								if k < bendAt {
									nr = r + d1[0]*k
									nc = col + d1[1]*k
								} else {
									nr = r + d1[0]*(bendAt-1) + d2[0]*(k-bendAt+1)
									nc = col + d1[1]*(bendAt-1) + d2[1]*(k-bendAt+1)
								}
								if nr < 0 || nr >= c.gridSize || nc < 0 || nc >= c.gridSize {
									valid = false
									break
								}
								path = append(path, [2]int{nr, nc})
							}
							if valid {
								options = append(options, path)
							}
						}
					}
				}
			}
		}
	}

	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	for _, path := range options {
		ok := true
		for k, pos := range path {
			cell := c.grid[pos[0]][pos[1]]
			if cell != 0 && cell != upper[k] {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}

		// Place word
		for k, pos := range path {
			c.grid[pos[0]][pos[1]] = upper[k]
		}

		c.words = append(c.words, WordPlacement{
			Word: upper,
			Path: path,
		})
		return true
	}
	return false
}
