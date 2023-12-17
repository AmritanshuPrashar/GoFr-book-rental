// model/book.go
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title          string             `json:"title,omitempty"`
	Author         string             `json:"author,omitempty"`
	AvailableCount int                `json:"availableCount" bson:"availableCount"`
}

type Rental struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	BookID string             `json:"bookID,omitempty" bson:"bookID,omitempty"`
}
