package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/neilZon/workout-logger-api/accesscontroller/accesscontrol"
	db "github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/helpers"
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

	acs := accesscontrol.NewAccessControllerService(db)
	srv := helpers.NewGqlServer(db, acs)
	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
		// notify bug tracker...maybe? idk too much mone˜
		if err != nil {
			fmt.Println(err)
		}
		return gqlerror.Errorf("Internal server error")
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1", "http://localhost:8080", "https://hoppscotch.io/"},
		AllowCredentials: true,
		Debug:            true,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", c.Handler(middleware.AuthMiddleware(srv)))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
