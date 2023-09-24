package main

import (
  "os"

  middleware "infinity/rms/middleware"
  routes "infinity/rms/routes"
  "github.com/gin-gonic/gin"
)

func main() {
  port := os.Getenv("PORT")

  if port == ""{
    port = "8000"
  }

  router := gin.New()
  router.Use(gin.Logger())
  routes.UserRoutes(router)
  router.Use(middleware.Auth())

  routes.FoodRoutes(router)
  routes.InvoiceRoutes(router)
  routes.MenuRoutes(router)
  routes.OrderRoutes(router)
  routes.OrderItemRoutes(router)
  routes.TableRoutes(router)

  router.Run(":" + port)
}
