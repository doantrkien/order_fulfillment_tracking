package mocks

import (
	"context"
	dto_api "main/internal/dto/api"

	mock "github.com/stretchr/testify/mock"
)

type OrderEventService struct {
	mock.Mock
}

func (_m *OrderEventService) ImportOrderEvents(ctx context.Context, reqs []dto_api.ImportOrderEventRequest) (dto_api.ImportOrderEventsResponse, error) {
	ret := _m.Called(ctx, reqs)

	if len(ret) == 0 {
		panic("no return value specified for ImportOrderEvents")
	}

	var r0 dto_api.ImportOrderEventsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []dto_api.ImportOrderEventRequest) (dto_api.ImportOrderEventsResponse, error)); ok {
		return rf(ctx, reqs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []dto_api.ImportOrderEventRequest) dto_api.ImportOrderEventsResponse); ok {
		r0 = rf(ctx, reqs)
	} else {
		r0 = ret.Get(0).(dto_api.ImportOrderEventsResponse)
	}

	if rf, ok := ret.Get(1).(func(context.Context, []dto_api.ImportOrderEventRequest) error); ok {
		r1 = rf(ctx, reqs)
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
