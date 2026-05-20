package mocks

import (
	"main/internal/models"
	"time"

	mock "github.com/stretchr/testify/mock"
)

type ReportService struct {
	mock.Mock
}

func (_m *ReportService) GetDailyReport(date time.Time) (*models.Report, error) {
	ret := _m.Called(date)

	if len(ret) == 0 {
		panic("no return value specified for GetDailyReport")
	}

	var r0 *models.Report
	var r1 error
	if rf, ok := ret.Get(0).(func(time.Time) (*models.Report, error)); ok {
		return rf(date)
	}
	if rf, ok := ret.Get(0).(func(time.Time) *models.Report); ok {
		r0 = rf(date)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Report)
		}
	}

	if rf, ok := ret.Get(1).(func(time.Time) error); ok {
		r1 = rf(date)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *ReportService) CreateDailyReport(date time.Time) (*models.Report, error) {
	ret := _m.Called(date)

	if len(ret) == 0 {
		panic("no return value specified for CreateDailyReport")
	}

	var r0 *models.Report
	var r1 error
	if rf, ok := ret.Get(0).(func(time.Time) (*models.Report, error)); ok {
		return rf(date)
	}
	if rf, ok := ret.Get(0).(func(time.Time) *models.Report); ok {
		r0 = rf(date)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Report)
		}
	}

	if rf, ok := ret.Get(1).(func(time.Time) error); ok {
		r1 = rf(date)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewReportService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ReportService {
	mock := &ReportService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
