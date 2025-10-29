package notes

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title"         json:"title"`
	Content   string             `bson:"content"       json:"content"`
	CreatedAt time.Time          `bson:"createdAt"     json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"     json:"updatedAt"`
}
