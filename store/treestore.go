package store

import (
	"context"
	"sort"
)

type Treeable interface {
	// GetParentID gets the ID of the parent node in the databases preferred
	// format.
	GetParentID() *string

	// Return the name of the field that stores the id of the parent node.
	GetParentIDField() string
}

// TreeStorable is a [Storable] that models a tree of [Storable]s connected via
// unique parent identifiers.
type TreeStorable interface {
	Storable
	Treeable
}

type AncestorLister[Model TreeStorable] interface {
	ListAncestors(ctx context.Context, id string) (TreeResponse[Model], error)
}

type DescendantLister[Model TreeStorable] interface {
	ListDescendants(ctx context.Context, id string) (TreeResponse[Model], error)
}

// TreeStore is an advanced store implementation that's capable of querying
// ancestor/descentant relationships between [TreeStorable] nodes.
type TreeStore[Model TreeStorable, Params Parameterized] interface {
	Store[Model, Params]
	AncestorLister[Model]
	DescendantLister[Model]
}

// A Layer is a set of nodes at the relative path length from the root of the
// request.
type Layer[Model any] struct {
	PathLength int
	Items      []Model
}

type TreeResponse[Model any] struct {
	// Layers are nodes of the result tree returned with their relative distance
	// from the supplied root.
	//
	// Thus, len(layers) is the depth of the tree, NOT the number of nodes.
	Layers []Layer[Model]

	// Count is the total number of items available across all layers
	Count int
}

// Flat returns the complete list of nodes in [TreeResponse] in path-length
// order, so, e.g.: len(t.Flat()) == t.Count
func (t TreeResponse[Model]) Flat() []Model {
	sortedLayers := t.Layers
	sort.Slice(sortedLayers, func(i, j int) bool {
		return sortedLayers[i].PathLength < sortedLayers[j].PathLength
	})

	var result []Model
	for _, layer := range sortedLayers {
		result = append(result, layer.Items...)
	}

	return result
}
