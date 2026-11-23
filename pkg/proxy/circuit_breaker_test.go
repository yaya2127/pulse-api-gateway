package proxy

import (
	"testing"
	"time"
)

func TestCircuitBreakerTransitions(t *testing.T) {
	cb := NewCircuitBreaker(3, 50*time.Millisecond)

	if !cb.Allow() {
		t.Fatal("expected circuit breaker to be closed initially")
	}

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != StateClosed {
		t.Fatalf("expected state Closed after 2 failures, got %v", cb.GetState())
	}

	cb.RecordFailure()
	if cb.GetState() != StateOpen {
		t.Fatalf("expected state Open after 3 failures, got %v", cb.GetState())
	}

	if cb.Allow() {
		t.Fatal("expected request rejected when circuit is Open")
	}

	time.Sleep(60 * time.Millisecond)
	if !cb.Allow() {
		t.Fatal("expected request permitted to probe Half-Open")
	}

	cb.RecordSuccess()
	cb.RecordSuccess()
	if cb.GetState() != StateClosed {
		t.Fatalf("expected state Closed after probationary successes, got %v", cb.GetState())
	}
}
