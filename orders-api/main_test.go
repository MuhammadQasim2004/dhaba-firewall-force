package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeStore is an in-memory OrderStore used only in tests, so these tests
// never touch a real Postgres or Redis connection.
type fakeStore struct {
	orders  []Order
	nextID  int64
	failNew bool // simulate a DB error on CreateOrder
	failAll bool // simulate a DB error on ListOrders
}

func (f *fakeStore) CreateOrder(ctx context.Context, dish string) (Order, error) {
	if f.failNew {
		return Order{}, errors.New("simulated db failure")
	}
	f.nextID++
	order := Order{ID: f.nextID, Dish: dish, Status: "new"}
	f.orders = append(f.orders, order)
	return order, nil
}

func (f *fakeStore) ListOrders(ctx context.Context) ([]Order, error) {
	if f.failAll {
		return nil, errors.New("simulated db failure")
	}
	return f.orders, nil
}

func TestHealthEndpoint(t *testing.T) {
	mux := newMux(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body)
	}
}

func TestInfoEndpoint(t *testing.T) {
	mux := newMux(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var body map[string]string
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["service"] != "orders-api" {
		t.Fatalf("expected service=orders-api, got %v", body)
	}
}

func TestCreateOrder_Success(t *testing.T) {
	store := &fakeStore{}
	mux := newMux(store)

	payload, _ := json.Marshal(orderRequest{Dish: "Chicken Karahi"})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var order Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if order.Dish != "Chicken Karahi" || order.Status != "new" {
		t.Fatalf("unexpected order: %+v", order)
	}
	if len(store.orders) != 1 {
		t.Fatalf("expected store to have 1 order, got %d", len(store.orders))
	}
}

func TestCreateOrder_MissingDish(t *testing.T) {
	mux := newMux(&fakeStore{})

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing dish, got %d", w.Code)
	}
}

func TestCreateOrder_StoreFailure(t *testing.T) {
	mux := newMux(&fakeStore{failNew: true})

	payload, _ := json.Marshal(orderRequest{Dish: "Daal Chawal"})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when store fails, got %d", w.Code)
	}
}

func TestListOrders(t *testing.T) {
	store := &fakeStore{orders: []Order{
		{ID: 1, Dish: "Seekh Kebab", Status: "ready"},
	}}
	mux := newMux(store)

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Orders []Order `json:"orders"`
	}
	_ = json.NewDecoder(w.Body).Decode(&body)
	if len(body.Orders) != 1 || body.Orders[0].Dish != "Seekh Kebab" {
		t.Fatalf("unexpected orders: %+v", body.Orders)
	}
}

func TestOrdersEndpoint_MethodNotAllowed(t *testing.T) {
	mux := newMux(&fakeStore{})
	req := httptest.NewRequest(http.MethodDelete, "/orders", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
