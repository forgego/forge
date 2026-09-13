import type React from "react";
import type { FilterMetadata } from "../../api/types";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

type ListFilterPanelProps = {
  filters: FilterMetadata[];
  activeFilters: Record<string, any>;
  setActiveFilters: React.Dispatch<React.SetStateAction<Record<string, any>>>;
  resetPage: () => void;
};

export function ListFilterPanel({
  filters,
  activeFilters,
  setActiveFilters,
  resetPage,
}: ListFilterPanelProps) {
  return (
    <div className="p-4 bg-background rounded-lg border border-border-subtle">
      <div className="flex items-center justify-between gap-2 mb-4">
        <p className="text-body font-medium text-muted-foreground">
          Filters
        </p>
        <Button
          variant="ghost"
          size="sm"
          data-testid="reset-filters"
          className="text-meta text-muted-foreground hover:text-foreground"
          onClick={() => {
            setActiveFilters({});
            resetPage();
          }}
        >
          Reset filters
        </Button>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {filters.map((filter) => (
          <div key={filter.name} className="space-y-1.5">
            <label
              htmlFor={`filter-${filter.name}`}
              className="text-meta font-semibold text-muted-foreground"
            >
              {filter.label}
            </label>
            {filter.type === "choice" ? (
              <Select
                value={activeFilters[filter.name] || "__all__"}
                onValueChange={(v) => {
                  setActiveFilters((prev) => ({ ...prev, [filter.name]: v === "__all__" ? "" : v }));
                  resetPage();
                }}
              >
                <SelectTrigger id={`filter-${filter.name}`} data-testid={`filter-${filter.name}`}>
                  <SelectValue placeholder="All" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">All</SelectItem>
                  {filter.choices?.map((c) => (
                    <SelectItem key={String(c.value)} value={String(c.value)}>{c.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            ) : filter.type === "date" ||
              filter.type === "datetime" ? (
              <div className="flex gap-2">
                <div className="flex-1 space-y-1">
                  <label
                    htmlFor={`filter-${filter.name}-from`}
                    className="text-micro text-muted-foreground"
                  >
                    From
                  </label>
                  <Input
                    id={`filter-${filter.name}-from`}
                    type="date"
                    className="h-9 text-meta"
                    value={activeFilters[`${filter.name}__gte`] || ""}
                    onChange={(e) => {
                      setActiveFilters((prev) => ({
                        ...prev,
                        [`${filter.name}__gte`]: e.target.value,
                      }));
                      resetPage();
                    }}
                  />
                </div>
                <div className="flex-1 space-y-1">
                  <label
                    htmlFor={`filter-${filter.name}-to`}
                    className="text-micro text-muted-foreground"
                  >
                    To
                  </label>
                  <Input
                    id={`filter-${filter.name}-to`}
                    type="date"
                    className="h-9 text-meta"
                    value={activeFilters[`${filter.name}__lte`] || ""}
                    onChange={(e) => {
                      setActiveFilters((prev) => ({
                        ...prev,
                        [`${filter.name}__lte`]: e.target.value,
                      }));
                      resetPage();
                    }}
                  />
                </div>
              </div>
            ) : filter.type === "boolean" ? (
              <Select
                value={activeFilters[filter.name] || "__all__"}
                onValueChange={(v) => {
                  setActiveFilters((prev) => ({ ...prev, [filter.name]: v === "__all__" ? "" : v }));
                  resetPage();
                }}
              >
                <SelectTrigger id={`filter-${filter.name}`} data-testid={`filter-${filter.name}`}>
                  <SelectValue placeholder="All" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__all__">All</SelectItem>
                  <SelectItem value="true">Yes</SelectItem>
                  <SelectItem value="false">No</SelectItem>
                </SelectContent>
              </Select>
            ) : filter.type === "number" ? (
              <div className="flex items-center gap-2">
                <div className="flex-1">
                  <label htmlFor={`filter-${filter.name}-min`} className="sr-only">
                    {filter.label} minimum
                  </label>
                  <Input
                    id={`filter-${filter.name}-min`}
                    data-testid={`filter-${filter.name}-min`}
                    type="number"
                    inputMode="decimal"
                    placeholder="Min"
                    className="h-9 text-ui"
                    value={activeFilters[`${filter.name}__gte`] || ""}
                    onChange={(e) => {
                      setActiveFilters((prev) => ({ ...prev, [`${filter.name}__gte`]: e.target.value }));
                      resetPage();
                    }}
                  />
                </div>
                <span aria-hidden className="text-meta text-muted-foreground">–</span>
                <div className="flex-1">
                  <label htmlFor={`filter-${filter.name}-max`} className="sr-only">
                    {filter.label} maximum
                  </label>
                  <Input
                    id={`filter-${filter.name}-max`}
                    data-testid={`filter-${filter.name}-max`}
                    type="number"
                    inputMode="decimal"
                    placeholder="Max"
                    className="h-9 text-ui"
                    value={activeFilters[`${filter.name}__lte`] || ""}
                    onChange={(e) => {
                      setActiveFilters((prev) => ({ ...prev, [`${filter.name}__lte`]: e.target.value }));
                      resetPage();
                    }}
                  />
                </div>
              </div>
            ) : (
              <Input
                id={`filter-${filter.name}`}
                data-testid={`filter-${filter.name}`}
                placeholder={filter.label}
                className="h-9 text-ui"
                value={activeFilters[filter.name] || ""}
                onChange={(e) => {
                  setActiveFilters((prev) => ({
                    ...prev,
                    [filter.name]: e.target.value,
                  }));
                  resetPage();
                }}
              />
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
