package admin

import (
	"bufio"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPasswordPipedInput(t *testing.T) {
	// Simulate piped input where reader already buffered stdin
	input := "user\nuser@example.com\nsecret123\nsecret123\n"
	r, w, err := os.Pipe()
	require.NoError(t, err)

	_, err = io.WriteString(w, input)
	require.NoError(t, err)
	w.Close()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	// Outer reader reads first two lines (buffering the rest)
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	require.NoError(t, err)
	assert.Equal(t, "user\n", username)

	email, err := reader.ReadString('\n')
	require.NoError(t, err)
	assert.Equal(t, "user@example.com\n", email)

	// Before fix, readPassword creates a new bufio.Reader(os.Stdin),
	// which fails with EOF because reader buffered the remaining pipe contents.
	pass, err := readPassword(reader)
	require.NoError(t, err)
	assert.Equal(t, "secret123", pass)

	passConfirm, err := readPassword(reader)
	require.NoError(t, err)
	assert.Equal(t, "secret123", passConfirm)
}

func TestReadPasswordNilReaderFallback(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	_, err = io.WriteString(w, "fallbackpassword\n")
	require.NoError(t, err)
	w.Close()

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = r

	pass, err := readPassword(nil)
	require.NoError(t, err)
	assert.Equal(t, "fallbackpassword", pass)
}
