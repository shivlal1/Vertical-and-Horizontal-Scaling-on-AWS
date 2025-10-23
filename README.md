# Horizontal Scaling Implementation Report

## 1. How the System Solved the Part II Bottleneck

The bounded search implementation creates a CPU-bound bottleneck - each request examines exactly 100 products, creating consistent computational load. With a single instance, this bottleneck limits throughput. 

Horizontal scaling solves this by:
- **Distributing requests** across multiple instances (2-4 tasks)
- **Independent processing** - each task handles its share of requests


## 2. Role of Each Component

### Application Load Balancer (ALB)
- Entry point for all traffic
- Distributes requests evenly across healthy tasks
- Performs health checks every 30 seconds
- Automatically stops sending traffic to unhealthy tasks

### Target Group
- Registry of available ECS tasks
- Tracks which tasks are healthy via health checks
- Provides the mapping between ALB and ECS tasks
- Uses IP targeting for Fargate compatibility

### Auto Scaling
- Monitors average CPU across all tasks
- Adds tasks when CPU > 70% (scale-out)
- Removes tasks when CPU < 70% (scale-in)
- Enforces min=2, max=4 task boundaries
- 300-second cooldown prevents flapping

## 3. Trade-offs: Horizontal vs Vertical Scaling

### Horizontal Scaling (Current Approach)

**Advantages:**
-  No downtime during scaling
-  Better fault tolerance (one task fails, others continue)
-  Cost-effective (pay for what you use)
-  Nearly unlimited scaling potential

**Disadvantages:**
-  More complex architecture (ALB, auto-scaling)
-  Requires stateless applications
-  Network overhead between components

### Vertical Scaling (Bigger Instances)

**Advantages:**
-  Simpler architecture (single instance)
-  No network overhead
-  Works with stateful applications

**Disadvantages:**
-  Requires downtime to resize
-  Single point of failure
-  Hard scaling limit (largest instance type)
-  Less cost-effective (paying for peak capacity always)

## 4. Experiments and Analysis of Different Load Patterns

## Response Time Comparison - Without Scaling 

| Users | Median (ms) | Average (ms) | 95%ile (ms) | 99%ile (ms) |
|-------|-------------|--------------|-------------|-------------|
| 5     | 1           | 2.76         | 3           | 57          |
| 20    | 3           | 9.65         | 59          | 86          |

(The table is based on values in the Part II /screenshots folder )

### Response Time Increase
- **Median**: 3x increase (1ms → 3ms)
- **Average**: 3.5x increase (2.76ms → 9.65ms)
- **95th percentile**: 19.7x increase (3ms → 59ms)
- **99th percentile**: 1.5x increase (57ms → 86ms)

As we increase the load, the CPU utilisation increases and the performance begins to degrade.  


## Response Time Comparison - with Scaling

| Configuration | Users | Median (ms) | Average (ms) | 95%ile (ms) | 99%ile (ms) | RPS |
|---------------|-------|-------------|--------------|-------------|-------------|------|
| 5 users | 5 | 2 | 1.99 | 3 | 8 | 2,228 |
| 20 users | 20 | 3 | 5.13 | 9 | 70 | 3,621 |

(The table is based on values in the Part III /Screenshots/5-users and 20users folder )


### Response Time Improvements at 20 Users
- **Average**: 9.65ms → 5.13ms (47% faster)
- **95th percentile**: 59ms → 9ms (85% faster)
- **99th percentile**: 86ms → 70ms (19% faster)

## Analysis

1. **Scaling significantly improves throughput** - especially under higher load 
2. **Response times improve dramatically** - 95th percentile drops from 59ms to 9ms with scaling
3. **System scales efficiently** - the ALB distributes load effectively across multiple tasks
4. **Higher concurrency benefits more** - 20 users see much larger improvements than 5 users

## Load Test Analysis - 50 Users with Modified Scaling

### Test Configuration
- **Users**: 50 concurrent users
- **Scaling Parameters**:
  - Max tasks: 6 (increased from 4)
  - CPU target: 50% (reduced from 70%)
  - Cooldown: 150s (reduced from 300s)

### Performance Results

| Metric | Value |
|--------|--------|
| Median Response Time | 6ms |
| Average Response Time | 9.87ms |
| 95th Percentile | 28ms |
| 99th Percentile | 84ms |
| Max Response Time | 333ms |

(The table is based on values in the Part III/Screenshots/50-users folder )

### Analysis

### 1. Performance Under Heavy Load
- System maintains excellent response times with 50 concurrent users
- Median of 6ms demonstrates consistent fast performance
- 95th percentile under 30ms shows reliable service for most requests

### 2. Impact of Modified Scaling Parameters
- **Lower CPU threshold (50%)**: Triggers scaling earlier, preventing performance degradation
- **Shorter cooldown (150s)**: Enables faster response to traffic fluctuations
- **Higher max capacity (6 tasks)**: Provides additional headroom for peak loads

### 3. Scaling Effectiveness
The system successfully handled 2.5x more users compared to the 20-user test while maintaining:
- Low response times (6ms median vs 3ms with 20 users)

## Conclusion
The modified scaling configuration effectively handles high concurrent load. The more aggressive scaling parameters (lower CPU target, shorter cooldown) ensure the system scales proactively rather than reactively, maintaining consistent performance even under heavy load.

## Resilience Experiment - Task Recovery Testing

### Experiment Description
Testing ECS service resilience by manually stopping tasks through the AWS console to observe auto-recovery behavior.
(The table is based on values in the Part III/Screenshots/otherExperiments folder )


### Test Scenarios and Results

### Scenario 1: Complete Task Termination
- **Action**: Stopped all 6 running tasks manually from UI
- **Result**: All 6 tasks respawned instantly
- **Recovery Time**: Near-instantaneous

### Scenario 2: Partial Task Termination
- **Action**: Stopped 3 out of 6 running tasks
- **Result**: Exactly 3 new tasks spawned to maintain desired count
- **Recovery Time**: Near-instantaneous

### Key Findings

1. **ECS Service Self-Healing**: The ECS service continuously monitors task count and immediately replaces any terminated tasks to maintain the desired count

2. **Precise Recovery**: The system spawns exactly the number of tasks needed - no more, no less

3. **No Cooldown for Recovery**: Unlike scaling actions, task recovery is not subject to the 150-second cooldown period

4. **High Availability**: This demonstrates that the service can quickly recover from failures, ensuring minimal downtime

### Technical Explanation

This behavior occurs because:
- ECS service has a `desired_count` parameter (set by auto-scaling)
- ECS continuously reconciles actual vs desired state
- When tasks are stopped/fail, ECS immediately launches replacements
- This is separate from auto-scaling - it's the core ECS service functionality

### Implications for Production

- **Fault Tolerance**: System can handle unexpected task failures
- **Zero Downtime Deployments**: Tasks can be replaced one-by-one
- **Resilience**: Even catastrophic failures (all tasks down) recover automatically
- **No Manual Intervention**: Self-healing reduces operational overhead
