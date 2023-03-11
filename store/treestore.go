package store

import (
	"context"
)

type Treeable interface {
	GetParentID() *string
}

// TreeStorable is a [Storable] that models a tree of [Storable]s connected via
// unique parent identifiers.
type TreeStorable interface {
	Storable
	Treeable
}

type AncestorLister[Model TreeStorable] interface {
	ListAncestors(ctx context.Context, id string) (ListResponse[Model], error)
}

type DescendantLister[Model TreeStorable] interface {
	ListDescendants(ctx context.Context, id string) (ListResponse[Model], error)
}

// TreeStore is an advanced store implementation that's capable of querying
// ancestor/descentant relationships between [TreeStorable] nodes.
type TreeStore[Model TreeStorable, Params Parameterized] interface {
	Store[Model, Params]
	AncestorLister[Model]
	DescendantLister[Model]
}
