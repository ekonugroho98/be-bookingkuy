package booking

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStateMachine tests creating a new state machine
func TestNewStateMachine(t *testing.T) {
	booking := &Booking{
		ID:     "test-booking-1",
		Status: StatusInit,
	}

	sm := NewStateMachine(booking)

	require.NotNil(t, sm)
	assert.Equal(t, booking, sm.booking)
}

// TestStateMachine_Transition_ValidTransitions tests all valid state transitions
func TestStateMachine_Transition_ValidTransitions(t *testing.T) {
	tests := []struct {
		name         string
		initialState BookingStatus
		newState     BookingStatus
	}{
		{
			name:         "INIT -> AWAITING_PAYMENT",
			initialState: StatusInit,
			newState:     StatusAwaitingPayment,
		},
		{
			name:         "INIT -> CANCELLED",
			initialState: StatusInit,
			newState:     StatusCancelled,
		},
		{
			name:         "AWAITING_PAYMENT -> PAID",
			initialState: StatusAwaitingPayment,
			newState:     StatusPaid,
		},
		{
			name:         "AWAITING_PAYMENT -> CANCELLED",
			initialState: StatusAwaitingPayment,
			newState:     StatusCancelled,
		},
		{
			name:         "PAID -> CONFIRMED",
			initialState: StatusPaid,
			newState:     StatusConfirmed,
		},
		{
			name:         "PAID -> CANCELLED",
			initialState: StatusPaid,
			newState:     StatusCancelled,
		},
		{
			name:         "CONFIRMED -> COMPLETED",
			initialState: StatusConfirmed,
			newState:     StatusCompleted,
		},
		{
			name:         "CONFIRMED -> CANCELLED",
			initialState: StatusConfirmed,
			newState:     StatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			booking := &Booking{
				ID:     "test-booking",
				Status: tt.initialState,
			}

			sm := NewStateMachine(booking)
			err := sm.Transition(tt.newState)

			require.NoError(t, err)
			assert.Equal(t, tt.newState, booking.Status)
			assert.Equal(t, tt.newState, sm.GetStatus())
		})
	}
}

// TestStateMachine_Transition_InvalidTransitions tests invalid state transitions
func TestStateMachine_Transition_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name         string
		initialState BookingStatus
		newState     BookingStatus
		errorMsg     string
	}{
		{
			name:         "AWAITING_PAYMENT -> INIT (reverse)",
			initialState: StatusAwaitingPayment,
			newState:     StatusInit,
			errorMsg:     "invalid state transition",
		},
		{
			name:         "PAID -> AWAITING_PAYMENT (reverse)",
			initialState: StatusPaid,
			newState:     StatusAwaitingPayment,
			errorMsg:     "invalid state transition",
		},
		{
			name:         "CONFIRMED -> PAID (reverse)",
			initialState: StatusConfirmed,
			newState:     StatusPaid,
			errorMsg:     "invalid state transition",
		},
		{
			name:         "COMPLETED -> CONFIRMED (reverse)",
			initialState: StatusCompleted,
			newState:     StatusConfirmed,
			errorMsg:     "invalid state transition",
		},
		{
			name:         "CANCELLED -> INIT (cannot resume)",
			initialState: StatusCancelled,
			newState:     StatusInit,
			errorMsg:     "invalid state transition",
		},
		{
			name:         "COMPLETED -> CANCELLED (too late)",
			initialState: StatusCompleted,
			newState:     StatusCancelled,
			errorMsg:     "invalid state transition",
		},
		{
			name:         "INIT -> PAID (skip AWAITING_PAYMENT)",
			initialState: StatusInit,
			newState:     StatusPaid,
			errorMsg:     "invalid state transition",
		},
		{
			name:         "INIT -> CONFIRMED (multiple steps)",
			initialState: StatusInit,
			newState:     StatusConfirmed,
			errorMsg:     "invalid state transition",
		},
		{
			name:         "AWAITING_PAYMENT -> COMPLETED (skip PAID, CONFIRMED)",
			initialState: StatusAwaitingPayment,
			newState:     StatusCompleted,
			errorMsg:     "invalid state transition",
		},
		{
			name:         "INIT -> COMPLETED (skip all)",
			initialState: StatusInit,
			newState:     StatusCompleted,
			errorMsg:     "invalid state transition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			booking := &Booking{
				ID:     "test-booking",
				Status: tt.initialState,
			}

			sm := NewStateMachine(booking)
			err := sm.Transition(tt.newState)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
			assert.Equal(t, tt.initialState, booking.Status, "Status should not change on invalid transition")
		})
	}
}

