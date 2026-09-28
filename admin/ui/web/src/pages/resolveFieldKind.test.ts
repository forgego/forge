import { describe, expect, it } from "vitest";
import { resolveFieldKind, isIntegerField } from "../components/form/resolve-field-kind";
import { modelIconComponent } from "../components/ModelIcon";
import { FolderTree, Shapes } from "lucide-react";

describe("resolveFieldKind", () => {
  it("maps Go-flavored backend types to form kinds", () => {
    expect(resolveFieldKind({ type: "Bool", widget: "checkbox" })).toBe("boolean");
    expect(resolveFieldKind({ type: "boolean", widget: "checkbox" })).toBe("boolean");
    expect(resolveFieldKind({ type: "Int64", widget: "number" })).toBe("number");
    expect(resolveFieldKind({ type: "Int32", widget: "number" })).toBe("number");
    expect(resolveFieldKind({ type: "Float64", widget: "number" })).toBe("number");
    expect(resolveFieldKind({ type: "String", widget: "text" })).toBe("text");
    expect(resolveFieldKind({ type: "String", widget: "textarea" })).toBe("textarea");
    expect(resolveFieldKind({ type: "Text", widget: "rich_text" })).toBe("textarea");
  });

  it("prefers widgets for dates, passwords and contact inputs", () => {
    expect(resolveFieldKind({ type: "String", widget: "date" })).toBe("date");
    expect(resolveFieldKind({ type: "String", widget: "datetime" })).toBe("datetime");
    expect(resolveFieldKind({ type: "String", widget: "password" })).toBe("password");
    expect(resolveFieldKind({ type: "String", widget: "email" })).toBe("email");
  });

  it("detects choices and relations", () => {
    expect(
      resolveFieldKind({ type: "String", widget: "text", choices: [{ value: "a", label: "A" }] })
    ).toBe("choice");
    expect(
      resolveFieldKind({ type: "Int64", widget: "number" }, { type: "ForeignKey" })
    ).toBe("fk");
    expect(
      resolveFieldKind({ type: "String", widget: "text" }, { type: "many_to_many" })
    ).toBe("m2m");
  });

  it("maps time types to time inputs and falls back to text otherwise", () => {
    expect(resolveFieldKind({ type: "Time", widget: "text" })).toBe("time");
    expect(resolveFieldKind({ type: "SomethingElse", widget: "text" })).toBe("text");
    expect(resolveFieldKind({})).toBe("text");
  });
});

describe("isIntegerField", () => {
  it("distinguishes ints from floats", () => {
    expect(isIntegerField({ type: "Int64", widget: "number" })).toBe(true);
    expect(isIntegerField({ type: "Float64", widget: "number" })).toBe(false);
    expect(isIntegerField({ type: "String", widget: "text" })).toBe(false);
  });
});

describe("modelIconComponent", () => {
  it("resolves backend icon names case-insensitively", () => {
    expect(modelIconComponent("FolderTree")).toBe(FolderTree);
    expect(modelIconComponent("folder-tree")).toBe(FolderTree);
    expect(modelIconComponent("shopping_cart")).toBeDefined();
  });

  it("falls back to a neutral shape for unknown icons", () => {
    expect(modelIconComponent("file")).toBeDefined();
    expect(modelIconComponent("nope-not-an-icon")).toBe(Shapes);
    expect(modelIconComponent(undefined)).toBe(Shapes);
  });
});
