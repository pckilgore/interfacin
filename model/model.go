package model

import (
	"strings"
)

type Kinder interface {
	// Kind returns a human-readable "type" for the object as a string. It's
	// usually just the name of the struct that implements it.
	Kind() string
}

// Model ids are opaque, unique identifiers for a particular Model. Under the
// hood, it's a combination of a human-readable domain-driven namespace, and an
// opaque identifier, that is, or maps to, a database identifier (but nothings
// requires that).
type ID[T Kinder] string

var separator = "_"

// String implements Stringer for a constructed ID. It does not validate that
// the ID corresponding to the returned string representation is valid.
func (id ID[any]) String() string {
	return string(id)
}

func ParseID[Model Kinder](id ID[Model]) string {
	model := *new(Model)
	if id, ok := strings.CutPrefix(string(id), model.Kind()+separator); ok {
		return id
	}

	// TODO ????
	return ""
}

// NewID accepts a model and a specifier -- some string that uniquely identifies
// the model, and constructs a Model ID.
func NewID[Model Kinder](specifier string) ID[Model] {
	m := *new(Model)
	return ID[Model](strings.Join([]string{m.Kind(), separator, specifier}, ""))
}
