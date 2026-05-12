package services

import (
	"main/internal/models"
	"main/internal/repositories"
)

type OrderEventService interface {
	ImportOrderEvents(events []models.OrderEvent) error
}

type orderEventService struct {
	eventRepo repositories.OrderEventRepository
}

func NewOrderEventService(eventRepo repositories.OrderEventRepository) OrderEventService {
	return &orderEventService{
		eventRepo: eventRepo,
	}
}

func (s *orderEventService) ImportOrderEvents(events []models.OrderEvent) error {
	// TODO
	return nil
}
