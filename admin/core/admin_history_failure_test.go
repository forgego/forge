package core

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/forgego/forge/audit"
	"github.com/stretchr/testify/require"
)

type failedHistoryManager struct{}

func (failedHistoryManager) LogAction(context.Context, LogEntry) error {
	return errors.New("history unavailable")
}
func (failedHistoryManager) GetHistory(context.Context, string, string) ([]LogEntry, error) {
	return nil, nil
}

func TestHistoryFailurePolicy(t *testing.T) {
	admin, err := NewAdmin[privateLabelModel](privateLabelModel{}, nil, &Config[privateLabelModel]{HistoryManager: failedHistoryManager{}})
	require.NoError(t, err)
	ctx := context.Background()
	require.Error(t, admin.writeHistory(ctx, nil, "1", "item", ActionAdd, "{}"), "fail-closed is the default")

	var output bytes.Buffer
	counter := &audit.Counter{}
	admin.Config().AuditPolicy = audit.FailOpen
	admin.Config().AuditLogger = slog.New(slog.NewTextHandler(&output, nil))
	admin.Config().AuditFailures = counter
	require.NoError(t, admin.writeHistory(ctx, nil, "1", "item", ActionAdd, "{}"))
	require.EqualValues(t, 1, counter.Count())
	require.Contains(t, output.String(), "correlation_id")
	require.True(t, strings.Contains(output.String(), "admin-"))
}
