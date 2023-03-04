// Memorystore is a toy implementation of store using a in-memory map.
package memorystore

import (
	"context"
	"pckilgore/app/store"
	"sort"

	"github.com/pkg/errors"
)

type memorystore[Storable store.Storable, Params store.Parameterized] map[string]Storable

func New[S store.Storable, P store.Parameterized]() memorystore[S, P] {
	return make(memorystore[S, P])
}

// Store a model.
func (s memorystore[D, P]) Create(_ context.Context, m D) (*D, error) {
	if _, exists := s[m.GetID()]; exists {
		return nil, errors.New("a record with that ID already exists")
	}

	s[m.GetID()] = m

	return &m, nil
}

// Retrieve a model.
func (s memorystore[D, P]) Retrieve(_ context.Context, id string) (*D, bool, error) {
	if model, exists := s[id]; exists {
		return &model, true, nil
	}

	return nil, false, nil
}

// Delete a model.
func (s memorystore[D, P]) Delete(c context.Context, id string) (bool, error) {
	if _, exists := s[id]; exists {
		delete(s, id)
		return true, nil
	}

	return false, nil
}

// List a model.
func (s memorystore[D, P]) List(c context.Context, params P) (store.ListResponse[D], error) {
	var result []D
	for _, m := range s {
		result = append(result, m)
	}

	sort.SliceStable(result, func(i, j int) bool { 
		return result[i].GetID() < result[j].GetID()
	})

	return store.ListResponse[D]{
		Items: result[:params.Limit()],
		Count: len(result),
	}, nil
}
