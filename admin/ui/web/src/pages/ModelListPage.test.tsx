import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import ModelListPage from "./ModelListPage";

const useNavigateMock = vi.hoisted(() => vi.fn());
const useParamsMock = vi.hoisted(() => vi.fn());
const useModelMetadataMock = vi.hoisted(() => vi.fn());
const useModelListMock = vi.hoisted(() => vi.fn());
const useDeleteObjectMock = vi.hoisted(() => vi.fn());
const useBulkDeleteMock = vi.hoisted(() => vi.fn());
const useSavedViewsMock = vi.hoisted(() => vi.fn());
const useSaveSavedViewMock = vi.hoisted(() => vi.fn());
const useDeleteSavedViewMock = vi.hoisted(() => vi.fn());
const useLogoutMock = vi.hoisted(() => vi.fn());

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => useNavigateMock,
  useParams: () => useParamsMock(),
  useLocation: () => ({ pathname: "/products" }),
  Link: ({ children, to }: any) => <a href={to}>{children}</a>,
}));

vi.mock("../api/hooks/adminHooks", () => ({
  useModelMetadata: (...args: any[]) => useModelMetadataMock(...args),
  useModelList: (...args: any[]) => useModelListMock(...args),
  useDeleteObject: (...args: any[]) => useDeleteObjectMock(...args),
  useBulkDelete: (...args: any[]) => useBulkDeleteMock(...args),
  useSavedViews: (...args: any[]) => useSavedViewsMock(...args),
  useSaveSavedView: (...args: any[]) => useSaveSavedViewMock(...args),
  useDeleteSavedView: (...args: any[]) => useDeleteSavedViewMock(...args),
  useLogout: (...args: any[]) => useLogoutMock(...args),
  useModels: () => ({ data: { models: [] } }),
  useConfig: () => ({ data: {} }),
  adminKeys: {
    model: () => ["admin", "model"],
    savedViews: () => ["admin", "saved-views"],
  },
}));

vi.mock("../hooks/use-toast", () => ({
  useToast: () => ({ toast: vi.fn() }),
}));

vi.mock("../hooks/useUIComponent", () => ({
  useUIComponent: (_override: any, defaultComp: any) => defaultComp,
}));

