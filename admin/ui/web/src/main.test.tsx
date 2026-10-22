import React from "react";
import { describe, expect, it, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";

const { bootstrapMock, createRouterMock } = vi.hoisted(() => ({
  bootstrapMock: vi.fn(),
  createRouterMock: vi.fn((options) => options),
}));

vi.mock("./lib/boot", () => ({ bootstrapAdmin: bootstrapMock }));
vi.mock("./routeTree.gen", () => ({ routeTree: { id: "root" } }));
vi.mock("@tanstack/react-router", () => ({
  createRouter: createRouterMock,
  RouterProvider: () => <main>Admin router ready</main>,
}));
vi.mock("./components/providers/ThemeProvider", () => ({
  ThemeProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));
vi.mock("./components/ui/tooltip", () => ({
  TooltipProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

describe("Admin boot", () => {
  it("starts React 19 with a TanStack router at the admin root", async () => {
    expect(React.version.split(".")[0]).toBe("19");
    const routerPackage = await vi.importActual<typeof import("@tanstack/react-router")>("@tanstack/react-router");
    expect(routerPackage.createRouter).toBeTypeOf("function");
    document.body.innerHTML = '<div id="root"></div>';

    await import("./main");

    expect(bootstrapMock).toHaveBeenCalledOnce();
    expect(createRouterMock).toHaveBeenCalledWith(
      expect.objectContaining({ routeTree: { id: "root" }, basepath: expect.any(String) })
    );
    await waitFor(() => expect(screen.getByRole("main")).toHaveTextContent("Admin router ready"));
  });
});
