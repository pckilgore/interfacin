package main_test

import (
	"context"
	"pckilgore/app/pointers"
	"pckilgore/app/store/gormstore"
	"pckilgore/app/store/memorystore"
	"pckilgore/app/widget"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

func BenchmarkMemorySqliteStore(b *testing.B) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&widget.DatabaseWidget{})

	if err != nil {
		b.Fatalf("couldn't open db connection")
	}

	widgetStore := gormstore.New[widget.DatabaseWidget, widget.WidgetParams](db)

	widgetService := widget.NewService(widgetStore)
	b.ResetTimer()

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

	widgetStore := gormstore.New[widget.DatabaseWidget, widget.WidgetParams](db)
	if err != nil {
		b.Fatalf("couldn't initialize store")
	}
	widgetService := widget.NewService(widgetStore)
	b.ResetTimer()

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

	b.StopTimer()
	err = os.Remove("testdb.sqlite")
	if err != nil {
		b.Fatalf("cleanup database connection")
	}
	b.StartTimer()
}

func BenchmarkMemoryStore(b *testing.B) {
	ctx := context.Background()
	widgetStore := memorystore.New[widget.DatabaseWidget, widget.WidgetParams]()
	widgetService := widget.NewService(widgetStore)
	b.ResetTimer()

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

var result string

func BenchmarkSmallConcat(b *testing.B) {
	prefix := "prefix_"

	var r string
	for i := 0; i < b.N; i++ {
		r = prefix + "sometinyishstring"
	}

	result = r
}

func BenchmarkSmallBuild(b *testing.B) {
	prefix := "prefix_"

	var r string
	for i := 0; i < b.N; i++ {
		var res strings.Builder
		res.WriteString(prefix)
		res.WriteString("sometinyishstring")
		r = res.String()
	}

	result = r
}
