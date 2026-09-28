import { Button } from "../ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

export type ListPaginationProps = {
  from: number;
  to: number;
  totalCount: number;
  page: number;
  totalPages: number;
  pageSize: number;
  pageSizeOptions: number[];
  onPageSizeChange: (size: string) => void;
  onPrev: () => void;
  onNext: () => void;
};

export function ListPagination({
  from,
  to,
  totalCount,
  page,
  totalPages,
  pageSize,
  pageSizeOptions,
  onPageSizeChange,
  onPrev,
  onNext,
}: ListPaginationProps) {
  return (
    <div className="px-6 py-4 bg-muted/30 border-t border-border-subtle flex flex-col lg:flex-row lg:items-center gap-4">
      <p
        className="text-meta text-muted-foreground font-medium"
        data-testid="pagination-status"
        aria-live="polite"
      >
        Showing <span className="font-mono tabular-nums text-foreground font-semibold">{from}–{to}</span>{" "}
        of <span className="font-mono tabular-nums text-foreground font-semibold">{totalCount.toLocaleString()}</span>
      </p>
      <div className="flex items-center gap-2 lg:ml-auto">
        <label htmlFor="page-size-select" className="text-meta text-muted-foreground whitespace-nowrap">
          Rows per page
        </label>
        <Select
          value={String(pageSize)}
          onValueChange={onPageSizeChange}
        >
          <SelectTrigger
            id="page-size-select"
            data-testid="page-size-select"
            aria-label="Rows per page"
            className="h-9 w-auto min-w-[80px] text-ui"
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {pageSizeOptions.map((size) => (
              <SelectItem key={size} value={String(size)}>
                {size}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="flex items-center gap-2">
        <p className="text-meta text-muted-foreground font-medium mr-1">
          Page <span className="font-mono tabular-nums text-foreground">{page}</span> of{" "}
          <span className="font-mono tabular-nums text-foreground">{totalPages}</span>
        </p>
        <Button
          variant="outline"
          size="sm"
          data-testid="pagination-prev"
          className="h-9 px-4 text-body font-medium"
          onClick={onPrev}
          disabled={page <= 1}
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          data-testid="pagination-next"
          className="h-9 px-4 text-body font-medium"
          onClick={onNext}
          disabled={page >= totalPages}
        >
          Next
        </Button>
      </div>
    </div>
  );
}
