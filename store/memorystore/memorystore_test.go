package memorystore_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"pckilgore/app/store"
	"pckilgore/app/store/memorystore"
	"pckilgore/app/store/pagination"
	"pckilgore/app/widget"
)

func TestMemorystore(t *testing.T) {
	t.Parallel()

	widgetStore := memorystore.New[widget.DatabaseWidget, widget.WidgetParams]()
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
