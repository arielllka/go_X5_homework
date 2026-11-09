package main

import "sync"

type Order struct {
	ID       int    `json:"id"`
	Customer string `json:"customer"`
	Status   string `json:"status"`
}

// Хранилище уже готово. Блокировка защищает данные общих HTTP-запросов.
type Store struct {
	mu     sync.Mutex
	nextID int
	orders map[int]Order
}

func NewStore() *Store { return &Store{nextID: 1, orders: make(map[int]Order)} }
func (s *Store) Create(customer string) Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	order := Order{ID: s.nextID, Customer: customer, Status: "new"}
	s.orders[order.ID] = order
	s.nextID++
	return order
}
func (s *Store) ByID(id int) (Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	order, ok := s.orders[id]
	return order, ok
}
