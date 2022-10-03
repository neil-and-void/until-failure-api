package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/neilZon/workout-logger-api/utils/token"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("Authorization")

		// decode token to get user
		claims, _ := token.Decode(t, []byte(os.Getenv("ACCESS_TOKEN")))
		user := token.ClaimsToStruct(claims)
		
		// put it in context
		ctx := context.WithValue(r.Context(), "user", user)

		// and call the next with our new context
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
