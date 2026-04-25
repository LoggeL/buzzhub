package creative

import (
	"context"
	"math/rand"
	"time"

	"github.com/logge/buzzhub/internal/game"
)

type variant struct {
	ID          string
	Name        string
	Description string
	Icon        string
	PromptLabel string
	InputLabel  string
	Placeholder string
	VoteTitle   string
	Prompts     []string
}

type submission struct {
	PlayerID string `json:"playerId"`
	Answer   string `json:"answer"`
}

type Creative struct {
	variant      variant
	players      []string
	prompts      []string
	currentRound int
	totalRounds  int
	submissions  map[string]string
	votes        map[string]string
	scores       map[string]int
	phase        string
}

func init() {
	for _, v := range variants {
		mode := v
		game.Register(mode.ID, func() game.Game { return &Creative{variant: mode} })
	}
}

func (c *Creative) Info() game.GameInfo {
	return game.GameInfo{
		ID:          c.variant.ID,
		Name:        c.variant.Name,
		Description: c.variant.Description,
		MinPlayers:  3,
		MaxPlayers:  16,
		Icon:        c.variant.Icon,
	}
}

func (c *Creative) Init(_ context.Context, players []string) error {
	c.players = players
	c.scores = make(map[string]int)
	c.totalRounds = 4
	if len(players) >= 6 {
		c.totalRounds = 5
	}
	c.prompts = pick(c.variant.Prompts, c.totalRounds)
	c.currentRound = -1
	c.phase = ""
	for _, p := range players {
		c.scores[p] = 0
	}
	return nil
}

func (c *Creative) Phases() []string {
	return []string{"prompt", "vote", "results", "scoreboard"}
}

func (c *Creative) NextPhase(_ context.Context) (*game.Phase, error) {
	switch c.phase {
	case "":
		c.currentRound = 0
		c.phase = "prompt"
		return c.promptPhase()
	case "prompt":
		c.phase = "vote"
		return c.votePhase()
	case "vote":
		c.phase = "results"
		return c.resultsPhase()
	case "results":
		c.currentRound++
		if c.currentRound >= c.totalRounds {
			c.phase = "scoreboard"
			return c.scoreboardPhase()
		}
		c.phase = "prompt"
		return c.promptPhase()
	case "scoreboard":
		return nil, nil
	}
	return nil, nil
}

