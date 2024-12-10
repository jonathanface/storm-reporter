package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonathanface/storm-reporter/API/dao"
	"github.com/jonathanface/storm-reporter/API/middleware"
	"github.com/jonathanface/storm-reporter/API/models"
	"github.com/stretchr/testify/assert"
)

func TestWithDAOContext(t *testing.T) {
	mockDAO := &dao.MockStormDAO{
		MockGetStormReports: func(start string, end string) ([]models.StormReport, error) {
			return []models.StormReport{
				{Date: "2024-12-09", Location: "Test City", Type: "tornado"},
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	w := httptest.NewRecorder()

	handler := middleware.WithDAOContext(mockDAO)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dao := middleware.GetDAO(r.Context())
		reports, err := dao.GetStormReports("2024-12-09", "2024-12-09")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(reports))
		assert.Equal(t, "Test City", reports[0].Location)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response code should be 200")
}
