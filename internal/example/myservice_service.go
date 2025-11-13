package example

import (
	"context"
	"fmt"
	"strings"
)

// MyServiceService handles business logic for myservice
type MyServiceService interface {
	DoSomething(ctx context.Context, input string) (string, error)
}

type myServiceService struct {
	repo MyServiceRepository
}

// NewMyServiceService creates a new instance of MyServiceService
func NewMyServiceService(repo MyServiceRepository) MyServiceService {
	return &myServiceService{
		repo: repo,
	}
}

// DoSomething processes the input and stores the result
func (s *myServiceService) DoSomething(ctx context.Context, input string) (string, error) {
	// Validate input
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	// Business logic: transform the input (example: uppercase and add prefix)
	output := fmt.Sprintf("PROCESSED: %s", strings.ToUpper(input))

	// Store in database
	record, err := s.repo.Create(ctx, input, output)
	if err != nil {
		return "", fmt.Errorf("failed to store record: %w", err)
	}

	// Return the processed result
	return fmt.Sprintf("Result (ID: %d): %s", record.ID, output), nil
}
