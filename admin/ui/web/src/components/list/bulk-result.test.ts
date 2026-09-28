import { describe, expect, it } from "vitest";
import { bulkFailures, bulkErrorsFromRejection } from "./bulk-result";

describe("bulkFailures", () => {
  it("uses the record id reported by bulk actions", () => {
    expect(
      bulkFailures(
        [{ id: 7, code: "permission_denied", message: "permission denied" }],
        [3, 7]
      )
    ).toEqual([
      {
        id: 7,
        code: "permission_denied",
        message: "You don't have permission to change this record",
      },
    ]);
  });

  it("maps bulk delete error indexes back to the requested ids", () => {
    expect(
      bulkFailures(
        [{ index: 1, code: "delete_failed", message: "still referenced" }],
        [3, 7]
      )
    ).toEqual([{ id: 7, code: "delete_failed", message: "still referenced" }]);
  });

  it("keeps failures it cannot attribute", () => {
    expect(
      bulkFailures([{ id: 0, code: "invalid_id", message: "invalid id x" }], [])
    ).toEqual([{ id: "?", code: "invalid_id", message: "invalid id x" }]);
    expect(bulkFailures(undefined, [1])).toEqual([]);
  });
});

describe("bulkErrorsFromRejection", () => {
  it("reads action and bulk-delete rejection bodies", () => {
    const actionBody = { response: { data: { errors: [{ id: 1, message: "x" }] } } };
    const deleteBody = {
      response: { data: { error: { details: { errors: [{ index: 0, message: "y" }] } } } },
    };
    expect(bulkErrorsFromRejection(actionBody)).toEqual([{ id: 1, message: "x" }]);
    expect(bulkErrorsFromRejection(deleteBody)).toEqual([{ index: 0, message: "y" }]);
    expect(bulkErrorsFromRejection(new Error("network"))).toBeUndefined();
  });
});
