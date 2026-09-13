import { Button } from "../ui/button";
import { Card, CardContent } from "../ui/card";
import { Plus, X } from "lucide-react";
import { InlineFieldRenderer } from "./InlineFieldRenderer";
import type { InlineRelationDetail } from "./types";

export function getInlineFields(
  inlineMeta: any,
  inlineConfig: any,
  relatedField?: string,
  mode?: "create" | "edit"
) {
  const allowedFields = inlineConfig?.fields?.length
    ? new Set(inlineConfig.fields)
    : null;
  return (inlineMeta?.fields ?? []).filter((field: any) => {
    if (field.name === "id") return false;
    if (relatedField && field.name === relatedField) return false;
    if (allowedFields && !allowedFields.has(field.name)) return false;
    if (field.read_only && mode === "create") return false;
    return true;
  });
}

interface InlineRelationsProps {
  details: InlineRelationDetail[];
  metaQueries: any[];
  inlineData: Record<string, any[]>;
  mode: "create" | "edit";
  onUpdateRow: (
    relationName: string,
    rowIndex: number,
    fieldName: string,
    value: any
  ) => void;
  onAddRow: (relationName: string) => void;
  onRemoveRow: (relationName: string, rowIndex: number) => void;
}

export function InlineRelations({
  details,
  metaQueries,
  inlineData,
  mode,
  onUpdateRow,
  onAddRow,
  onRemoveRow,
}: InlineRelationsProps) {
  return (
    <>
      {details.map((detail, index) => {
        const inlineMeta = metaQueries[index]?.data;
        const inlineRows = inlineData[detail.relation.name] ?? [];
        const inlineLabel =
          detail.inlineConfig.label ||
          detail.relation.label ||
          detail.relation.name;
        const allowMultiple = detail.inlineConfig.type === "one_to_many";
        const inlineFields = getInlineFields(
          inlineMeta,
          detail.inlineConfig,
          detail.relatedField,
          mode
        );
        const inlineRelByName = new Map(
          (inlineMeta?.relations ?? []).map((rel: any) => [rel.name, rel])
        );

        return (
          <Card
            key={detail.relation.name}
            className="overflow-hidden max-w-4xl mx-auto"
          >
            <CardContent className="p-4 sm:p-8 space-y-6">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-semibold text-foreground">
                    {inlineLabel}
                  </h2>
                  {detail.inlineConfig.type && (
                    <p className="text-xs text-muted-foreground uppercase tracking-widest mt-1">
                      {detail.inlineConfig.type.replace(/_/g, " ")}
                    </p>
                  )}
                </div>
                {allowMultiple && (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => onAddRow(detail.relation.name)}
                    className="gap-2"
                  >
                    <Plus className="h-4 w-4" />
                    Add {inlineLabel}
                  </Button>
                )}
              </div>

              {inlineFields.length === 0 && (
                <p className="text-sm text-muted-foreground">
                  No editable fields configured for this inline relation.
                </p>
              )}

              <div className="space-y-6">
                {inlineRows.length === 0 && allowMultiple && (
                  <div className="rounded-lg border border-dashed border-border/60 p-6 text-center text-sm text-muted-foreground">
                    No {inlineLabel} added yet.
                  </div>
                )}

                {inlineRows.map((row, rowIndex) => (
                  <div
                    key={row?.id || rowIndex}
                    className="rounded-lg border border-border/50 bg-background/40 p-6 space-y-6"
                  >
                    <div className="flex items-center justify-between">
                      <h3 className="text-sm font-semibold text-foreground">
                        {inlineLabel} {rowIndex + 1}
                      </h3>
                      {allowMultiple && (
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() =>
                            onRemoveRow(detail.relation.name, rowIndex)
                          }
                          className="text-muted-foreground hover:text-destructive"
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      )}
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6">
                      {inlineFields.map((field: any) => (
                        <div key={field.name} className="space-y-2">
                          <div className="flex items-center justify-between px-1">
                            <label
                              htmlFor={`${detail.relation.name}-${rowIndex}-${field.name}`}
                              className="text-xs font-bold uppercase tracking-widest text-muted-foreground/80"
                            >
                              {field.label}{" "}
                              {field.required && (
                                <span className="text-destructive font-normal">
                                  *
                                </span>
                              )}
                            </label>
                          </div>
                          <InlineFieldRenderer
                            // Prefix the DOM id namespace so inline inputs
                            // never collide with main-form ids.
                            field={{
                              ...field,
                              name: `${detail.relation.name}-${rowIndex}-${field.name}`,
                            }}
                            value={row?.[field.name]}
                            onChange={(val) =>
                              onUpdateRow(
                                detail.relation.name,
                                rowIndex,
                                field.name,
                                val
                              )
                            }
                            relation={inlineRelByName.get(field.name)}
                          />
                          {field.help_text && (
                            <p className="text-micro text-muted-foreground px-1 leading-relaxed">
                              {field.help_text}
                            </p>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        );
      })}
    </>
  );
}
