package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonathanface/storm-reporter/API/dao"
	"github.com/jonathanface/storm-reporter/API/middleware"
	"github.com/jonathanface/storm-reporter/API/models"
	"github.com/jonathanface/storm-reporter/API/routes"
)

func TestGetMessagesHandler(t *testing.T) {
	mockDAO := &dao.MockStormDAO{
		MockGetStormReports: func(start string, end string) ([]models.StormReport, error) {
			return []models.StormReport{
				{Date: "2024-12-09", Location: "Test City", Type: "tornado"},
			}, nil
		},
	}

	req, err := http.NewRequest("GET", "/messages?date=1733775461", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := middleware.WithDAOContext(mockDAO)(http.HandlerFunc(routes.GetMessagesHandler))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status OK; got %v", rr.Code)
	}

	var reports []models.StormReport
	err = json.Unmarshal(rr.Body.Bytes(), &reports)
	if err != nil {
		t.Fatalf("Could not parse response: %v", err)
	}

	if len(reports) != 1 || reports[0].Location != "Test City" {
		t.Errorf("Unexpected response: %v", reports)
	}
}
