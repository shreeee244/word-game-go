//go:build integration

// Package tests holds the integration tests. They run against real Postgres
// and Redis. In CircleCI those are service containers, so the host is
// localhost; in docker-compose the host is the service name. The URLs are
// taken from the environment, so both work.
package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// waitForApp polls /health until the app responds or the deadline expires.
func waitForApp(t *testing.T, url string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	var body map[string]any
	for time.Now().Before(deadline) {
		resp, err := http.Get(url + "/health")
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				if err := json.NewDecoder(resp.Body).Decode(&body); err == nil {
					return body
				}
			}
			resp.Body.Close()
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("app never became healthy at %s", url)
	return nil
}

func TestHealthReportsBothBackends(t *testing.T) {
	url := env("APP_URL", "http://localhost:8080")
	body := waitForApp(t, url)

	if body["postgres"] != true {
		t.Fatalf("postgres not reachable: %v", body)
	}
	if body["redis"] != true {
		t.Fatalf("redis not reachable: %v", body)
	}
	if body["status"] != "ok" {
		t.Fatalf("status not ok: %v", body)
	}
}

func TestPostgresReachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, env("DATABASE_URL",
		"postgres://wordgame:wordgame@localhost:5432/wordgame?sslmode=disable"))
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	var n int
	if err := pool.QueryRow(ctx, "SELECT 1").Scan(&n); err != nil {
		t.Fatalf("SELECT 1: %v", err)
	}
	if n != 1 {
		t.Fatalf("got %d, want 1", n)
	}
}

func TestRedisReachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opt, err := redis.ParseURL(env("REDIS_URL", "redis://localhost:6379/0"))
	if err != nil {
		t.Fatalf("redis.ParseURL: %v", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	if err := rdb.Set(ctx, "it:ping", "pong", 10*time.Second).Err(); err != nil {
		t.Fatalf("SET: %v", err)
	}
	got, err := rdb.Get(ctx, "it:ping").Result()
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	if got != "pong" {
		t.Fatalf("got %q, want pong", got)
	}
}

// TestGameEndpoint exercises /game for a fresh player; it must return an
// empty board without error even before any guesses.
func TestGameEndpoint(t *testing.T) {
	url := env("APP_URL", "http://localhost:8080")
	waitForApp(t, url)

	resp, err := http.Get(url + "/game?player=integration-test")
	if err != nil {
		t.Fatalf("GET /game: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/game status = %d, want 200", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["player"] != "integration-test" {
		t.Fatalf("player = %v, want integration-test", body["player"])
	}
	if body["done"] != false {
		t.Fatalf("done = %v, want false", body["done"])
	}
}
