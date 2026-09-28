import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import DynamicModelPage from "./DynamicModelPage";

const useModelMetadataMock = vi.hoisted(() => vi.fn());
const useParamsMock = vi.hoisted(() => vi.fn());

vi.mock("../api/hooks/adminHooks", () => ({
  useModelMetadata: useModelMetadataMock,
}));

vi.mock("@tanstack/react-router", () => ({
  useParams: useParamsMock,
}));

vi.mock("../pages/ModelListPage", () => ({
  default: () => <div>LIST_PAGE</div>,
}));

vi.mock("../pages/ModelUpsertPage", () => ({
  default: ({ mode }: { mode: string }) => <div>UPSERT_{mode}</div>,
}));

vi.mock("../pages/ModelViewPage", () => ({
  default: () => <div>VIEW_PAGE</div>,
}));

describe("DynamicModelPage", () => {
  beforeEach(() => {
    useParamsMock.mockReturnValue({ model: "products" });
    useModelMetadataMock.mockReturnValue({
      data: { name: "products" },
      isLoading: false,
      error: null,
    });
  });

  it("renders loading state while metadata is loading", () => {
    useModelMetadataMock.mockReturnValue({
      data: null,
      isLoading: true,
      error: null,
    });

    const { container } = render(<DynamicModelPage mode="list" />);

    expect(container.querySelector(".animate-spin")).toBeInTheDocument();
  });

  it("renders error state when metadata fails", () => {
    useModelMetadataMock.mockReturnValue({
      data: null,
      isLoading: false,
      error: new Error("failed"),
    });

    render(<DynamicModelPage mode="list" />);

    expect(screen.getByText('Error loading metadata for "products"')).toBeInTheDocument();
  });

  it("renders list mode", () => {
    render(<DynamicModelPage mode="list" />);
    expect(screen.getByText("LIST_PAGE")).toBeInTheDocument();
  });

  it("renders create mode", () => {
    render(<DynamicModelPage mode="create" />);
    expect(screen.getByText("UPSERT_create")).toBeInTheDocument();
  });

  it("renders edit mode", () => {
    render(<DynamicModelPage mode="edit" />);
    expect(screen.getByText("UPSERT_edit")).toBeInTheDocument();
  });

  it("renders detail mode with model view page", () => {
    render(<DynamicModelPage mode="detail" />);
    expect(screen.getByText("VIEW_PAGE")).toBeInTheDocument();
  });
});
