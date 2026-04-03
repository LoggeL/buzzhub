package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/logge/buzzhub/internal/engine"
	"github.com/logge/buzzhub/internal/lobby"
	redisclient "github.com/logge/buzzhub/internal/redis"
	sockethandler "github.com/logge/buzzhub/internal/socket"

	// Register all games
	_ "github.com/logge/buzzhub/internal/games/bluff"
	_ "github.com/logge/buzzhub/internal/games/drawing"
	_ "github.com/logge/buzzhub/internal/games/quiz"
	_ "github.com/logge/buzzhub/internal/games/voting"

	"github.com/zishang520/socket.io/v2/socket"
)

func main() {
	redisAddr := getEnv("REDIS_URL", "localhost:6379")
	port := getEnv("PORT", "3000")
	staticDir := getEnv("STATIC_DIR", "../frontend/build")

	rdb := redisclient.New(redisAddr)
	store := lobby.NewRedisStore(rdb)
	mgr := lobby.NewManager(store)
	eng := engine.New(mgr)

	io := socket.NewServer(nil, nil)
	handler := sockethandler.New(io, mgr, eng)
	handler.Setup()

	mux := http.NewServeMux()
	mux.Handle("/socket.io/", io.ServeHandler(nil))
	mux.Handle("/", spaHandler(staticDir))

	log.Printf("BuzzHub starting on :%s", port)
	go func() {
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Fatal(err)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit

	log.Println("Shutting down...")
	io.Close(nil)
}

func spaHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := dir + r.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, dir+"/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
