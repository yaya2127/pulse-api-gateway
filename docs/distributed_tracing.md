# W3C Distributed Tracing Specification

Pulse API Gateway implements standard W3C TraceContext propagation.

## Headers
* `traceparent`: `00-{trace-id}-{span-id}-{flags}`
* `tracestate`: Vendor-specific state propagation
