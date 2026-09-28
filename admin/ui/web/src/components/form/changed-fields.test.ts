import { describe, expect, it } from "vitest";
import { changedFields } from "./changed-fields";

const fields = [
  { name: "id", read_only: true },
  { name: "name", read_only: false },
  { name: "parent_id", read_only: false },
  { name: "level", read_only: true },
  { name: "tags", read_only: false },
  { name: "is_active", read_only: false },
];

describe("changedFields", () => {
  const loaded = {
    id: 7,
    name: "Shoes",
    parent_id: 0,
    level: 0,
    tags: [1, 2],
    is_active: true,
    created_at: "",
  };

  it("sends only the fields the user changed", () => {
    expect(changedFields(loaded, { ...loaded, name: "Boots" }, fields)).toEqual({
      name: "Boots",
    });
  });

  it("never re-sends untouched values such as an unset foreign key", () => {
    expect(changedFields(loaded, { ...loaded }, fields)).toEqual({});
  });

  it("drops read-only fields even when their value differs", () => {
    expect(
      changedFields(loaded, { ...loaded, level: 9, id: 8 }, fields)
    ).toEqual({});
  });

  it("keeps falsy changes and compares arrays by value", () => {
    expect(
      changedFields(
        loaded,
        { ...loaded, is_active: false, tags: [1, 2, 3] },
        fields
      )
    ).toEqual({ is_active: false, tags: [1, 2, 3] });
    expect(changedFields(loaded, { ...loaded, tags: [1, 2] }, fields)).toEqual(
      {}
    );
  });
});
