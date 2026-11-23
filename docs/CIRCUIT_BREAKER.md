# Circuit Breaker Specification

Prevents cascading downstream service failures:
- **CLOSED**: Requests flow normally. Consecutive failures increment error counter.
- **OPEN**: Trips after threshold (default: 5) failures. Rejects incoming traffic with 503.
- **HALF-OPEN**: After cooldown timeout, probationary requests probe upstream health.
