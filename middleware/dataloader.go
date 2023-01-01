package middleware

import (
	"context"
	"net/http"

	"github.com/neilZon/workout-logger-api/loader"
)

type ctxKey string

const (
	LoadersKey = ctxKey("DATALOADERS")
)

// Middleware injects data loaders into the context
func DataloaderMiddleware(loaders *loader.Loaders, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCtx := context.WithValue(r.Context(), LoadersKey, loaders)
		r = r.WithContext(nextCtx)
		next.ServeHTTP(w, r)
	})
}

func GetLoaders(ctx context.Context) *loader.Loaders {
	return ctx.Value(LoadersKey).(*loader.Loaders)
}
