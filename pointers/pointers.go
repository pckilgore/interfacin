package pointers

// Reference anything!
func Make[T any](t T) *T {
  return &t
}

// Safely dereference.
func GetWithDefault[T any](t *T, d T) T {
  if t == nil {
    return d
  }

  return *t
}
