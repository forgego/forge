import { useState } from 'react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../components/ui/dropdown-menu';
import { Button } from '../components/ui/button';
import { Bookmark, BookmarkPlus, Trash2, Clock } from 'lucide-react';
import { cn } from '../lib/utils';

export interface SavedFilter {
  id: string;
  name: string;
  filter: Record<string, any>;
  createdAt: string;
  isDefault?: boolean;
}

interface SavedFiltersProps {
  model: string;
  currentFilters: Record<string, any>;
  onApplyFilter: (filter: Record<string, any>) => void;
  onSaveFilter: (name: string) => void;
  onDeleteFilter: (id: string) => void;
  onSetDefault?: (id: string) => void;
}

function readStoredFilters(modelName: string): SavedFilter[] {
  const key = `forge-saved-filters-${modelName}`;
  const stored = localStorage.getItem(key);
  if (stored) {
    try {
      return JSON.parse(stored);
    } catch {
      return [];
    }
  }
  return [];
}

export function SavedFilters({
  model,
  currentFilters,
  onApplyFilter,
  onSaveFilter,
  onDeleteFilter,
  onSetDefault,
}: SavedFiltersProps) {
  const [savedFilters, setSavedFilters] = useState<SavedFilter[]>(() =>
    readStoredFilters(model),
  );
  const [showSaveDialog, setShowSaveDialog] = useState(false);
  const [newFilterName, setNewFilterName] = useState('');

  const [prevModel, setPrevModel] = useState(model);
  if (prevModel !== model) {
    setPrevModel(model);
    setSavedFilters(readStoredFilters(model));
  }

  const saveFilters = (filters: SavedFilter[]) => {
    const key = `forge-saved-filters-${model}`;
    localStorage.setItem(key, JSON.stringify(filters));
    setSavedFilters(filters);
  };

  const handleSave = () => {
    if (!newFilterName.trim()) return;

    const newFilter: SavedFilter = {
      id: Date.now().toString(),
      name: newFilterName.trim(),
      filter: { ...currentFilters },
      createdAt: new Date().toISOString(),
    };

    saveFilters([...savedFilters, newFilter]);
    onSaveFilter(newFilterName.trim());
    setNewFilterName('');
    setShowSaveDialog(false);
  };

  const handleDelete = (id: string) => {
    saveFilters(savedFilters.filter((f) => f.id !== id));
    onDeleteFilter(id);
  };

  const handleSetDefault = (id: string) => {
    const updated = savedFilters.map((f) => ({
      ...f,
      isDefault: f.id === id,
    }));
    saveFilters(updated);
    onSetDefault?.(id);
  };

  const handleApply = (filter: SavedFilter) => {
    onApplyFilter(filter.filter);
  };

  const getDefaultFilter = () => savedFilters.find((f) => f.isDefault);

  return (
    <div className="relative">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm" className="h-8">
            <Bookmark className="h-3.5 w-3.5 mr-1.5" />
            Filters
            {getDefaultFilter() && (
              <span className="ml-1 px-1.5 py-0.5 bg-primary/10 text-primary text-xs rounded">
                Default
              </span>
            )}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-64">
          <DropdownMenuLabel className="flex items-center justify-between">
            <span>Saved Filters</span>
            <Button
              variant="ghost"
              size="sm"
              className="h-6 px-2 text-xs"
              onClick={() => setShowSaveDialog(true)}
            >
              <BookmarkPlus className="h-3 w-3 mr-1" />
              Save Current
            </Button>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />

          {savedFilters.length === 0 ? (
            <div className="py-4 text-center text-sm text-muted-foreground">
              <p>No saved filters</p>
              <p className="text-xs mt-1">Save your current filters for quick access</p>
            </div>
          ) : (
            savedFilters.map((filter) => (
              <DropdownMenuItem
                key={filter.id}
                className="flex items-center justify-between py-2 cursor-pointer"
                onClick={() => handleApply(filter)}
              >
                <div className="flex items-center gap-2">
                  <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                  <div>
                    <div className="font-medium text-sm">{filter.name}</div>
                    <div className="text-xs text-muted-foreground">
                      {new Date(filter.createdAt).toLocaleDateString()}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  {filter.isDefault && (
                    <span className="px-1.5 py-0.5 bg-primary/10 text-primary text-xs rounded">
                      Default
                    </span>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleSetDefault(filter.id);
                    }}
                    title="Set as default"
                  >
                    <Bookmark className={cn("h-3 w-3", filter.isDefault && "fill-current")} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0 text-destructive hover:text-destructive"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDelete(filter.id);
                    }}
                    title="Delete"
                  >
                    <Trash2 className="h-3 w-3" />
                  </Button>
                </div>
              </DropdownMenuItem>
            ))
          )}
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Save Dialog */}
      {showSaveDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-background rounded-lg p-6 w-80 shadow-xl">
            <h3 className="font-semibold mb-4">Save Filter</h3>
            <input
              type="text"
              value={newFilterName}
              onChange={(e) => setNewFilterName(e.target.value)}
              placeholder="Filter name..."
              className="w-full px-3 py-2 border rounded-md mb-4"
              autoFocus
            />
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowSaveDialog(false)}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={handleSave}
                disabled={!newFilterName.trim()}
              >
                Save
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// Hook for managing filter state with localStorage persistence
export function useSavedFilters(model: string) {
  const [filters, setFilters] = useState<Record<string, any>>({});
  const [savedFilters, setSavedFilters] = useState<SavedFilter[]>(() =>
    readStoredFilters(model),
  );

  const [prevModel, setPrevModel] = useState(model);
  if (prevModel !== model) {
    setPrevModel(model);
    setSavedFilters(readStoredFilters(model));
  }

  const saveCurrentFilter = (name: string) => {
    const newFilter: SavedFilter = {
      id: Date.now().toString(),
      name,
      filter: { ...filters },
      createdAt: new Date().toISOString(),
    };
    const updated = [...savedFilters, newFilter];
    localStorage.setItem(`forge-saved-filters-${model}`, JSON.stringify(updated));
    setSavedFilters(updated);
  };

  const deleteFilter = (id: string) => {
    const updated = savedFilters.filter((f) => f.id !== id);
    localStorage.setItem(`forge-saved-filters-${model}`, JSON.stringify(updated));
    setSavedFilters(updated);
  };

  return {
    filters,
    setFilters,
    savedFilters,
    saveCurrentFilter,
    deleteFilter,
  };
}
