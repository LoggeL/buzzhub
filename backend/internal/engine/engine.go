package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/logge/buzzhub/internal/game"
	"github.com/logge/buzzhub/internal/lobby"
)

type EmitFunc func(code string, update *game.StateUpdate, phase *game.Phase)
type PhaseEndFunc func(code string)

type GameSession struct {
	Game      game.Game
	Code      string
	Timer     *time.Timer
	Phase     int
	mu        sync.Mutex
}

type Engine struct {
	lobbyMgr *lobby.Manager
	sessions map[string]*GameSession
	mu       sync.RWMutex
	onEmit   EmitFunc
	onEnd    PhaseEndFunc
}

func New(lobbyMgr *lobby.Manager) *Engine {
	return &Engine{
		lobbyMgr: lobbyMgr,
		sessions: make(map[string]*GameSession),
	}
}

func (e *Engine) SetEmitFunc(fn EmitFunc) {
	e.onEmit = fn
}

func (e *Engine) SetPhaseEndFunc(fn PhaseEndFunc) {
	e.onEnd = fn
}

func (e *Engine) StartGame(ctx context.Context, code, gameID string, settings map[string]any) error {
	l, err := e.lobbyMgr.Get(ctx, code)
	if err != nil {
		return err
	}

	g, err := game.Create(gameID)
	if err != nil {
		return err
	}

	var playerIDs []string
	for _, p := range l.Players {
		playerIDs = append(playerIDs, p.ID)
	}

	if err := g.Init(ctx, playerIDs); err != nil {
		return err
	}

	// Apply game-specific settings if supported
	if cfg, ok := g.(game.Configurable); ok && settings != nil {
		cfg.Configure(settings)
	}

	if err := e.lobbyMgr.SetStatus(ctx, code, "playing"); err != nil {
		return err
	}

	session := &GameSession{
		Game: g,
		Code: code,
	}

	e.mu.Lock()
	e.sessions[code] = session
	e.mu.Unlock()

	// Delay first phase so clients have time to navigate to the game page
	go func() {
		time.Sleep(2 * time.Second)
		if err := e.advancePhase(context.Background(), session); err != nil {
			log.Printf("first phase error for %s: %v", code, err)
		}
	}()

	return nil
}

func (e *Engine) HandleEvent(ctx context.Context, code string, event game.PlayerEvent) error {
	session := e.getSession(code)
	if session == nil {
		return nil
	}

	session.mu.Lock()

	update, err := session.Game.HandleEvent(ctx, event)
	if err != nil {
		session.mu.Unlock()
		return err
	}

	if update == nil {
		session.mu.Unlock()
		return nil
	}

	if e.onEmit != nil {
		e.onEmit(code, update, nil)
	}

	if update.GameOver {
		session.mu.Unlock()
		e.endGame(ctx, session)
		return nil
	}

	if update.PhaseComplete {
		if session.Timer != nil {
			session.Timer.Stop()
		}
		session.mu.Unlock()
		// Small delay so clients see the "all answered" state
		go func() {
			time.Sleep(500 * time.Millisecond)
			if err := e.advancePhase(context.Background(), session); err != nil {
				log.Printf("auto-advance error for %s: %v", code, err)
			}
		}()
		return nil
	}

	session.mu.Unlock()
	return nil
}

func (e *Engine) advancePhase(ctx context.Context, session *GameSession) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	phase, err := session.Game.NextPhase(ctx)
	if err != nil {
		return err
	}

	if phase == nil {
		e.endGame(ctx, session)
		return nil
	}

	session.Phase++

	if e.onEmit != nil {
		e.onEmit(session.Code, nil, phase)
	}

	if phase.Duration > 0 {
		if session.Timer != nil {
			session.Timer.Stop()
		}
		session.Timer = time.AfterFunc(phase.Duration, func() {
			e.handleTimerExpired(session)
		})
	}

	return nil
}

func (e *Engine) handleTimerExpired(session *GameSession) {
	ctx := context.Background()
	session.mu.Lock()

	update, err := session.Game.TimerExpired(ctx)
	if err != nil {
		log.Printf("timer error for %s: %v", session.Code, err)
		session.mu.Unlock()
		return
	}

	if update != nil && e.onEmit != nil {
		e.onEmit(session.Code, update, nil)
	}
	session.mu.Unlock()

	if update != nil && update.GameOver {
		e.endGame(ctx, session)
		return
	}

	// Advance to next phase
	if err := e.advancePhase(ctx, session); err != nil {
		log.Printf("phase advance error for %s: %v", session.Code, err)
	}
}

func (e *Engine) endGame(ctx context.Context, session *GameSession) {
	scores := session.Game.Scores()
	if len(scores) > 0 {
		e.lobbyMgr.UpdateScores(ctx, session.Code, scores)
	}
	e.lobbyMgr.SetStatus(ctx, session.Code, "waiting")
	session.Game.Cleanup()

	if session.Timer != nil {
		session.Timer.Stop()
	}

	e.mu.Lock()
	delete(e.sessions, session.Code)
	e.mu.Unlock()

	if e.onEnd != nil {
		e.onEnd(session.Code)
	}
}

func (e *Engine) getSession(code string) *GameSession {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.sessions[code]
}

func (e *Engine) GetCurrentGame(code string) game.Game {
	s := e.getSession(code)
	if s == nil {
		return nil
	}
	return s.Game
}