// TestStateMachine_GetStatus tests GetStatus method
func TestStateMachine_GetStatus(t *testing.T) {
	tests := []struct {
		name   string
		status BookingStatus
	}{
		{name: "INIT status", status: StatusInit},
		{name: "AWAITING_PAYMENT status", status: StatusAwaitingPayment},
		{name: "PAID status", status: StatusPaid},
		{name: "CONFIRMED status", status: StatusConfirmed},
		{name: "COMPLETED status", status: StatusCompleted},
		{name: "CANCELLED status", status: StatusCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			booking := &Booking{
				ID:     "test-booking",
				Status: tt.status,
			}

			sm := NewStateMachine(booking)
			assert.Equal(t, tt.status, sm.GetStatus())
		})
	}
}

// TestStateMachine_IsFinal tests IsFinal method
func TestStateMachine_IsFinal(t *testing.T) {
	tests := []struct {
		name     string
		status   BookingStatus
		expected bool
	}{
		{
			name:     "INIT is not final",
			status:   StatusInit,
			expected: false,
		},
		{
			name:     "AWAITING_PAYMENT is not final",
			status:   StatusAwaitingPayment,
			expected: false,
		},
		{
			name:     "PAID is not final",
			status:   StatusPaid,
			expected: false,
		},
		{
			name:     "CONFIRMED is not final",
			status:   StatusConfirmed,
			expected: false,
		},
		{
			name:     "COMPLETED is final",
			status:   StatusCompleted,
			expected: true,
		},
		{
			name:     "CANCELLED is final",
			status:   StatusCancelled,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			booking := &Booking{
				ID:     "test-booking",
				Status: tt.status,
			}

			sm := NewStateMachine(booking)
			assert.Equal(t, tt.expected, sm.IsFinal())
		})
	}
}

