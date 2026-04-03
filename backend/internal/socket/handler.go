package socket

import (
	"context"
	"log"
	"sync"

	"github.com/logge/buzzhub/internal/engine"
	gamepkg "github.com/logge/buzzhub/internal/game"
	"github.com/logge/buzzhub/internal/lobby"

	"github.com/google/uuid"
	"github.com/zishang520/socket.io/v2/socket"
)

type Handler struct {
	io       *socket.Server
	lobbyMgr *lobby.Manager
	engine   *engine.Engine

	// Maps socketID -> {lobbyCode, playerID}
	connections sync.Map
}

type connInfo struct {
	LobbyCode string
	PlayerID  string
	IsHost    bool
}

func New(io *socket.Server, lobbyMgr *lobby.Manager, eng *engine.Engine) *Handler {
	return &Handler{
		io:       io,
		lobbyMgr: lobbyMgr,
		engine:   eng,
	}
}

func (h *Handler) Setup() {
	h.engine.SetEmitFunc(h.emitGameUpdate)
	h.engine.SetPhaseEndFunc(h.emitGameEnd)

	h.io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		log.Printf("Client connected: %s", client.Id())

		client.On("lobby:create", func(args ...any) {
			h.handleLobbyCreate(client, args)
		})

		client.On("lobby:join", func(args ...any) {
			h.handleLobbyJoin(client, args)
		})

		client.On("lobby:rejoin", func(args ...any) {
			h.handleLobbyRejoin(client, args)
		})

		client.On("host:join", func(args ...any) {
			h.handleHostJoin(client, args)
		})

		client.On("lobby:select-game", func(args ...any) {
			h.handleSelectGame(client, args)
		})

		client.On("lobby:start-game", func(args ...any) {
			h.handleStartGame(client, args)
		})

		client.On("lobby:kick", func(args ...any) {
			h.handleKick(client, args)
		})

		client.On("lobby:leave", func(args ...any) {
			h.handleLeave(client, args)
		})

		client.On("player:action", func(args ...any) {
			h.handlePlayerAction(client, args)
		})

		client.On("disconnect", func(args ...any) {
			h.handleDisconnect(client)
		})
	})
}

func (h *Handler) handleLobbyCreate(client *socket.Socket, args []any) {
	data := toMap(args)
	name, _ := data["name"].(string)
	if name == "" {
		client.Emit("lobby:error", map[string]any{"message": "Name ist erforderlich"})
		return
	}

	ctx := context.Background()
	l, playerID, err := h.lobbyMgr.Create(ctx, name)
	if err != nil {
		client.Emit("lobby:error", map[string]any{"message": err.Error()})
		return
	}

	token := uuid.New().String()
	h.lobbyMgr.SaveSession(ctx, token, &lobby.Session{
		LobbyCode: l.Code,
		PlayerID:  playerID,
		Name:      name,
	})

	h.connections.Store(string(client.Id()), &connInfo{
		LobbyCode: l.Code,
		PlayerID:  playerID,
	})

	client.Join(roomKey(l.Code))
	client.Join(roomPlayersKey(l.Code))
	client.Join(playerRoom(playerID))

	client.Emit("lobby:created", map[string]any{
		"code":     l.Code,
		"playerId": playerID,
		"token":    token,
		"lobby":    l,
	})
}

func (h *Handler) handleLobbyJoin(client *socket.Socket, args []any) {
	data := toMap(args)
	code, _ := data["code"].(string)
	name, _ := data["name"].(string)
	if code == "" || name == "" {
		client.Emit("lobby:error", map[string]any{"message": "Code und Name sind erforderlich"})
		return
	}

	ctx := context.Background()
	l, playerID, err := h.lobbyMgr.Join(ctx, code, name)
	if err != nil {
		client.Emit("lobby:error", map[string]any{"message": err.Error()})
		return
	}

	token := uuid.New().String()
	h.lobbyMgr.SaveSession(ctx, token, &lobby.Session{
		LobbyCode: l.Code,
		PlayerID:  playerID,
		Name:      name,
	})

	h.connections.Store(string(client.Id()), &connInfo{
		LobbyCode: l.Code,
		PlayerID:  playerID,
	})

	client.Join(roomKey(l.Code))
	client.Join(roomPlayersKey(l.Code))
	client.Join(playerRoom(playerID))

	client.Emit("lobby:joined", map[string]any{
		"playerId": playerID,
		"token":    token,
		"lobby":    l,
	})

	h.io.To(roomKey(l.Code)).Emit("lobby:player-joined", map[string]any{
		"player": lobby.Player{ID: playerID, Name: name, Connected: true},
		"lobby":  l,
	})
}

