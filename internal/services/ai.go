package services

import (
	"main/internal/repositories"
)

type AIService interface {
	AnalyzeException() error
}

type aiService struct {
	aiRepo repositories.AIRepository
}

func NewAIService(aiRepo repositories.AIRepository) AIService {
	return &aiService{aiRepo: aiRepo}
}

func (s *aiService) AnalyzeException() error {
	// Implementation for analyzing exceptions
	return nil
}
