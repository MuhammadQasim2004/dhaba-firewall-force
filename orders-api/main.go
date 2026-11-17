package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	gitCommit = envOr("GIT_COMMIT", "unknown")
	podName   = envOr("POD_NAME", "local")
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type Order struct {
	ID     int64  `json:"id"`
	Dish   string `json:"dish"`
	Status string `json:"status"`
}

type orderRequest struct {
	Dish string `json:"dish"`
}

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe /health and exit 0/1 (used by Docker HEALTHCHECK)")
	flag.Parse()
	if *healthcheck {
		resp, err := http.Get("http://127.0.0.1:8080/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	store := NewStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{
			"service": "orders-api",
			"version": gitCommit,
			"pod":     podName,
		})
	})

	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req orderRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Dish == "" {
				http.Error(w, "dish is required", http.StatusBadRequest)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			order, err := store.CreateOrder(ctx, req.Dish)
			if err != nil {
				log.Printf("create order failed: %v", err)
				http.Error(w, "could not save order", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			writeJSON(w, order)
		case http.MethodGet:
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			orders, err := store.ListOrders(ctx)
			if err != nil {
				log.Printf("list orders failed: %v", err)
				http.Error(w, "could not list orders", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]any{"orders": orders})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	log.Printf("orders-api %s starting on :8080 (pod=%s)", gitCommit, podName)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
