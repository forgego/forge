// The API sends timestamps as RFC 3339 ("2026-09-19T14:30:05Z"); a
// datetime-local input only takes "YYYY-MM-DDTHH:mm[:ss]" in local time.

const pad = (n: number) => String(n).padStart(2, "0");

/** Converts an API timestamp to the value a datetime-local input shows. */
export function toDateTimeLocal(value: unknown): string {
  if (typeof value !== "string" || value === "") return "";
  // Already a local value (no zone): leave it alone.
  if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2})?$/.test(value)) return value;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return (
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` +
    `T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
  );
}

/** Converts a datetime-local value back to an RFC 3339 UTC timestamp. */
export function fromDateTimeLocal(value: string): string {
  if (value === "") return "";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toISOString();
}
