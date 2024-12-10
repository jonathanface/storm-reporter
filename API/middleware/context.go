package middleware

import (
	"context"
	"net/http"

	"github.com/jonathanface/storm-reporter/API/models"
)

type contextKey string

const daoKey contextKey = "dao"

func WithDAOContext(dao models.StormDAOInterface) func(http.HandlerFunc) http.HandlerFunc {
	// Normally would put some kind of auth here, but for now just return a middleware that sets the DAO in the context
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), daoKey, dao)
			next(w, r.WithContext(ctx))
		}
	}
}

func GetDAO(ctx context.Context) models.StormDAOInterface {
	dao, ok := ctx.Value(daoKey).(models.StormDAOInterface)
	if !ok {
		panic("DAO not found in context")
	}
	return dao
}
