import React from "react";
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import DashboardPage from "./DashboardPage";

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => vi.fn(),
  Link: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
vi.mock("../api/hooks/adminHooks", () => ({
  useConfig: () => ({ data: {}, isLoading: false, error: null }),
  useMetadata: () => ({
    data: { models: [{
      name: "orders", verbose_name: "Order", verbose_name_plural: "Orders", count: 7,
      permissions: { view: true, add: true, change: true, delete: true },
    }], plugins: [] },
    isLoading: false,
  }),
}));
vi.mock("../components/layout/AdminLayout", () => ({
  default: ({ children }: { children: React.ReactNode }) => <main>{children}</main>,
}));

describe("Default dashboard", () => {
  it("shows model and record summaries without an analytics chart", () => {
    const { container } = render(<DashboardPage />);
    expect(screen.getByText("Registered Models")).toBeInTheDocument();
    expect(screen.getByText("Total Records")).toBeInTheDocument();
    expect(screen.getByText("Orders")).toBeInTheDocument();
    expect(container.querySelector(".recharts-wrapper")).toBeNull();
  });
});
