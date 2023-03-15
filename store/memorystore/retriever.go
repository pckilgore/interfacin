package memorystore

import (
	"context"

	"pckilgore/app/store"
)

type Retriever[D store.Storable] struct {
	d *data[D]
}

func NewRetriever[D store.Storable](d *data[D]) *Retriever[D] {
	return &Retriever[D]{d: d}
}

func (r *Retriever[D]) Retrieve(_ context.Context, id string) (*D, bool, error) {
	r.d.mu.RLock()
	defer r.d.mu.RUnlock()

	if model, exists := r.d.store[id]; exists {
		return &model, true, nil
	}

	return nil, false, nil
}
