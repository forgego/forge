import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { ListFilterPanel } from "./ListFilterPanel";
import type { FilterMetadata } from "../../api/types";

function renderFilter(filter: FilterMetadata) {
  const resetPage = vi.fn();
  function Harness() {
    const [activeFilters, setActiveFilters] = useState<Record<string, any>>({});
    return <>
      <ListFilterPanel filters={[filter]} activeFilters={activeFilters}
        setActiveFilters={setActiveFilters} resetPage={resetPage} />
      <output data-testid="active-filters">{JSON.stringify(activeFilters)}</output>
    </>;
  }
  render(<Harness />);
  return resetPage;
}

describe("List filter controls", () => {
  it("uses a plain text field for relation filters and applies each edit immediately", () => {
    const resetPage = renderFilter({ name: "owner", type: "relation", label: "Owner" });
    const input = screen.getByRole("textbox", { name: "Owner" });
    fireEvent.change(input, { target: { value: "42" } });
    expect(screen.getByTestId("active-filters")).toHaveTextContent('"owner":"42"');
    expect(resetPage).toHaveBeenCalledOnce();
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
  });

  it("shows date bounds without preset controls", () => {
    renderFilter({ name: "created", type: "date", label: "Created" });
    expect(screen.getByLabelText("From")).toHaveAttribute("type", "date");
    expect(screen.getByLabelText("To")).toHaveAttribute("type", "date");
    expect(screen.queryByRole("button", { name: /today|last week|last month/i })).toBeNull();
  });

  it("shows configured choice labels without matching counts", () => {
    renderFilter({ name: "status", type: "choice", label: "Status", choices: [
      { value: "open", label: "Open" }, { value: "closed", label: "Closed" },
    ] });
    fireEvent.keyDown(screen.getByTestId("filter-status"), { key: "ArrowDown" });
    expect(screen.getByRole("option", { name: "Open" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "Closed" })).toBeInTheDocument();
    expect(screen.queryByText(/Open\s*\(\d+\)/)).toBeNull();
  });
});
