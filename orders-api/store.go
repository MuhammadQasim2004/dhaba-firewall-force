package main

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	pg       *pgxpool.Pool
	rdb      *redis.Client
	mu       sync.Mutex
	fallback []Order
	nextID   int64
}

func NewStore() *Store {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dbURL := envOr("DATABASE_URL", "postgresql://dhaba:dhaba@postgres:5432/dhaba")
	redisURL := envOr("REDIS_URL", "redis://redis:6379/0")

	pool, _ := pgxpool.New(ctx, dbURL)

	opts, _ := redis.ParseURL(redisURL)
	rdb := redis.NewClient(opts)

	return &Store{pg: pool, rdb: rdb, nextID: 1}
}

func (s *Store) CreateOrder(ctx context.Context, dish string) (Order, error) {
	var id int64
	if s.pg != nil {
		err := s.pg.QueryRow(ctx,
			"INSERT INTO orders (dish, status) VALUES ($1, 'new') RETURNING id", dish,
		).Scan(&id)
		if err == nil {
			if s.rdb != nil {
				s.rdb.LPush(ctx, "orders:queue", id)
			}
			return Order{ID: id, Dish: dish, Status: "new"}, nil
		}
	}
	// fallback in-memory (local/dev only)
	s.mu.Lock()
	defer s.mu.Unlock()
	id = atomic.AddInt64(&s.nextID, 1)
	order := Order{ID: id, Dish: dish, Status: "new"}
	s.fallback = append(s.fallback, order)
	return order, nil
}

func (s *Store) ListOrders(ctx context.Context) ([]Order, error) {
	if s.pg != nil {
		rows, err := s.pg.Query(ctx, "SELECT id, dish, status FROM orders ORDER BY id DESC LIMIT 50")
		if err == nil {
			defer rows.Close()
			var out []Order
			for rows.Next() {
				var o Order
				if err := rows.Scan(&o.ID, &o.Dish, &o.Status); err == nil {
					out = append(out, o)
				}
			}
			return out, nil
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fallback, nil
}

var _ = os.Getenv
