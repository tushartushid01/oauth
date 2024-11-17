package main

import (
	"Oauth/database"
	"Oauth/handler"
	"Oauth/middleware"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

var chiLambda *chiadapter.ChiLambda

func main() {
	// Database configuration from environment variables
	host := "localhost"
	port := "5432"
	databaseName := "oauth"
	user := "postgres"
	password := "1234"

	// Initialize the database connection
	err := database.ConnectAndMigrate(host, port, databaseName, user, password, database.SSLModeDisable)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize router and setup routes
	r := chi.NewRouter()
	r.Use(middleware.CommonMiddlewares()...) // Add any common middlewares

	// Define routes
	r.Route("/v1", func(audiophile chi.Router) {
		// Health check endpoint
		audiophile.Route("/health", func(r chi.Router) {
			r.Get("/api", func(w http.ResponseWriter, r *http.Request) {
				_, err := fmt.Fprintf(w, "health")
				if err != nil {
					log.Printf("Health check failed: %v", err)
				}
			})
		})

		// Authentication routes
		audiophile.Post("/log-in", handler.Login)
		audiophile.Post("/register", handler.Register)

		// Protected routes
		audiophile.Route("/auth", func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware) // Add authentication middleware

			auth.Put("/update-password", handler.UpdatePassword)
			auth.Get("/products", handler.GetProducts)
			auth.Post("/products", handler.BuyProduct)
			auth.Post("/feedback", handler.CreateFeedback)
			auth.Post("/log-out", handler.Logout)

			// Admin-specific routes
			auth.Route("/admin", func(admin chi.Router) {
				admin.Use(middleware.AdminMiddleware) // Admin-only middleware
				admin.Post("/sell-product", handler.CreateProduct)
			})
		})
	})

	// Wrap Chi router with AWS Lambda adapter
	chiLambda = chiadapter.New(r)

	// Start Lambda handler
	lambda.Start(LambdaHandler)
}

// LambdaHandler handles requests when run in Lambda
func LambdaHandler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return chiLambda.Proxy(req)
}
