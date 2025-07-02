# Swarm 
Swarm is a Go-based load testing tool designed to simplify performance testing for developers and engineers. Our primary goal is to reduce the learning curve typically associated with load testing tools by providing YAML configuration support and an intuitive command-line interface.

### objective
1. CLI implementation
2. YAML configuration support
3. Multi path load testing
4. Time-based and percentile analysis of results

## Input/Output

- Input
  - Target URL
  - Duration
  - Users
  - (Optional) YAML configuration file
- Output
  - Success Rate(Percentage of successful HTTP responses (200-300 status codes))
  - Failure Rate
    - if not responded to within 1 second
  - Time-based statistics (requests, success, fail, latency per interval)
  - Percentile latency (e.g., 50th, 90th, 99th percentile)
  - RPS (Requests Per Second) trend

## Installation

```sh
```

## How to Use

### Basic CLI Usage

```sh
swarm --users 10 --duration 60 -H http://localhost:8080
swarm --config=testdata/config.yaml
```

### Analysis Command

After running a test, you can analyze the results (assuming results are saved as `results.json`):

```sh
swarm analysis --input results.json
```

This will print a detailed analysis including total requests, success/failure rates, latency percentiles, and time-based statistics.

## YAML Configuration 
```yaml
host: http://localhost:8080
duration: 1h # example 1m, 10s
users: 50
paths:
  - path: /test1
    ratio: 30
  - path: /test2
    ratio: 70
```

## Output Example

```
📊 Load Test Analysis
────────────────────────────
Total Requests:        1000
Successful (2xx):      950 (95.00%)
Failed:                50 (5.00%)
Average Latency:       120 ms
P50 Latency:           100 ms
P90 Latency:           200 ms
P99 Latency:           400 ms

📈 Time-based Analysis
────────────────────────────
Total Duration: 1m0s
Average RPS: 16.67

Time-based Statistics:
Timestamp    Requests  Success  Fail  Avg Latency  Min Latency  Max Latency
12:00:00     20        19       1     110 ms       90 ms        150 ms
...

RPS Trend:
12:00:00 | █████████████████████ 20
...
```

## Reference

* [grafana/k6](https://github.com/grafana/k6)
* [spf13/cobra](https://github.com/spf13/cobra)
* [spf13/viper](https://github.com/spf13/viper)
* [tsenart/vegeta](https://github.com/tsenart/vegeta)
* [valyala/fasthttp](https://github.com/valyala/fasthttp)
