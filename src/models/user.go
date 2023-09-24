package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
type User struct {
	ID           primitive.ObjectID `json:"id,omitempty"`
	FirstName    *string            `json:"first_name,omitempty"`
	LastName     *string            `json:"last_name,omitempty"`
	Password     *string            `json:"password,omitempty"`
	Email        *string            `json:"email,omitempty"`
	Avatar       *string            `json:"avatar,omitempty"`
	Phone        *string            `json:"phone,omitempty"`
	Token        *string            `json:"token,omitempty"`
	RefreshToken *string            `json:"refresh_token,omitempty"`
	UserID       string            `json:"user_id,omitempty"`
	CreatedAt    time.Time          `json:"created_at,omitempty"`
	UpdatedAt    time.Time          `json:"updated_at,omitempty"`
}
