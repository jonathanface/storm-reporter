package middleware

import (
	"context"
	"net/http"

	"github.com/jonathanface/storm-reporter/API/dao"
)

type contextKey string

const daoKey contextKey = "dao"

func WithDAOContext(dao *dao.StormDAO) func(http.HandlerFunc) http.HandlerFunc {
	// Normally would put some kind of auth here, but for now just return a middleware that sets the DAO in the context
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), daoKey, dao)
			next(w, r.WithContext(ctx))
		}
	}
}

func GetDAO(ctx context.Context) *dao.StormDAO {
	return ctx.Value(daoKey).(*dao.StormDAO)
}
