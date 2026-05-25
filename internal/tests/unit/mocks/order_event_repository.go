package mocks

import (
	"context"
	"main/internal/models"
	"main/internal/repositories"

	mock "github.com/stretchr/testify/mock"
)

type OrderEventRepository struct {
	mock.Mock
}

func (_m *OrderEventRepository) ProcessSingleEventTx(ctx context.Context, event models.OrderEvent) (repositories.ProcessResultDetail, error) {
	ret := _m.Called(ctx, event)

	if len(ret) == 0 {
		panic("no return value specified for ProcessSingleEventTx")
	}

	var r0 repositories.ProcessResultDetail
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, models.OrderEvent) (repositories.ProcessResultDetail, error)); ok {
		return rf(ctx, event)
	}
	if rf, ok := ret.Get(0).(func(context.Context, models.OrderEvent) repositories.ProcessResultDetail); ok {
		r0 = rf(ctx, event)
	} else {
		r0 = ret.Get(0).(repositories.ProcessResultDetail)
	}

	if rf, ok := ret.Get(1).(func(context.Context, models.OrderEvent) error); ok {
		r1 = rf(ctx, event)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewOrderEventRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *OrderEventRepository {
	mock := &OrderEventRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
