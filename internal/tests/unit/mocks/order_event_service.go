package mocks

import (
	"context"
	dto "main/internal/dto"

	mock "github.com/stretchr/testify/mock"
)

type OrderEventService struct {
	mock.Mock
}

func (_m *OrderEventService) ImportOrderEvents(ctx context.Context, reqs []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error) {
	ret := _m.Called(ctx, reqs)

	if len(ret) == 0 {
		panic("no return value specified for ImportOrderEvents")
	}

	var r0 dto.ImportOrderEventsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []dto.ImportOrderEventRequest) (dto.ImportOrderEventsResponse, error)); ok {
		return rf(ctx, reqs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []dto.ImportOrderEventRequest) dto.ImportOrderEventsResponse); ok {
		r0 = rf(ctx, reqs)
	} else {
		r0 = ret.Get(0).(dto.ImportOrderEventsResponse)
	}

	if rf, ok := ret.Get(1).(func(context.Context, []dto.ImportOrderEventRequest) error); ok {
		r1 = rf(ctx, reqs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *OrderEventService) UpdateDriverNote(ctx context.Context, orderID int64, note string) error {
	ret := _m.Called(ctx, orderID, note)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string) error); ok {
		r0 = rf(ctx, orderID, note)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

func (_m *OrderEventService) DriverUpdateStatus(ctx context.Context, orderID int64, newStatus string, updatedBy string) error {
	ret := _m.Called(ctx, orderID, newStatus, updatedBy)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string, string) error); ok {
		r0 = rf(ctx, orderID, newStatus, updatedBy)
	} else {
		r0 = ret.Error(0)
	}

	return r0
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
