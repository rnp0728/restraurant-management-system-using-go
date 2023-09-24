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

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		result, err := tableCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching all tables",
			})
		}

		var allTables []bson.M

		if err = result.All(c, &allTables); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allTables)
	}
}

func GetTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		tableId := ctx.Param("table_id")

		var table models.Table

		err := tableCollection.FindOne(curCtx, bson.M{"table_id": tableId}).Decode(&table)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching",
			})
		}

		ctx.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table

		if err := ctx.BindJSON(&table); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		validationErr := validate.Struct(table)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Error(),
			})
			return
		}

		table.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.ID = primitive.NewObjectID()
		table.TableID = table.ID.Hex()

		result, insertErr := orderCollection.InsertOne(curCtx, table)
		if insertErr != nil {
			msg := fmt.Sprintf("Failed to create a table item")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table

		if err := ctx.BindJSON(&table); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		var updatedObj primitive.D

		tableId := ctx.Param("table_id")

		if table.NumberOfGuests != nil {
			updatedObj = append(updatedObj, bson.E{Key: "number_of_guests", Value: table.NumberOfGuests})
		}

		if table.TableNumber != nil {
			updatedObj = append(updatedObj, bson.E{Key: "table_number", Value: table.TableNumber})
		}

		table.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		updatedObj = append(updatedObj, bson.E{Key: "updated_at", Value: table.UpdatedAt})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		filter := bson.M{"table_id": tableId}
		result, err := tableCollection.UpdateOne(
			curCtx, filter, bson.D{
				{Key: "$set", Value: updatedObj},
			},
			&opt,
		)

		if err != nil {
			msg := "Table updation failed"
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
