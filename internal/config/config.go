package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	APIKey      string
	RESTBase    string
	WSURL       string
	RESTPerMin  int
	DatabaseURL string
	Addr        string
	Dev         bool
}

func Load() (Config, error) {
	loadDotEnv(".env")

	key := os.Getenv("MASSIVE_API_KEY")
	if key == "" {
		return Config{}, errors.New("MASSIVE_API_KEY is required")
	}

	ws := "wss://delayed.massive.com/stocks"
	if os.Getenv("MASSIVE_FEED") == "realtime" {
		ws = "wss://socket.massive.com/stocks"
	}

	rpm, err := strconv.Atoi(env("MASSIVE_REST_PER_MIN", "5"))
	if err != nil || rpm < 1 {
		return Config{}, errors.New("MASSIVE_REST_PER_MIN must be a positive integer")
	}

	return Config{
		APIKey:      key,
		RESTBase:    "https://api.massive.com",
		WSURL:       ws,
		RESTPerMin:  rpm,
		DatabaseURL: env("DATABASE_URL", "postgres://momo:momo@localhost:5437/momo?sslmode=disable"),
		Addr:        env("ADDR", ":7778"),
		Dev:         os.Getenv("MOMO_DEV") == "1",
	}, nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}

	return def
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}
