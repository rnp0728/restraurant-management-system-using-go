package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type OrderItem struct {
	ID          primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Quantity    *string            `json:"quantity,omitempty" validate:"required,eq=S|eq=M|eq=L"`
	UnitPrice   *float64           `json:"unit_price,omitempty"`
	FoodID      *string            `json:"food_id,omitempty"`
	OrderItemID string             `json:"order_item_id,omitempty"`
	OrderID     string             `json:"order_id,omitempty"`
	CreatedAt   time.Time          `json:"created_at,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at,omitempty"`
}
