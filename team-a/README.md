# Swarm 
Swarm is a Go-based load testing tool designed to simplify performance testing for developers and engineers. Our primary goal is to reduce the learning curve typically associated with load testing tools by providing YAML configuration support and an intuitive command-line interface.

### objective
1. CLI implementation
2. YAML configuration support
3. Multi path load testing

## Input/Output
- Input
  - Target URL
  - Duration
  - Users
- Output
  - Success Rate(Percentage of successful HTTP responses (200-300 status codes))
  - Failure Rate
    - if not responded to within 1 second

## Installation
```sh
```

## How to Use 
```sh
swarm --users 10 --duration 60 -H http://localhost:8080
swarm --config=./config.yaml
```

## YAML Configuration 
```yaml

```

## Reference

* [grafana/k6](https://github.com/grafana/k6)
* [spf13/cobra](https://github.com/spf13/cobra)
* [tsenart/vegeta](https://github.com/tsenart/vegeta)
* [valyala/fasthttp](https://github.com/valyala/fasthttp)
