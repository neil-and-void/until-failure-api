package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/neilZon/workout-logger-api/accesscontroller/accesscontrol"
	"github.com/neilZon/workout-logger-api/database"
	db "github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/helpers"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/gorm"
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
	srv.Use(extension.Introspection{})
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
		Debug:            false,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
	})

	loaders := helpers.NewLoaders(db)

	dataloaderMiddleware := middleware.DataloaderMiddleware(loaders, srv)
	authMiddleware := middleware.AuthMiddleware(dataloaderMiddleware)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", c.Handler(authMiddleware))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))

	basehandler := &BaseHandler{
		DB: db,
	}
	http.HandleFunc("/verify", basehandler.verify)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type BaseHandler struct {
	DB *gorm.DB
}

func (b *BaseHandler) verify(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Redirect(w, r, "http://localhost:8080/static/verification-failure.html", http.StatusSeeOther)
		}

		expiryTime := time.Now().Add(24 * time.Hour)
		user, err := database.GetUserByCode(b.DB, code)
		if err != nil || user == nil || user.VerificationCode != code || time.Now().After(expiryTime) {
			http.Redirect(w, r, "http://localhost:8080/static/verification-failure.html", http.StatusSeeOther)
			return
		}

		if user.Verified {
			http.Redirect(w, r, "http://localhost:8080/static/verification-failure.html", http.StatusSeeOther)
			return
		}

		err = database.VerifyUser(b.DB, fmt.Sprintf("%d", user.ID), code)
		if err != nil {
			http.Redirect(w, r, "http://localhost:8080/static/verification-failure.html", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "http://localhost:8080/static/verification-success.html", http.StatusSeeOther)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 Method not allowed"))
		return
	}
}
