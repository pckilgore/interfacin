// Interfaces for writing to database.
package store

import (
	"context"
)

type Tabler interface {
	TableName() string
}

type Storable interface {
	Tabler
	GetID() string
	NewID() string
}

type Retriever[Model Storable] interface {
	Retrieve(ctx context.Context, id string) (*Model, bool, error)
}

type Deleter[Model Storable] interface {
	Delete(ctx context.Context, id string) (bool, error)
}

type Creator[Model Storable] interface {
	Create(ctx context.Context, m Model) (*Model, error)
}

type Lister[Model Storable, Params Parameterized] interface {
	List(ctx context.Context, p Params) (ListResponse[Model], error)
}

type Store[Model Storable, Params Parameterized] interface {
	Retriever[Model]
	Creator[Model]
	Deleter[Model]
	Lister[Model, Params]
}
