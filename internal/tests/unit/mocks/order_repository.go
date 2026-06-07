package mocks

import (
	dto "main/internal/dto"
	models "main/internal/models"

	mock "github.com/stretchr/testify/mock"
)

type OrderRepository struct {
	mock.Mock
}

func (_m *OrderRepository) CreateOrder(order models.Order, updatedBy string) (*models.Order, error) {
	ret := _m.Called(order, updatedBy)

	if len(ret) == 0 {
		panic("no return value specified for CreateOrder")
	}

	var r0 *models.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(models.Order, string) (*models.Order, error)); ok {
		return rf(order, updatedBy)
	}
	if rf, ok := ret.Get(0).(func(models.Order, string) *models.Order); ok {
		r0 = rf(order, updatedBy)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(models.Order, string) error); ok {
		r1 = rf(order, updatedBy)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *OrderRepository) GetAllOrder(query dto.OrderQuery) ([]models.Order, int64, error) {
	ret := _m.Called(query)

	if len(ret) == 0 {
		panic("no return value specified for GetAllOrder")
	}

	var r0 []models.Order
	var r1 int64
	var r2 error
	if rf, ok := ret.Get(0).(func(dto.OrderQuery) ([]models.Order, int64, error)); ok {
		return rf(query)
	}
	if rf, ok := ret.Get(0).(func(dto.OrderQuery) []models.Order); ok {
		r0 = rf(query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(dto.OrderQuery) int64); ok {
		r1 = rf(query)
	} else {
		r1 = ret.Get(1).(int64)
	}

	if rf, ok := ret.Get(2).(func(dto.OrderQuery) error); ok {
		r2 = rf(query)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

func (_m *OrderRepository) GetOrderDetail(id int64) (*models.Order, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for GetOrderDetail")
	}

	var r0 *models.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(int64) (*models.Order, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(int64) *models.Order); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *OrderRepository) UpdateOrderStatus(id int64, status string, updatedBy string, driverID *int64) (*models.Order, error) {
	ret := _m.Called(id, status, updatedBy, driverID)

	if len(ret) == 0 {
		panic("no return value specified for UpdateOrderStatus")
	}

	var r0 *models.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(int64, string, string, *int64) (*models.Order, error)); ok {
		return rf(id, status, updatedBy, driverID)
	}
	if rf, ok := ret.Get(0).(func(int64, string, string, *int64) *models.Order); ok {
		r0 = rf(id, status, updatedBy, driverID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(int64, string, string, *int64) error); ok {
		r1 = rf(id, status, updatedBy, driverID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewOrderRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *OrderRepository {
	mock := &OrderRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
