package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/neilZon/workout-logger-api/common"
	"github.com/neilZon/workout-logger-api/config"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/token"
	"gorm.io/gorm"
)

const UserCtxKey string = "USER"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("Authorization")

		// decode token to get user
		claims, _ := token.Decode(t, []byte(os.Getenv(config.ACCESS_SECRET)))

		// put it in context
		ctx := context.WithValue(r.Context(), UserCtxKey, claims)

		// and call the next with our new context
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func GetUser(ctx context.Context) (*token.Claims, error) {
	u, ok := ctx.Value(UserCtxKey).(*token.Claims)
	if !ok || u == nil || (token.Claims{}) == *u {
		return nil, &common.UnauthorizedError{}
	}
	return u, nil
}

func VerifyUser(db *gorm.DB, userId string) error {
	user, err := database.GetUserById(db, userId)
	if err != nil {
		return errors.New("could not verify user")
	}
	if !user.Verified {
		return errors.New("user not verified")
	}
	return nil
}
