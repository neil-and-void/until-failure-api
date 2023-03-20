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
	"github.com/neilZon/workout-logger-api/config"
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

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		// Open the file specified by the request path
		file, err := os.Open("." + r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		// Get the file information, including the modification time
		info, err := file.Stat()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header to "text/html"
		w.Header().Set("Content-Type", "text/html")

		// Serve the file content using http.ServeContent
		http.ServeContent(w, r, "", info.ModTime(), file)
	})

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
		host := os.Getenv(config.HOST)

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Redirect(w, r, fmt.Sprintf("%s/static/verification-failure.html", host), http.StatusSeeOther)
		}

		expiryTime := time.Now().Add(24 * time.Hour)
		user, err := database.GetUserByVerificationCode(b.DB, code)
		if err != nil || user == nil || user.VerificationCode == nil || *user.VerificationCode != code || user.VerificationSentAt == nil || user.VerificationSentAt.After(expiryTime) {
			http.Redirect(w, r, fmt.Sprintf("%s/static/verification-failure.html", host), http.StatusSeeOther)
			return
		}

		if user.Verified {
			http.Redirect(w, r, fmt.Sprintf("%s/static/verification-failure.html", host), http.StatusSeeOther)
			return
		}

		err = database.VerifyUser(b.DB, fmt.Sprintf("%d", user.ID), code)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s/static/verification-failure.html", host), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("%s/static/verification-success.html", host), http.StatusSeeOther)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 Method not allowed"))
		return
	}
}