func (c *Creative) promptPhase() (*game.Phase, error) {
	c.submissions = make(map[string]string)
	c.votes = make(map[string]string)

	data := c.baseData()
	return &game.Phase{
		Name:          "prompt",
		Duration:      35 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (c *Creative) votePhase() (*game.Phase, error) {
	var subs []submission
	for pid, answer := range c.submissions {
		subs = append(subs, submission{PlayerID: pid, Answer: answer})
	}

	playerData := make(map[string]any)
	for _, pid := range c.players {
		filtered := make([]submission, 0, len(subs))
		for _, s := range subs {
			if s.PlayerID != pid {
				filtered = append(filtered, s)
			}
		}
		data := c.baseData()
		data["submissions"] = filtered
		playerData[pid] = data
	}

	hostData := c.baseData()
	hostData["submissions"] = subs
	return &game.Phase{
		Name:       "vote",
		Duration:   25 * time.Second,
		HostData:   hostData,
		PlayerData: playerData,
	}, nil
}

func (c *Creative) resultsPhase() (*game.Phase, error) {
	voteCounts := make(map[string]int)
	for _, votedFor := range c.votes {
		voteCounts[votedFor]++
	}

	roundScores := make(map[string]int)
	for pid, count := range voteCounts {
		points := count * 100
		roundScores[pid] = points
		c.scores[pid] += points
	}

	data := c.baseData()
	data["submissions"] = c.submissions
	data["votes"] = c.votes
	data["voteCounts"] = voteCounts
	data["roundScores"] = roundScores
	data["scores"] = c.scores
	return &game.Phase{
		Name:          "results",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (c *Creative) scoreboardPhase() (*game.Phase, error) {
	data := map[string]any{
		"modeName": c.variant.Name,
		"scores":   c.scores,
		"final":    true,
	}
	return &game.Phase{
		Name:          "scoreboard",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (c *Creative) HandleEvent(_ context.Context, event game.PlayerEvent) (*game.StateUpdate, error) {
	switch event.Type {
	case "submit":
		if c.phase != "prompt" {
			return nil, nil
		}
		answer, _ := event.Data["answer"].(string)
		if answer == "" {
			return nil, nil
		}
		c.submissions[event.PlayerID] = answer
		return &game.StateUpdate{
			BroadcastUpdate: map[string]any{
				"submittedCount": len(c.submissions),
				"totalPlayers":   len(c.players),
			},
			PhaseComplete: len(c.submissions) >= len(c.players),
		}, nil

	case "vote":
		if c.phase != "vote" {
			return nil, nil
		}
		votedFor, _ := event.Data["playerId"].(string)
		if votedFor == "" || votedFor == event.PlayerID {
			return nil, nil
		}
		c.votes[event.PlayerID] = votedFor
		return &game.StateUpdate{
			BroadcastUpdate: map[string]any{
				"votedCount":   len(c.votes),
				"totalPlayers": len(c.players),
			},
			PhaseComplete: len(c.votes) >= len(c.players),
		}, nil
	}
	return nil, nil
}

func (c *Creative) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	return nil, nil
}

func (c *Creative) Scores() map[string]int {
	return c.scores
}

func (c *Creative) Cleanup() {}

func (c *Creative) baseData() map[string]any {
	return map[string]any{
		"modeId":      c.variant.ID,
		"modeName":    c.variant.Name,
		"prompt":      c.prompts[c.currentRound],
		"promptLabel": c.variant.PromptLabel,
		"inputLabel":  c.variant.InputLabel,
		"placeholder": c.variant.Placeholder,
		"voteTitle":   c.variant.VoteTitle,
		"roundNum":    c.currentRound + 1,
		"totalRounds": c.totalRounds,
	}
}

func pick(items []string, n int) []string {
	shuffled := make([]string, len(items))
	copy(shuffled, items)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	if n > len(shuffled) {
		n = len(shuffled)
	}
	return shuffled[:n]
}

var variants = []variant{
	{
		ID: "headlines", Name: "Clickbait After Dark", Icon: "devil",
		Description: "Ergaenze schamlose Schlagzeilen und vote den groessten Teaser.",
		PromptLabel: "Schlagzeile", InputLabel: "Fehlende Passage", Placeholder: "Exklusiv: ...",
		VoteTitle: "Welcher Teaser bekommt den Klick?",
		Prompts: []string{
			"Lokaler Mann wird beruehmt, nachdem er _____",
			"Studie beweist: Wer _____, lebt angeblich laenger",
			"Niemand im Club rechnete damit, dass _____",
			"Influencer entschuldigt sich fuer _____",
			"Deutschland diskutiert ueber neues Verbot von _____",
			"Forscher finden heraus: _____ ist doch gesund",
		},
	},
	{
		ID: "redflags", Name: "Red Flags Only", Icon: "flag",
		Description: "Schreibe Warnsignale fuer Profile, Chats und fragwuerdige Anzeigen.",
		PromptLabel: "Profil", InputLabel: "Rote Flagge", Placeholder: "Hat definitiv ...",
		VoteTitle: "Welche rote Flagge ist am heftigsten?",
		Prompts: []string{
			"Dating-Profil: 'Ich bin spontan, ehrlich und liebe Abenteuer.'",
			"WG-Anzeige: 'Wir sind wie eine kleine Familie.'",
			"Jobanzeige: 'Wir suchen belastbare Teamplayer mit Humor.'",
			"Airbnb-Beschreibung: 'Rustikal, zentral, authentisch.'",
			"Kleinanzeige: 'Fast neu, nur einmal benutzt.'",
			"Festival-Gruppe: 'Wir sind komplett entspannt organisiert.'",
		},
	},
	{
		ID: "courtroom", Name: "Dirty Gericht", Icon: "scale",
		Description: "Verteidige oder zerlege absurde Party-Vergehen vor der Runde.",
		PromptLabel: "Anklage", InputLabel: "Statement", Placeholder: "Euer Ehren, ...",
		VoteTitle: "Welches Statement ueberzeugt das Gericht?",
		Prompts: []string{
			"Angeklagt wegen AUX-Kaperung um 3:12 Uhr.",
			"Angeklagt wegen 'kurz weg' und drei Stunden verschwunden.",
			"Angeklagt wegen leerem Akku trotz Powerbank.",
			"Angeklagt wegen Gruppenfoto-Sabotage.",
			"Angeklagt wegen Pizza mit Ananas auf der Hausparty.",
			"Angeklagt wegen Spoiler im falschen Moment.",
		},
	},
	{
		ID: "emoji", Name: "Emoji Beichte", Icon: "smirk",
		Description: "Deute zweideutige Geschichten aus reinem Emoji-Chaos.",
		PromptLabel: "Emoji-Beichte", InputLabel: "Interpretation", Placeholder: "Das sieht sehr danach aus, dass ...",
		VoteTitle: "Welche Deutung passt am besten?",
		Prompts: []string{
			"🍕🚕😭🕺",
			"📱💔🍻🧦",
			"🚿👟⏰💀",
			"🎤🐸🚨🙈",
			"🛒🥒👀🏃",
			"🧃💻🔥🤷",
		},
	},
	{
		ID: "wouldrather", Name: "Wer Wuerde Eher: After Dark", Icon: "eyes",
		Description: "Begruende anonym, wer am ehesten eskaliert.",
		PromptLabel: "Frage", InputLabel: "Begruendung", Placeholder: "Ganz klar, weil ...",
		VoteTitle: "Welche Begruendung trifft am haertesten?",
		Prompts: []string{
			"Wer wuerde eher einen fremden Hund auf einer Party adoptieren?",
			"Wer wuerde eher im Urlaub aus Versehen eine neue Identitaet anfangen?",
			"Wer wuerde eher bei Karaoke zu ernst werden?",
			"Wer wuerde eher eine Entschuldigung als PowerPoint halten?",
			"Wer wuerde eher einen Streit mit Google Maps verlieren?",
			"Wer wuerde eher eine Sprachnachricht als Podcast produzieren?",
		},
	},
	{
		ID: "cvlies", Name: "Sexy Lebenslauf", Icon: "briefcase",
		Description: "Erfinde fragwuerdige Qualifikationen fuer noch fragwuerdigere Jobs.",
		PromptLabel: "Jobtitel", InputLabel: "Qualifikation", Placeholder: "Zertifiziert in ...",
		VoteTitle: "Welche Qualifikation klingt am glaubwuerdigsten?",
		Prompts: []string{
			"Professioneller Warteschlangen-Berater",
			"Chief Snack Officer",
			"Senior Balkon-Philosoph",
			"Freiberuflicher Entschuldigungsdesigner",
			"Festival-Logistik-Orakel",
			"Influencer-Krisenmanager fuer Haustiere",
		},
	},
	{
		ID: "memecourt", Name: "Meme Lust", Icon: "camera",
		Description: "Schreibe Captions fuer Bilder, die niemand erklaeren will.",
		PromptLabel: "Vorlage", InputLabel: "Caption", Placeholder: "Wenn du ...",
		VoteTitle: "Welche Caption wuerde viral gehen?",
		Prompts: []string{
			"Bild: Eine Person laechelt, waehrend im Hintergrund alles brennt.",
			"Bild: Drei Freunde schauen schuldig auf ein Handy.",
			"Bild: Ein leerer Teller und sehr viel Erklaerungsbedarf.",
			"Bild: Jemand zeigt stolz auf eine komplett falsche Loesung.",
			"Bild: Gruppenchat explodiert, eine Person schreibt nur 'ok'.",
			"Bild: Der DJ schaut panisch auf den Laptop.",
		},
	},
	{
		ID: "therapy", Name: "Aftercare Therapie", Icon: "couch",
		Description: "Gib fragwuerdige Ratschlaege fuer kleine Dramen nach der Eskalation.",
		PromptLabel: "Problem", InputLabel: "Ratschlag", Placeholder: "Ich fuehle da rein und sage ...",
		VoteTitle: "Welcher Ratschlag ist am schlimmsten hilfreich?",
		Prompts: []string{
			"Mein Nachbar grillt um 3 Uhr morgens.",
			"Ich habe aus Versehen 'Liebe Gruesse' an meinen Chef geschrieben.",
			"Meine Playlist wurde auf einer Party uebersprungen.",
			"Ich kann nicht aufhoeren, Lieferstatus zu aktualisieren.",
			"Meine Pflanze sieht mich enttaeuscht an.",
			"Ich habe eine Sprachnachricht abgebrochen und jetzt ist alles komisch.",
		},
	},
	{
		ID: "passwordpanic", Name: "Passwort Panik", Icon: "key",
		Description: "Errate Login-Geheimnisse, die besser verborgen geblieben waeren.",
		PromptLabel: "Account", InputLabel: "Passwort", Placeholder: "Sommer2024! ...",
		VoteTitle: "Welches Passwort ist leider realistisch?",
		Prompts: []string{
			"Omas TikTok-Account",
			"Der alte WG-Netflix-Account",
			"Das geheime Vereins-WLAN",
			"Der Laptop vom DJ",
			"Ein Fitnessstudio-Login aus 2017",
			"Der Familien-Drucker im Homeoffice",
		},
	},
	{
		ID: "lastwords", Name: "Letzter Hoehepunkt", Icon: "message",
		Description: "Schreibe den letzten Satz direkt vor der kompletten Eskalation.",
		PromptLabel: "Szene", InputLabel: "Letzter Satz", Placeholder: "Keine Sorge, ich ...",
		VoteTitle: "Welcher letzte Satz trifft am besten?",
		Prompts: []string{
			"Du wachst im Clubklo auf und dein Handy zeigt 47 verpasste Anrufe.",
			"Der Grill explodiert fast, alle schauen dich an.",
			"Du stehst mit einem Einkaufswagen im Aufzug.",
			"Die Musik stoppt genau in deinem lautesten Moment.",
			"Du findest einen fremden Hut in deiner Tasche.",
			"Der Gruppenchat hat deinen Screenshot gesehen.",
		},
	},
}
