## Performance Issue Template

```markdown
---
name: ⚡ Performance Issue
about: Report a performance problem
title: '[PERF] '
labels: 'performance'
assignees: ''
---

## ⚡ Performance Issue

###  Performance Problem
Describe the performance issue you're experiencing.

###  Benchmarks/Metrics
Please provide specific metrics:

- **Response time**: [e.g. 2.5 seconds for search]
- **Throughput**: [e.g. 10 requests/second]
- **Memory usage**: [e.g. 2GB for 100k vectors]
- **CPU usage**: [e.g. 80% constantly]

###  Steps to Reproduce
1. Set up environment with: [describe setup]
2. Load data: [describe data size/type]
3. Execute: [specific operations]
4. Measure: [how you measured performance]

###  Dataset Information
- **Number of vectors**: [e.g. 100,000]
- **Vector dimensions**: [e.g. 768]
- **Data size**: [e.g. 500MB]
- **Concurrent requests**: [e.g. 50]

###  Environment
- **Hardware**: [e.g. 4 CPU cores, 8GB RAM, SSD]
- **OS**: [e.g. Ubuntu 20.04]
- **Deployment**: [e.g. Docker, Kubernetes, bare metal]
- **Same-Same Version**: [e.g. v1.0.0]

###  Expected Performance
What performance were you expecting?

###  Configuration
Share relevant configuration:

# docker-compose.yml resource limits
deploy:
  resources:
    limits:
      memory: 512M
      cpus: '0.5'

# Environment variables
RATE_LIMIT_REQUESTS_PER_MINUTE=60


###  Profiling Data
If you have profiling data, please include it:

go tool pprof results or other profiling output

### 💡 Potential Solutions
If you have ideas for performance improvements, share them here.

###  Additional Context
Any other information about the performance issue.
```