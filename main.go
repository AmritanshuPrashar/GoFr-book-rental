// main.go
package main

import (
	"context"
	"test-server/controller"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gofr.dev/pkg/gofr"
)

func main() {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.Background())

	bookController := controller.NewBookController(client)

	app := gofr.New()
	app.POST("/books", bookController.AddBook)
	app.GET("/books", bookController.GetBooks)
	app.GET("/books/{id}", bookController.GetBookByID)
	app.PUT("/books/{id}", bookController.UpdateBook)    
	app.DELETE("/books/{id}", bookController.DeleteBook)
	app.POST("/books/rent/{id}", bookController.RentBook)
	app.POST("/books/return/{id}", bookController.ReturnBook)
	app.GET("/rentals", bookController.GetRentals)

	app.Start()
}
