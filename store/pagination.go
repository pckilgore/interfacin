package store

import "github.com/pkg/errors"

type ListResponse[Model any] struct {
	// Items includes models matching the query parameters up to the limit,
	// wherein it represents a single page of responses matching the query.
	Items  []Model

	// Count is the total number of items available, irrespective of any limit or
	// pagination.
	Count  int

	// After can be provided by parameters to retreive the page of results
	// following the page of [Items].
	After  *Cursor

	// Before can be provided by parameters to retreive the page of results
	// preceding the page of [Items].
	Before *Cursor
}

type Parameterized interface {
	Limit() int
	After() *Cursor
	Before() *Cursor
}

type InvalidPaginationParamsErr error

func NewInvalidPaginationParamsErr(msg string) InvalidPaginationParamsErr {
	return errors.New(msg)
}
