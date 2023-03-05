package store

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"encoding/base64"
)

const prefix = "cursor_"

// Cursor is an opaque identifier for a record in a data store. Determining what
// the cursor does or how to decode it is left to the implementing store. It
// serializes into a URL-safe token.
type Cursor struct {
	value []byte
}

// NewCursor instantiates [Cursor] with a known good value.
func NewCursor(value string) Cursor {
	return Cursor{value: []byte(value)}
}

// Parse converts a valid serialized Cursor token back to a Cursor, else errors.
func Parse(maybeCursorToken string) (*Cursor, error) {
	sansPrefix, found := strings.CutPrefix(maybeCursorToken, prefix)
	if !found {
		return nil, errors.New("not a cursor")
	}

	enc, err := base64.URLEncoding.DecodeString(sansPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "not a cursor")
	}
	
	cursor := NewCursor(string(enc))

	return &cursor, nil
}

// Value returns the value of the [Cursor].
func (c Cursor) Value() string {
	return string(c.value)
}

// Token converts a Cursor into a URL-Safe string token.
func (c Cursor) Token() string {
	return prefix + base64.URLEncoding.EncodeToString(c.value)
}

func (c Cursor) String() string {
	return fmt.Sprintf(`[Cursor: %s]`, string(c.value))
}

func (c Cursor) GoString() string {
	return fmt.Sprintf(`"%s => %s"`, c.String(), c.Token())
}

