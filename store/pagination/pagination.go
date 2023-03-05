package pagination

import "pckilgore/app/store"

// Pagination is a basic implementation of store.Parameterized.
type Pagination struct {
	limit  int
	before *store.Cursor
	after  *store.Cursor
}

// Options are parameters to construct [Params].
type Params struct {
	Limit  int
	Before *store.Cursor
	After  *store.Cursor
}

func New(p Params) Pagination {
	limit := 100
	if p.Limit > 0 && p.Limit < 100 {
		limit = p.Limit
	}

	return Pagination{limit: limit, before: p.Before, after: p.After}
}

func (p Pagination) Limit() int {
	return p.limit
}

func (p Pagination) After() *store.Cursor {
	return p.after
}

func (p Pagination) Before() *store.Cursor {
	return p.before
}
