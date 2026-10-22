import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { ListFilterPanel } from "./ListFilterPanel";
import type { FilterMetadata } from "../../api/types";

function renderFilters(filters: FilterMetadata[]) {
  const resetPage = vi.fn();
  function Harness() {
    const [activeFilters, setActiveFilters] = useState<Record<string, any>>({});
    return <>
      <ListFilterPanel filters={filters} activeFilters={activeFilters}
        setActiveFilters={setActiveFilters} resetPage={resetPage} />
      <output data-testid="active-filters">{JSON.stringify(activeFilters)}</output>
    </>;
  }
  render(<Harness />);
  return resetPage;
}

describe("ListFilterPanel", () => {
  it("offers All, Yes, and No for Boolean filters", () => {
    renderFilters([{ name: "active", type: "boolean", label: "Active" }]);
    fireEvent.keyDown(screen.getByTestId("filter-active"), { key: "ArrowDown" });
    expect(screen.getByRole("option", { name: "All" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Yes" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("option", { name: "No" }));
    expect(screen.getByTestId("active-filters")).toHaveTextContent('"active":"false"');
  });

  it("offers date range inputs that set lower and upper bounds", () => {
    renderFilters([{ name: "created_at", type: "date", label: "Created" }]);
    fireEvent.change(screen.getByLabelText("From"), { target: { value: "2026-01-01" } });
    fireEvent.change(screen.getByLabelText("To"), { target: { value: "2026-01-31" } });
    expect(screen.getByTestId("active-filters")).toHaveTextContent('"created_at__gte":"2026-01-01"');
    expect(screen.getByTestId("active-filters")).toHaveTextContent('"created_at__lte":"2026-01-31"');
  });

  it("offers numeric minimum and maximum inputs", () => {
    renderFilters([{ name: "price", type: "number", label: "Price" }]);
    fireEvent.change(screen.getByLabelText("Price minimum"), { target: { value: "10" } });
    fireEvent.change(screen.getByLabelText("Price maximum"), { target: { value: "50" } });
    expect(screen.getByTestId("active-filters")).toHaveTextContent('"price__gte":"10"');
    expect(screen.getByTestId("active-filters")).toHaveTextContent('"price__lte":"50"');
  });

  it("offers configured choices without inventing matching counts", () => {
    renderFilters([{ name: "status", type: "choice", label: "Status", choices: [
      { value: "open", label: "Open" }, { value: "closed", label: "Closed" },
    ] }]);
    fireEvent.keyDown(screen.getByTestId("filter-status"), { key: "ArrowDown" });
    expect(screen.getByRole("option", { name: "Open" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Closed" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("option", { name: "Open" }));
    expect(screen.getByTestId("active-filters")).toHaveTextContent('"status":"open"');
  });
});
