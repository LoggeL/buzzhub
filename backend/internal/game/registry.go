package game

import "fmt"

var registry = map[string]Factory{}

func Register(id string, factory Factory) {
	if _, exists := registry[id]; exists {
		panic(fmt.Sprintf("game %q already registered", id))
	}
	registry[id] = factory
}

func Create(id string) (Game, error) {
	factory, ok := registry[id]
	if !ok {
		return nil, fmt.Errorf("unknown game: %s", id)
	}
	return factory(), nil
}

func List() []GameInfo {
	var games []GameInfo
	for _, factory := range registry {
		g := factory()
		games = append(games, g.Info())
	}
	return games
}
