package database

import (
	"context"
	"net/http"

	"gorm.io/gorm"
)

const dbContextKey = "DB_CONTEXT"

type DBContext struct {
	Database *gorm.DB
}

func CreateContext(args *DBContext, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbContext := &DBContext{
			Database: args.Database,
		}
		requestWithCtx := r.WithContext(context.WithValue(r.Context(), dbContextKey, dbContext))
		next.ServeHTTP(w, requestWithCtx)
	})
}

func GetContext(ctx context.Context) *DBContext {
	dbContext, ok := ctx.Value(dbContextKey).(*DBContext)
	if !ok {
		return nil
	}
	return dbContext
}
