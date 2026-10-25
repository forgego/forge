import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { FieldRenderer } from "./FieldRenderer";
import type { FieldMetadata } from "../../api/types";
import { ForgeAdmin, clearRegisteredFieldEditors } from "../../lib/extensions/api";

describe("FieldRenderer", () => {
  const uploadUnsupportedField: FieldMetadata = {
    name: "avatar",
    type: "file",
    label: "Avatar",
    required: false,
    read_only: false,
    widget: "file",
    upload_supported: false,
  };

  const charField: FieldMetadata = {
    name: "title",
    type: "string",
    label: "Title",
    required: false,
    read_only: false,
    widget: "text",
  };

  it("renders the note and no input/textbox when upload_supported is false", () => {
    const onChange = vi.fn();
    const { container } = render(
      <FieldRenderer
        field={uploadUnsupportedField}
        value=""
        onChange={onChange}
        mode="edit"
      />
    );

    expect(
      screen.getByText(
        "File uploads are not supported by the admin. Set this value in code or through your own upload endpoint."
      )
    ).toBeInTheDocument();
    expect(screen.getByText("—")).toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(container.querySelector("input")).toBeNull();
    expect(onChange).not.toHaveBeenCalled();
  });

  it("renders existing value as plain read-only text when upload_supported is false", () => {
    const onChange = vi.fn();
    const { container } = render(
      <FieldRenderer
        field={uploadUnsupportedField}
        value="documents/spec.pdf"
        onChange={onChange}
        mode="edit"
      />
    );

    expect(screen.getByText("documents/spec.pdf")).toBeInTheDocument();
    expect(
      screen.getByText(
        "File uploads are not supported by the admin. Set this value in code or through your own upload endpoint."
      )
    ).toBeInTheDocument();
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    expect(container.querySelector("input")).toBeNull();
    expect(onChange).not.toHaveBeenCalled();
  });

  it("renders an input for a char field without upload_supported flag", () => {
    const onChange = vi.fn();
    const { container } = render(
      <FieldRenderer
        field={charField}
        value="Initial Title"
        onChange={onChange}
        mode="edit"
      />
    );

    const input = screen.getByRole("textbox");
    expect(input).toBeInTheDocument();
    expect(container.querySelector("input")).toBeInTheDocument();
    expect(
      screen.queryByText(
        "File uploads are not supported by the admin. Set this value in code or through your own upload endpoint."
      )
    ).not.toBeInTheDocument();
  });

  it("renders a registered custom extension field editor when available", () => {
    ForgeAdmin.registerFieldEditor(
      { model: "products", field: "tags" },
      ({ value }) => <div data-testid="custom-tags-editor">Tags: {String(value)}</div>
    );

    const tagField: FieldMetadata = {
      name: "tags",
      type: "string",
      label: "Tags",
      required: false,
      read_only: false,
    };

    render(
      <FieldRenderer
        field={tagField}
        value="electronics, computers"
        onChange={vi.fn()}
        mode="edit"
        metadata={{ name: "products" } as any}
      />
    );

    expect(screen.getByTestId("custom-tags-editor")).toBeInTheDocument();
    expect(screen.getByText("Tags: electronics, computers")).toBeInTheDocument();
    clearRegisteredFieldEditors();
  });
});
