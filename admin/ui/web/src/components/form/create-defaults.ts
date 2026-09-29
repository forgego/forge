import type { FieldMetadata } from "../../api/types";

/**
 * Returns the initial values of a create form: every writable field whose
 * metadata declares a `default_value` starts at that value, so a field the
 * user never touches is submitted with the schema default instead of the
 * widget's empty state (an unchecked box would otherwise store false).
 */
export function createDefaults(
  fields: Pick<FieldMetadata, "name" | "read_only" | "default_value">[]
): Record<string, any> {
  const values: Record<string, any> = {};
  for (const field of fields) {
    if (field.name === "id" || field.read_only) continue;
    if (field.default_value === undefined || field.default_value === null) continue;
    values[field.name] = field.default_value;
  }
  return values;
}
