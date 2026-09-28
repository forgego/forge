import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, within } from "@testing-library/react";
import ModelRegistryPage from "./ModelRegistryPage";

const useModelsMock = vi.hoisted(() => vi.fn());

vi.mock("@tanstack/react-router", () => ({
  Link: ({ children, to }: any) => <a href={to}>{children}</a>,
  useNavigate: () => vi.fn(),
  useLocation: () => ({ pathname: "/models" }),
  useParams: () => ({}),
}));

vi.mock("../api/hooks/adminHooks", () => ({
  useModels: () => useModelsMock(),
  useConfig: () => ({ data: {} }),
  useLogout: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe("ModelRegistryPage", () => {
  const mockModels = [
    {
      name: "products",
      verbose_name: "Product",
      verbose_name_plural: "Products",
      count: 42,
      permissions: { view: true, add: true, change: true, delete: true },
    },
    {
      name: "orders",
      verbose_name: "Order",
      verbose_name_plural: "Orders",
      count: 15,
      permissions: { view: true, add: false, change: true, delete: false },
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    useModelsMock.mockReturnValue({
      data: { models: mockModels },
      isLoading: false,
      error: null,
    });
  });

  it("renders model registry with search and model cards", () => {
    render(<ModelRegistryPage />);

    const main = within(screen.getByRole("main"));
    expect(main.getByText("Model Registry")).toBeInTheDocument();
    expect(main.getByText("Products")).toBeInTheDocument();
    expect(main.getByText("Orders")).toBeInTheDocument();
    expect(main.getByText("42 records")).toBeInTheDocument();
  });

  it("filters models by search input", () => {
    render(<ModelRegistryPage />);

    const searchInput = screen.getByPlaceholderText("Search registered models...");
    fireEvent.change(searchInput, { target: { value: "prod" } });

    const main = within(screen.getByRole("main"));
    expect(main.getByText("Products")).toBeInTheDocument();
    expect(main.queryByText("Orders")).not.toBeInTheDocument();
  });
});
