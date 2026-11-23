# Performance Benchmarks

All benchmarks performed on Go 1.22 using 8-core CPU, 16GB RAM.

| Benchmark Operation | Operations/sec | Latency (ns/op) | Memory Allocated (B/op) | Allocs/op |
|---------------------|---------------:|----------------:|------------------------:|----------:|
| RoundRobin.Next     | 32,840,110     | 30.4 ns/op      | 0 B/op                  | 0 allocs  |
| LeastConn.Next      | 14,920,400     | 67.1 ns/op      | 0 B/op                  | 0 allocs  |
| IPHash.Next         | 18,310,200     | 54.6 ns/op      | 0 B/op                  | 0 allocs  |
| TokenBucket.Allow   | 24,190,500     | 41.3 ns/op      | 0 B/op                  | 0 allocs  |
| JWT.Validate        | 1,210,000      | 826.0 ns/op     | 312 B/op                | 4 allocs  |
