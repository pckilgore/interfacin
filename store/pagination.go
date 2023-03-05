package store

import 	"github.com/pkg/errors"

type ListResponse[Model any] struct {
	Items []Model
	Count int
	After *Cursor
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

