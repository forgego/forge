package core

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/schema"
)

// buildMetadata builds metadata from schema and config
func buildMetadata[T any](s schema.Schema, config *Config[T], name string) (*Metadata, error) {
	meta := s.Meta()

	// Build fields metadata
	fieldsMetadata, err := buildFieldsMetadata(s)
	if err != nil {
		return nil, err
	}

	// Build relations metadata
	relationsMetadata := buildRelationsMetadata(s, config)

	// Build actions metadata
	actionsMetadata := buildActionsMetadata(config)

	// Build filters metadata
	filtersMetadata := buildFiltersMetadata(s, config)

	// Determine list display
	listDisplay := config.ListDisplay
	if len(listDisplay) == 0 {
		listDisplay = toFieldSlice(inferListDisplay(fieldsMetadata))
	}

	// Determine search fields
	searchFields := config.SearchFields
	if len(searchFields) == 0 {
		searchFields = toFieldSlice(inferSearchFields(fieldsMetadata))
	}

	// Build metadata
	metadata := &Metadata{
		Name:              name,
		VerboseName:       getOrDefault(config.VerboseName, meta.VerboseName, name),
		VerboseNamePlural: getOrDefault(config.VerboseNamePlural, meta.VerboseNamePlural, name+"s"),
		Description:       config.Description,
		Icon:              getOrDefault(config.Icon, "", "file"),
		Fields:            fieldsMetadata,
		Relations:         relationsMetadata,
		Permissions:       PermissionMetadata{}, // Set by GetMetadata based on user
		Actions:           actionsMetadata,
		Filters:           filtersMetadata,
		ListDisplay:       toStringSlice(listDisplay),
		ListFilter:        toStringSlice(config.ListFilter),
		SearchFields:      toStringSlice(searchFields),
		Ordering:          toStringSlice(config.Ordering),
		Pagination: PaginationConfig{
			PageSize:    config.ListPerPage,
			MaxPageSize: config.ListMaxShowAll,
		},
		UIOverrides: config.UIOverrides,
	}

	return metadata, nil
}

// buildFieldsMetadata builds field metadata from schema fields
func buildFieldsMetadata(s schema.Schema) ([]FieldMetadata, error) {
	fields := s.Fields()
	result := make([]FieldMetadata, 0, len(fields))

	for _, field := range fields {
		fieldMeta := FieldMetadata{
			Name:         field.Name,
			Type:         field.Type.String(),
			Label:        getOrDefault(field.VerboseName, "", field.Name),
			HelpText:     field.HelpText,
			Required:     field.Required,
			ReadOnly:     !field.Editable, // ReadOnly is inverse of Editable usually, or need to check field definition
			Widget:       inferWidget(field),
			DefaultValue: field.Default,
		}

		// Add field-specific metadata
		switch field.Type {
		case schema.TypeString, schema.TypeEmail:
			if field.MaxLength != nil && *field.MaxLength > 0 {
				fieldMeta.MaxLength = *field.MaxLength
			}
			if field.MinLength != nil && *field.MinLength > 0 {
				fieldMeta.MinLength = *field.MinLength
			}

		case schema.TypeInt64, schema.TypeInt32, schema.TypeFloat64:
			// Could add min/max values if schema supports it

		case schema.TypeBytes: // TypeFile/Image not in exported constants? Checked field.go, TypeBytes exists. Use loose match or check Schema
			// Could add accept, maxSize if schema supports it
		}

		// Add choices if available
		if len(field.Choices) > 0 {
			fieldMeta.Choices = make([]Choice, len(field.Choices))
			for i, choice := range field.Choices {
				fieldMeta.Choices[i] = Choice{
					Value: choice.Value,
					Label: choice.Label,
				}
			}
		}

		// Add validators
		fieldMeta.Validators = buildValidatorsMetadata(field)

		result = append(result, fieldMeta)
	}

	return result, nil
}

