# gRPC-Web Proxying & Distributed Sliding-Window Rate Limiting

Pulse API Gateway v1.2.0 introduces:
* **gRPC-Web Protocol Translation**: Transparent binary framing and trailer streaming for browser frontend clients.
* **Sliding-Window Rate Limiting**: Sub-millisecond evaluation with cluster backplane sync.
* **Speculative Request Hedging**: Automatic concurrent upstream dispatch to squash p99/p99.9 latency spikes.
* **WebAssembly (WASM) Filter Hooks**: Zero-overhead request/response header interception.
