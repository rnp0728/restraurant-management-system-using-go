package main

import (
  "os"

  "infinity/rms/routes"
  "infinity/rms/middleware"
  "infinity/rms/database"
  "go.mongodb.org/mongo-driver/mongo"
  "github.com/gin-gonic/gin"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {
  port := os.Getenv("PORT")

  if port == ""{
    port = "8000"
  }

  router := gin.New()
  router.Use(gin.Logger())
  router.UserRoutes(router)
  router.Use(middleware.Auth())

  router.FoodRoutes(router)
  router.InvoiceRoutes(router)
  router.MenuRoutes(router)
  // router.NoteRoutes(router)
  router.OrderRoutes(router)
  router.OrderItemRoutes(router)
  router.TableRoutes(router)

  router.Run(":" + port)
}
