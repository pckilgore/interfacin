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

type ListResponse[Model any] struct {
	Items []Model
	Count int
}

type Parameterized interface {
	Limit() int
}

// Pagination is a simple implementation of Parameterized.
type Pagination struct {
	limit int
}

func NewPagination(limit int) Pagination {
	return Pagination{limit: limit}
}

func (p Pagination) Limit() int {
	return p.limit
}

func DeserializeAll[Out any, In Deserializer[Out]](list []In) []Out {
	o := make([]Out, len(list))

	for _, m := range list {
		o = append(o, m.Deserialize())
	}

	return o
} 

type Serializer interface {
	Serialize(any) any
}

type Deserializer[Out any] interface {
	Deserialize() Out
}

//type Operation int
//const (
	//eq = iota
//)

//type Operator interface {
	//Operate(o Operation) 
//}

//type Param struct {
	//Field string
	//Operation Operation
	//Value comparable
//}

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
