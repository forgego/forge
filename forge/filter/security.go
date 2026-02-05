package filter

import (
	"fmt"
	"time"
)

// SecurityConfig contains security settings
type SecurityConfig struct {
	AllowedFields   map[string][]string // Per-model, per-role
	AllowedLookups  map[string][]string // Per-field allowed lookups
	MaxJoinDepth    int
	MaxConditions   int
	MaxORBranches   int
	TimeoutDuration time.Duration
	CostThreshold   int
}

// NewSecurityConfig creates a new security config with defaults
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedFields:   make(map[string][]string),
		AllowedLookups:  make(map[string][]string),
		MaxJoinDepth:    3,
		MaxConditions:   50,
		MaxORBranches:   20,
		TimeoutDuration: 30 * time.Second,
		CostThreshold:   100,
	}
}

// ValidateComplexity validates filter complexity against security config
func (sc *SecurityConfig) ValidateComplexity(ast *FilterNode) error {
	if ast == nil {
		return nil
	}

	// Check max conditions
	conditionCount := countConditions(ast)
	if conditionCount > sc.MaxConditions {
		return fmt.Errorf("filter exceeds maximum conditions: %d > %d", conditionCount, sc.MaxConditions)
	}

	// Check max OR branches
	orBranches := countORBranches(ast)
	if orBranches > sc.MaxORBranches {
		return fmt.Errorf("filter exceeds maximum OR branches: %d > %d", orBranches, sc.MaxORBranches)
	}

	// Check max join depth
	maxDepth := getMaxRelationDepth(ast)
	if maxDepth > sc.MaxJoinDepth {
		return fmt.Errorf("filter exceeds maximum join depth: %d > %d", maxDepth, sc.MaxJoinDepth)
	}

	return nil
}

// ValidateCost validates filter cost against threshold
func (sc *SecurityConfig) ValidateCost(cost int) error {
	if cost > sc.CostThreshold {
		return fmt.Errorf("filter cost %d exceeds threshold %d", cost, sc.CostThreshold)
	}
	return nil
}

// countConditions counts the number of conditions in the AST
func countConditions(node *FilterNode) int {
	if node == nil {
		return 0
	}

	if node.Op == OpField {
		return 1
	}

	count := 0
	for _, child := range node.Children {
		count += countConditions(child)
	}

	return count
}

// countORBranches counts the number of OR branches
func countORBranches(node *FilterNode) int {
	if node == nil {
		return 0
	}

	if node.Op == OpOr {
		return len(node.Children)
	}

	count := 0
	for _, child := range node.Children {
		count += countORBranches(child)
	}

	return count
}

// getMaxRelationDepth gets the maximum relation depth
func getMaxRelationDepth(node *FilterNode) int {
	if node == nil {
		return 0
	}

	maxDepth := 0
	if node.Op == OpField && node.Field != "" {
		// Count __ in field path
		depth := countDepthInPath(node.Field)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	for _, child := range node.Children {
		depth := getMaxRelationDepth(child)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

// countDepthInPath counts the depth of relations in a field path
func countDepthInPath(path string) int {
	if path == "" {
		return 0
	}

	count := 0
	for i := 0; i < len(path)-1; i++ {
		if path[i] == '_' && path[i+1] == '_' {
			count++
			i++ // Skip next underscore
		}
	}

	return count
}

// AuditLog represents an audit log entry
type AuditLog struct {
	FilterID      string
	UserID        string
	Action        string // "execute", "save", "load"
	Cost          int
	ExecutionTime time.Duration
	Denied        bool
	Reason        string
	Parameters    map[string]interface{} // Masked
	Timestamp     time.Time
}

// AuditLogger logs filter operations
type AuditLogger interface {
	Log(entry *AuditLog) error
}

// DefaultAuditLogger is a simple audit logger
type DefaultAuditLogger struct {
	logs []*AuditLog
}

// NewDefaultAuditLogger creates a new default audit logger
func NewDefaultAuditLogger() *DefaultAuditLogger {
	return &DefaultAuditLogger{
		logs: make([]*AuditLog, 0),
	}
}

// Log logs an audit entry
func (l *DefaultAuditLogger) Log(entry *AuditLog) error {
	entry.Timestamp = time.Now()
	l.logs = append(l.logs, entry)
	return nil
}

// GetLogs returns all audit logs
func (l *DefaultAuditLogger) GetLogs() []*AuditLog {
	return l.logs
}

// MaskParameters masks sensitive parameter values
func MaskParameters(params map[string]interface{}) map[string]interface{} {
	masked := make(map[string]interface{})
	for k, v := range params {
		if v != nil {
			// Mask the value (show first/last chars only)
			masked[k] = maskValue(v)
		} else {
			masked[k] = nil
		}
	}
	return masked
}

// maskValue masks a value for logging
func maskValue(v interface{}) string {
	s := fmt.Sprintf("%v", v)
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}
