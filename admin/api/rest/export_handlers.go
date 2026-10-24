package rest

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/forgego/forge/admin/core"
	apicore "github.com/forgego/forge/api/core"
	"github.com/forgego/forge/internal/csvsafe"
)

// handleExport returns list results as CSV or JSON using current filters/search.
func (r *Router) handleExport(admin core.AdminInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user, _ := apicore.UserFromContext(ctx)

		if !admin.HasViewPermission(ctx, user, nil) {
			respondError(w, http.StatusForbidden, "permission_denied", "You don't have permission to export this model", nil)
			return
		}

		format := strings.ToLower(req.URL.Query().Get("format"))
		if format == "" {
			format = "json"
		}
		if format != "json" && format != "csv" {
			respondError(w, http.StatusBadRequest, "invalid_format", "Unsupported export format", nil)
			return
		}

		meta, err := admin.GetMetadata(ctx, user)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		listFields, labelMap := exportListFields(meta)
		hidden := make(map[string]bool)
		for _, name := range hiddenResponseFields(admin) {
			hidden[name] = true
		}
		visibleFields := make([]string, 0, len(listFields))
		for _, name := range listFields {
			if !hidden[name] {
				visibleFields = append(visibleFields, name)
			}
		}
		listFields = visibleFields
		if len(listFields) == 0 {
			respondError(w, http.StatusBadRequest, "export_failed", "No fields available for export", nil)
			return
		}

		params := parseListParams(req)
		maxPageSize := meta.Pagination.MaxPageSize
		if maxPageSize <= 0 {
			maxPageSize = 100
		}
		params.PageSize = maxPageSize
		params.Page = 1

		results := make([]map[string]interface{}, 0)
		for {
			response, err := admin.ListObjects(ctx, params)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "export_failed", err.Error(), nil)
				return
			}

			if response.Count > core.MaxExportRows {
				respondError(w, http.StatusBadRequest, "export_too_large",
					fmt.Sprintf("export exceeds %d rows; narrow the filters", core.MaxExportRows),
					map[string]interface{}{"count": response.Count, "limit": core.MaxExportRows})
				return
			}

			items := interfaceSlice(response.Results)
			if len(items) == 0 {
				break
			}

			for _, item := range visibleAdminRows(ctx, admin, user, items) {
				rowMap := make(map[string]interface{}, len(listFields))
				itemMap := objectToMap(item)
				for _, field := range listFields {
					rowMap[field] = getMapValue(itemMap, field)
				}
				results = append(results, rowMap)
				if len(results) > core.MaxExportRows {
					respondError(w, http.StatusBadRequest, "export_too_large",
						fmt.Sprintf("export exceeds %d rows; narrow the filters", core.MaxExportRows),
						map[string]interface{}{"count": int64(len(results)), "limit": core.MaxExportRows})
					return
				}
			}

			if int64(params.Page)*int64(params.PageSize) >= response.Count || len(items) < params.PageSize {
				break
			}
			params.Page++
		}

		switch format {
		case "csv":
			if err := writeCSVExport(w, admin.ModelName(), listFields, labelMap, results); err != nil {
				respondError(w, http.StatusInternalServerError, "export_failed", err.Error(), nil)
			}
		case "json":
			respondJSON(w, http.StatusOK, results)
		}
	}
}

func exportListFields(meta *core.Metadata) ([]string, map[string]string) {
	listFields := meta.ListDisplay
	if len(listFields) == 0 {
		listFields = make([]string, 0, len(meta.Fields))
		for _, field := range meta.Fields {
			listFields = append(listFields, field.Name)
		}
	}

	labelMap := make(map[string]string, len(meta.Fields))
	for _, field := range meta.Fields {
		labelMap[field.Name] = field.Label
	}

	return listFields, labelMap
}

func interfaceSlice(results interface{}) []interface{} {
	if results == nil {
		return nil
	}
	val := reflect.ValueOf(results)
	if val.Kind() != reflect.Slice {
		return nil
	}
	items := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		items[i] = val.Index(i).Interface()
	}
	return items
}

func objectToMap(obj interface{}) map[string]interface{} {
	if obj == nil {
		return map[string]interface{}{}
	}
	if mapObj, ok := obj.(map[string]interface{}); ok {
		return mapObj
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return map[string]interface{}{}
	}
	var result map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return map[string]interface{}{}
	}
	return result
}

func getMapValue(data map[string]interface{}, field string) interface{} {
	if data == nil {
		return nil
	}
	parts := strings.Split(field, ".")
	var current interface{} = data
	for _, part := range parts {
		next, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current, ok = next[part]
		if !ok {
			return nil
		}
	}
	return current
}

func writeCSVExport(
	w http.ResponseWriter,
	modelName string,
	fields []string,
	labelMap map[string]string,
	rows []map[string]interface{},
) error {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	headers := make([]string, len(fields))
	for i, field := range fields {
		if label, ok := labelMap[field]; ok && label != "" {
			headers[i] = label
		} else {
			headers[i] = field
		}
	}

	for i, header := range headers {
		headers[i] = csvsafe.Cell(header, header)
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, row := range rows {
		record := make([]string, len(fields))
		for i, field := range fields {
			record[i] = csvsafe.Cell(row[field], formatExportValue(row[field]))
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s-export.csv", modelName)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(buf.Bytes())
	return err
}

func formatExportValue(value interface{}) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case time.Time:
		return typed.Format(time.RFC3339)
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(value)
	}
}
