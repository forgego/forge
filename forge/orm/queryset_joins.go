package orm

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
)

// warnedSelectRelated stores model+relation keys that have already emitted a warning
// for skipped select_related joins where no foreign key column could be found.
// A package-level sync.Map cache is used so that applications with repeated queries do not
// spam logs with duplicate warnings for the same model and relation over the process lifetime.
var warnedSelectRelated sync.Map

// fkColumnFor resolves the FK column on schema pointing to rel.
// It returns rel.FKColumn when set; when empty, it falls back to the resolver.
func fkColumnFor(schema *ModelSchema, rel *RelationInfo) string {
	if rel != nil && rel.FKColumn != "" {
		return rel.FKColumn
	}
	col, _ := resolveRelationFK(schema, rel)
	return col
}

// buildJoinClause builds the JOIN clause
func (qs *BaseQuerySet[T]) buildJoinClause(builder *SQLBuilder) {
	qs.joins = []string{}
	qs.joinMap = make(map[string]bool)

	for _, path := range qs.selectRelated {
		if qs.joinMap[path] {
			continue
		}

		rel := qs.schema.GetRelation(path)
		if rel == nil {
			continue
		}

		fkColumn := fkColumnFor(qs.schema, rel)
		if fkColumn == "" {
			qs.warnSkippedSelectRelated(rel.Name)
			continue
		}

		targetSchema, err := GetModelSchemaByName(rel.TargetModel)
		if err != nil {
			continue
		}

		qs.appendJoinClause(path, rel, targetSchema, fkColumn)
	}
}

func (qs *BaseQuerySet[T]) warnSkippedSelectRelated(relName string) {
	modelName := ""
	if qs.schema != nil {
		modelName = qs.schema.GetModelName()
	}
	if modelName == "" {
		modelName = qs.table
	}
	warnKey := modelName + "." + relName
	if _, loaded := warnedSelectRelated.LoadOrStore(warnKey, struct{}{}); !loaded {
		slog.Warn("select_related skipped: no foreign key column",
			"model", modelName,
			"relation", relName,
			"expected_column", strcase.ToSnake(relName)+"_id",
		)
	}
}

func (qs *BaseQuerySet[T]) appendJoinClause(path string, rel *RelationInfo, targetSchema *ModelSchema, fkColumn string) {
	joinTable := EscapeIdentifier(targetSchema.TableName)
	alias := EscapeIdentifier(rel.Name)
	mainTable := EscapeIdentifier(qs.table)

	targetPk := targetSchema.PrimaryKey
	if targetPk == "" {
		targetPk = "id"
	}

	joinSQL := fmt.Sprintf("LEFT OUTER JOIN %s %s ON %s.%s = %s.%s",
		joinTable, alias,
		mainTable, EscapeIdentifier(fkColumn),
		alias, EscapeIdentifier(targetPk))

	qs.joins = append(qs.joins, joinSQL)
	qs.recordJoinAliases(path, rel, alias)
}

func (qs *BaseQuerySet[T]) recordJoinAliases(path string, rel *RelationInfo, alias string) {
	qs.joinMap[path] = true
	qs.joinMap[rel.Name] = true
	qs.joinMap[strings.ToLower(rel.Name)] = true
	qs.joinMap[rel.TargetModel] = true
	qs.joinMap[strings.ToLower(rel.TargetModel)] = true
	qs.joinMap[alias] = true
}

// reverseFKFor resolves the foreign key column on targetSchema pointing back to currentSchema.
func reverseFKFor(currentSchema, targetSchema *ModelSchema) string {
	if currentSchema == nil || targetSchema == nil {
		return ""
	}
	for _, r := range targetSchema.Relations {
		if (currentSchema.ModelType != nil && r.TargetModel == currentSchema.ModelType.Name()) || r.TargetModel == currentSchema.TableName {
			if r.FKColumn != "" {
				return r.FKColumn
			}
			if col := fkColumnFor(targetSchema, &r); col != "" {
				return col
			}
		}
	}
	guess := strings.ToLower(currentSchema.TableName) + "_id"
	if f := targetSchema.GetField(guess); f != nil {
		return f.DBColumn
	}
	if currentSchema.ModelType != nil {
		guess = strings.ToLower(currentSchema.ModelType.Name()) + "_id"
		if f := targetSchema.GetField(guess); f != nil {
			return f.DBColumn
		}
	}
	return ""
}

