package main

import (
	"context"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pckilgore/app/pointers"
	"pckilgore/app/store/gormstore"
	"pckilgore/app/widget"
)

func main() {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&widget.DatabaseWidget{})

	widgetStore := gormstore.New[widget.DatabaseWidget](db)
	widgetService := widget.NewService(widgetStore)

	w, err := widgetService.Create(
		ctx,
		widget.WidgetTemplate{Name: pointers.Make("My Widget")},
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("gorm: %#v\n", w)
}
