package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store interface {
	Save(ctx context.Context, lobby *Lobby) error
	Get(ctx context.Context, code string) (*Lobby, error)
	Delete(ctx context.Context, code string) error
	CodeExists(ctx context.Context, code string) (bool, error)
	SaveSession(ctx context.Context, token string, session *Session) error
	GetSession(ctx context.Context, token string) (*Session, error)
}

type Session struct {
	LobbyCode string `json:"lobbyCode"`
	PlayerID  string `json:"playerId"`
	Name      string `json:"name"`
}

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

func (s *RedisStore) Save(ctx context.Context, l *Lobby) error {
	data, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, lobbyKey(l.Code), data, 2*time.Hour).Err()
}

func (s *RedisStore) Get(ctx context.Context, code string) (*Lobby, error) {
	data, err := s.rdb.Get(ctx, lobbyKey(code)).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("lobby %s not found", code)
	}
	if err != nil {
		return nil, err
	}
	var l Lobby
	return &l, json.Unmarshal(data, &l)
}

func (s *RedisStore) Delete(ctx context.Context, code string) error {
	return s.rdb.Del(ctx, lobbyKey(code)).Err()
}

func (s *RedisStore) CodeExists(ctx context.Context, code string) (bool, error) {
	n, err := s.rdb.Exists(ctx, lobbyKey(code)).Result()
	return n > 0, err
}

func (s *RedisStore) SaveSession(ctx context.Context, token string, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, sessionKey(token), data, 1*time.Hour).Err()
}

func (s *RedisStore) GetSession(ctx context.Context, token string) (*Session, error) {
	data, err := s.rdb.Get(ctx, sessionKey(token)).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, err
	}
	var sess Session
	return &sess, json.Unmarshal(data, &sess)
}

func lobbyKey(code string) string  { return "buzzhub:lobby:" + code }
func sessionKey(token string) string { return "buzzhub:player:" + token }