func (h *Handler) handleLobbyRejoin(client *socket.Socket, args []any) {
	data := toMap(args)
	token, _ := data["token"].(string)
	if token == "" {
		client.Emit("lobby:error", map[string]any{"message": "Token fehlt"})
		return
	}

	ctx := context.Background()
	sess, err := h.lobbyMgr.GetSession(ctx, token)
	if err != nil {
		client.Emit("lobby:error", map[string]any{"message": "Session abgelaufen"})
		return
	}

	h.lobbyMgr.SetPlayerConnected(ctx, sess.LobbyCode, sess.PlayerID, true)

	h.connections.Store(string(client.Id()), &connInfo{
		LobbyCode: sess.LobbyCode,
		PlayerID:  sess.PlayerID,
	})

	client.Join(roomKey(sess.LobbyCode))
	client.Join(roomPlayersKey(sess.LobbyCode))
	client.Join(playerRoom(sess.PlayerID))

	l, _ := h.lobbyMgr.Get(ctx, sess.LobbyCode)
	client.Emit("lobby:joined", map[string]any{
		"playerId": sess.PlayerID,
		"token":    token,
		"lobby":    l,
	})
}

func (h *Handler) handleHostJoin(client *socket.Socket, args []any) {
	data := toMap(args)
	code, _ := data["code"].(string)
	if code == "" {
		client.Emit("lobby:error", map[string]any{"message": "Code fehlt"})
		return
	}

	ctx := context.Background()
	l, err := h.lobbyMgr.Get(ctx, code)
	if err != nil {
		client.Emit("lobby:error", map[string]any{"message": "Lobby nicht gefunden"})
		return
	}

	h.connections.Store(string(client.Id()), &connInfo{
		LobbyCode: code,
		IsHost:    true,
	})

	client.Join(roomKey(code))
	client.Join(roomHostKey(code))

	client.Emit("lobby:state", map[string]any{"lobby": l})
}

func (h *Handler) handleSelectGame(client *socket.Socket, args []any) {
	info := h.getConn(client)
	if info == nil {
		return
	}

	data := toMap(args)
	gameID, _ := data["gameId"].(string)

	ctx := context.Background()
	l, err := h.lobbyMgr.SetGame(ctx, info.LobbyCode, info.PlayerID, gameID)
	if err != nil {
		client.Emit("lobby:error", map[string]any{"message": err.Error()})
		return
	}

	h.io.To(roomKey(info.LobbyCode)).Emit("lobby:state", map[string]any{"lobby": l})
}

func (h *Handler) handleStartGame(client *socket.Socket, args []any) {
	info := h.getConn(client)
	if info == nil {
		return
	}

	ctx := context.Background()
	l, err := h.lobbyMgr.Get(ctx, info.LobbyCode)
	if err != nil || l.HostID != info.PlayerID {
		client.Emit("lobby:error", map[string]any{"message": "Nur der Host kann das Spiel starten"})
		return
	}

	if l.GameID == "" {
		client.Emit("lobby:error", map[string]any{"message": "Kein Spiel ausgewählt"})
		return
	}

	g, _ := gamepkg.Create(l.GameID)
	gi := g.Info()
	if len(l.Players) < gi.MinPlayers {
		client.Emit("lobby:error", map[string]any{
			"message": "Zu wenige Spieler (mindestens " + string(rune('0'+gi.MinPlayers)) + ")",
		})
		return
	}

	h.io.To(roomKey(info.LobbyCode)).Emit("game:start", map[string]any{
		"gameId": l.GameID,
		"game":   gi,
	})

	if err := h.engine.StartGame(ctx, info.LobbyCode, l.GameID); err != nil {
		client.Emit("lobby:error", map[string]any{"message": err.Error()})
	}
}

func (h *Handler) handleKick(client *socket.Socket, args []any) {
	info := h.getConn(client)
	if info == nil {
		return
	}

	data := toMap(args)
	targetID, _ := data["playerId"].(string)

	ctx := context.Background()
	l, err := h.lobbyMgr.Kick(ctx, info.LobbyCode, info.PlayerID, targetID)
	if err != nil {
		client.Emit("lobby:error", map[string]any{"message": err.Error()})
		return
	}

	h.io.To(roomKey(info.LobbyCode)).Emit("lobby:player-left", map[string]any{
		"playerId": targetID,
		"lobby":    l,
		"kicked":   true,
	})
}

