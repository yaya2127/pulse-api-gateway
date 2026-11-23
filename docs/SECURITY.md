# Security Policy & Implementation

- **Zero External Dependencies**: Pure Go standard library minimizes attack surface.
- **Constant-Time Verification**: HMAC signatures checked with `crypto/hmac.Equal` to prevent timing attacks.
- **Panic Isolation**: All panics safely trapped with stack traces stripped from client responses.
