package mocks

import (
	dto "main/internal/dto"

	mock "github.com/stretchr/testify/mock"
)

type OrderEventService struct {
	mock.Mock
}

func (_m *OrderEventService) ImportOrderEvents(reqs []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error) {
	ret := _m.Called(reqs)

	if len(ret) == 0 {
		panic("no return value specified for ImportOrderEvents")
	}

	var r0 dto.ImportOrderEventsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func([]dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error)); ok {
		return rf(reqs)
	}
	if rf, ok := ret.Get(0).(func([]dto.ImportOrderEventRequest) dto.ImportOrderEventsResponse); ok {
		r0 = rf(reqs)
	} else {
		r0 = ret.Get(0).(dto.ImportOrderEventsResponse)
	}

	if rf, ok := ret.Get(1).(func([]dto.ImportOrderEventRequest) error); ok {
		r1 = rf(reqs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewOrderEventService(t interface {
	mock.TestingT
	Cleanup(func())
}) *OrderEventService {
	mock := &OrderEventService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
