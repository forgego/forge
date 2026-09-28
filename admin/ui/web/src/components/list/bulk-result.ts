import type { BulkActionError } from "../../api/types";

export interface BulkItemFailure {
  id: string | number;
  code?: string;
  message: string;
}

export interface BulkResult {
  /** Operation label, e.g. "Activate Warehouses" or "Delete". */
  label: string;
  /** Number of records the operation changed. */
  succeeded: number;
  failures: BulkItemFailure[];
}

/**
 * Maps per-item errors from a bulk endpoint back to record ids.
 *
 * Action errors carry the record `id`; bulk delete/update errors carry the
 * `index` of the id in the request. Errors with neither are kept with a
 * placeholder id so no failure is dropped from the report.
 */
export function bulkFailures(
  errors: BulkActionError[] | undefined,
  requestedIds: (string | number)[]
): BulkItemFailure[] {
  if (!Array.isArray(errors)) return [];
  return errors.map((err) => {
    let id: string | number = "?";
    if (err.id !== undefined && err.id !== null && err.id !== 0) {
      id = err.id;
    } else if (
      typeof err.index === "number" &&
      err.index >= 0 &&
      err.index < requestedIds.length
    ) {
      id = requestedIds[err.index];
    }
    return { id, code: err.code, message: humanizeBulkMessage(err) };
  });
}

function humanizeBulkMessage(err: BulkActionError): string {
  switch (err.code) {
    case "permission_denied":
      return "You don't have permission to change this record";
    case "not_found":
      return "Record no longer exists";
    default:
      return err.message || "Failed";
  }
}

/** Extracts per-item errors from a rejected bulk request (4xx body). */
export function bulkErrorsFromRejection(err: unknown): BulkActionError[] | undefined {
  const data = (err as any)?.response?.data;
  if (Array.isArray(data?.errors)) return data.errors;
  if (Array.isArray(data?.error?.details?.errors)) return data.error.details.errors;
  return undefined;
}
