package widget

import (
	"context"
	"github.com/pkg/errors"
	"pckilgore/app/pointers"
	"pckilgore/app/store"
)

type WidgetStore = store.Store[DatabaseWidget]

type Service struct {
	store store.Store[DatabaseWidget]
}

func NewService(store WidgetStore) Service {
	return Service{store: store}
}

func (s Service) Create(c context.Context, t WidgetTemplate) (*widget, error) {
	var w widget
	w.Name = pointers.GetWithDefault(t.Name, "Some Widget")
	w.ID = createID()

	// Persist
	res, err := s.store.Create(c, Serialize(w))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to save created widget")
	}

	return Deserialize(res)
}
