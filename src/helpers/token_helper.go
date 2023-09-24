package helpers

import (
	"context"
	"infinity/rms/database"
	"infinity/rms/models"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")
var SECRETKEY string = os.Getenv("SECRET_KEY")

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	UID       string
	jwt.StandardClaims
}

func GenerateAllTokens(user *models.User) (string, string, error) {
	claims := &SignedDetails{
		Email:     *user.Email,
		FirstName: *user.FirstName,
		LastName:  *user.LastName,
		UID:       user.UserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString([]byte(SECRETKEY))
	if err != nil {
		log.Panic(err)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims).SignedString([]byte(SECRETKEY))
	if err != nil {
		log.Panic(err)
	}
	return token, refreshToken, err

}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updatedObj primitive.D

	updatedObj = append(updatedObj, bson.E{Key: "token", Value: signedToken})
	updatedObj = append(updatedObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	UpdatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	updatedObj = append(updatedObj, bson.E{Key: "updated_at", Value: UpdatedAt})

	upsert := true

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	filter := bson.M{"user_id": userId}

	_, err := userCollection.UpdateOne(
		curCtx, filter, bson.D{
			{Key: "$set", Value: updatedObj},
		},
		&opt,
	)

	if err != nil {
		log.Panic(err)
	}
}

func ValidateToken(signedToken string) (claims * SignedDetails, msg string) {
	token , err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRETKEY), nil
		},
	)

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "The token is invalid"
		return
	}
	
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "Token is Expaired"
		return
	}

	msg = err.Error()

	return claims, msg
}
