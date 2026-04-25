package quiz

import (
	"context"
	"math/rand"
	"time"

	"github.com/logge/buzzhub/internal/game"
)

func init() {
	game.Register("quiz", func() game.Game { return &Quiz{} })
}

type Question struct {
	Text    string   `json:"text"`
	Options []string `json:"options"`
	Correct int      `json:"correct"`
}

type Quiz struct {
	players     []string
	questions   []Question
	currentQ    int
	answers     map[string]int // playerID -> chosen option index
	scores      map[string]int
	totalRounds int
	phase       string
}

func (q *Quiz) Info() game.GameInfo {
	return game.GameInfo{
		ID:          "quiz",
		Name:        "Heisser Quickie",
		Description: "Drueck den Buzzer schneller als die anderen.",
		MinPlayers:  2,
		MaxPlayers:  16,
		Icon:        "fire",
	}
}

func (q *Quiz) Init(_ context.Context, players []string) error {
	q.players = players
	q.scores = make(map[string]int)
	q.totalRounds = 8
	if len(players) <= 4 {
		q.totalRounds = 6
	}
	q.questions = pickQuestions(q.totalRounds)
	q.currentQ = -1
	for _, p := range players {
		q.scores[p] = 0
	}
	return nil
}

func (q *Quiz) Phases() []string {
	return []string{"question", "reveal", "scoreboard"}
}

func (q *Quiz) NextPhase(_ context.Context) (*game.Phase, error) {
	switch q.phase {
	case "":
		q.phase = "question"
		return q.nextQuestion()
	case "question":
		q.phase = "reveal"
		return q.revealPhase()
	case "reveal":
		q.currentQ++
		if q.currentQ >= len(q.questions) {
			q.phase = "scoreboard"
			return q.scoreboardPhase()
		}
		q.phase = "question"
		return q.questionPhase()
	case "scoreboard":
		return nil, nil // game over
	}
	return nil, nil
}

func (q *Quiz) nextQuestion() (*game.Phase, error) {
	q.currentQ = 0
	q.answers = make(map[string]int)
	return q.questionPhase()
}

