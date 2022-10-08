package main

import (
	"context"
	"fmt"
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
		if err != nil { 
			fmt.Println(err)
		}
		return gqlerror.Errorf("Internal server error")
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1", "http://localhost:8080", "https://hoppscotch.io/"},
		AllowCredentials: true,
		Debug:            true,
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))	
	http.Handle("/query", c.Handler(middleware.AuthMiddleware(srv)))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
