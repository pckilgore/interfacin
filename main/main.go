package main

import (
	"context"
	"fmt"
	"pckilgore/app/pointers"
	"pckilgore/app/store/pagination"
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

	_, err = widgetService.Create(
		ctx,
		widget.WidgetTemplate{Name: pointers.Make("My Widget")},
	)

	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		widgetService.Create(
			ctx,
			widget.WidgetTemplate{Name: pointers.Make(fmt.Sprintf("widget %d", i))},
		)
	}

	list, err := widgetService.List(
		ctx,
		widget.WidgetParams{Pagination: pagination.New(
			pagination.Params{Limit: 6},
		)},
	)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Response: %#v\n", list)

	after, err := widgetService.List(
		ctx,
		widget.WidgetParams{Pagination: pagination.New(
			pagination.Params{After: list.After},
		)},
	)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Response: %#v\n", after)
}
