import type { FieldMetadata } from "../../api/types";

/**
 * Returns the PATCH body for an edit form: only the writable fields whose
 * value differs from the loaded object.
 *
 * Sending the whole loaded object back re-submits values the user never
 * touched, such as read-only timestamps or an unset foreign key that the
 * API reports as 0, which the database then rejects.
 */
export function changedFields(
  original: Record<string, any> | undefined,
  current: Record<string, any>,
  fields: Pick<FieldMetadata, "name" | "read_only">[]
): Record<string, any> {
  const readOnly = new Set(
    fields.filter((f) => f.read_only || f.name === "id").map((f) => f.name)
  );
  const base = original ?? {};
  const changes: Record<string, any> = {};
  for (const [name, value] of Object.entries(current)) {
    if (readOnly.has(name)) continue;
    if (sameValue(base[name], value)) continue;
    changes[name] = value;
  }
  return changes;
}

function sameValue(a: unknown, b: unknown): boolean {
  if (a === b) return true;
  if (typeof a === "object" || typeof b === "object") {
    return JSON.stringify(a) === JSON.stringify(b);
  }
  return false;
}
