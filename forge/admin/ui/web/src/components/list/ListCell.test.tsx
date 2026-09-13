import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import type { useNavigate } from "@tanstack/react-router";
import { ListCell } from "./ListCell";
import type { RelationMetadata, FieldMetadata } from "../../api/types";

describe("ListCell", () => {
  const mockNavigate = vi.fn() as unknown as ReturnType<typeof useNavigate>;

  const customerRelation: RelationMetadata = {
    name: "customer",
    type: "foreign_key",
    related_model: "customers",
    label: "Customer",
  };

  const customerField: FieldMetadata = {
    name: "customer_id",
    type: "integer",
    label: "Customer",
    required: false,
    read_only: false,
    widget: "foreign_key",
  };

  it("renders related-object label and muted #id when display mapping is present", () => {
    const navigate = vi.fn() as unknown as ReturnType<typeof useNavigate>;
    const obj = { id: 7, customer_id: 91 };
    const display = { customer: { "91": "Acme GmbH" } };

    render(
      <table>
        <tbody>
          <tr>
            <ListCell
              fieldName="customer_id"
              colIdx={1}
              obj={obj}
              field={customerField}
              relation={customerRelation}
              navigate={navigate}
              display={display}
            />
          </tr>
        </tbody>
      </table>
    );

    const button = screen.getByTestId("fk-customer_id-7");
    expect(button).toBeInTheDocument();
    expect(button).toHaveTextContent("Acme GmbH");
    expect(button).toHaveTextContent("#91");
    expect(button).toHaveAttribute("title", "Acme GmbH");

    fireEvent.click(button);
    expect(navigate).toHaveBeenCalledWith({
      to: "/$model/$id/view",
      params: {
        model: "customers",
        id: "91",
      },
    });
  });

  it("renders fallback #id button when no display prop is provided", () => {
    const navigate = vi.fn() as unknown as ReturnType<typeof useNavigate>;
    const obj = { id: 7, customer_id: 91 };

    render(
      <table>
        <tbody>
          <tr>
            <ListCell
              fieldName="customer_id"
              colIdx={1}
              obj={obj}
              field={customerField}
              relation={customerRelation}
              navigate={navigate}
            />
          </tr>
        </tbody>
      </table>
    );

    const button = screen.getByTestId("fk-customer_id-7");
    expect(button).toBeInTheDocument();
    expect(button.textContent).toBe("#91");
    expect(button).toHaveAttribute("title", "View related Customer");

    fireEvent.click(button);
    expect(navigate).toHaveBeenCalledWith({
      to: "/$model/$id/view",
      params: {
        model: "customers",
        id: "91",
      },
    });
  });

  it("renders EmptyValue em dash and no button when FK value is null", () => {
    const obj = { id: 7, customer_id: null };
    const display = { customer: { "91": "Acme GmbH" } };

    render(
      <table>
        <tbody>
          <tr>
            <ListCell
              fieldName="customer_id"
              colIdx={1}
              obj={obj}
              field={customerField}
              relation={customerRelation}
              navigate={mockNavigate}
              display={display}
            />
          </tr>
        </tbody>
      </table>
    );

    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    expect(screen.getByText("—")).toBeInTheDocument();
    expect(screen.getByLabelText("No value")).toBeInTheDocument();
  });

  it("resolves label using fieldName when relation.name is not in display", () => {
    const obj = { id: 7, customer_id: 91 };
    const display = { customer_id: { "91": "Acme GmbH" } };

    render(
      <table>
        <tbody>
          <tr>
            <ListCell
              fieldName="customer_id"
              colIdx={1}
              obj={obj}
              field={customerField}
              relation={customerRelation}
              navigate={mockNavigate}
              display={display}
            />
          </tr>
        </tbody>
      </table>
    );

    const button = screen.getByTestId("fk-customer_id-7");
    expect(button).toHaveTextContent("Acme GmbH");
    expect(button).toHaveTextContent("#91");
  });

  it("falls back to #id button when display entry is an empty string", () => {
    const obj = { id: 7, customer_id: 91 };
    const display = { customer: { "91": "" } };

    render(
      <table>
        <tbody>
          <tr>
            <ListCell
              fieldName="customer_id"
              colIdx={1}
              obj={obj}
              field={customerField}
              relation={customerRelation}
              navigate={mockNavigate}
              display={display}
            />
          </tr>
        </tbody>
      </table>
    );

    const button = screen.getByTestId("fk-customer_id-7");
    expect(button.textContent).toBe("#91");
  });

  it("falls back to #id button when id is not found in display", () => {
    const obj = { id: 7, customer_id: 91 };
    const display = { customer: { "92": "Other GmbH" } };

    render(
      <table>
        <tbody>
          <tr>
            <ListCell
              fieldName="customer_id"
              colIdx={1}
              obj={obj}
              field={customerField}
              relation={customerRelation}
              navigate={mockNavigate}
              display={display}
            />
          </tr>
        </tbody>
      </table>
    );

    const button = screen.getByTestId("fk-customer_id-7");
    expect(button.textContent).toBe("#91");
  });
});
