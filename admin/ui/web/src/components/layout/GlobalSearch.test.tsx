import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";
import { GlobalSearch } from "./GlobalSearch";

const navigateMock = vi.hoisted(() => vi.fn());
const globalSearchMock = vi.hoisted(() => vi.fn());

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => navigateMock,
}));

vi.mock("../../api/client", () => ({
  adminAPI: {
    globalSearch: (...args: any[]) => globalSearchMock(...args),
  },
}));

const models = [
  {
    name: "products",
    verbose_name: "Product",
    verbose_name_plural: "Products",
    count: 10,
    permissions: { view: true, add: true, change: true, delete: true },
  },
  {
    name: "orders",
    verbose_name: "Order",
    verbose_name_plural: "Orders",
    count: 5,
    permissions: { view: true, add: true, change: true, delete: true },
  },
];

describe("GlobalSearch", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    globalSearchMock.mockResolvedValue({ results: [] });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("opens the palette and lists models", () => {
    render(<GlobalSearch models={models as any} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

    fireEvent.click(screen.getByTestId("global-search-trigger"));
    expect(
      screen.getByRole("dialog", { name: "Global search" })
    ).toBeInTheDocument();
    expect(screen.getByTestId("global-search-input")).toBeInTheDocument();
    expect(screen.getByText("Products")).toBeInTheDocument();
    expect(screen.getByText("Orders")).toBeInTheDocument();
  });

  it("filters the model list as the user types", () => {
    render(<GlobalSearch models={models as any} />);
    fireEvent.click(screen.getByTestId("global-search-trigger"));

    fireEvent.change(screen.getByTestId("global-search-input"), {
      target: { value: "ord" },
    });
    expect(screen.queryByText("Products")).not.toBeInTheDocument();
    expect(screen.getByText("Orders")).toBeInTheDocument();
  });

  it("shows record results from the API and navigates with Enter", async () => {
    globalSearchMock.mockResolvedValue({
      results: [
        {
          model: "products",
          count: 1,
          items: [{ id: 7, title: "Laptop Pro", url: "/admin/products/7/view" }],
        },
      ],
    });

    render(<GlobalSearch models={models as any} />);
    fireEvent.click(screen.getByTestId("global-search-trigger"));
    fireEvent.change(screen.getByTestId("global-search-input"), {
      target: { value: "laptop" },
    });

    await act(async () => {
      vi.advanceTimersByTime(300);
    });

    expect(globalSearchMock).toHaveBeenCalledWith({ query: "laptop" });
    expect(screen.getByText("Laptop Pro")).toBeInTheDocument();

    fireEvent.keyDown(screen.getByTestId("global-search-input"), {
      key: "Enter",
    });
    expect(navigateMock).toHaveBeenCalledWith({ to: "/products/7/view" });
  });

  it("closes on Escape", () => {
    render(<GlobalSearch models={models as any} />);
    fireEvent.click(screen.getByTestId("global-search-trigger"));
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    fireEvent.keyDown(screen.getByTestId("global-search-input"), {
      key: "Escape",
    });
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
