package codenames

import (
	"context"
	"math/rand"
	"time"

	"github.com/logge/buzzhub/internal/game"
)

func init() {
	game.Register("codenames", func() game.Game { return &Codenames{} })
}

type Card struct {
	Word     string `json:"word"`
	Color    string `json:"color"`    // "red", "blue", "neutral", "assassin"
	Revealed bool   `json:"revealed"`
}

type Codenames struct {
	players     []string
	teams       map[string]string // playerID -> "red"/"blue"
	spymasters  map[string]bool   // playerID -> true if spymaster
	cards       [25]Card
	currentTeam string // "red"/"blue"
	currentHint string
	currentNum  int
	guessesLeft int
	scores      map[string]int
	phase       string // "assign", "hint", "guess", "results"
	winner      string
	redCount    int // remaining red cards
	blueCount   int // remaining blue cards
}

func (c *Codenames) Info() game.GameInfo {
	return game.GameInfo{
		ID:          "codenames",
		Name:        "Geheime Woerter",
		Description: "Zwei Teams. Ein Hinweis. Wer findet alle seine Agenten zuerst?",
		MinPlayers:  4,
		MaxPlayers:  16,
		Icon:        "spy",
	}
}

func (c *Codenames) Init(_ context.Context, players []string) error {
	c.players = players
	c.scores = make(map[string]int)
	c.teams = make(map[string]string)
	c.spymasters = make(map[string]bool)
	for _, p := range players {
		c.scores[p] = 0
	}
	c.setupCards()
	c.currentTeam = "red" // red starts (has 9 cards)
	return nil
}

func (c *Codenames) setupCards() {
	words := pickWords(25)
	// 9 red, 8 blue, 7 neutral, 1 assassin
	colors := make([]string, 25)
	for i := 0; i < 9; i++ {
		colors[i] = "red"
	}
	for i := 9; i < 17; i++ {
		colors[i] = "blue"
	}
	for i := 17; i < 24; i++ {
		colors[i] = "neutral"
	}
	colors[24] = "assassin"
	rand.Shuffle(len(colors), func(i, j int) { colors[i], colors[j] = colors[j], colors[i] })

	for i := 0; i < 25; i++ {
		c.cards[i] = Card{Word: words[i], Color: colors[i], Revealed: false}
	}
	c.redCount = 9
	c.blueCount = 8
}

func (c *Codenames) Phases() []string {
	return []string{"assign", "hint", "guess", "results"}
}

func (c *Codenames) NextPhase(_ context.Context) (*game.Phase, error) {
	switch c.phase {
	case "":
		c.phase = "assign"
		return c.assignPhase()
	case "assign":
		c.autoAssign()
		c.phase = "hint"
		return c.hintPhase()
	case "hint":
		c.phase = "guess"
		return c.guessPhase()
	case "guess":
		if c.winner != "" {
			c.phase = "results"
			return c.resultsPhase()
		}
		// Switch teams
		if c.currentTeam == "red" {
			c.currentTeam = "blue"
		} else {
			c.currentTeam = "red"
		}
		c.currentHint = ""
		c.currentNum = 0
		c.guessesLeft = 0
		c.phase = "hint"
		return c.hintPhase()
	case "results":
		return nil, nil // game over
	}
	return nil, nil
}

func (c *Codenames) assignPhase() (*game.Phase, error) {
	playerData := make(map[string]any)
	for _, p := range c.players {
		playerData[p] = map[string]any{
			"players":    c.players,
			"teams":      c.teams,
			"spymasters": c.spymasterList(),
		}
	}
	return &game.Phase{
		Name:     "assign",
		Duration: 30 * time.Second,
		HostData: map[string]any{
			"players":    c.players,
			"teams":      c.teams,
			"spymasters": c.spymasterList(),
		},
		PlayerData: playerData,
	}, nil
}

