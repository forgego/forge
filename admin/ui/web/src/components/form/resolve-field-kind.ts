import type { FieldKind } from "./types";

export type { FieldKind } from "./types";

const INT_TYPES = new Set([
  "integer", "int", "int8", "int16", "int32", "int64",
  "uint", "uint8", "uint16", "uint32", "uint64",
]);
const FLOAT_TYPES = new Set([
  "float", "float32", "float64", "double", "decimal", "numeric", "number",
]);
const FK_TYPES = new Set([
  "foreign_key", "foreignkey", "fk", "one_to_one", "onetoone", "many_to_one",
]);
const M2M_TYPES = new Set(["many_to_many", "manytomany", "m2m"]);

export function resolveFieldKind(field: any, relation?: any): FieldKind {
  const t = String(field?.type ?? "").toLowerCase();
  const w = String(field?.widget ?? "").toLowerCase();
  const relType = String(relation?.type ?? "").toLowerCase();

  if (Array.isArray(field?.choices) && field.choices.length > 0) return "choice";
  if (FK_TYPES.has(t) || FK_TYPES.has(relType)) return "fk";
  if (M2M_TYPES.has(t) || M2M_TYPES.has(relType)) return "m2m";
  if (t === "boolean" || t === "bool" || ["checkbox", "switch", "toggle"].includes(w)) {
    return "boolean";
  }
  if (w === "number" || INT_TYPES.has(t) || FLOAT_TYPES.has(t)) return "number";
  if (w === "date" || t === "date") return "date";
  if (w === "datetime" || w === "datetime-local" || ["datetime", "timestamp", "timestamptz"].includes(t)) {
    return "datetime";
  }
  if (w === "time" || t === "time") return "time";
  if (w === "password" || t === "password") return "password";
  if (w === "email" || t === "email") return "email";
  if (w === "url" || t === "url") return "url";
  if (w === "tel" || t === "tel" || t === "phone") return "tel";
  if (w === "color" || t === "color") return "color";
  if (w === "select" || t === "choice" || t === "enum") return "choice";
  if (["json", "jsonb", "object", "dict", "map"].includes(t) || w === "json") return "json";
  if (w === "textarea" || w === "rich_text" || t === "text") return "textarea";
  return "text";
}

export function isIntegerField(field: any): boolean {
  const t = String(field?.type ?? "").toLowerCase();
  const w = String(field?.widget ?? "").toLowerCase();
  if (w === "number") return INT_TYPES.has(t) || !FLOAT_TYPES.has(t);
  return INT_TYPES.has(t);
}
