package gormstore_test

import (
	"fmt"
	"math/rand"
	"testing"

	"pckilgore/app/node"
	"pckilgore/app/pointers"
	"pckilgore/app/store"
	"pckilgore/app/store/gormstore"
	"pckilgore/app/store/pagination"
	storetest "pckilgore/app/store/test"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormstore(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.Nil(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&node.DatabaseNode{})
	require.Nil(t, err)

	nodeStore, err := gormstore.New[node.DatabaseNode, node.NodeParams](db)
	require.Nil(t, err)

	storetest.CreateTreeStoreTest[node.DatabaseNode, node.NodeParams](
		t,
		nodeStore,
		func(nonce int, parentId *string) node.DatabaseNode {
			return node.DatabaseNode{
				ID:       fmt.Sprintf("%03d", nonce),
				Name:     fmt.Sprintf("testing node %d", nonce),
				ParentID: parentId,
			}
		},
		func(t *testing.T, model node.DatabaseNode) {
			require.NotNil(t, model.ID)
			require.Contains(t, model.Name, "")
		},
		func(limit int, after *store.Cursor, before *store.Cursor) node.NodeParams {
			return node.NodeParams{
				Pagination: pagination.New(pagination.Params{Limit: limit, After: after, Before: before}),
			}
		},
		func(d []node.DatabaseNode) node.NodeParams {
			var ids []node.ID
			for _, item := range d {
				w, err := node.Deserialize(&item)
				require.Nil(t, err)
				ids = append(ids, node.ID(w.ID))
			}
			rand.Shuffle(len(ids), func(i, j int) {
				ids[i], ids[j] = ids[j], ids[i]
			})

			ids = ids[:15]

			return node.NodeParams{
				IDs:        pointers.Make(ids),
				Pagination: pagination.New(pagination.Params{}),
			}
		},
		func(t *testing.T, params node.NodeParams, d []node.DatabaseNode) {
			require.NotNil(t, params.IDs)
			var gotIds []node.ID
			for _, i := range d {
				m, err := node.Deserialize(&i)
				require.Nil(t, err)
				gotIds = append(gotIds, node.ID(m.ID))
			}
			require.Equal(t, len(d), len(*params.IDs))
			require.Subset(t, gotIds, *params.IDs)
		},
	)
}

func TestHelpers(t *testing.T) {
	t.Parallel()
	var wmodels []node.DatabaseNode

	// The tests here assert against the sqlite dialect, but should hold
	// regardless of dialector.
	t.Run("ColumnInIds", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.Nil(t, err)

		t.Run("without nulls", func(t *testing.T) {
			t.Parallel()
			want := "SELECT * FROM `nodes` WHERE `id` IN (\"123\",\"456\")"
			got := db.ToSQL(func(db *gorm.DB) *gorm.DB {
				return gormstore.ColumnInIDs("id", &[]string{"123", "456"})(db).Find(&wmodels)
			})
			require.Equal(t, want, got)
		})

		t.Run("with nulls", func(t *testing.T) {
			t.Parallel()
			want := "SELECT * FROM `nodes` WHERE (`id` IS NULL OR `id` IN (\"123\",\"456\"))"
			got := db.ToSQL(func(db *gorm.DB) *gorm.DB {
				return gormstore.ColumnInIDs(
					"id",
					&[]string{"123", "456", gormstore.Null},
				)(db).Find(&wmodels)
			})
			require.Equal(t, want, got)
		})

		t.Run("empty ids", func(t *testing.T) {
			t.Parallel()
			want := "SELECT * FROM `nodes`"
			got := db.ToSQL(func(db *gorm.DB) *gorm.DB {
				return gormstore.ColumnInIDs(
					"id",
					&[]string{},
				)(db).Find(&wmodels)
			})
			require.Equal(t, want, got)
		})

		t.Run("nil ids", func(t *testing.T) {
			t.Parallel()
			want := "SELECT * FROM `nodes`"
			got := db.ToSQL(func(db *gorm.DB) *gorm.DB {
				return gormstore.ColumnInIDs[string](
					"id",
					nil,
				)(db).Find(&wmodels)
			})
			require.Equal(t, want, got)
		})
	})
}
