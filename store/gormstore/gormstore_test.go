package gormstore_test

import (
	"fmt"
	"math/rand"
	"testing"

	"pckilgore/app/pointers"
	"pckilgore/app/store"
	"pckilgore/app/store/gormstore"
	"pckilgore/app/store/pagination"
	storetest "pckilgore/app/store/test"
	"pckilgore/app/widget"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormstore(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.Nil(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&widget.DatabaseWidget{})
	require.Nil(t, err)

	widgetStore := gormstore.New[widget.DatabaseWidget, widget.WidgetParams](db)
	storetest.CreateStoreTest[widget.DatabaseWidget, widget.WidgetParams](
		t,
		widgetStore,
		func(nonce int) widget.DatabaseWidget {
			return widget.DatabaseWidget{
				ID:   fmt.Sprintf("%03d", nonce),
				Name: fmt.Sprintf("testing widget %d", nonce),
			}
		},
		func(t *testing.T, model widget.DatabaseWidget) {
			require.NotNil(t, model.ID)
			require.Contains(t, model.Name, "")
		},
		func(limit int, after *store.Cursor, before *store.Cursor) widget.WidgetParams {
			return widget.WidgetParams{
				Pagination: pagination.New(pagination.Params{Limit: limit, After: after, Before: before}),
			}
		},
		func(d []widget.DatabaseWidget) widget.WidgetParams {
			var ids []widget.ID
			for _, item := range d {
				w, err := widget.Deserialize(&item)
				require.Nil(t, err)
				ids = append(ids, widget.ID(w.ID))
			}
			rand.Shuffle(len(ids), func(i, j int) {
				ids[i], ids[j] = ids[j], ids[i]
			})

			ids = ids[:15]

			return widget.WidgetParams{
				IDs:        pointers.Make(ids),
				Pagination: pagination.New(pagination.Params{}),
			}
		},
		func(t *testing.T, params widget.WidgetParams, d []widget.DatabaseWidget) {
			require.NotNil(t, params.IDs)
			var gotIds []widget.ID
			for _, i := range d {
				m, err := widget.Deserialize(&i)
				require.Nil(t, err)
				gotIds = append(gotIds, widget.ID(m.ID))
			}
			require.Equal(t, len(d), len(*params.IDs))
			require.Subset(t, gotIds, *params.IDs)
		},
	)
}

func TestHelpers(t *testing.T) {
	t.Parallel()
	var wmodels []widget.DatabaseWidget

	// The tests here assert against the sqlite dialect, but should hold
	// regardless of dialector.
	t.Run("ColumnInIds", func(t *testing.T) {
		t.Parallel()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.Nil(t, err)

		t.Run("without nulls", func(t *testing.T) {
			t.Parallel()
			want := "SELECT * FROM `widgets` WHERE `id` IN (\"123\",\"456\")"
			got := db.ToSQL(func(db *gorm.DB) *gorm.DB {
				return gormstore.ColumnInIDs("id", &[]string{"123", "456"})(db).Find(&wmodels)
			})
			require.Equal(t, want, got)
		})

		t.Run("with nulls", func(t *testing.T) {
			t.Parallel()
			want := "SELECT * FROM `widgets` WHERE (`id` IS NULL OR `id` IN (\"123\",\"456\"))"
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
			want := "SELECT * FROM `widgets`"
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
			want := "SELECT * FROM `widgets`"
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
