export interface ApiErrorInfo {
  code?: string;
  message: string;
  fieldErrors: Record<string, string[]>;
  formErrors: string[];
}

function isObject(val: unknown): val is Record<string, unknown> {
  return typeof val === "object" && val !== null;
}

export function parseApiError(error: unknown, fallback: string): ApiErrorInfo {
  try {
    if (!isObject(error)) {
      return {
        message: fallback,
        fieldErrors: {},
        formErrors: [],
      };
    }

    const responseData =
      "response" in error && isObject(error.response) && "data" in error.response
        ? error.response.data
        : undefined;

    const body =
      responseData !== undefined
        ? responseData
        : "error" in error || "details" in error
          ? error
          : undefined;

    let code: string | undefined;
    if (isObject(body)) {
      if (isObject(body.error) && typeof body.error.code === "string") {
        code = body.error.code;
      } else if (typeof body.code === "string") {
        code = body.code;
      }
    }

    let message: string | undefined;
    if (isObject(body)) {
      if (
        isObject(body.error) &&
        typeof body.error.message === "string" &&
        body.error.message.trim() !== ""
      ) {
        message = body.error.message;
      } else if (
        typeof body.message === "string" &&
        body.message.trim() !== ""
      ) {
        message = body.message;
      }
    }

    if (!message) {
      if (
        "message" in error &&
        typeof error.message === "string" &&
        error.message.trim() !== ""
      ) {
        const isAxiosGeneric = /^Request failed with status code \d+$/i.test(
          error.message.trim()
        );
        if (!isAxiosGeneric) {
          message = error.message;
        }
      }
    }

    if (!message) {
      message = fallback;
    }

    let details: unknown;
    if (isObject(body)) {
      if (isObject(body.error) && "details" in body.error) {
        details = body.error.details;
      } else if ("details" in body) {
        details = body.details;
      }
    }

    const fieldErrors: Record<string, string[]> = {};
    const formErrors: string[] = [];

    if (isObject(details) && !Array.isArray(details)) {
      for (const [key, val] of Object.entries(details)) {
        let strings: string[] = [];
        if (Array.isArray(val)) {
          strings = val.filter(
            (item): item is string => typeof item === "string"
          );
        } else if (typeof val === "string") {
          strings = [val];
        } else {
          continue;
        }

        if (strings.length === 0) {
          continue;
        }

        if (key === "non_field_errors") {
          formErrors.push(...strings);
        } else {
          fieldErrors[key] = strings;
        }
      }
    }

    const result: ApiErrorInfo = {
      message,
      fieldErrors,
      formErrors,
    };

    if (code !== undefined) {
      result.code = code;
    }

    return result;
  } catch {
    return {
      message: fallback,
      fieldErrors: {},
      formErrors: [],
    };
  }
}
