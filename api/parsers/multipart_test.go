package parsers

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type customContentTypeReader struct {
	io.Reader
	contentType string
}

func (c *customContentTypeReader) ContentType() string {
	return c.contentType
}

func TestMultiPartParser_Parse_RealMultipartBody(t *testing.T) {
	parser := NewMultiPartParser()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writer.WriteField("title", "Test Title")
	require.NoError(t, err)

	part, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("hello multipart"))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	var result map[string]interface{}
	err = parser.Parse(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "Test Title", result["title"])
	fileHeader, ok := result["file"].(*multipart.FileHeader)
	require.True(t, ok)
	assert.Equal(t, "test.txt", fileHeader.Filename)
}

func TestMultiPartParser_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		setupParser func(writer *multipart.Writer) *MultiPartParser
		buildReader func(body *bytes.Buffer, writer *multipart.Writer) io.Reader
		expectErr   bool
		errContains string
	}{
		{
			name: "boundary from parser ContentType field",
			setupParser: func(w *multipart.Writer) *MultiPartParser {
				p := NewMultiPartParser()
				p.ContentType = w.FormDataContentType()
				return p
			},
			buildReader: func(body *bytes.Buffer, _ *multipart.Writer) io.Reader {
				return body
			},
			expectErr: false,
		},
		{
			name: "boundary from parser Boundary field",
			setupParser: func(w *multipart.Writer) *MultiPartParser {
				p := NewMultiPartParser()
				p.Boundary = w.Boundary()
				return p
			},
			buildReader: func(body *bytes.Buffer, _ *multipart.Writer) io.Reader {
				return body
			},
			expectErr: false,
		},
		{
			name: "boundary from reader ContentType interface",
			setupParser: func(_ *multipart.Writer) *MultiPartParser {
				return NewMultiPartParser()
			},
			buildReader: func(body *bytes.Buffer, w *multipart.Writer) io.Reader {
				return &customContentTypeReader{
					Reader:      body,
					contentType: w.FormDataContentType(),
				}
			},
			expectErr: false,
		},
		{
			name: "missing boundary in content type",
			setupParser: func(_ *multipart.Writer) *MultiPartParser {
				p := NewMultiPartParser()
				p.ContentType = "multipart/form-data"
				return p
			},
			buildReader: func(body *bytes.Buffer, _ *multipart.Writer) io.Reader {
				return body
			},
			expectErr:   true,
			errContains: "missing boundary",
		},
		{
			name: "invalid reader stream with no boundary",
			setupParser: func(_ *multipart.Writer) *MultiPartParser {
				return NewMultiPartParser()
			},
			buildReader: func(_ *bytes.Buffer, _ *multipart.Writer) io.Reader {
				return strings.NewReader("not a multipart body")
			},
			expectErr:   true,
			errContains: "missing multipart boundary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			require.NoError(t, writer.WriteField("foo", "bar"))
			require.NoError(t, writer.Close())

			parser := tt.setupParser(writer)
			reader := tt.buildReader(body, writer)

			var result map[string]interface{}
			err := parser.Parse(reader, &result)
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "bar", result["foo"])
			}
		})
	}
}
