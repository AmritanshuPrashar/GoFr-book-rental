// controller/book_controller.go
package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"test-server/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gofr.dev/pkg/gofr"
)

type BookController struct {
	client *mongo.Client
}

func NewBookController(client *mongo.Client) *BookController {
	return &BookController{client: client}
}

func (bc *BookController) AddBook(ctx *gofr.Context) (interface{}, error) {
	collection := bc.client.Database("BookRental").Collection("books")

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return nil, err
	}

	var newBook model.Book
	if err := json.Unmarshal(body, &newBook); err != nil {
		return nil, err
	}

	existingBook := model.Book{}
	err = collection.FindOne(context.Background(), bson.M{"title": newBook.Title}).Decode(&existingBook)
	if err == nil {
		update := bson.M{"$inc": bson.M{"availableCount": newBook.AvailableCount}}

		_, err := collection.UpdateOne(context.Background(), bson.M{"title": newBook.Title}, update)
		if err != nil {
			return nil, err
		}

		existingBook.AvailableCount += newBook.AvailableCount

		fmt.Println("Updated existing book availability for title:", newBook.Title)
		return existingBook, nil
	}

	if newBook.AvailableCount == 0 {
		newBook.AvailableCount = 1
	}

	insertResult, err := collection.InsertOne(context.Background(), newBook)
	if err != nil {
		return nil, err
	}

	fmt.Println("Inserted book document ID:", insertResult.InsertedID)

	return newBook, nil
}

func (bc *BookController) GetBooks(ctx *gofr.Context) (interface{}, error) {
	collection := bc.client.Database("BookRental").Collection("books")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var books []model.Book
	if err := cursor.All(context.Background(), &books); err != nil {
		return nil, err
	}

	return books, nil
}

func ExtractBookID(ctx *gofr.Context) (primitive.ObjectID, error) {
	rawPath := ctx.Request().URL.Path
	fmt.Println("Raw Path:", rawPath)
	pathParts := strings.Split(rawPath, "/")
	if len(pathParts) > 2 {
		bookID := pathParts[2]
		fmt.Println("Book ID:", bookID)
		objectID, err := primitive.ObjectIDFromHex(bookID)
		if err != nil {
			return primitive.NilObjectID, fmt.Errorf("invalid book ID format")
		}

		return objectID, nil
	}

	return primitive.NilObjectID, fmt.Errorf("book ID not found in the path")
}

func (bc *BookController) GetBookByID(ctx *gofr.Context) (interface{}, error) {
	collection := bc.client.Database("BookRental").Collection("books")

	objID, err := ExtractBookID(ctx)
	if err != nil {
		return nil, err
	}

	var book model.Book
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&book)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("book not found")
		}
		return nil, err
	}

	return book, nil
}

func (bc *BookController) UpdateBook(ctx *gofr.Context) (interface{}, error) {
	collection := bc.client.Database("BookRental").Collection("books")

	objID, err := ExtractBookID(ctx)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return nil, err
	}

	var updateData map[string]interface{}
	if err := json.Unmarshal(body, &updateData); err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": updateData}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}

	if result.ModifiedCount == 0 {
		return nil, fmt.Errorf("no book found for update")
	}

	var updatedBook model.Book
	err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&updatedBook)
	if err != nil {
		return nil, err
	}

	return updatedBook, nil
}

func (bc *BookController) DeleteBook(ctx *gofr.Context) (interface{}, error) {
	collection := bc.client.Database("BookRental").Collection("books")
	rentalsCollection := bc.client.Database("BookRental").Collection("rentals")

	objID, err := ExtractBookID(ctx)
	if err != nil {
		return nil, err
	}

	rentalFilter := bson.M{"bookID": objID.Hex()}
	rental := model.Rental{}
	err = rentalsCollection.FindOne(context.Background(), rentalFilter).Decode(&rental)
	if err == nil {
		_, deleteErr := rentalsCollection.DeleteOne(context.Background(), rentalFilter)
		if deleteErr != nil {
			return nil, deleteErr
		}
	}

	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		return nil, err
	}

	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("no book found for deletion")
	}

	return map[string]string{"message": "Book deleted successfully"}, nil
}

func (bc *BookController) RentBook(ctx *gofr.Context) (interface{}, error) {
	collection := bc.client.Database("BookRental").Collection("books")
	rentalsCollection := bc.client.Database("BookRental").Collection("rentals")

	rawPath := ctx.Request().URL.Path
	fmt.Println("Raw Path:", rawPath)

	pathParts := strings.Split(rawPath, "/")
	if len(pathParts) > 2 {
		bookID := pathParts[3]
		fmt.Println("Book ID:", bookID)

		objID, err := primitive.ObjectIDFromHex(bookID)
		if err != nil {
			return nil, fmt.Errorf("invalid book ID format")
		}

		filter := bson.M{"_id": objID, "availableCount": bson.M{"$gt": 0}}
		update := bson.M{"$inc": bson.M{"availableCount": -1}}

		result, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return nil, err
		}

		if result.ModifiedCount == 0 {
			return nil, fmt.Errorf("no available books for rental")
		}

		rental := model.Rental{BookID: bookID}
		_, insertErr := rentalsCollection.InsertOne(context.Background(), rental)
		if insertErr != nil {
			return nil, insertErr
		}

		var updatedBook model.Book
		err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&updatedBook)
		if err != nil {
			return nil, err
		}

		return updatedBook, nil
	} else {
		return nil, fmt.Errorf("book ID not found in the path")
	}
}

func (bc *BookController) ReturnBook(ctx *gofr.Context) (interface{}, error) {
	collection := bc.client.Database("BookRental").Collection("books")
	rentalsCollection := bc.client.Database("BookRental").Collection("rentals")

	rawPath := ctx.Request().URL.Path
	fmt.Println("Raw Path:", rawPath)

	pathParts := strings.Split(rawPath, "/")
	if len(pathParts) > 2 {
		bookID := pathParts[3]
		fmt.Println("Book ID:", bookID)

		objID, err := primitive.ObjectIDFromHex(bookID)
		if err != nil {
			return nil, fmt.Errorf("invalid book ID format")
		}

		rentalFilter := bson.M{"bookID": bookID}
		rental := model.Rental{}
		err = rentalsCollection.FindOne(context.Background(), rentalFilter).Decode(&rental)
		if err != nil {
			return nil, fmt.Errorf("book is not rented yet")
		}

		filter := bson.M{"_id": objID}
		update := bson.M{"$inc": bson.M{"availableCount": 1}}

		result, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return nil, err
		}

		if result.ModifiedCount == 0 {
			return nil, fmt.Errorf("error updating available count")
		}

		_, deleteErr := rentalsCollection.DeleteOne(context.Background(), rentalFilter)
		if deleteErr != nil {
			return nil, deleteErr
		}

		var updatedBook model.Book
		err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&updatedBook)
		if err != nil {
			return nil, err
		}

		return updatedBook, nil
	} else {
		return nil, fmt.Errorf("book ID not found in the path")
	}
}

func (bc *BookController) GetRentals(ctx *gofr.Context) (interface{}, error) {
	collection := bc.client.Database("BookRental").Collection("rentals")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var rentals []model.Rental
	if err := cursor.All(context.Background(), &rentals); err != nil {
		return nil, err
	}

	return rentals, nil
}
