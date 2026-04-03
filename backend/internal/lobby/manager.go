package lobby

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Manager struct {
	store Store
	mu    sync.RWMutex
}

func NewManager(store Store) *Manager {
	return &Manager{store: store}
}

func (m *Manager) Create(ctx context.Context, hostName string) (*Lobby, string, error) {
	code, err := m.generateCode(ctx)
	if err != nil {
		return nil, "", err
	}

	playerID := uuid.New().String()
	l := &Lobby{
		Code:       code,
		HostID:     playerID,
		Status:     "waiting",
		MaxPlayers: 16,
		Players: []Player{
			{ID: playerID, Name: hostName, Connected: true},
		},
		CreatedAt: time.Now(),
	}

	if err := m.store.Save(ctx, l); err != nil {
		return nil, "", err
	}

	return l, playerID, nil
}

func (m *Manager) Join(ctx context.Context, code, name string) (*Lobby, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := m.store.Get(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("Lobby nicht gefunden")
	}

	if l.Status != "waiting" {
		return nil, "", fmt.Errorf("Spiel hat bereits begonnen")
	}

	if len(l.Players) >= l.MaxPlayers {
		return nil, "", fmt.Errorf("Lobby ist voll")
	}

	for _, p := range l.Players {
		if strings.EqualFold(p.Name, name) {
			return nil, "", fmt.Errorf("Name '%s' ist bereits vergeben", name)
		}
	}

	playerID := uuid.New().String()
	l.Players = append(l.Players, Player{
		ID:        playerID,
		Name:      name,
		Connected: true,
	})

	if err := m.store.Save(ctx, l); err != nil {
		return nil, "", err
	}

	return l, playerID, nil
}

func (m *Manager) Leave(ctx context.Context, code, playerID string) (*Lobby, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := m.store.Get(ctx, code)
	if err != nil {
		return nil, err
	}

	for i, p := range l.Players {
		if p.ID == playerID {
			l.Players = append(l.Players[:i], l.Players[i+1:]...)
			break
		}
	}

	if len(l.Players) == 0 {
		return nil, m.store.Delete(ctx, code)
	}

	if l.HostID == playerID && len(l.Players) > 0 {
		l.HostID = l.Players[0].ID
	}

	return l, m.store.Save(ctx, l)
}

func (m *Manager) Kick(ctx context.Context, code, hostID, targetID string) (*Lobby, error) {
	l, err := m.store.Get(ctx, code)
	if err != nil {
		return nil, err
	}
	if l.HostID != hostID {
		return nil, fmt.Errorf("nur der Host kann Spieler kicken")
	}
	return m.Leave(ctx, code, targetID)
}

func (m *Manager) Get(ctx context.Context, code string) (*Lobby, error) {
	return m.store.Get(ctx, code)
}

func (m *Manager) SetGame(ctx context.Context, code, hostID, gameID string) (*Lobby, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := m.store.Get(ctx, code)
	if err != nil {
		return nil, err
	}
	if l.HostID != hostID {
		return nil, fmt.Errorf("nur der Host kann das Spiel wechseln")
	}
	l.GameID = gameID
	return l, m.store.Save(ctx, l)
}

func (m *Manager) SetStatus(ctx context.Context, code, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := m.store.Get(ctx, code)
	if err != nil {
		return err
	}
	l.Status = status
	return m.store.Save(ctx, l)
}

func (m *Manager) SetPlayerConnected(ctx context.Context, code, playerID string, connected bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := m.store.Get(ctx, code)
	if err != nil {
		return err
	}
	for i, p := range l.Players {
		if p.ID == playerID {
			l.Players[i].Connected = connected
			break
		}
	}
	return m.store.Save(ctx, l)
}

func (m *Manager) UpdateScores(ctx context.Context, code string, scores map[string]int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := m.store.Get(ctx, code)
	if err != nil {
		return err
	}
	for i, p := range l.Players {
		if s, ok := scores[p.ID]; ok {
			l.Players[i].Score += s
		}
	}
	return m.store.Save(ctx, l)
}

func (m *Manager) SaveSession(ctx context.Context, token string, session *Session) error {
	return m.store.SaveSession(ctx, token, session)
}

func (m *Manager) GetSession(ctx context.Context, token string) (*Session, error) {
	return m.store.GetSession(ctx, token)
}

func (m *Manager) generateCode(ctx context.Context) (string, error) {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ" // no I, O to avoid confusion
	for attempts := 0; attempts < 100; attempts++ {
		var b strings.Builder
		for i := 0; i < 4; i++ {
			b.WriteByte(letters[rand.Intn(len(letters))])
		}
		code := b.String()
		exists, err := m.store.CodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("could not generate unique code")
}
