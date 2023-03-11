package memorystore_test

import (
	"fmt"
	"math/rand"
	"testing"

	"pckilgore/app/node"
	"pckilgore/app/pointers"
	"pckilgore/app/store"
	"pckilgore/app/store/memorystore"
	"pckilgore/app/store/pagination"
	storetest "pckilgore/app/store/test"

	"github.com/stretchr/testify/require"
)

func TestMemorystore(t *testing.T) {
	t.Parallel()

	nodeStore := memorystore.New[node.DatabaseNode, node.NodeParams]()

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
			require.Contains(t, model.Name, "testing node")
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
