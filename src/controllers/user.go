package controllers

import (
	"infinity/rms/models"

	"github.com/gin-gonic/gin"
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

}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	
} 