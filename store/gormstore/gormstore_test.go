package gormstore_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pckilgore/app/store"
	"pckilgore/app/store/gormstore"
	"pckilgore/app/store/pagination"
	"pckilgore/app/widget"
)

func TestGormstore(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&widget.DatabaseWidget{})

	widgetStore := gormstore.New[widget.DatabaseWidget, widget.WidgetParams](db)
	store.CreateStoreTest[widget.DatabaseWidget, widget.WidgetParams](
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
	)
}
