// Interfaces for writing to database.
package store

import (
	"context"
)

type Tabler interface {
	TableName() string
}

type Storable[Model any] interface {
	Tabler
	GetID() string
	NewID() string
}

type Serializer[Model any, DatabaseModel any] interface {
	Serialize(m Model) DatabaseModel
}

type Deserializer[Model any, DatabaseModel any] interface {
	// Error because the database is not necessarily a trustworthy source.
	Deserialize(d DatabaseModel) (*Model, error)
}

type Serder[Model any, DatabaseModel any] interface {
	Serializer[Model, DatabaseModel]
	Deserializer[Model, DatabaseModel]
}

type Retriever[Model any] interface {
	Retrieve(ctx context.Context, id string) (*Model, bool, error)
}

type Deleter[Model any] interface {
	Delete(ctx context.Context, id string) (bool, error)
}

type Creator[Model any] interface {
	Create(ctx context.Context, m Model) (*Model, error)
}

type Store[Model any] interface {
	Retriever[Model]
	Creator[Model]
	Deleter[Model]
}