func (q *Quiz) questionPhase() (*game.Phase, error) {
	q.answers = make(map[string]int)
	question := q.questions[q.currentQ]

	data := map[string]any{
		"question":       question.Text,
		"options":        question.Options,
		"questionNum":    q.currentQ + 1,
		"totalQuestions": len(q.questions),
	}

	return &game.Phase{
		Name:          "question",
		Duration:      15 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (q *Quiz) revealPhase() (*game.Phase, error) {
	question := q.questions[q.currentQ]

	// Calculate points for this round
	roundScores := make(map[string]int)
	for pid, answer := range q.answers {
		if answer == question.Correct {
			roundScores[pid] = 100
			q.scores[pid] += 100
		}
	}

	data := map[string]any{
		"correct":     question.Correct,
		"correctText": question.Options[question.Correct],
		"answers":     q.answers,
		"roundScores": roundScores,
		"scores":      q.scores,
		"questionNum": q.currentQ + 1,
	}

	return &game.Phase{
		Name:          "reveal",
		Duration:      5 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (q *Quiz) scoreboardPhase() (*game.Phase, error) {
	data := map[string]any{
		"scores": q.scores,
		"final":  true,
	}
	return &game.Phase{
		Name:          "scoreboard",
		Duration:      8 * time.Second,
		HostData:      data,
		BroadcastData: data,
	}, nil
}

func (q *Quiz) HandleEvent(_ context.Context, event game.PlayerEvent) (*game.StateUpdate, error) {
	if q.phase != "question" {
		return nil, nil
	}

	if event.Type != "answer" {
		return nil, nil
	}

	answerIdx, ok := event.Data["answer"].(float64)
	if !ok {
		return nil, nil
	}

	if _, alreadyAnswered := q.answers[event.PlayerID]; alreadyAnswered {
		return nil, nil
	}

	q.answers[event.PlayerID] = int(answerIdx)

	allAnswered := len(q.answers) >= len(q.players)

	update := &game.StateUpdate{
		BroadcastUpdate: map[string]any{
			"answeredCount": len(q.answers),
			"totalPlayers":  len(q.players),
		},
		PhaseComplete: allAnswered,
	}

	return update, nil
}

func (q *Quiz) TimerExpired(_ context.Context) (*game.StateUpdate, error) {
	return nil, nil
}

func (q *Quiz) Scores() map[string]int {
	return q.scores
}

func (q *Quiz) Cleanup() {}

var questionBank = []Question{
	{Text: "Welches Land hat die meisten Einwohner?", Options: []string{"Indien", "China", "USA", "Indonesien"}, Correct: 0},
	{Text: "Wie viele Planeten hat unser Sonnensystem?", Options: []string{"7", "8", "9", "10"}, Correct: 1},
	{Text: "Wer malte die Mona Lisa?", Options: []string{"Michelangelo", "Da Vinci", "Raphael", "Rembrandt"}, Correct: 1},
	{Text: "Was ist die Hauptstadt von Australien?", Options: []string{"Sydney", "Melbourne", "Canberra", "Brisbane"}, Correct: 2},
	{Text: "Welches Element hat das Symbol 'Au'?", Options: []string{"Silber", "Aluminium", "Gold", "Argon"}, Correct: 2},
	{Text: "In welchem Jahr fiel die Berliner Mauer?", Options: []string{"1987", "1988", "1989", "1990"}, Correct: 2},
	{Text: "Wie viele Knochen hat ein erwachsener Mensch?", Options: []string{"186", "206", "226", "246"}, Correct: 1},
	{Text: "Welcher Ozean ist der groesste?", Options: []string{"Atlantik", "Indischer Ozean", "Pazifik", "Arktischer Ozean"}, Correct: 2},
	{Text: "Wer schrieb 'Faust'?", Options: []string{"Schiller", "Goethe", "Lessing", "Heine"}, Correct: 1},
	{Text: "Was ist die chemische Formel fuer Wasser?", Options: []string{"H2O", "CO2", "NaCl", "O2"}, Correct: 0},
	{Text: "Welches Tier ist das schnellste an Land?", Options: []string{"Loewe", "Gepard", "Pferd", "Antilope"}, Correct: 1},
	{Text: "Wie viele Bundeslaender hat Deutschland?", Options: []string{"14", "15", "16", "17"}, Correct: 2},
	{Text: "Was ist die Quadratwurzel von 144?", Options: []string{"10", "11", "12", "13"}, Correct: 2},
	{Text: "Welche Farbe entsteht aus Blau und Gelb?", Options: []string{"Orange", "Gruen", "Lila", "Braun"}, Correct: 1},
	{Text: "In welcher Stadt steht der Eiffelturm?", Options: []string{"London", "Berlin", "Paris", "Rom"}, Correct: 2},
	{Text: "Wie viele Seiten hat ein Wuerfel?", Options: []string{"4", "6", "8", "12"}, Correct: 1},
	{Text: "Welches Instrument hat 88 Tasten?", Options: []string{"Gitarre", "Klavier", "Orgel", "Akkordeon"}, Correct: 1},
	{Text: "Was ist die Hauptstadt von Japan?", Options: []string{"Osaka", "Kyoto", "Tokio", "Yokohama"}, Correct: 2},
	{Text: "Wie viele Minuten hat eine Stunde?", Options: []string{"30", "45", "60", "90"}, Correct: 2},
	{Text: "Welcher Planet ist der groesste?", Options: []string{"Saturn", "Jupiter", "Neptun", "Uranus"}, Correct: 1},
}

func pickQuestions(n int) []Question {
	shuffled := make([]Question, len(questionBank))
	copy(shuffled, questionBank)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	if n > len(shuffled) {
		n = len(shuffled)
	}
	return shuffled[:n]
}
