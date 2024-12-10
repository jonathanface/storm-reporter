package dao_test

import (
	"context"
	"testing"

	"github.com/jonathanface/storm-reporter/API/dao"
	"github.com/jonathanface/storm-reporter/API/models"
	"github.com/stretchr/testify/assert"
)

type MockMongoCollection struct{}

func (m *MockMongoCollection) Find(ctx context.Context, filter interface{}) ([]models.StormReport, error) {
	return []models.StormReport{
		{Date: "2024-12-09", Location: "Test City", Type: "tornado"},
	}, nil
}

func TestGetStormReports(t *testing.T) {
	mockDAO := &dao.MockStormDAO{
		MockGetStormReports: func(start string, end string) ([]models.StormReport, error) {
			return []models.StormReport{
				{Date: "2024-12-09", Location: "Test City", Type: "tornado"},
			}, nil
		},
	}

	startDate := "1733773445"
	endDate := "1733777109"

	reports, err := mockDAO.GetStormReports(startDate, endDate)

	assert.NoError(t, err, "Expected no error")
	assert.Len(t, reports, 1, "Expected one report")
	assert.Equal(t, "2024-12-09", reports[0].Date)
	assert.Equal(t, "Test City", reports[0].Location)
	assert.Equal(t, "tornado", string(reports[0].Type))
}
