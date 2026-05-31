package mocks

import (
	dto "main/internal/dto"
	models "main/internal/models"

	mock "github.com/stretchr/testify/mock"
)

type OrderService struct {
	mock.Mock
}

func (_m *OrderService) CreateOrder(_a0 dto.OrderRequest) (*models.Order, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for CreateOrder")
	}

	var r0 *models.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(dto.OrderRequest) (*models.Order, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(dto.OrderRequest) *models.Order); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(dto.OrderRequest) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *OrderService) GetAllOrder(query dto.OrderQuery, role string, userID int64) ([]dto.OrderReponse, int64, error) {
	ret := _m.Called(query, role, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetAllOrder")
	}

	var r0 []dto.OrderReponse
	var r1 int64
	var r2 error
	if rf, ok := ret.Get(0).(func(dto.OrderQuery, string, int64) ([]dto.OrderReponse, int64, error)); ok {
		return rf(query, role, userID)
	}
	if rf, ok := ret.Get(0).(func(dto.OrderQuery) []dto.OrderReponse); ok {
		r0 = rf(query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]dto.OrderReponse)
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

func (_m *OrderService) GetOrder(id int64, role string, userID int64) (*dto.OrderReponse, error) {
	ret := _m.Called(id, role, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetOrder")
	}

	var r0 *dto.OrderReponse
	var r1 error
	if rf, ok := ret.Get(0).(func(int64, string, int64) (*dto.OrderReponse, error)); ok {
		return rf(id, role, userID)
	}
	if rf, ok := ret.Get(0).(func(int64) *dto.OrderReponse); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*dto.OrderReponse)
		}
	}

	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func (_m *OrderService) UpdateOrderStatus(id int64, status string, updatedBy string) (*models.Order, error) {
	ret := _m.Called(id, status, updatedBy)

	if len(ret) == 0 {
		panic("no return value specified for UpdateOrderStatus")
	}

	var r0 *models.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(int64, string, string) (*models.Order, error)); ok {
		return rf(id, status, updatedBy)
	}
	if rf, ok := ret.Get(0).(func(int64, string, string) *models.Order); ok {
		r0 = rf(id, status, updatedBy)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(int64, string, string) error); ok {
		r1 = rf(id, status, updatedBy)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

func NewOrderService(t interface {
	mock.TestingT
	Cleanup(func())
}) *OrderService {
	mock := &OrderService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
