package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	db "github.com/neilZon/workout-logger-api/common/database"
	"github.com/neilZon/workout-logger-api/graph"
	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const defaultPort = "8080"

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	db, err := db.InitDb()
	if err != nil {
		log.Fatal(err)
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		DB: db,
	}}))
	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		// notify bug tracker...maybe? idk too much money
		return gqlerror.Errorf("Internal server error")
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", 
	cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080", "https://studio.apollographql.com"},
		AllowCredentials: true,
		Debug:            true,
		AllowedHeaders: []string{"Content-Type","Authorization"},
	}).Handler(middleware.AuthMiddleware(srv)))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
