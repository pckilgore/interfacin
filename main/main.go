package main

import (
  "fmt"
  "context"
	"gorm.io/gorm"
  "gorm.io/driver/sqlite"
  "pckilgore/app/widget"
	"pckilgore/app/pointers"
)

func main() {
  ctx := context.Background()
  db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
  if err != nil {
    panic("failed to connect database")
  }

  // Migrate the schema
  db.AutoMigrate(&widget.DatabaseWidget{})

  widgetService := widget.NewService(db)

  w, err := widgetService.Create(
    ctx, 
    widget.WidgetTemplate{Name: pointers.Make("My Widget")},
  )

  if err != nil {
    panic(err)
  }

  fmt.Printf("--->#%v", w)
}
