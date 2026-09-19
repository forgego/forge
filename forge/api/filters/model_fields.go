package filters

import "github.com/forgego/forge/orm"

type modelSchemaProvider interface {
	GetModelSchema() *orm.ModelSchema
}

func visibleDatabaseField(queryset interface{}, name string) (string, bool) {
	provider, ok := queryset.(modelSchemaProvider)
	if !ok || provider.GetModelSchema() == nil {
		return name, true
	}
	field := provider.GetModelSchema().GetField(name)
	if field == nil {
		return name, true
	}
	if !field.Serialize {
		return "", false
	}
	if field.DBColumn != "" {
		return field.DBColumn, true
	}
	return field.Name, true
}