func (c *Codenames) autoAssign() {
	// Auto-assign players who haven't chosen
	unassigned := []string{}
	for _, p := range c.players {
		if c.teams[p] == "" {
			unassigned = append(unassigned, p)
		}
	}

	redCount := 0
	blueCount := 0
	for _, t := range c.teams {
		if t == "red" {
			redCount++
		} else if t == "blue" {
			blueCount++
		}
	}

	rand.Shuffle(len(unassigned), func(i, j int) { unassigned[i], unassigned[j] = unassigned[j], unassigned[i] })
	for _, p := range unassigned {
		if redCount <= blueCount {
			c.teams[p] = "red"
			redCount++
		} else {
			c.teams[p] = "blue"
			blueCount++
		}
	}

	// Ensure each team has a spymaster
	hasRedSpy := false
	hasBlueSpy := false
	for p, isSpy := range c.spymasters {
		if isSpy && c.teams[p] == "red" {
			hasRedSpy = true
		}
		if isSpy && c.teams[p] == "blue" {
			hasBlueSpy = true
		}
	}
	if !hasRedSpy {
		for _, p := range c.players {
			if c.teams[p] == "red" {
				c.spymasters[p] = true
				break
			}
		}
	}
	if !hasBlueSpy {
		for _, p := range c.players {
			if c.teams[p] == "blue" {
				c.spymasters[p] = true
				break
			}
		}
	}
}

func (c *Codenames) hintPhase() (*game.Phase, error) {
	playerData := make(map[string]any)
	for _, p := range c.players {
		pd := map[string]any{
			"cards":       c.playerCards(p),
			"currentTeam": c.currentTeam,
			"myTeam":      c.teams[p],
			"isSpymaster": c.spymasters[p],
			"teams":       c.teams,
			"spymasters":  c.spymasterList(),
			"redLeft":     c.redCount,
			"blueLeft":    c.blueCount,
			"waiting":     true,
		}
		playerData[p] = pd
	}

	return &game.Phase{
		Name:     "hint",
		Duration: 90 * time.Second,
		HostData: map[string]any{
			"cards":       c.allCards(),
			"currentTeam": c.currentTeam,
			"teams":       c.teams,
			"spymasters":  c.spymasterList(),
			"redLeft":     c.redCount,
			"blueLeft":    c.blueCount,
		},
		PlayerData: playerData,
	}, nil
}

func (c *Codenames) guessPhase() (*game.Phase, error) {
	playerData := make(map[string]any)
	for _, p := range c.players {
		pd := map[string]any{
			"cards":       c.playerCards(p),
			"currentTeam": c.currentTeam,
			"myTeam":      c.teams[p],
			"isSpymaster": c.spymasters[p],
			"teams":       c.teams,
			"spymasters":  c.spymasterList(),
			"hint":        c.currentHint,
			"hintNum":     c.currentNum,
			"guessesLeft": c.guessesLeft,
			"redLeft":     c.redCount,
			"blueLeft":    c.blueCount,
		}
		playerData[p] = pd
	}

	return &game.Phase{
		Name:     "guess",
		Duration: 120 * time.Second,
		HostData: map[string]any{
			"cards":       c.allCards(),
			"currentTeam": c.currentTeam,
			"teams":       c.teams,
			"spymasters":  c.spymasterList(),
			"hint":        c.currentHint,
			"hintNum":     c.currentNum,
			"guessesLeft": c.guessesLeft,
			"redLeft":     c.redCount,
			"blueLeft":    c.blueCount,
		},
		PlayerData: playerData,
	}, nil
}

