package saga

import (
	"context"
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// Step represents a saga step
type Step struct {
	Name       string
	Execute    func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

// Saga orchestrates distributed transactions
type Saga struct {
	name  string
	steps []Step
}

// New creates a new saga
func New(name string) *Saga {
	return &Saga{
		name:  name,
		steps: make([]Step, 0),
	}
}

// AddStep adds a step to the saga
func (s *Saga) AddStep(name string, execute, compensate func(ctx context.Context) error) *Saga {
	s.steps = append(s.steps, Step{
		Name:       name,
		Execute:    execute,
		Compensate: compensate,
	})
	return s
}

// Execute executes the saga
func (s *Saga) Execute(ctx context.Context) error {
	logger.Infof("Starting saga: %s with %d steps", s.name, len(s.steps))

	// Track completed steps for compensation
	completedSteps := 0

	// Execute each step
	for i, step := range s.steps {
		logger.Infof("Executing step %d/%d: %s", i+1, len(s.steps), step.Name)

		if err := step.Execute(ctx); err != nil {
			logger.Errorf("Step %s failed: %v", step.Name, err)

			// Compensate completed steps in reverse order
			if err := s.compensate(ctx, completedSteps); err != nil {
				logger.Errorf("Compensation failed: %v", err)
				return fmt.Errorf("step failed and compensation also failed: %w", err)
			}

			return fmt.Errorf("saga failed at step %s: %w", step.Name, err)
		}

		completedSteps++
		logger.Infof("Step %s completed successfully", step.Name)
	}

	logger.Infof("Saga %s completed successfully", s.name)
	return nil
}

// compensate compensates completed steps in reverse order
func (s *Saga) compensate(ctx context.Context, completedSteps int) error {
	logger.Infof("Starting compensation for %d steps", completedSteps)

	// Compensate in reverse order
	for i := completedSteps - 1; i >= 0; i-- {
		step := s.steps[i]
		logger.Infof("Compensating step %d: %s", i+1, step.Name)

		if err := step.Compensate(ctx); err != nil {
			logger.Errorf("Compensation failed for step %s: %v", step.Name, err)
			return err
		}

		logger.Infof("Step %s compensated successfully", step.Name)
	}

	logger.Infof("Compensation completed successfully")
	return nil
}
