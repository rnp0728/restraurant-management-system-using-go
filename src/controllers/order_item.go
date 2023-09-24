package controllers

import (
	"context"
	"fmt"
	"infinity/rms/database"
	"infinity/rms/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderItemPack struct {
	TableID    *string
	OrderItems []models.OrderItem
}

// DB
var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		result, err := orderItemCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching order items",
			})
			return
		}

		var allOrderItems []bson.M

		if err = result.All(c, &allOrderItems); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allOrderItems)
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		order_item_id := ctx.Param("order_item_id")

		var orderItem models.OrderItem

		err := orderItemCollection.FindOne(curCtx, bson.M{"order_item_id": order_item_id}).Decode(&orderItem)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching",
			})
		}

		ctx.JSON(http.StatusOK, orderItem)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("order_id")

		allOrderItems, err := ItemsByOrder(orderId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unable to get all the order items",
			})
			return
		}
		ctx.JSON(http.StatusOK, allOrderItems)
	}
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {
	curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	matchStage := bson.D{
		{Key: "$match", Value: bson.D{{Key: "order_id", Value: id}}},
	}

	foodLookupStage := bson.D{
		{Key: "$lookup", Value: bson.D{{Key: "from", Value: "food"}, {Key: "localField", Value: "food_id"}, {Key: "foreignField", Value: "food_id"}, {Key: "as", Value: "food"}}},
	}

	foodUnwindStage := bson.D{
		{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$food"}, {Key: "preserveNullAndEmptyArrays", Value: true}}},
	}

	orderLookupStage := bson.D{
		{Key: "$lookup", Value: bson.D{{Key: "from", Value: "order"}, {Key: "localField", Value: "order_id"}, {Key: "foreignField", Value: "order_id"}, {Key: "as", Value: "order"}}},
	}

	orderUnwindStage := bson.D{
		{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$order"}, {Key: "preserveNullAndEmptyArrays", Value: true}}},
	}

	tableLookupStage := bson.D{
		{Key: "$lookup", Value: bson.D{{Key: "from", Value: "table"}, {Key: "localField", Value: "order.table_id"}, {Key: "foreignField", Value: "table_id"}, {Key: "as", Value: "order"}}},
	}

	tableUnwindStage := bson.D{
		{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$table"}, {Key: "preserveNullAndEmptyArrays", Value: true}}},
	}

	projectStage := bson.D{
		{
			Key: "$project", Value: bson.D{
				{Key: "id", Value: 0},
				{Key: "amount", Value: "$food.price"},
				{Key: "total_count", Value: 1},
				{Key: "food_name", Value: "$food.name"},
				{Key: "food_image", Value: "$food.food_image"},
				{Key: "table_number", Value: "$table.table_number"},
				{Key: "table_id", Value: "$table.table_id"},
				{Key: "order_id", Value: "$order.order_id"},
				{Key: "price", Value: "$food.price"},
				{Key: "quantity", Value: 1},
			},
		},
	}
	groupStage := bson.D{
		{
			Key: "$group", Value: bson.D{{
				Key: "_id", Value: bson.D{{Key: "order_id", Value: "$order_id"}, {Key: "table_id", Value: "$table_id"}, {Key: "table_number", Value: "$table_number"}, {Key: "payment_due_date", Value: bson.D{
					{Key: "$num", Value: "$amount"},
				}},{Key: "total_count", Value: bson.D{
					{Key: "$num", Value: 1},{Key: "order_items", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
				}},
				},
			}},
		},
	}

	secondProjectStage := bson.D{
		{
			Key: "$project", Value: bson.D{
				{
					Key: "id", Value: 0,
				},
				{
					Key: "payment_due", Value: 1,
				},
				{
					Key: "total_count", Value: 1,
				},
				{
					Key: "table_number", Value: "$_id.table_number",
				},
				{
					Key: "order_items", Value: 1,
				},
			},
		},
	}

	result , err := orderItemCollection.Aggregate(
		curCtx, mongo.Pipeline{
			matchStage,
			foodLookupStage,
			foodUnwindStage,
			orderLookupStage,
			orderUnwindStage,
			tableLookupStage,
			tableUnwindStage,
			projectStage,
			groupStage,
			secondProjectStage,
		},
	)

	if err != nil {
		panic(err)
	}

	if err = result.All(curCtx, &OrderItems); err != nil {
		panic(err)
	}

	return OrderItems, err
}

func CreateOrderItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItemPack OrderItemPack
		var order models.Order

		if err := ctx.BindJSON(&orderItemPack); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		order.OrderDate, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		orderItemsToBeInserted := []interface{}{}
		order.TableID = orderItemPack.TableID
		order_id := orderItemOrderCreator(order)

		for _, orderItem := range orderItemPack.OrderItems {
			orderItem.OrderID = order_id
			validationErr := validate.Struct(orderItem)
			if validationErr != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": validationErr.Error(),
				})
				return
			}
			orderItem.ID = primitive.NewObjectID()
			orderItem.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.OrderItemID = orderItem.ID.Hex()

			var num = toFixed(*orderItem.UnitPrice, 2)
			orderItem.UnitPrice = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}

		result, err := orderItemCollection.InsertMany(curCtx, orderItemsToBeInserted)
		if err != nil {
			msg := fmt.Sprintf("Failed to create order items")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItem models.OrderItem

		if err := ctx.BindJSON(&orderItem); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		var updatedObj primitive.D

		orderItemId := ctx.Param("order_item_id")

		if orderItem.UnitPrice != nil {
			updatedObj = append(updatedObj, bson.E{Key: "unit_price", Value: orderItem.UnitPrice})
		}
		if orderItem.Quantity != nil {
			updatedObj = append(updatedObj, bson.E{Key: "quantity", Value: orderItem.Quantity})
		}
		if orderItem.FoodID != nil {
			updatedObj = append(updatedObj, bson.E{Key: "food_id", Value: orderItem.FoodID})
		}
		orderItem.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		updatedObj = append(updatedObj, bson.E{Key: "updated_at", Value: orderItem.UpdatedAt})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		filter := bson.M{"order_item_id": orderItemId}
		result, err := menuCollection.UpdateOne(
			curCtx, filter, bson.D{
				{Key: "$set", Value: updatedObj},
			},
			&opt,
		)

		if err != nil {
			msg := "Order Item updation failed"
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
