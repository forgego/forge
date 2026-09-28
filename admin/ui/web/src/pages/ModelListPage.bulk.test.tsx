import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import ModelListPage from "./ModelListPage";

const executeActionMock = vi.hoisted(() => vi.fn());
const toastMock = vi.hoisted(() => vi.fn());

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => vi.fn(),
  useParams: () => ({ model: "warehouses" }),
  useLocation: () => ({ pathname: "/warehouses" }),
  Link: ({ children, to }: any) => <a href={to}>{children}</a>,
}));

vi.mock("../api/client", () => ({
  adminAPI: { executeAction: (...args: any[]) => executeActionMock(...args) },
}));

vi.mock("../api/hooks/adminHooks", () => ({
  useModelMetadata: () => ({
    data: {
      name: "warehouses",
      verbose_name: "Warehouse",
      verbose_name_plural: "Warehouses",
      list_display: ["name"],
      fields: [{ name: "name", label: "Name", type: "String", widget: "text", required: true, read_only: false }],
      permissions: { view: true, add: true, change: true, delete: true },
      actions: [{ name: "activate", label: "Activate Warehouses" }],
      filters: [],
      pagination: { page_size: 20, max_page_size: 100 },
    },
    isLoading: false,
  }),
  useModelList: () => ({
    data: {
      count: 2,
      total_pages: 1,
      results: [
        { id: 3, name: "Overflow" },
        { id: 7, name: "Primary" },
      ],
    },
    isLoading: false,
    error: null,
  }),
  useDeleteObject: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useBulkDelete: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useSavedViews: () => ({ data: { views: [] } }),
  useSaveSavedView: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteSavedView: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useLogout: () => ({ mutate: vi.fn(), isPending: false }),
  useModels: () => ({ data: { models: [] } }),
  useConfig: () => ({ data: {} }),
  adminKeys: { model: () => ["admin", "model"] },
}));

vi.mock("../hooks/use-toast", () => ({
  useToast: () => ({ toast: toastMock }),
}));

vi.mock("../hooks/useUIComponent", () => ({
  useUIComponent: (_override: any, defaultComp: any) => defaultComp,
}));

function renderPage() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <ModelListPage />
    </QueryClientProvider>
  );
}

describe("ModelListPage bulk action feedback", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("reports each skipped record when an action partially succeeds", async () => {
    executeActionMock.mockResolvedValue({
      success: false,
      affected: 1,
      message: "Executed Activate Warehouses on 1 objects; 1 skipped",
      errors: [{ id: 7, code: "permission_denied", message: "permission denied" }],
    });

    renderPage();
    fireEvent.click(screen.getByTestId("select-all"));
    fireEvent.click(screen.getByTestId("bulk-action-activate"));

    const panel = await screen.findByTestId("bulk-result");
    expect(panel).toHaveTextContent("Activate Warehouses: applied to 1 of 2 selected");
    expect(screen.getByTestId("bulk-result-item-7")).toHaveTextContent(
      "#7 — You don't have permission to change this record"
    );
    // Only the skipped record stays selected.
    expect(screen.getByText("1 selected")).toBeInTheDocument();
    expect(toastMock).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Partially applied", variant: "destructive" })
    );
  });

  it("does not report success when every record is rejected", async () => {
    executeActionMock.mockRejectedValue({
      message: "Request failed with status code 400",
      response: {
        status: 400,
        data: {
          success: false,
          affected: 0,
          message: "No permitted objects found for action",
          errors: [{ id: 7, code: "permission_denied", message: "permission denied" }],
        },
      },
    });

    renderPage();
    fireEvent.click(screen.getByTestId("select-7"));
    fireEvent.click(screen.getByTestId("bulk-action-activate"));

    await screen.findByTestId("bulk-result-item-7");
    await waitFor(() =>
      expect(toastMock).toHaveBeenCalledWith(
        expect.objectContaining({
          title: "Action failed",
          description: "No permitted objects found for action",
        })
      )
    );
  });

  it("clears the report after a fully successful action", async () => {
    executeActionMock.mockResolvedValue({ success: true, affected: 2, message: "ok" });

    renderPage();
    fireEvent.click(screen.getByTestId("select-all"));
    fireEvent.click(screen.getByTestId("bulk-action-activate"));

    await waitFor(() =>
      expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({ title: "Success" }))
    );
    expect(screen.queryByTestId("bulk-result")).toBeNull();
  });
});