func (qs *BaseQuerySet[T]) createJoinResolver(pathJoins *[]string, seen map[string]bool, multiValued *bool) JoinResolver {
	if seen == nil {
		seen = make(map[string]bool)
	}
	return func(parts []string) (string, string, error) {
		if len(parts) <= 1 {
			return "", "", fmt.Errorf("invalid relation path: %v", parts)
		}
		currentSchema := qs.schema
		parentRef := EscapeIdentifier(qs.table)
		var currentAlias string

		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]
			if currentSchema == nil {
				return "", "", fmt.Errorf("cannot traverse relation %s: schema is nil", part)
			}
			rel := currentSchema.GetRelation(part)
			if rel == nil {
				return "", "", fmt.Errorf("relation %s not found in model %s", part, currentSchema.TableName)
			}

			if multiValued != nil && rel.Type == RelationManyToMany {
				*multiValued = true
			}

			targetSchema, err := GetModelSchemaByName(rel.TargetModel)
			if err != nil {
				return "", "", fmt.Errorf("failed to resolve target model %s for relation %s: %w", rel.TargetModel, part, err)
			}

			hopAlias := strings.Join(parts[:i+1], "__")
			currentAlias = hopAlias

			// Check if already joined (dedupe with qs.joinMap)
			if i == 0 && qs.joinMap[rel.Name] {
				currentAlias = rel.Name
				hopAlias = rel.Name
			}

			if !seen[hopAlias] && !(i == 0 && qs.joinMap[rel.Name]) {
				targetPk := targetSchema.PrimaryKey
				if targetPk == "" {
					targetPk = "id"
				}

				fkColumn := fkColumnFor(currentSchema, rel)
				var joinSQL string
				if fkColumn != "" {
					joinSQL = fmt.Sprintf("LEFT JOIN %s AS %s ON %s.%s = %s.%s",
						EscapeIdentifier(targetSchema.TableName),
						EscapeIdentifier(hopAlias),
						EscapeIdentifier(hopAlias),
						EscapeIdentifier(targetPk),
						parentRef,
						EscapeIdentifier(fkColumn),
					)
				} else {
					if multiValued != nil {
						*multiValued = true
					}
					reverseFK := reverseFKFor(currentSchema, targetSchema)
					if reverseFK == "" {
						return "", "", fmt.Errorf("cannot determine join condition for relation %s", part)
					}
					currentPk := currentSchema.PrimaryKey
					if currentPk == "" {
						currentPk = "id"
					}
					joinSQL = fmt.Sprintf("LEFT JOIN %s AS %s ON %s.%s = %s.%s",
						EscapeIdentifier(targetSchema.TableName),
						EscapeIdentifier(hopAlias),
						EscapeIdentifier(hopAlias),
						EscapeIdentifier(reverseFK),
						parentRef,
						EscapeIdentifier(currentPk),
					)
				}

				*pathJoins = append(*pathJoins, joinSQL)
				seen[hopAlias] = true
			}

			parentRef = EscapeIdentifier(currentAlias)
			currentSchema = targetSchema
		}

		lastPart := parts[len(parts)-1]
		if currentSchema == nil {
			return "", "", fmt.Errorf("target schema is nil for field %s", lastPart)
		}
		field := currentSchema.GetField(lastPart)
		if field == nil {
			return "", "", fmt.Errorf("field %s not found in model %s", lastPart, currentSchema.TableName)
		}
		column := field.DBColumn
		if column == "" {
			column = field.Name
		}
		return currentAlias, column, nil
	}
}
