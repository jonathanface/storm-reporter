package dao

import "github.com/jonathanface/storm-reporter/API/models"

type MockStormDAO struct {
	MockGetStormReports func(start string, end string) ([]models.StormReport, error)
}

func (m *MockStormDAO) GetStormReports(start string, end string) ([]models.StormReport, error) {
	return m.MockGetStormReports(start, end)
}

func (m *MockStormDAO) Disconnect() error {
	return nil
}
