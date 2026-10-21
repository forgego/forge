package renderers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/forgego/forge/internal/csvsafe"
)

// CSVRenderer renders data as CSV
type CSVRenderer struct{}

// NewCSVRenderer creates a new CSV renderer
func NewCSVRenderer() *CSVRenderer {
	return &CSVRenderer{}
}

// Render renders data to CSV bytes
func (r *CSVRenderer) Render(data interface{}) ([]byte, error) {
	var b bytes.Buffer
	if err := r.RenderToWriter(&b, data); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// MediaType returns the CSV media type
func (r *CSVRenderer) MediaType() string {
	return "text/csv"
}

// RenderToWriter renders data directly to a writer
func (r *CSVRenderer) RenderToWriter(w io.Writer, data interface{}) error {
	writer := csv.NewWriter(w)
	v := reflect.ValueOf(data)
	if !v.IsValid() || (v.Kind() != reflect.Slice && v.Kind() != reflect.Array) {
		return fmt.Errorf("CSV data must be a slice or array")
	}
	if v.Len() == 0 {
		return nil
	}
	headers, err := csvHeaders(v.Index(0))
	if err != nil {
		return err
	}
	safeHeaders := make([]string, len(headers))
	for i, header := range headers {
		safeHeaders[i] = csvsafe.Cell(header, header)
	}
	if err := writer.Write(safeHeaders); err != nil {
		return err
	}
	for i := 0; i < v.Len(); i++ {
		row, err := csvRow(v.Index(i), headers)
		if err != nil {
			return err
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func csvValue(v reflect.Value) (reflect.Value, error) {
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return v, fmt.Errorf("CSV row is nil")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct && (v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String) {
		return v, fmt.Errorf("CSV rows must be structs or string-keyed maps")
	}
	return v, nil
}

func csvHeaders(v reflect.Value) ([]string, error) {
	v, err := csvValue(v)
	if err != nil {
		return nil, err
	}
	var headers []string
	if v.Kind() == reflect.Map {
		for _, k := range v.MapKeys() {
			headers = append(headers, k.String())
		}
		sort.Strings(headers)
	} else {
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
			headers = append(headers, name)
		}
	}
	return headers, nil
}

func csvRow(v reflect.Value, headers []string) ([]string, error) {
	v, err := csvValue(v)
	if err != nil {
		return nil, err
	}
	row := make([]string, len(headers))
	for i, name := range headers {
		var value reflect.Value
		if v.Kind() == reflect.Map {
			value = v.MapIndex(reflect.ValueOf(name).Convert(v.Type().Key()))
		} else {
			for j := 0; j < v.NumField(); j++ {
				field := v.Type().Field(j)
				if !field.IsExported() {
					continue
				}
				jsonName, _, _ := strings.Cut(field.Tag.Get("json"), ",")
				if jsonName == "-" {
					continue
				}
				if jsonName == "" {
					jsonName = field.Name
				}
				if jsonName == name {
					value = v.Field(j)
					break
				}
			}
		}
		if value.IsValid() {
			raw := value.Interface()
			row[i] = csvsafe.Cell(raw, fmt.Sprint(raw))
		}
	}
	return row, nil
}