func (h *Handler) handleLeave(client *socket.Socket, args []any) {
	info := h.getConn(client)
	if info == nil {
		return
	}

	ctx := context.Background()
	l, err := h.lobbyMgr.Leave(ctx, info.LobbyCode, info.PlayerID)
	if err != nil {
		return
	}

	client.Leave(roomKey(info.LobbyCode))
	client.Leave(roomPlayersKey(info.LobbyCode))
	client.Leave(playerRoom(info.PlayerID))
	h.connections.Delete(string(client.Id()))

	if l != nil {
		h.io.To(roomKey(info.LobbyCode)).Emit("lobby:player-left", map[string]any{
			"playerId": info.PlayerID,
			"lobby":    l,
		})
	}
}

func (h *Handler) handlePlayerAction(client *socket.Socket, args []any) {
	info := h.getConn(client)
	if info == nil {
		return
	}

	data := toMap(args)
	actionType, _ := data["type"].(string)

	event := gamepkg.PlayerEvent{
		PlayerID: info.PlayerID,
		Type:     actionType,
		Data:     data,
	}

	ctx := context.Background()
	if err := h.engine.HandleEvent(ctx, info.LobbyCode, event); err != nil {
		client.Emit("game:error", map[string]any{"message": err.Error()})
	}
}

func (h *Handler) handleDisconnect(client *socket.Socket) {
	info := h.getConn(client)
	if info == nil {
		return
	}

	log.Printf("Client disconnected: %s (lobby: %s)", client.Id(), info.LobbyCode)

	if !info.IsHost {
		ctx := context.Background()
		h.lobbyMgr.SetPlayerConnected(ctx, info.LobbyCode, info.PlayerID, false)

		h.io.To(roomKey(info.LobbyCode)).Emit("lobby:state", map[string]any{
			"type": "player-disconnected",
			"playerId": info.PlayerID,
		})
	}

	h.connections.Delete(string(client.Id()))
}

func (h *Handler) emitGameUpdate(code string, update *gamepkg.StateUpdate, phase *gamepkg.Phase) {
	if phase != nil {
		if phase.HostData != nil {
			h.io.To(roomHostKey(code)).Emit("game:phase", map[string]any{
				"phase": phase.Name,
				"data":  phase.HostData,
			})
		}
		if phase.PlayerData != nil {
			for pid, data := range phase.PlayerData {
				h.io.To(playerRoom(pid)).Emit("game:phase", map[string]any{
					"phase": phase.Name,
					"data":  data,
				})
			}
		}
		if phase.BroadcastData != nil {
			h.io.To(roomPlayersKey(code)).Emit("game:phase", map[string]any{
				"phase": phase.Name,
				"data":  phase.BroadcastData,
			})
		}
		if phase.Duration > 0 {
			h.io.To(roomKey(code)).Emit("game:timer", map[string]any{
				"duration": phase.Duration.Seconds(),
			})
		}
		return
	}

	if update != nil {
		if update.HostUpdate != nil {
			h.io.To(roomHostKey(code)).Emit("game:update", update.HostUpdate)
		}
		if update.PlayerUpdates != nil {
			for pid, data := range update.PlayerUpdates {
				h.io.To(playerRoom(pid)).Emit("game:update", data)
			}
		}
		if update.BroadcastUpdate != nil {
			h.io.To(roomPlayersKey(code)).Emit("game:update", update.BroadcastUpdate)
		}
		if update.Scores != nil {
			h.io.To(roomKey(code)).Emit("game:update", map[string]any{
				"scores": update.Scores,
			})
		}
		if update.GameOver {
			h.io.To(roomKey(code)).Emit("game:end", map[string]any{
				"scores": update.Scores,
			})
		}
	}
}

func (h *Handler) emitGameEnd(code string) {
	ctx := context.Background()
	l, _ := h.lobbyMgr.Get(ctx, code)
	if l != nil {
		h.io.To(roomKey(code)).Emit("lobby:state", map[string]any{"lobby": l})
	}
}

func (h *Handler) getConn(client *socket.Socket) *connInfo {
	val, ok := h.connections.Load(string(client.Id()))
	if !ok {
		return nil
	}
	return val.(*connInfo)
}

func roomKey(code string) socket.Room         { return socket.Room("room:" + code) }
func roomHostKey(code string) socket.Room     { return socket.Room("room:" + code + ":host") }
func roomPlayersKey(code string) socket.Room  { return socket.Room("room:" + code + ":players") }
func playerRoom(playerID string) socket.Room  { return socket.Room("player:" + playerID) }

func toMap(args []any) map[string]any {
	if len(args) == 0 {
		return map[string]any{}
	}
	if m, ok := args[0].(map[string]any); ok {
		return m
	}
	return map[string]any{}
}
