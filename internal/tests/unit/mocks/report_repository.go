package mocks

import (
	"main/internal/models"
	"time"

	mock "github.com/stretchr/testify/mock"
)

type ReportRepository struct {
	mock.Mock
}

func (_m *ReportRepository) GetDailyReport(date time.Time) (*models.Report, error) {
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

func (_m *ReportRepository) BuildDailyReport(start, end time.Time) (*models.Report, error) {
	ret := _m.Called(start, end)

	if len(ret) == 0 {
		panic("no return value specified for BuildDailyReport")
	}

	var r0 *models.Report
	var r1 error
	if rf, ok := ret.Get(0).(func(time.Time, time.Time) (*models.Report, error)); ok {
		return rf(start, end)
	}
	if rf, ok := ret.Get(0).(func(time.Time, time.Time) *models.Report); ok {
		r0 = rf(start, end)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Report)
		}
	}

	if rf, ok := ret.Get(1).(func(time.Time, time.Time) error); ok {
		r1 = rf(start, end)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *ReportRepository) SaveReport(report *models.Report) (*models.Report, error) {
	ret := _m.Called(report)

	if len(ret) == 0 {
		panic("no return value specified for SaveReport")
	}

	var r0 *models.Report
	var r1 error
	if rf, ok := ret.Get(0).(func(*models.Report) (*models.Report, error)); ok {
		return rf(report)
	}
	if rf, ok := ret.Get(0).(func(*models.Report) *models.Report); ok {
		r0 = rf(report)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Report)
		}
	}

	if rf, ok := ret.Get(1).(func(*models.Report) error); ok {
		r1 = rf(report)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewReportRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *ReportRepository {
	mock := &ReportRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
