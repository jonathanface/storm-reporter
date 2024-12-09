package middleware

import (
	"context"
	"net/http"

	"github.com/jonathanface/storm-reporter/dao"
)

type contextKey string

const daoKey contextKey = "dao"

func WithDAOContext(dao *dao.StormDAO) func(http.HandlerFunc) http.HandlerFunc {
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
