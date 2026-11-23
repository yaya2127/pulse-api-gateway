# Token Bucket Rate Limiting

Pulse employs a thread-safe token bucket algorithm with floating-point elapsed time accumulation:

$$\text{tokens} = \min(\text{capacity}, \text{tokens} + \Delta t \times \text{refillRate})$$

### HTTP Response Headers
- `X-RateLimit-Limit`: Maximum burst capacity.
- `X-RateLimit-Remaining`: Tokens currently remaining in client's bucket.
- `Retry-After`: Seconds until at least 1 token is refilled upon 429 status.
