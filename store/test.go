package store

import (
	"context"
	"math/rand"
	"pckilgore/app/pointers"
	"testing"

	"github.com/stretchr/testify/require"
)

func CreateStoreTest[D Storable, P Parameterized](
	t *testing.T,
	s Store[D, P],
	// Build a model. for each call, nonce is guaranteed to be unique.
	modelBuilder func(nonce int) D,
	// Validate that no deserialization errors occured.
	isValid func(t *testing.T, model D),
	// Generate search parameters.
	parameterBuilder func(limit int, after *Cursor, before *Cursor) P,
) {
	ctx := context.Background()
	var ids []string

	next := (func() func() int {
		count := 0
		return func() int {
			count++
			return count
		}
	})()

	t.Run("Create", func(t *testing.T) {
		c, err := s.Create(ctx, modelBuilder(next()))
		require.Nil(t, err, "store.Create should not error")
		isValid(t, *c)
		ids = append(ids, (*c).GetID())
	})

	t.Run("Retrieve", func(t *testing.T) {
		for i := 0; i < 150; i++ {
			m, err := s.Create(ctx, modelBuilder(next()))
			require.Nil(t, err, "failed to create multiple models for testing")
			ids = append(ids, (*m).GetID())
		}

		randomId := ids[rand.Intn(len(ids))]
		r, found, err := s.Retrieve(ctx, randomId)
		require.Nil(t, err, "store.Retrieve should not error")
		require.Truef(t, found, "store.Retrieve should be able to find id=%s in the store", randomId)
		isValid(t, *r)
	})

	t.Run("parameterBuilder contract", func(t *testing.T) {
		t.Parallel()
		for limit := 1; limit < 100; limit++ {
			var before *Cursor
			var after *Cursor
			if limit%2 == 0 {
				before = pointers.Make(NewCursor(ids[rand.Intn(len(ids))]))
			} else if limit%5 == 0 {
				after = pointers.Make(NewCursor(ids[rand.Intn(len(ids))]))
			}

			params := parameterBuilder(limit, after, before)
			require.Equal(t, limit, params.Limit(), "parameterBuilder did not set limit")
			require.Equal(t, before, params.Before(), "parameterBuilder did not set Before")
			require.Equal(t, after, params.After(), "parameterBuilder did not set After")
		}
	})

	t.Run("List", func(t *testing.T) {
		t.Run("pagination", func(t *testing.T) {
			limit := 77
			params := parameterBuilder(limit, nil, nil)
			list, err := s.List(ctx, params)
			require.Nil(t, err, "store.Lister should not error")
			require.Equal(t, len(ids), list.Count)
			require.Equal(t, limit, len(list.Items))

			params = parameterBuilder(100, list.After, nil)
			list, err = s.List(ctx, params)
			require.Nil(t, err, "store.Lister should not error")
			require.Equal(t, len(ids), list.Count, "still expect same total count")
			require.Equal(t, len(ids)-limit, len(list.Items), "the next page should only contain this many items")

			// Get first page of ten.
			params = parameterBuilder(10, nil, nil)
			firstPage, err := s.List(ctx, params)
			require.Nil(t, err, "store.Lister should not error")

			// Get next page of ten.
			params = parameterBuilder(10, firstPage.After, nil)
			secondPage, err := s.List(ctx, params)
			require.Nil(t, err, "store.Lister should not error")

			// Get first page (again).
			params = parameterBuilder(10, nil, secondPage.Before)
			firstPageRedux, err := s.List(ctx, params)
			require.Nil(t, err, "store.Lister should not error")
			require.Equal(t, firstPage, firstPageRedux, "first page should be the same")
		})
	})
}
