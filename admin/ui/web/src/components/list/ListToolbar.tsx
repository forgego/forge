import type { SavedView } from "../../api/types";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Badge } from "../ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Search, X, Trash2, Filter } from "lucide-react";

export type ListToolbarProps = {
  searchLabel: string;
  searchInput: string;
  onSearchChange: (value: string) => void;
  savedViews: SavedView[];
  selectedSavedView: string;
  onApplySavedView: (viewId: string) => void;
  onOpenSaveView: () => void;
  onDeleteSavedView: () => void;
  deleteViewPending: boolean;
  isFilterOpen: boolean;
  onToggleFilters: () => void;
  activeFilterCount: number;
};

export function ListToolbar({
  searchLabel,
  searchInput,
  onSearchChange,
  savedViews,
  selectedSavedView,
  onApplySavedView,
  onOpenSaveView,
  onDeleteSavedView,
  deleteViewPending,
  isFilterOpen,
  onToggleFilters,
  activeFilterCount,
}: ListToolbarProps) {
  return (
    <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
      <div className="relative flex-1 w-full">
        <label htmlFor="list-search" className="sr-only">
          Search {searchLabel}
        </label>
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
        <Input
          id="list-search"
          data-testid="search-input"
          placeholder={`Search ${searchLabel.toLowerCase()}...`}
          value={searchInput}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-10 pr-9 h-10"
        />
        {searchInput && (
          <button
            type="button"
            aria-label="Clear search"
            data-testid="clear-search"
            onClick={() => onSearchChange("")}
            className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
      <div className="flex items-center gap-2 flex-wrap">
        <label htmlFor="saved-view-select" className="sr-only">
          Saved views
        </label>
        <Select
          value={selectedSavedView || "__none__"}
          onValueChange={(v) => onApplySavedView(v === "__none__" ? "" : v)}
        >
          <SelectTrigger
            id="saved-view-select"
            data-testid="saved-view-select"
            aria-label="Saved views"
            className="h-9 w-auto min-w-[160px] text-ui"
          >
            <SelectValue placeholder="Saved views" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">Saved views</SelectItem>
            {savedViews.map((view) => (
              <SelectItem key={view.id} value={view.id}>
                {view.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          variant="outline"
          size="sm"
          data-testid="save-view-button"
          className="h-10 px-4 text-body font-medium"
          onClick={onOpenSaveView}
        >
          Save view
        </Button>
        {selectedSavedView && (
          <Button
            variant="ghost"
            size="sm"
            data-testid="delete-view-button"
            className="h-10 px-3 text-meta text-muted-foreground hover:text-destructive"
            onClick={onDeleteSavedView}
            disabled={deleteViewPending}
            title="Delete this saved view"
          >
            <Trash2 className="h-4 w-4" />
            <span className="sr-only">Delete saved view</span>
          </Button>
        )}
        <Button
          data-testid="filter-button"
          variant={isFilterOpen ? "secondary" : "ghost"}
          size="sm"
          className="h-10 px-4 text-body font-medium gap-2"
          onClick={onToggleFilters}
          aria-expanded={isFilterOpen}
        >
          <Filter className="h-4 w-4" />
          Filters
          {activeFilterCount > 0 && (
            <Badge variant="secondary" className="ml-1 px-1.5">
              {activeFilterCount}
            </Badge>
          )}
        </Button>
      </div>
    </div>
  );
}