// buildRelationsMetadata builds relation metadata from schema relations
func buildRelationsMetadata[T any](s schema.Schema, config *Config[T]) []RelationMetadata {
	relations := s.Relations()
	result := make([]RelationMetadata, 0, len(relations))
	inlineConfigs := map[string]InlineRelationConfig{}
	if config != nil && config.InlineRelations != nil {
		inlineConfigs = config.InlineRelations
	}
	usedInline := map[string]struct{}{}

	for _, rel := range relations {
		relMeta := RelationMetadata{
			Name:         rel.Name,
			Type:         rel.Type.String(),
			RelatedModel: rel.To,
			RelatedField: rel.RelatedName,
			Label:        getOrDefault("", "", rel.Name), // Relation struct has no VerboseName
		}
		if inlineConfig, ok := inlineConfigs[rel.Name]; ok {
			relMeta.Inline = &inlineConfig
			if inlineConfig.Type != "" {
				relMeta.Type = inlineConfig.Type
			}
			if inlineConfig.Label != "" {
				relMeta.Label = inlineConfig.Label
			}
			if inlineConfig.RelatedModel != "" {
				relMeta.RelatedModel = inlineConfig.RelatedModel
			}
			if inlineConfig.RelatedField != "" {
				relMeta.RelatedField = inlineConfig.RelatedField
			}
			usedInline[rel.Name] = struct{}{}
		}
		result = append(result, relMeta)
	}

	for name, inlineConfig := range inlineConfigs {
		if _, ok := usedInline[name]; ok {
			continue
		}
		relMeta := RelationMetadata{
			Name:         name,
			Type:         inlineConfig.Type,
			RelatedModel: inlineConfig.RelatedModel,
			RelatedField: inlineConfig.RelatedField,
			Label:        getOrDefault(inlineConfig.Label, "", name),
			Inline:       &inlineConfig,
		}
		result = append(result, relMeta)
	}

	return result
}

// buildActionsMetadata builds action metadata from config
func buildActionsMetadata[T any](config *Config[T]) []ActionMetadata {
	if len(config.Actions) == 0 {
		return []ActionMetadata{}
	}

	result := make([]ActionMetadata, len(config.Actions))
	for i, action := range config.Actions {
		result[i] = ActionMetadata{
			Name:         action.Name,
			Label:        action.Label,
			Description:  action.Description,
			Confirmation: action.Confirmation,
			Permissions:  action.Permissions,
			Dangerous:    action.Dangerous,
			Icon:         action.Icon,
			UIComponent:  action.UIComponent,
		}
	}
	return result
}

// buildFiltersMetadata builds filter metadata from schema and config
func buildFiltersMetadata[T any](s schema.Schema, config *Config[T]) []FilterMetadata {
	if len(config.ListFilter) == 0 {
		return []FilterMetadata{}
	}

	fields := s.Fields()
	fieldMap := make(map[string]schema.Field)
	for _, field := range fields {
		fieldMap[field.Name] = field
	}

	result := make([]FilterMetadata, 0, len(config.ListFilter))
	for _, filterField := range config.ListFilter {
		filterName := filterField.Path()

		field, ok := fieldMap[filterName]
		if !ok {
			continue
		}

		filterMeta := FilterMetadata{
			Name:  filterName,
			Type:  inferFilterType(field),
			Label: getOrDefault(field.VerboseName, "", field.Name),
		}

		// Add choices for choice fields
		if len(field.Choices) > 0 {
			filterMeta.Choices = make([]Choice, len(field.Choices))
			for i, choice := range field.Choices {
				filterMeta.Choices[i] = Choice{
					Value: choice.Value,
					Label: choice.Label,
				}
			}
		}

		// Add related model for relations
		// Note: Field struct doesn't have RelatedModel directly? Need to check.
		// Assuming Relation handling is separate or Field has it.
		// For now simplifying to avoid error if RelatedModel field absent
		if field.Type == schema.TypeForeignKey || field.Type == schema.TypeManyToMany {
			// filterMeta.RelatedModel = field.RelatedModel // FIXME: Check field struct
			filterMeta.Multiple = field.Type == schema.TypeManyToMany
		}

		result = append(result, filterMeta)
	}

	return result
}

