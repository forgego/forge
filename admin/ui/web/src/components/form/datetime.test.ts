import { describe, expect, it } from "vitest";
import { fromDateTimeLocal, toDateTimeLocal } from "./datetime";

describe("datetime-local conversion", () => {
  it("shows an RFC 3339 timestamp in local time and round-trips it", () => {
    const shown = toDateTimeLocal("2026-09-19T14:30:05Z");
    expect(shown).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$/);
    expect(fromDateTimeLocal(shown)).toBe("2026-09-19T14:30:05.000Z");
  });

  it("leaves a zone-less local value alone", () => {
    expect(toDateTimeLocal("2026-09-19T14:30")).toBe("2026-09-19T14:30");
  });

  it("shows nothing for empty or invalid input", () => {
    expect(toDateTimeLocal("")).toBe("");
    expect(toDateTimeLocal(null)).toBe("");
    expect(toDateTimeLocal("garbage")).toBe("");
    expect(fromDateTimeLocal("")).toBe("");
  });
});
