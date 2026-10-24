package core

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/forgego/forge/api/errors"
	"github.com/forgego/forge/audit"
	"github.com/stretchr/testify/require"
)

func TestHistoryFailurePoliciesCoverCRUD(t *testing.T) {
	for _, policyCase := range []struct {
		name   string
		policy audit.Policy
	}{
		{name: "closed", policy: audit.FailClosed},
		{name: "open", policy: audit.FailOpen},
	} {
		for _, operation := range []string{"create", "update", "delete"} {
			t.Run(policyCase.name+"/"+operation, func(t *testing.T) {
				ctx := errors.WithRequestID(context.Background(), "request-123")
				counter := &audit.Counter{}
				admin, manager, id, _ := setupWriteFieldsDB(t, &Config[writeFieldsTestItem]{
					HistoryManager: failedHistoryManager{}, AuditPolicy: policyCase.policy, AuditFailures: counter,
				})
				var output bytes.Buffer
				admin.Config().AuditLogger = slog.New(slog.NewTextHandler(&output, nil))

				var err error
				switch operation {
				case "create":
					_, err = admin.CreateObject(ctx, map[string]interface{}{"name": "created"})
				case "update":
					_, err = admin.UpdateObject(ctx, id, map[string]interface{}{"name": "changed"})
				case "delete":
					err = admin.DeleteObject(ctx, id)
				}

				if policyCase.policy == audit.FailClosed {
					require.Error(t, err)
					require.Zero(t, counter.Count())
				} else {
					require.NoError(t, err)
					require.EqualValues(t, 1, counter.Count())
					require.Contains(t, output.String(), "correlation_id=request-123")
				}

				items, listErr := manager.All(ctx)
				require.NoError(t, listErr)
				switch operation {
				case "create":
					want := 1
					if policyCase.policy == audit.FailOpen {
						want = 2
					}
					require.Len(t, items, want)
				case "update":
					stored, getErr := manager.Get(ctx, id)
					require.NoError(t, getErr)
					want := "Original Name"
					if policyCase.policy == audit.FailOpen {
						want = "changed"
					}
					require.Equal(t, want, stored.Name)
				case "delete":
					_, getErr := manager.Get(ctx, id)
					if policyCase.policy == audit.FailClosed {
						require.NoError(t, getErr)
					} else {
						require.Error(t, getErr)
					}
				}
			})
		}
	}
}
