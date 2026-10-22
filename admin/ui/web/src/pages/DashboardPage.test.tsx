import React from "react";
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import DashboardPage from "./DashboardPage";

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => vi.fn(),
  Link: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
vi.mock("../api/hooks/adminHooks", () => ({
  useConfig: () => ({
    data: {
      dashboard: {
        Layout: {
          type: "chart",
          props: {
            title: "Orders by day",
            chartType: "line",
            data: [
              { name: "Mon", orders: 4 },
              { name: "Tue", orders: 7 },
            ],
          },
        },
      },
    },
    isLoading: false,
    error: null,
  }),
  useMetadata: () => ({ data: { models: [], plugins: [] }, isLoading: false }),
}));
vi.mock("../components/layout/AdminLayout", () => ({
  default: ({ children }: { children: React.ReactNode }) => <main>{children}</main>,
}));
vi.mock("recharts", async (importOriginal) => {
  const actual = await importOriginal<typeof import("recharts")>();
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactElement }) =>
      React.cloneElement(children, { width: 640, height: 300 } as React.Attributes),
  };
});

describe("DashboardPage", () => {
  it("renders a configured Recharts analytics chart", async () => {
    const { container } = render(<DashboardPage />);

    expect(await screen.findByText("Orders by day")).toBeInTheDocument();
    await waitFor(() => {
      expect(container.querySelector(".recharts-line")).toBeInTheDocument();
    });
  });
});
