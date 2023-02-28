// Memorystore is a toy implementation of store using a in-memory map.
package memorystore

import (
	"context"
	"github.com/pkg/errors"
	"pckilgore/app/store"
)

type memorystore[Storable store.Storable] map[string]Storable

func New[Storable store.Storable]() memorystore[Storable] {
	return make(memorystore[Storable])
}

// Store a model.
func (s memorystore[Storable]) Create(_ context.Context, m Storable) (*Storable, error) {
	if _, exists := s[m.GetID()]; exists {
		return nil, errors.New("a record with that ID already exists")
	}

	s[m.GetID()] = m

	return &m, nil
}

// Retrieve a model.
func (s memorystore[Storable]) Retrieve(_ context.Context, id string) (*Storable, bool, error) {
	if model, exists := s[id]; exists {
		return &model, true, nil
	}

	return nil, false, nil
}

// Delete a model.
func (s memorystore[DatabaseModel]) Delete(c context.Context, id string) (bool, error) {
	if _, exists := s[id]; exists {
		delete(s, id)
		return true, nil
	}

	return false, nil
}
