package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	NumberOfGuests *int               `json:"number_of_guests,omitempty"`
	TableNumber    *int               `json:"table_number,omitempty"`
	TableID        string             `json:"table_id,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}
