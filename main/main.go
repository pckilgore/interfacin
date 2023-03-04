package main

import (
	"context"
	"fmt"
	"pckilgore/app/pointers"
	"pckilgore/app/store"
	"pckilgore/app/store/memorystore"
	"pckilgore/app/widget"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&widget.DatabaseWidget{})

	widgetStore := memorystore.New[widget.DatabaseWidget, widget.WidgetParams]()
	widgetService := widget.NewService(widgetStore)

	w, err := widgetService.Create(
		ctx,
		widget.WidgetTemplate{Name: pointers.Make("My Widget")},
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("mem: %#v\n", w)

	for i := 0; i < 10; i++ {
		widgetService.Create(
			ctx,
			widget.WidgetTemplate{Name: pointers.Make(fmt.Sprintf("widget %d", i))},
		)
	}

	list, _ := widgetService.List(
		ctx, 
		widget.WidgetParams{ Pagination: store.NewPagination(6)},
	)
	
	for i, li := range list.Items {
		fmt.Printf("Item %d: %#v\n", i, li)
	}
}
