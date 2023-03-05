package widget

import (
	"context"
	"pckilgore/app/pointers"
	"pckilgore/app/store"

	"github.com/pkg/errors"
)

type WidgetStore = store.Store[DatabaseWidget, WidgetParams]

type Service struct {
	store WidgetStore
}

func NewService(store WidgetStore) Service {
	return Service{store: store}
}

func (s Service) List(c context.Context, p WidgetParams) (*store.ListResponse[widget], error) {
	dbw, err := s.store.List(c, p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list widget")
	}

	var items []widget
	for _, i := range dbw.Items {
		item, err := Deserialize(&i)
		if err != nil {
			return nil, errors.Wrap(err, "failed to deserialize widget")
		}
		items = append(items, *item)
	}

	return &store.ListResponse[widget]{
		Count:  dbw.Count,
		Items:  items,
		After:  dbw.After,
		Before: dbw.Before,
	}, nil
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
