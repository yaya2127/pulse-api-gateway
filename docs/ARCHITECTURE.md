# Pulse API Gateway Architecture

Pulse API Gateway is designed for sub-millisecond reverse proxying, dynamic load balancing, and edge policy enforcement.

```
Client Requests
      |
      v
+-------------------------------------------------------------+
|                     Pulse API Gateway                       |
|                                                             |
|  [ Middleware Pipeline ]                                    |
|    1. Request ID Injection (UUID / hex)                     |
|    2. Panic Recovery & Sanitization                         |
|    3. CORS Headers & Preflight                              |
|    4. Access Logger & Telemetry                             |
|    5. Token-Bucket Rate Limiter (per client IP)             |
|    6. JWT HMAC-SHA256 Auth & Claims Extraction              |
|                                                             |
|  [ Proxy Core ]                                             |
|    * Circuit Breaker (Closed -> Open -> Half-Open)           |
|    * Balancer: Round-Robin / Least-Conn / IP-Hash           |
+-------------------------------------------------------------+
      |                     |                     |
      v                     v                     v
[ Upstream 1 ]        [ Upstream 2 ]        [ Upstream 3 ]
```
