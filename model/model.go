package model

type ID[T any] string

func CastID[T any](s string) ID[T] {
  return ID[T](s)
}

type Readier interface {
  Ready() bool
}

type Templater[T any] interface {
  Template() T
}

// CoreObject describes domain objects in the system.
type CoreObject[Templ any, Obj any] interface {
  Readier
  Templater[Templ]
  SetID(string)
  GetID() ID[Obj]
  Name() string
}

