package filter

import (
	"fmt"
	"sync"
	"time"
)

// Metrics tracks filter system metrics
type Metrics struct {
	executions      int64
	executionTime   []time.Duration
	denials         int64
	savedCount      int64
	costDistribution []int
	mu              sync.RWMutex
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		executionTime:   make([]time.Duration, 0),
		costDistribution: make([]int, 0),
	}
}

// RecordExecution records a filter execution
func (m *Metrics) RecordExecution(duration time.Duration, cost int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.executions++
	m.executionTime = append(m.executionTime, duration)
	m.costDistribution = append(m.costDistribution, cost)

	// Keep only last 1000 entries
	if len(m.executionTime) > 1000 {
		m.executionTime = m.executionTime[len(m.executionTime)-1000:]
	}
	if len(m.costDistribution) > 1000 {
		m.costDistribution = m.costDistribution[len(m.costDistribution)-1000:]
	}
}

// RecordDenial records a filter denial
func (m *Metrics) RecordDenial() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.denials++
}

// RecordSavedFilter records a saved filter usage
func (m *Metrics) RecordSavedFilter() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.savedCount++
}

// GetStats returns current statistics
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgTime := time.Duration(0)
	if len(m.executionTime) > 0 {
		total := time.Duration(0)
		for _, d := range m.executionTime {
			total += d
		}
		avgTime = total / time.Duration(len(m.executionTime))
	}

	avgCost := 0
	if len(m.costDistribution) > 0 {
		total := 0
		for _, c := range m.costDistribution {
			total += c
		}
		avgCost = total / len(m.costDistribution)
	}

	return map[string]interface{}{
		"executions":        m.executions,
		"average_time":       avgTime.String(),
		"denials":           m.denials,
		"saved_count":       m.savedCount,
		"average_cost":      avgCost,
		"execution_samples": len(m.executionTime),
	}
}

// AlertChecker checks for alert conditions
type AlertChecker struct {
	slowQueryThreshold time.Duration
	highDenialThreshold int64
}

// NewAlertChecker creates a new alert checker
func NewAlertChecker(slowQueryThreshold time.Duration, highDenialThreshold int64) *AlertChecker {
	return &AlertChecker{
		slowQueryThreshold:  slowQueryThreshold,
		highDenialThreshold: highDenialThreshold,
	}
}

// CheckSlowQueries checks for slow queries
func (ac *AlertChecker) CheckSlowQueries(metrics *Metrics) []string {
	alerts := make([]string, 0)

	stats := metrics.GetStats()
	avgTimeStr := stats["average_time"].(string)
	avgTime, _ := time.ParseDuration(avgTimeStr)

	if avgTime > ac.slowQueryThreshold {
		alerts = append(alerts, fmt.Sprintf("Average filter execution time %v exceeds threshold %v", avgTime, ac.slowQueryThreshold))
	}

	return alerts
}

// CheckDenials checks for high denial rates
func (ac *AlertChecker) CheckDenials(metrics *Metrics) []string {
	alerts := make([]string, 0)

	stats := metrics.GetStats()
	denials := stats["denials"].(int64)

	if denials > ac.highDenialThreshold {
		alerts = append(alerts, fmt.Sprintf("Filter denials %d exceed threshold %d", denials, ac.highDenialThreshold))
	}

	return alerts
}
