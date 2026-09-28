import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { Breadcrumbs } from "./Breadcrumbs";

const useNavigateMock = vi.hoisted(() => vi.fn());
const useParamsMock = vi.hoisted(() => vi.fn());
const useLocationMock = vi.hoisted(() => vi.fn());

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => useNavigateMock,
  useParams: () => useParamsMock(),
  useLocation: () => useLocationMock(),
}));

describe("Breadcrumbs", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useParamsMock.mockReturnValue({});
    useLocationMock.mockReturnValue({ pathname: "/" });
  });

  it("renders home breadcrumb at root path", () => {
    render(<Breadcrumbs />);
    expect(screen.getByText("Home")).toBeInTheDocument();
  });

  it("renders model list breadcrumbs automatically from route params", () => {
    useParamsMock.mockReturnValue({ model: "products" });
    useLocationMock.mockReturnValue({ pathname: "/products" });

    render(<Breadcrumbs />);
    expect(screen.getByText("Home")).toBeInTheDocument();
    expect(screen.getByText("Products")).toBeInTheDocument();
  });

  it("renders create mode breadcrumb when in create route", () => {
    useParamsMock.mockReturnValue({ model: "products" });
    useLocationMock.mockReturnValue({ pathname: "/products/create" });

    render(<Breadcrumbs />);
    expect(screen.getByText("Products")).toBeInTheDocument();
    expect(screen.getByText("Create")).toBeInTheDocument();
  });

  it("renders detail mode breadcrumb when in view route", () => {
    useParamsMock.mockReturnValue({ model: "products", id: "42" });
    useLocationMock.mockReturnValue({ pathname: "/products/42/view" });

    render(<Breadcrumbs />);
    expect(screen.getByText("Products")).toBeInTheDocument();
    expect(screen.getByText("View")).toBeInTheDocument();
  });

  it("renders custom breadcrumb items when provided", () => {
    render(
      <Breadcrumbs
        customItems={[
          { label: "Dashboard", path: "/" },
          { label: "Settings" },
        ]}
      />
    );
    expect(screen.getByText("Dashboard")).toBeInTheDocument();
    expect(screen.getByText("Settings")).toBeInTheDocument();
  });
});
