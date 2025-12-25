package booking

import (
	"fmt"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/logger"
)

// StateMachine manages booking state transitions
type StateMachine struct {
	booking *Booking
}

// NewStateMachine creates a new state machine
func NewStateMachine(booking *Booking) *StateMachine {
	return &StateMachine{
		booking: booking,
	}
}

// Transition transitions the booking to a new state
func (sm *StateMachine) Transition(newStatus BookingStatus) error {
	if !sm.booking.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid state transition from %s to %s", sm.booking.Status, newStatus)
	}

	oldStatus := sm.booking.Status
	sm.booking.Status = newStatus

	logger.Infof("Booking %s transitioned from %s to %s", sm.booking.ID, oldStatus, newStatus)
	return nil
}

// GetStatus returns current booking status
func (sm *StateMachine) GetStatus() BookingStatus {
	return sm.booking.Status
}

// IsFinal checks if current state is final
func (sm *StateMachine) IsFinal() bool {
	return sm.booking.Status == StatusCompleted || sm.booking.Status == StatusCancelled
}
