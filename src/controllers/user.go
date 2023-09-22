package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers() gin.HandlerFunc{
	return func(ctx *gin.Context) {

	}
}

func GetUser() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		
	}
}

func SignUp() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		
	}
}

func LogIn() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		
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