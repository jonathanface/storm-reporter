package dao

import "github.com/jonathanface/storm-reporter/API/models"

type MockStormDAO struct {
	StormReports []models.StormReport
}

func (m *MockStormDAO) GetStormReports(start, end string) ([]models.StormReport, error) {
	return m.StormReports, nil
}

func (m *MockStormDAO) InsertStormReport(report models.StormReport) error {
	m.StormReports = append(m.StormReports, report)
	return nil
}
