package exporters

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteWriterDropsBufferAfterSendFailure(t *testing.T) {
	server := httptest.NewServer(nil)
	server.Close()
	writer := &remoteWriter{exporter: NewRemoteExporter(RemoteConfig{URL: server.URL}, 0)}

	for range 1000 {
		_, err := writer.Write([]byte("entry\n"))
		require.NoError(t, err)
		require.Empty(t, writer.buffer)
	}
}
