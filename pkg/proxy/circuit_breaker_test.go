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

func TestCircuitBreakerConcurrency(t *testing.T) {
	cb := NewCircuitBreaker(10, 100*time.Millisecond)
	done := make(chan bool)

	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				cb.Allow()
				cb.RecordSuccess()
			}
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}

	if cb.GetState() != StateClosed {
		t.Errorf("expected circuit to remain Closed, got %v", cb.GetState())
	}
}
