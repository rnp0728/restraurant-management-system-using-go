package controllers

import (
	"context"
	"fmt"
	"infinity/rms/database"
	"infinity/rms/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Food Collection
var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(ctx.Query("startIndex"))

		// match stage, group stage, project stage
		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}}, {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
		projectStage := bson.D{
			{
				Key: "$projeValue: ct", Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "total_count", Value: 1},
					{
						Key: "food_items", Value: bson.D{
							{
								Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage},
							},
						},
					},
				},
			},
		}

		result, err := foodCollection.Aggregate(c, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching",
			})
		}

		var foods []bson.M

		if err = result.All(c, &foods); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, foods)
	}
}

func GetFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		foodId := ctx.Param("food_id")

		var food models.Food

		err := foodCollection.FindOne(curCtx, bson.M{"food_id": foodId}).Decode(&food)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching",
			})
		}

		ctx.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var menu models.Menu
		var food models.Food

		if err := ctx.BindJSON(&food); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		validationErr := validate.Struct(food)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Error(),
			})
			return
		}

		err := menuCollection.FindOne(curCtx, bson.M{"menu_id": food.MenuId}).Decode(&menu)
		if err != nil {
			msg := fmt.Sprintf("Menu not founf")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
		}

		food.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.FoodId = food.ID.Hex()
		num := toFixed(*food.Price, 2)
		food.Price = &num

		result, insertErr := foodCollection.InsertOne(curCtx, food)
		if insertErr != nil {
			msg := fmt.Sprintf("Food item was not created")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateFood() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var menu models.Menu
		var food models.Food

		if err := ctx.BindJSON(&menu); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		foodId := ctx.Param("food_id")

		filter := bson.M{"food_id": foodId}

		var updateObj primitive.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		}
		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
		}
		if food.FoodImage != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.FoodImage})
		}
		if food.MenuId != nil {
			err := menuCollection.FindOne(curCtx, bson.M{"menu_id": food.MenuId}).Decode(&menu)
			if err != nil {
				msg := fmt.Sprintf("Menu was not found")
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "menu_id", Value: food.MenuId})
		}

		food.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.UpdatedAt})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		result, err := menuCollection.UpdateOne(
			curCtx, filter, bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)
		if err != nil {
			msg := "Food update failed"
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})

		}
		ctx.JSON(http.StatusOK, result)
	}
}

func round(num float64) int {
	return int(num * math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
