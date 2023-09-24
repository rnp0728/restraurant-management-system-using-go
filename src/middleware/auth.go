package middleware

import (
	"infinity/rms/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		clientToken := ctx.Request.Header.Get("token")

		if clientToken == "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Token is Empty",
			})
			ctx.Abort()
			return
		}

		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
		}

		ctx.Set("email", claims.Email)
		ctx.Set("first_name", claims.FirstName)
		ctx.Set("last_name", claims.LastName)
		ctx.Set("uid", claims.UID)

		ctx.Next()
	}
}