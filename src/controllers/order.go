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

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func GetOrders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		result, err := orderCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching orders",
			})
		}

		var allOrders []bson.M

		if err = result.All(c, &allOrders); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allOrders)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderId := ctx.Param("order_id")

		var order models.Order

		err := orderCollection.FindOne(curCtx, bson.M{"order_id": orderId}).Decode(&order)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching",
			})
		}

		ctx.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var table models.Table
		var order models.Order

		if err := ctx.BindJSON(&order); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		validationErr := validate.Struct(order)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Error(),
			})
			return
		}

		err := tableCollection.FindOne(curCtx, bson.M{"table_id": order.TableID}).Decode(&table)
		if err != nil {
			msg := fmt.Sprintf("Table wasn't found")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
		}

		order.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.ID = primitive.NewObjectID()
		order.OrderID = order.ID.Hex()
		order.OrderDate, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		result, insertErr := orderCollection.InsertOne(curCtx, order)
		if insertErr != nil {
			msg := fmt.Sprintf("Failed to create an order")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table
		var order models.Order

		if err := ctx.BindJSON(&order); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		var updatedObj primitive.D

		orderId := ctx.Param("order_id")

		if order.TableID != nil {
			err := orderCollection.FindOne(curCtx, bson.M{
				"table_id": order.TableID,
			}).Decode(&table)

			if err != nil {
				msg := fmt.Sprintf("Table was not found")
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updatedObj = append(updatedObj, bson.E{Key: "table_id", Value: order.TableID})
		}

		order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		updatedObj = append(updatedObj, bson.E{Key: "updated_at", Value: order.UpdatedAt})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		filter := bson.M{"order_id": orderId}
		result, err := orderCollection.UpdateOne(
			curCtx, filter, bson.D{
				{Key: "$set", Value: updatedObj},
			},
			&opt,
		)

		if err != nil {
			msg := "Order updation failed"
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func OrderItemOrderCreator(order models.Order) string {
	curCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	order.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.OrderID = order.ID.Hex()

	orderCollection.InsertOne(curCtx, order)
	defer cancel()
	return order.OrderID
}
