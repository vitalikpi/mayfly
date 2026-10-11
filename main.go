package main

import (
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

//go:embed index.html
var indexHTML []byte

//go:embed read.html
var readHTML []byte

//go:embed app.js
var appJS []byte

//go:embed style.css
var styleCSS []byte

type entry struct {
	blob      []byte
	expiresAt time.Time
}

type store struct {
	mu         sync.Mutex
	secrets    map[string]entry
	ttl        time.Duration
	maxEntries int
}

func newStore(ttl time.Duration, maxEntries int) *store {
	return &store{
		secrets:    make(map[string]entry),
		ttl:        ttl,
		maxEntries: maxEntries,
	}
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (s *store) put(blob []byte) (id string, ok bool) {
	id = newID()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.secrets) >= s.maxEntries {
		return "", false
	}
	s.secrets[id] = entry{blob: blob, expiresAt: time.Now().Add(s.ttl)}
	return id, true
}

func (s *store) take(id string) (blob []byte, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.secrets[id]
	if !ok {
		return nil, false
	}
	delete(s.secrets, id)
	if time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.blob, true
}

func (s *store) sweep() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, e := range s.secrets {
		if now.After(e.expiresAt) {
			delete(s.secrets, id)
		}
	}
}

func main() {
	mux := http.NewServeMux()

	maxSize := int64(500 * 1024)
	if v := os.Getenv("MAX_SIZE"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			log.Fatalf("invalid MAX_SIZE %q: want a positive integer (bytes)", v)
		}
		maxSize = n
	}

	ttl := 24 * time.Hour
	if v := os.Getenv("TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			log.Fatalf("invalid TTL %q: want a positive duration like 24h or 30m", v)
		}
		ttl = d
	}

	maxEntries := 10000
	if v := os.Getenv("MAX_ENTRIES"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			log.Fatalf("invalid MAX_ENTRIES %q: want a positive integer", v)
		}
		maxEntries = n
	}

	store := newStore(ttl, maxEntries)
	log.Printf("max secret size: %d bytes, ttl: %s, max entries: %d", maxSize, ttl, maxEntries)

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			store.sweep()
		}
	}()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	mux.HandleFunc("GET /s/{secretid}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store") // never cache a page that reveals a secret
		w.Write(readHTML)
	})

	mux.HandleFunc("GET /app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Write(appJS)
	})

	mux.HandleFunc("GET /style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write(styleCSS)
	})

	mux.HandleFunc("POST /api/secrets", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		blob, err := io.ReadAll(r.Body)
		if err != nil {
			var tooBig *http.MaxBytesError
			if errors.As(err, &tooBig) {
				http.Error(w, "secret too large", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if len(blob) == 0 {
			http.Error(w, "empty secret", http.StatusBadRequest)
			return
		}
		id, ok := store.put(blob)
		if !ok {
			http.Error(w, "service at capacity, try again later", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(id))
	})

	mux.HandleFunc("GET /api/secrets/{secretid}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("secretid")
		secret, ok := store.take(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff") // hey client! Trust me bro!
		w.Header().Set("Cache-Control", "no-store")         // the secret must never be cached
		w.Write(secret)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
