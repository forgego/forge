import { describe, it, expect } from "vitest";
import { parseApiError } from "./errors";

describe("parseApiError", () => {
  it("parses nested error shape correctly", () => {
    const error = {
      response: {
        data: {
          error: {
            code: "validation_error",
            message: "validation failed: ...",
            details: {
              email: ["Enter a valid email"],
              non_field_errors: ["..."],
            },
          },
        },
      },
    };

    const info = parseApiError(error, "Default fallback");

    expect(info.fieldErrors.email).toEqual(["Enter a valid email"]);
    expect(info.formErrors).toEqual(["..."]);
    expect(info.message).toBe("validation failed: ...");
    expect(info.code).toBe("validation_error");
  });

  it("parses nested error shape when passed without response wrapper", () => {
    const error = {
      error: {
        code: "validation_error",
        message: "validation failed: ...",
        details: {
          email: ["Enter a valid email"],
          non_field_errors: ["..."],
        },
      },
    };

    const info = parseApiError(error, "Default fallback");

    expect(info.fieldErrors.email).toEqual(["Enter a valid email"]);
    expect(info.formErrors).toEqual(["..."]);
    expect(info.message).toBe("validation failed: ...");
    expect(info.code).toBe("validation_error");
  });

  it("parses legacy top-level {message, details}", () => {
    const error = {
      response: {
        data: {
          code: "legacy_error",
          message: "Legacy error occurred",
          details: {
            username: ["Username is required"],
            non_field_errors: ["Global form issue"],
          },
        },
      },
    };

    const info = parseApiError(error, "Default fallback");

    expect(info.fieldErrors.username).toEqual(["Username is required"]);
    expect(info.formErrors).toEqual(["Global form issue"]);
    expect(info.message).toBe("Legacy error occurred");
    expect(info.code).toBe("legacy_error");
  });

  it("wraps string detail values in an array", () => {
    const error = {
      response: {
        data: {
          error: {
            message: "Invalid input",
            details: {
              email: "Enter a valid email",
              non_field_errors: "Form error message",
            },
          },
        },
      },
    };

    const info = parseApiError(error, "Default fallback");

    expect(info.fieldErrors.email).toEqual(["Enter a valid email"]);
    expect(info.formErrors).toEqual(["Form error message"]);
    expect(info.message).toBe("Invalid input");
  });

  it("skips non-string values and filters non-string elements from detail arrays", () => {
    const error = {
      response: {
        data: {
          error: {
            message: "Invalid input",
            details: {
              email: ["Enter a valid email", 123, null],
              age: 42,
              valid: true,
              extra: { nested: "object" },
              empty: [],
            },
          },
        },
      },
    };

    const info = parseApiError(error, "Default fallback");

    expect(info.fieldErrors.email).toEqual(["Enter a valid email"]);
    expect(info.fieldErrors.age).toBeUndefined();
    expect(info.fieldErrors.valid).toBeUndefined();
    expect(info.fieldErrors.extra).toBeUndefined();
    expect(info.fieldErrors.empty).toBeUndefined();
  });

  it("falls back when axios error has no body and generic status message", () => {
    const error = {
      message: "Request failed with status code 500",
      response: undefined,
    };

    const info = parseApiError(error, "Failed to save changes. Please try again.");

    expect(info.message).toBe("Failed to save changes. Please try again.");
    expect(info.fieldErrors).toEqual({});
    expect(info.formErrors).toEqual([]);
    expect(info.code).toBeUndefined();
  });

  it("falls back when axios error has generic status message and empty response body", () => {
    const error = {
      message: "Request failed with status code 404",
      response: { data: null },
    };

    const info = parseApiError(error, "Custom fallback");

    expect(info.message).toBe("Custom fallback");
    expect(info.fieldErrors).toEqual({});
    expect(info.formErrors).toEqual([]);
  });

  it("returns plain Error message when not an axios generic status code", () => {
    const error = new Error("boom");

    const info = parseApiError(error, "Default fallback");

    expect(info.message).toBe("boom");
    expect(info.fieldErrors).toEqual({});
    expect(info.formErrors).toEqual([]);
    expect(info.code).toBeUndefined();
  });

  it("handles undefined / null / 'x' by returning fallback and empty maps", () => {
    const testCases = [undefined, null, "x", 123, true, () => {}, Symbol("test")];

    for (const input of testCases) {
      const info = parseApiError(input, "Default fallback");

      expect(info.message).toBe("Default fallback");
      expect(info.fieldErrors).toEqual({});
      expect(info.formErrors).toEqual([]);
      expect(info.code).toBeUndefined();
    }
  });

  it("prefers body error message over axios generic message", () => {
    const error = {
      message: "Request failed with status code 400",
      response: {
        data: {
          error: {
            message: "validation failed: password too short",
          },
        },
      },
    };

    const info = parseApiError(error, "Default fallback");

    expect(info.message).toBe("validation failed: password too short");
  });
});
