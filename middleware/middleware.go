package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/neilZon/workout-logger-api/utils/token"
)

const UserCtxKey string = "USER"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("Authorization")

		// decode token to get user
		claims, _ := token.Decode(t, []byte(os.Getenv("ACCESS_SECRET")))
		
		// put it in context
		ctx := context.WithValue(r.Context(), UserCtxKey, claims)

		// and call the next with our new context
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func GetUser(ctx context.Context) (*token.Claims, error) {
	u, ok := ctx.Value(UserCtxKey).(*token.Claims)
	if !ok || (token.Claims{}) == *u {
		return nil, errors.New("Invalid Token")	
	}
	return u, nil
}
