package memorystore

import (
	"context"

	"github.com/pkg/errors"
	"pckilgore/app/store"
)

type Creator[D store.Storable] struct {
	d *data[D]
}

func NewCreator[D store.Storable](d *data[D]) *Creator[D] {
	return &Creator[D]{d: d}
}

func (c *Creator[D]) Create(_ context.Context, storable D) (*D, error) {
	c.d.mu.Lock()
	defer c.d.mu.Unlock()

	if _, exists := c.d.store[storable.GetID()]; exists {
		return nil, errors.New("a record with that ID already exists")
	}

	c.d.store[storable.GetID()] = storable

	return &storable, nil
}