func (c *Codenames) resultsPhase() (*game.Phase, error) {
	// Award scores to winning team
	for _, p := range c.players {
		if c.teams[p] == c.winner {
			c.scores[p] = 100
		}
	}

	data := map[string]any{
		"cards":   c.allCards(),
		"winner":  c.winner,
		"scores":  c.scores,
		"teams":   c.teams,
		"final":   true,
	}
	return &game.Phase{
		Name:          "results",
		Duration:      10 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (c *Codenames) HandleEvent(_ context.Context, event game.PlayerEvent) (*game.StateUpdate, error) {
	switch event.Type {
	case "join-team":
		return c.handleJoinTeam(event)
	case "set-spymaster":
		return c.handleSetSpymaster(event)
	case "give-hint":
		return c.handleGiveHint(event)
	case "guess":
		return c.handleGuess(event)
	case "end-turn":
		return c.handleEndTurn(event)
	}
	return nil, nil
}

func (c *Codenames) handleJoinTeam(event game.PlayerEvent) (*game.StateUpdate, error) {
	if c.phase != "assign" {
		return nil, nil
	}
	team, ok := event.Data["team"].(string)
	if !ok || (team != "red" && team != "blue") {
		return nil, nil
	}
	c.teams[event.PlayerID] = team

	update := c.assignUpdate()
	return update, nil
}

func (c *Codenames) handleSetSpymaster(event game.PlayerEvent) (*game.StateUpdate, error) {
	if c.phase != "assign" {
		return nil, nil
	}
	team := c.teams[event.PlayerID]
	if team == "" {
		return nil, nil
	}
	// Remove previous spymaster from same team
	for p, isSpy := range c.spymasters {
		if isSpy && c.teams[p] == team {
			c.spymasters[p] = false
		}
	}
	c.spymasters[event.PlayerID] = true

	update := c.assignUpdate()
	return update, nil
}

func (c *Codenames) assignUpdate() *game.StateUpdate {
	playerUpdates := make(map[string]any)
	for _, p := range c.players {
		playerUpdates[p] = map[string]any{
			"teams":      c.teams,
			"spymasters": c.spymasterList(),
		}
	}
	return &game.StateUpdate{
		HostUpdate: map[string]any{
			"teams":      c.teams,
			"spymasters": c.spymasterList(),
		},
		PlayerUpdates: playerUpdates,
	}
}

func (c *Codenames) handleGiveHint(event game.PlayerEvent) (*game.StateUpdate, error) {
	if c.phase != "hint" {
		return nil, nil
	}
	// Only current team's spymaster can give hint
	if c.teams[event.PlayerID] != c.currentTeam || !c.spymasters[event.PlayerID] {
		return nil, nil
	}

	word, ok := event.Data["word"].(string)
	if !ok || word == "" {
		return nil, nil
	}
	count, ok := event.Data["count"].(float64)
	if !ok || count < 1 || count > 9 {
		return nil, nil
	}

	c.currentHint = word
	c.currentNum = int(count)
	c.guessesLeft = int(count) + 1 // can guess one extra

	return &game.StateUpdate{
		PhaseComplete: true,
	}, nil
}

func (c *Codenames) handleGuess(event game.PlayerEvent) (*game.StateUpdate, error) {
	if c.phase != "guess" {
		return nil, nil
	}
	// Only current team's non-spymaster members can guess
	if c.teams[event.PlayerID] != c.currentTeam || c.spymasters[event.PlayerID] {
		return nil, nil
	}

	idx, ok := event.Data["index"].(float64)
	if !ok || int(idx) < 0 || int(idx) >= 25 {
		return nil, nil
	}
	cardIdx := int(idx)
	if c.cards[cardIdx].Revealed {
		return nil, nil
	}

	// Reveal the card
	c.cards[cardIdx].Revealed = true
	color := c.cards[cardIdx].Color

	// Update remaining counts
	if color == "red" {
		c.redCount--
	} else if color == "blue" {
		c.blueCount--
	}

	c.guessesLeft--
	endTurn := false
	gameOver := false

	switch {
	case color == "assassin":
		// Guessing team loses
		if c.currentTeam == "red" {
			c.winner = "blue"
		} else {
			c.winner = "red"
		}
		gameOver = true
	case color == c.currentTeam:
		// Correct guess - check win
		if (color == "red" && c.redCount == 0) || (color == "blue" && c.blueCount == 0) {
			c.winner = c.currentTeam
			gameOver = true
		} else if c.guessesLeft <= 0 {
			endTurn = true
		}
	default:
		// Wrong team or neutral - check if other team won
		if color == "red" && c.redCount == 0 {
			c.winner = "red"
			gameOver = true
		} else if color == "blue" && c.blueCount == 0 {
			c.winner = "blue"
			gameOver = true
		} else {
			endTurn = true
		}
	}

	if gameOver {
		return &game.StateUpdate{
			PhaseComplete: true,
		}, nil
	}

	if endTurn {
		return &game.StateUpdate{
			PhaseComplete: true,
		}, nil
	}

	// Continue guessing - send updates to all players
	playerUpdates := make(map[string]any)
	for _, p := range c.players {
		playerUpdates[p] = map[string]any{
			"cards":       c.playerCards(p),
			"guessesLeft": c.guessesLeft,
			"redLeft":     c.redCount,
			"blueLeft":    c.blueCount,
		}
	}

	return &game.StateUpdate{
		HostUpdate: map[string]any{
			"cards":       c.allCards(),
			"guessesLeft": c.guessesLeft,
			"redLeft":     c.redCount,
			"blueLeft":    c.blueCount,
		},
		PlayerUpdates: playerUpdates,
	}, nil
}

func (c *Codenames) handleEndTurn(event game.PlayerEvent) (*game.StateUpdate, error) {
	if c.phase != "guess" {
		return nil, nil
	}
	if c.teams[event.PlayerID] != c.currentTeam {
		return nil, nil
	}
	return &game.StateUpdate{
		PhaseComplete: true,
	}, nil
}

func (c *Codenames) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	if c.phase == "assign" || c.phase == "hint" {
		return &game.StateUpdate{PhaseComplete: true}, nil
	}
	return nil, nil
}

func (c *Codenames) Scores() map[string]int {
	return c.scores
}

func (c *Codenames) Cleanup() {}

// Helper: cards visible to a specific player
func (c *Codenames) playerCards(playerID string) []map[string]any {
	isSpymaster := c.spymasters[playerID]
	cards := make([]map[string]any, 25)
	for i, card := range c.cards {
		m := map[string]any{
			"word":     card.Word,
			"revealed": card.Revealed,
		}
		if card.Revealed || isSpymaster {
			m["color"] = card.Color
		}
		cards[i] = m
	}
	return cards
}

// Helper: all cards with colors (for host)
func (c *Codenames) allCards() []map[string]any {
	cards := make([]map[string]any, 25)
	for i, card := range c.cards {
		cards[i] = map[string]any{
			"word":     card.Word,
			"color":    card.Color,
			"revealed": card.Revealed,
		}
	}
	return cards
}

// Helper: list of spymaster player IDs
func (c *Codenames) spymasterList() []string {
	list := []string{}
	for p, isSpy := range c.spymasters {
		if isSpy {
			list = append(list, p)
		}
	}
	return list
}

// Word bank - ~200 German nouns
var wordBank = []string{
	"Apfel", "Auto", "Baum", "Berg", "Blume", "Boot", "Brief", "Bruecke",
	"Buch", "Burg", "Dach", "Diamant", "Drache", "Eis", "Elefant", "Engel",
	"Erde", "Feder", "Feuer", "Fisch", "Flagge", "Flugzeug", "Frosch", "Fuchs",
	"Garten", "Geist", "Glocke", "Gold", "Gras", "Guertel", "Hafen", "Hammer",
	"Handschuh", "Haus", "Herz", "Himmel", "Hund", "Hut", "Insel", "Juwel",
	"Kabel", "Kaktus", "Kamera", "Kanone", "Karte", "Katze", "Kerze", "Kette",
	"Kirche", "Knochen", "Koenig", "Koffer", "Kopf", "Korn", "Kran", "Krone",
	"Kuchen", "Lampe", "Lanze", "Laser", "Laub", "Lava", "Leiter", "Licht",
	"Loewe", "Luft", "Maske", "Mauer", "Messer", "Mond", "Motor", "Muehle",
	"Muenze", "Mutter", "Nadel", "Nagel", "Nase", "Nest", "Netz", "Nuss",
	"Ofen", "Palme", "Panzer", "Papier", "Perle", "Pfeife", "Pferd", "Pilz",
	"Pirat", "Planet", "Platte", "Rakete", "Ratte", "Regen", "Ring", "Ritter",
	"Robot", "Rose", "Sack", "Salz", "Sand", "Saege", "Schatz", "Schere",
	"Schild", "Schlange", "Schloss", "Schluessel", "Schnee", "Schuh", "Schwert", "See",
	"Seife", "Seil", "Sonne", "Spinne", "Spiegel", "Stein", "Stern", "Stiefel",
	"Stift", "Stuhl", "Tasche", "Teller", "Tier", "Tiger", "Tor", "Traktor",
	"Turm", "Uhr", "Vampir", "Vase", "Vogel", "Vulkan", "Waage", "Wagen",
	"Wald", "Wand", "Wasser", "Welle", "Welt", "Wiese", "Wind", "Wolke",
	"Wurm", "Zahn", "Zaun", "Zelt", "Ziege", "Zug", "Zwerg", "Anker",
	"Ballon", "Bank", "Bart", "Biene", "Birne", "Blatt", "Blitz", "Dose",
	"Eule", "Fahne", "Falle", "Fass", "Finger", "Flamme", "Fliege", "Fluss",
	"Gabel", "Gift", "Glas", "Grenze", "Harfe", "Helm", "Hexe", "Horn",
	"Igel", "Kamin", "Kissen", "Klinge", "Kreuz", "Kugel", "Lager", "Markt",
	"Nebel", "Oper", "Orgel", "Peitsche", "Pfeil", "Quelle", "Rad", "Rauch",
	"Ruine", "Sattel", "Schatten", "Sense", "Socke", "Stab", "Tempel", "Tinte",
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
