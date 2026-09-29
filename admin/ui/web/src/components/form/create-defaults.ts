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

/**
 * Returns the field as the form should validate it. On create, a field with
 * a schema default (including a callable one such as time.Now, which has no
 * `default_value`) is filled by the server when left empty, so the native
 * `required` check must not block the submit.
 */
export function formField<T extends Pick<FieldMetadata, "required" | "has_default">>(
  field: T,
  mode: string
): T {
  if (mode === "create" && field.required && field.has_default) {
    return { ...field, required: false };
  }
  return field;
}
