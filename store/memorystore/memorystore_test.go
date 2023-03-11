package memorystore_test

import (
	"fmt"
	"math/rand"
	"testing"

	"pckilgore/app/pointers"
	"pckilgore/app/store"
	storetest "pckilgore/app/store/test"
	"pckilgore/app/store/memorystore"
	"pckilgore/app/store/pagination"
	"pckilgore/app/widget"

	"github.com/stretchr/testify/require"
)

func TestMemorystore(t *testing.T) {
	t.Parallel()

	widgetStore := memorystore.New[widget.DatabaseWidget, widget.WidgetParams]()

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
