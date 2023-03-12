// Memorystore is a toy implementation of store using a in-memory map.
package memorystore

import (
	"context"
	"sync"

	"pckilgore/app/store"
)

type MemoryParams[D store.Storable] interface {
	store.Parameterized

	// MemoryFilter applies parameters to data.
	MemoryFilter(pre []D) (post []D)
}

type memorystore[S store.Storable, P MemoryParams[S]] struct {
	mu   sync.RWMutex
	data map[string]S
}

func New[S store.Storable, P MemoryParams[S]]() *memorystore[S, P] {
	return &memorystore[S, P]{data: make(map[string]S)}
}

// Store a model.
func (s *memorystore[D, P]) Create(_ context.Context, m D) (*D, error) {
	return createImpl(&s.mu, s.data, m)
}

// Retrieve a model.
func (s *memorystore[D, P]) Retrieve(_ context.Context, id string) (*D, bool, error) {
	return retrieveImpl(&s.mu, s.data, id)
}

// Delete a model.
func (s *memorystore[D, P]) Delete(_ context.Context, id string) (bool, error) {
	return deleteImpl(&s.mu, s.data, id)
}

// List a model.
func (s *memorystore[D, P]) List(_ context.Context, params P) (store.ListResponse[D], error) {
	return listImpl(&s.mu, s.data, params)
}
