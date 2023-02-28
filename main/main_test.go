package main_test

import (
	"context"
	"pckilgore/app/pointers"
	"pckilgore/app/store/gormstore"
	"pckilgore/app/store/memorystore"
	"pckilgore/app/widget"
	"testing"

	"os"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func BenchmarkMemorySqliteStore(b *testing.B) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&widget.DatabaseWidget{})

	if err != nil {
		b.Fatalf("couldn't open db connection")
	}

	widgetStore := gormstore.New[widget.DatabaseWidget](db)
	widgetService := widget.NewService(widgetStore)

	for i := 0; i < b.N; i++ {
		m, err := widgetService.Create(
			ctx,
			widget.WidgetTemplate{Name: pointers.Make("My Widget")},
		)

		if err != nil {
			panic(err)
		}

		if m.Name != "My Widget" {
			b.Fatalf("write failed")
		}
	}
}

func BenchmarkFileSqliteStore(b *testing.B) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open("testdb.sqlite"), &gorm.Config{})
	db.AutoMigrate(&widget.DatabaseWidget{})

	if err != nil {
		b.Fatalf("couldn't open db connection")
	}

	widgetStore := gormstore.New[widget.DatabaseWidget](db)
	widgetService := widget.NewService(widgetStore)

	for i := 0; i < b.N; i++ {
		m, err := widgetService.Create(
			ctx,
			widget.WidgetTemplate{Name: pointers.Make("My Widget")},
		)

		if err != nil {
			panic(err)
		}

		if m.Name != "My Widget" {
			b.Fatalf("write failed")
		}
	}

  err = os.Remove("testdb.sqlite")
	if err != nil {
		b.Fatalf("cleanup database connection")
	}
}

func BenchmarkMemoryStore(b *testing.B) {
	ctx := context.Background()
	widgetStore := memorystore.New[widget.DatabaseWidget]()
	widgetService := widget.NewService(widgetStore)
	
	for i := 0; i < b.N; i++ {
		m, err := widgetService.Create(
			ctx,
			widget.WidgetTemplate{Name: pointers.Make("My Widget")},
		)

		if err != nil {
			panic(err)
		}

		if m.Name != "My Widget" {
			b.Fatalf("write failed")
		}
	}
}