// TestBooking_CanTransitionTo tests the CanTransitionTo method directly
func TestBooking_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name         string
		currentState BookingStatus
		targetState  BookingStatus
		expected     bool
	}{
		// Valid transitions
		{currentState: StatusInit, targetState: StatusAwaitingPayment, expected: true},
		{currentState: StatusInit, targetState: StatusCancelled, expected: true},
		{currentState: StatusAwaitingPayment, targetState: StatusPaid, expected: true},
		{currentState: StatusAwaitingPayment, targetState: StatusCancelled, expected: true},
		{currentState: StatusPaid, targetState: StatusConfirmed, expected: true},
		{currentState: StatusPaid, targetState: StatusCancelled, expected: true},
		{currentState: StatusConfirmed, targetState: StatusCompleted, expected: true},
		{currentState: StatusConfirmed, targetState: StatusCancelled, expected: true},

		// Invalid transitions
		{currentState: StatusInit, targetState: StatusPaid, expected: false},
		{currentState: StatusInit, targetState: StatusConfirmed, expected: false},
		{currentState: StatusInit, targetState: StatusCompleted, expected: false},
		{currentState: StatusAwaitingPayment, targetState: StatusInit, expected: false},
		{currentState: StatusAwaitingPayment, targetState: StatusConfirmed, expected: false},
		{currentState: StatusPaid, targetState: StatusAwaitingPayment, expected: false},
		{currentState: StatusPaid, targetState: StatusCompleted, expected: false},
		{currentState: StatusConfirmed, targetState: StatusPaid, expected: false},
		{currentState: StatusCompleted, targetState: StatusConfirmed, expected: false},
		{currentState: StatusCompleted, targetState: StatusCancelled, expected: false},
		{currentState: StatusCancelled, targetState: StatusInit, expected: false},
		{currentState: StatusCancelled, targetState: StatusAwaitingPayment, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			booking := &Booking{
				ID:     "test-booking",
				Status: tt.currentState,
			}

			result := booking.CanTransitionTo(tt.targetState)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestStateMachine_FullBookingFlow tests a complete booking lifecycle
func TestStateMachine_FullBookingFlow(t *testing.T) {
	booking := &Booking{
		ID:     "test-booking-full-flow",
		Status: StatusInit,
	}

	sm := NewStateMachine(booking)

	// Step 1: INIT -> AWAITING_PAYMENT
	err := sm.Transition(StatusAwaitingPayment)
	require.NoError(t, err)
	assert.Equal(t, StatusAwaitingPayment, sm.GetStatus())
	assert.False(t, sm.IsFinal())

	// Step 2: AWAITING_PAYMENT -> PAID
	err = sm.Transition(StatusPaid)
	require.NoError(t, err)
	assert.Equal(t, StatusPaid, sm.GetStatus())
	assert.False(t, sm.IsFinal())

	// Step 3: PAID -> CONFIRMED
	err = sm.Transition(StatusConfirmed)
	require.NoError(t, err)
	assert.Equal(t, StatusConfirmed, sm.GetStatus())
	assert.False(t, sm.IsFinal())

	// Step 4: CONFIRMED -> COMPLETED
	err = sm.Transition(StatusCompleted)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, sm.GetStatus())
	assert.True(t, sm.IsFinal())
}

// TestStateMachine_CancellationFlow tests booking cancellation at various stages
func TestStateMachine_CancellationFlow(t *testing.T) {
	cancellationStages := []struct {
		name         string
		initialState BookingStatus
	}{
		{name: "Cancel at INIT", initialState: StatusInit},
		{name: "Cancel at AWAITING_PAYMENT", initialState: StatusAwaitingPayment},
		{name: "Cancel at PAID", initialState: StatusPaid},
		{name: "Cancel at CONFIRMED", initialState: StatusConfirmed},
	}

	for _, tt := range cancellationStages {
		t.Run(tt.name, func(t *testing.T) {
			booking := &Booking{
				ID:     "test-booking-cancel",
				Status: tt.initialState,
			}

			sm := NewStateMachine(booking)

			// Cancel the booking
			err := sm.Transition(StatusCancelled)
			require.NoError(t, err)
			assert.Equal(t, StatusCancelled, sm.GetStatus())
			assert.True(t, sm.IsFinal())

			// Verify no further transitions possible
			err = sm.Transition(StatusInit)
			assert.Error(t, err)
			assert.Equal(t, StatusCancelled, sm.GetStatus())
		})
	}
}

// TestStateMachine_ConcurrentTransitions tests concurrent state transitions (safety)
func TestStateMachine_ConcurrentTransitions(t *testing.T) {
	booking := &Booking{
		ID:     "test-booking-concurrent",
		Status: StatusInit,
	}

	sm := NewStateMachine(booking)
	errChan := make(chan error, 2)

	// Attempt two concurrent transitions
	go func() {
		errChan <- sm.Transition(StatusAwaitingPayment)
	}()

	go func() {
		errChan <- sm.Transition(StatusCancelled)
	}()

	// Collect results
	err1 := <-errChan
	err2 := <-errChan

	// One should succeed, one should fail
	successCount := 0
	if err1 == nil {
		successCount++
	}
	if err2 == nil {
		successCount++
	}

	assert.Equal(t, 1, successCount, "Only one transition should succeed")
}
