package controllers

import (
	"context"
	"fmt"
	"infinity/rms/database"
	"infinity/rms/helpers"
	"infinity/rms/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

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
					{Key: "user_items", Value: bson.D{{
						Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
					},
					},
				},
			},
		}

		result, err := foodCollection.Aggregate(curCtx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching users",
			})
		}

		var allUsers []bson.M

		if err = result.All(curCtx, &allUsers); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allUsers)
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userId := ctx.Param("user_id")

		var user models.User

		err := userCollection.FindOne(curCtx, bson.M{"user_id": userId}).Decode(&user)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching",
			})
		}

		ctx.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		// Conversion
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Validation
		validationErr := validate.Struct(user)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Error(),
			})
			return
		}

		// Check singleton for email & phone
		count , err := userCollection.CountDocuments(curCtx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occurred while checking the email",
			})
			return
		}

		count , err = userCollection.CountDocuments(curCtx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occurred while checking the phone",
			})
			return
		}

		if count > 1 {
			ctx.JSON(
				http.StatusInternalServerError, gin.H{
					"error": "This email or phone number already in Use",
				},
			)
		}
		// hash password
		password := HashPassword(*user.Password)
		user.Password = &password

		// create the time stamps & id
		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()

		// generate token & refresh token
		token , refreshToken, _ := helpers.GenerateAllTokens(&user)
		user.Token = &token
		user.RefreshToken = &refreshToken

		// then insertion

		result, insertErr := userCollection.InsertOne(curCtx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User not created")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		ctx.JSON(http.StatusOK, result)

	}
}

func LogIn() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User
		// Conversion
		if err := ctx.BindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// find the user in mongo
		err := userCollection.FindOne(curCtx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user not found, login seems to be incorrect"})
			return
		}

		// verify password
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		// generate tokens

		token, refreshToken, _ := helpers.GenerateAllTokens(&foundUser)

		//update tokens - token and refersh token
		helpers.UpdateAllTokens(token, refreshToken, foundUser.UserID)

		//return statusOK
		ctx.JSON(http.StatusOK, foundUser)
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		fmt.Println("Hashing failed")
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))
	return err == nil, err.Error()
}
