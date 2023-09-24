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

type InvoiceViewFormat struct {
	InvoiceID      string
	PaymentMethod  string
	OrderID        string
	PaymentStatus  *string
	PaymentDue     interface{}
	TableNumber    interface{}
	PaymentDueDate time.Time
	OrderDetails   interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		result, err := invoiceCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching invoices",
			})
			return
		}

		var allInvoices []bson.M

		if err = result.All(c, &allInvoices); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		invoiceId := ctx.Param("invoice_id")

		var invoice models.Invoice

		err := invoiceCollection.FindOne(curCtx, bson.M{"order_id": invoiceId}).Decode(&invoice)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error occured while fetching invoice",
			})
		}
		var invoiceView InvoiceViewFormat
		allOrderItems, err := ItemsByOrder(invoice.OrderId)

		invoiceView.OrderID = invoice.OrderId
		invoiceView.InvoiceID = invoice.InvoiceId
		invoiceView.PaymentDueDate = invoice.PaymentDueData
		invoiceView.PaymentStatus = *&invoice.PaymentStatus

		invoiceView.PaymentMethod = "null"
		if invoice.PaymentMethod != nil {
			invoiceView.PaymentMethod = *invoice.PaymentMethod
		}

		invoiceView.PaymentDue = allOrderItems[0]["payment_due"]
		invoiceView.OrderDetails = allOrderItems[0]["table_number"]
		invoiceView.TableNumber = allOrderItems[0]["order_items"]

		ctx.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var invoice models.Invoice
		if err := ctx.BindJSON(&invoice); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		var order models.Order
		err := orderCollection.FindOne(curCtx, bson.M{
			"order_id": invoice.OrderId,
		}).Decode(&order)

		if err != nil {
			msg := fmt.Sprintf("Order was not found")
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Error(),
			})
			return
		}

		// setting up the values
		status := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &status
		}
		invoice.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.PaymentDueData, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.InvoiceId = invoice.ID.Hex()

		result, insertErr := invoiceCollection.InsertOne(curCtx, invoice)
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

func UpdateInvoice() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		curCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var invoice models.Invoice
		if err := ctx.BindJSON(&invoice); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		var updatedObj primitive.D

		invoiceId := ctx.Param("invoice_id")
		filter := bson.M{"invoice_id": invoiceId}

		if invoice.PaymentMethod != nil {
			updatedObj = append(updatedObj, bson.E{Key: "payment_method", Value: invoice.PaymentMethod})
		}
		if invoice.PaymentStatus != nil {
			updatedObj = append(updatedObj, bson.E{Key: "payment_status", Value: invoice.PaymentStatus})
		}

		invoice.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updatedObj = append(updatedObj, bson.E{Key: "updated_at", Value: invoice.UpdatedAt})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		status := "PENDING"
		if invoice.PaymentStatus == nil {
			invoice.PaymentStatus = &status
			updatedObj = append(updatedObj, bson.E{Key: "payment_status", Value: invoice.PaymentStatus})
		}

		result, err := menuCollection.UpdateOne(
			curCtx, filter, bson.D{
				{Key: "$set", Value: updatedObj},
			},
			&opt,
		)

		if err != nil {
			msg := "Invoice updation failed"
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