describe("ModelListPage", () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  const mockMetadata = {
    name: "products",
    verbose_name: "Product",
    verbose_name_plural: "Products",
    description: "Product management",
    list_display: ["id", "title", "price"],
    fields: [
      { name: "id", label: "ID", type: "integer", widget: "number", required: true, read_only: true },
      { name: "title", label: "Title", type: "text", widget: "text", required: true, read_only: false },
      { name: "price", label: "Price", type: "number", widget: "number", required: true, read_only: false },
    ],
    permissions: { view: true, add: true, change: true, delete: true },
    actions: [{ name: "publish", label: "Publish Products" }],
    filters: [],
    pagination: { page_size: 20, max_page_size: 100 },
  };

  const mockListData = {
    count: 2,
    total_pages: 1,
    page: 1,
    page_size: 20,
    results: [
      { id: 1, title: "Laptop Pro", price: 1299 },
      { id: 2, title: "Wireless Mouse", price: 49 },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    useParamsMock.mockReturnValue({ model: "products" });
    useModelMetadataMock.mockReturnValue({ data: mockMetadata, isLoading: false, error: null });
    useModelListMock.mockReturnValue({ data: mockListData, isLoading: false, error: null });
    useDeleteObjectMock.mockReturnValue({ mutateAsync: vi.fn(), isPending: false });
    useBulkDeleteMock.mockReturnValue({ mutateAsync: vi.fn(), isPending: false });
    useSavedViewsMock.mockReturnValue({ data: { views: [] } });
    useSaveSavedViewMock.mockReturnValue({ mutateAsync: vi.fn(), isPending: false });
    useDeleteSavedViewMock.mockReturnValue({ mutateAsync: vi.fn(), isPending: false });
    useLogoutMock.mockReturnValue({ mutate: vi.fn(), isPending: false });
  });

  it("renders table rows and column headers", () => {
    render(
      <QueryClientProvider client={queryClient}>
        <ModelListPage />
      </QueryClientProvider>
    );
    expect(screen.getAllByText("Products").length).toBeGreaterThan(0);
    expect(screen.getByText("Laptop Pro")).toBeInTheDocument();
    expect(screen.getByText("Wireless Mouse")).toBeInTheDocument();
  });

  it("enables bulk actions when items are selected", () => {
    render(
      <QueryClientProvider client={queryClient}>
        <ModelListPage />
      </QueryClientProvider>
    );

    const selectAllCheckbox = screen.getByTestId("select-all");
    fireEvent.click(selectAllCheckbox);

    expect(screen.getByText("2 selected")).toBeInTheDocument();
  });

  it("shows bulk delete button when items are selected and delete permission is true", () => {
    render(
      <QueryClientProvider client={queryClient}>
        <ModelListPage />
      </QueryClientProvider>
    );

    const selectAllCheckbox = screen.getByTestId("select-all");
    fireEvent.click(selectAllCheckbox);

    const bulkDeleteBtn = screen.getByTestId("bulk-delete-button");
    expect(bulkDeleteBtn).toBeInTheDocument();
    expect(bulkDeleteBtn).toHaveTextContent("Delete (2)");
  });

  it("toggles column sorting when header sort buttons are clicked", () => {
    render(
      <QueryClientProvider client={queryClient}>
        <ModelListPage />
      </QueryClientProvider>
    );

    const sortButton = screen.getByTestId("sort-title");
    expect(sortButton).toHaveAttribute("title", "Sort by Title");
    fireEvent.click(sortButton);

    // The list query should now include ascending ordering for title.
    const calls = useModelListMock.mock.calls;
    const lastParams = calls[calls.length - 1][1];
    expect(lastParams.ordering).toBe("title");

    fireEvent.click(sortButton);
    const calls2 = useModelListMock.mock.calls;
    expect(calls2[calls2.length - 1][1].ordering).toBe("-title");
  });

  it("renders pagination status, page-size control and prev/next buttons", () => {
    render(
      <QueryClientProvider client={queryClient}>
        <ModelListPage />
      </QueryClientProvider>
    );

    expect(screen.getByTestId("pagination-status")).toHaveTextContent(
      /of 2$/
    );
    expect(screen.getByTestId("page-size-select")).toBeInTheDocument();
    const prev = screen.getByTestId("pagination-prev");
    const next = screen.getByTestId("pagination-next");
    expect(prev).toBeDisabled();
    // Single page of results: next is disabled too.
    expect(next).toBeDisabled();
  });

  it("opens the save-view dialog instead of a native prompt", () => {
    render(
      <QueryClientProvider client={queryClient}>
        <ModelListPage />
      </QueryClientProvider>
    );

    expect(
      screen.queryByTestId("save-view-dialog")
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByTestId("save-view-button"));
    expect(screen.getByTestId("save-view-dialog")).toBeInTheDocument();
    expect(screen.getByTestId("save-view-name")).toBeInTheDocument();
    expect(screen.getByTestId("save-view-confirm")).toBeDisabled();
  });

  it("shows a retry button when the list query fails", () => {
    useModelListMock.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error("boom"),
      refetch: vi.fn(),
    });
    render(
      <QueryClientProvider client={queryClient}>
        <ModelListPage />
      </QueryClientProvider>
    );

    expect(screen.getByText("Could not load records")).toBeInTheDocument();
    expect(screen.getByTestId("list-retry")).toBeInTheDocument();
  });

  it("offers to clear search and filters when nothing matches", () => {
    vi.useFakeTimers();
    try {
      useModelListMock.mockReturnValue({
        data: { count: 0, total_pages: 0, page: 1, page_size: 20, results: [] },
        isLoading: false,
        error: null,
      });
      render(
        <QueryClientProvider client={queryClient}>
          <ModelListPage />
        </QueryClientProvider>
      );

      fireEvent.change(screen.getByTestId("search-input"), {
        target: { value: "zzz-no-match" },
      });
      act(() => {
        vi.advanceTimersByTime(400);
      });
      expect(screen.getByText("No results found")).toBeInTheDocument();
      expect(screen.getByTestId("clear-filters")).toBeInTheDocument();
    } finally {
      vi.useRealTimers();
    }
  });

  it("clears the search input via the clear button", () => {
    vi.useFakeTimers();
    try {
      render(
        <QueryClientProvider client={queryClient}>
          <ModelListPage />
        </QueryClientProvider>
      );

      const input = screen.getByTestId("search-input") as HTMLInputElement;
      fireEvent.change(input, { target: { value: "lap" } });
      expect(screen.getByTestId("clear-search")).toBeInTheDocument();
      fireEvent.click(screen.getByTestId("clear-search"));
      expect(input.value).toBe("");
    } finally {
      vi.useRealTimers();
    }
  });
});
