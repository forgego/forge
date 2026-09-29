import { describe, expect, it } from "vitest";
import { createDefaults } from "./create-defaults";

describe("createDefaults", () => {
  it("prefills writable fields from default_value", () => {
    expect(
      createDefaults([
        { name: "is_active", read_only: false, default_value: true },
        { name: "sort_order", read_only: false, default_value: 0 },
        { name: "status", read_only: false, default_value: "draft" },
      ])
    ).toEqual({ is_active: true, sort_order: 0, status: "draft" });
  });

  it("keeps false and zero defaults", () => {
    expect(
      createDefaults([
        { name: "is_featured", read_only: false, default_value: false },
        { name: "rating", read_only: false, default_value: 0 },
      ])
    ).toEqual({ is_featured: false, rating: 0 });
  });

  it("skips id, read-only fields and fields without a default", () => {
    expect(
      createDefaults([
        { name: "id", read_only: false, default_value: 1 },
        { name: "view_count", read_only: true, default_value: 0 },
        { name: "name", read_only: false },
        { name: "notes", read_only: false, default_value: null },
      ])
    ).toEqual({});
  });
});