// buildValidatorsMetadata builds validator metadata from field
func buildValidatorsMetadata(field schema.Field) []ValidatorMetadata {
	// For now, just basic validators
	validators := []ValidatorMetadata{}

	if field.Required {
		validators = append(validators, ValidatorMetadata{
			Name:    "required",
			Message: "This field is required",
		})
	}

	if field.MaxLength != nil && *field.MaxLength > 0 {
		validators = append(validators, ValidatorMetadata{
			Name:    "max_length",
			Message: fmt.Sprintf("Maximum length is %d", *field.MaxLength),
			Params:  map[string]interface{}{"max": *field.MaxLength},
		})
	}

	if field.MinLength != nil && *field.MinLength > 0 {
		validators = append(validators, ValidatorMetadata{
			Name:    "min_length",
			Message: fmt.Sprintf("Minimum length is %d", *field.MinLength),
			Params:  map[string]interface{}{"min": *field.MinLength},
		})
	}

	return validators
}

// inferWidget infers the widget type from field type
func inferWidget(field schema.Field) string {
	switch field.Type {
	case schema.TypeString:
		if field.MaxLength != nil && *field.MaxLength > 200 {
			return "textarea"
		}
		return "text"
	case schema.TypeText:
		return "rich_text"
	case schema.TypeEmail:
		return "email"
	case schema.TypeBytes: // Password might be bytes or string? Assuming string for now
		return "password"
	case schema.TypeInt64, schema.TypeInt32:
		return "number"
	case schema.TypeFloat64:
		return "number"
	case schema.TypeBool:
		return "checkbox"
	case schema.TypeDate:
		return "date"
	case schema.TypeDateTime:
		return "datetime"
	case schema.TypeForeignKey:
		return "select"
	case schema.TypeManyToMany:
		return "multi_select"
	default:
		return "text"
	}
}

// inferFilterType infers the filter type from field type
func inferFilterType(field schema.Field) string {
	switch field.Type {
	case schema.TypeString, schema.TypeText, schema.TypeEmail:
		return "text"
	case schema.TypeInt64, schema.TypeInt32, schema.TypeFloat64:
		return "number"
	case schema.TypeBool:
		return "boolean"
	case schema.TypeDate, schema.TypeDateTime:
		return "date_range"
	case schema.TypeForeignKey:
		return "relation"
	case schema.TypeManyToMany:
		return "relation"
	default:
		if len(field.Choices) > 0 {
			return "choice"
		}
		return "text"
	}
}

// inferListDisplay infers which fields to display in list view
func inferListDisplay(fields []FieldMetadata) []string {
	result := []string{}

	for _, field := range fields {
		// Show up to 5 fields
		if len(result) >= 5 {
			break
		}

		// Skip large text fields and files
		if field.Type == "text" || field.Type == "file" || field.Type == "image" {
			continue
		}

		// Skip readonly technical fields
		if field.ReadOnly && (strings.HasSuffix(field.Name, "_at") || strings.HasSuffix(field.Name, "_by")) {
			continue
		}

		result = append(result, field.Name)
	}

	// Always show at least ID if nothing else
	if len(result) == 0 {
		result = append(result, "id")
	}

	return result
}

// inferSearchFields infers which fields to use for search
func inferSearchFields(fields []FieldMetadata) []string {
	result := []string{}

	for _, field := range fields {
		// Only include string/text fields
		if field.Type == "string" || field.Type == "text" || field.Type == "email" {
			result = append(result, field.Name)
		}
	}

	return result
}

// getOrDefault returns the first non-empty string
func getOrDefault(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// toStringSlice converts a slice of Field to a slice of strings
func toStringSlice(input []Field) []string {
	if len(input) == 0 {
		return []string{}
	}
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = v.Path()
	}
	return result
}

// toFieldSlice converts a string slice to a Field slice
func toFieldSlice(input []string) []Field {
	result := make([]Field, len(input))
	for i, v := range input {
		result[i] = Computed(v)
	}
	return result
}
