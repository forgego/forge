import React from "react";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import {
  ForgeAdmin,
  registerPage,
  registerFieldEditor,
  registerActionForm,
  getRegisteredPages,
  getRegisteredPage,
  clearRegisteredPages,
  getFieldEditor,
  clearRegisteredFieldEditors,
  getActionForm,
  clearRegisteredActionForms,
} from "./api";
import { loadExtensions, getExtensionFailures, clearExtensionFailures } from "./loader";
import { ExtensionBoundary } from "./ExtensionBoundary";
import { ExtensionErrorBanner } from "./ExtensionErrorBanner";
import { widgetRegistry } from "../widgets";
import ExtensionPageRoute from "../../pages/ExtensionPageRoute";

// Mock router hooks for testing ExtensionPageRoute
const mockLocation = vi.hoisted(() => ({ pathname: "/ext/demo-ext/overview" }));
const mockParams = vi.hoisted(() => ({ _splat: "demo-ext/overview" }));

vi.mock("@tanstack/react-router", () => ({
  useLocation: () => mockLocation,
  useParams: () => mockParams,
  useNavigate: () => vi.fn(),
  Link: ({ children, to }: any) => <a href={to}>{children}</a>,
  createFileRoute: () => (config: any) => config,
}));

vi.mock("../../api/hooks/adminHooks", () => ({
  useModels: () => ({ data: { models: [] }, isLoading: false }),
  useConfig: () => ({ data: { plugins: [] }, isLoading: false }),
  useLogout: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe("Admin UI Extensions", () => {
  beforeEach(() => {
    clearRegisteredPages();
    clearRegisteredFieldEditors();
    clearRegisteredActionForms();
    clearExtensionFailures();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("Registration API (window.ForgeAdmin)", () => {
    it("api frozen and registerWidget reaches widgetRegistry", () => {
      // 1. ForgeAdmin is exposed on window and is frozen
      expect(window.ForgeAdmin).toBeDefined();
      expect(window.ForgeAdmin).toBe(ForgeAdmin);
      expect(Object.isFrozen(window.ForgeAdmin)).toBe(true);
      expect(ForgeAdmin.version).toBe(1);
      expect(ForgeAdmin.react).toBe(React);
      expect(ForgeAdmin.ui).toBeDefined();
      expect(ForgeAdmin.ui.Button).toBeDefined();
      expect(React.isValidElement(<ForgeAdmin.ui.Button>Click</ForgeAdmin.ui.Button>)).toBe(true);

      // Modifying frozen object fails (throws in strict mode or is no-op)
      expect(() => {
        (window.ForgeAdmin as any).version = 999;
      }).toThrow();

      // 2. registerWidget delegates to widgetRegistry
      const TestWidget = () => <div>Test Widget Content</div>;
      window.ForgeAdmin?.registerWidget("custom-analytics", TestWidget);

      const registered = widgetRegistry.get("custom-analytics");
      expect(registered).toBeDefined();

      // Wrapped component renders properly
      const { getByText } = render(React.createElement(registered! as any));
      expect(getByText("Test Widget Content")).toBeInTheDocument();
    });

    it("registers field editors and action forms with registries", () => {
      const CustomFieldEditor = () => <div>Custom Field Editor</div>;
      ForgeAdmin.registerFieldEditor(
        { model: "products", field: "category" },
        CustomFieldEditor
      );

      const editor = getFieldEditor("products", "category");
      expect(editor).toBeDefined();
      const { getByText } = render(React.createElement(editor!));
      expect(getByText("Custom Field Editor")).toBeInTheDocument();

      const CustomActionForm = () => <div>Custom Action Form</div>;
      ForgeAdmin.registerActionForm("bulk_archive", CustomActionForm);

      const actionForm = getActionForm("bulk_archive");
      expect(actionForm).toBeDefined();
      const renderedForm = render(React.createElement(actionForm!));
      expect(renderedForm.getByText("Custom Action Form")).toBeInTheDocument();
    });
  });

  describe("Extension Loader", () => {
    it("loader calls default(api), records a failing import and an unsupported version as failures without throwing", async () => {
      const consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
      const consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      let apiReceivedByExtension: any = null;
      // We can use an inline module data URL or mocked file
      const initFunction = (api: any) => {
        apiReceivedByExtension = api;
        api.registerWidget("data-url-widget", () => "Data Widget");
      };

      // Create a test data URL module that calls initFunction
      const moduleCode = `
        let callback;
        export function setCallback(cb) { callback = cb; }
        export default function(api) {
          if (globalThis.__testExtensionInit) {
            globalThis.__testExtensionInit(api);
          }
        }
      `;
      (globalThis as any).__testExtensionInit = initFunction;
      const validEntryUrl = `data:text/javascript;base64,${Buffer.from(moduleCode).toString("base64")}`;

      const failures = await loadExtensions({
        models: [],
        plugins: [],
        extensions: [
          // 1. Valid extension with supported version
          {
            name: "valid-extension",
            entry: validEntryUrl,
            version: 1,
          },
          // 2. Unsupported version
          {
            name: "unsupported-ext",
            entry: "/ext/unsupported/index.js",
            version: 99,
          },
          // 3. Failing import
          {
            name: "failing-ext",
            entry: "invalid://protocol/non-existent-module-path.js",
            version: 1,
          },
        ],
      });

      // loader called default(api) for the valid extension
      expect(apiReceivedByExtension).toBe(ForgeAdmin);
      expect(widgetRegistry.get("data-url-widget")).toBeDefined();

      // Recorded unsupported version as failure without throwing
      const unsupported = failures.find((f) => f.name === "unsupported-ext");
      expect(unsupported).toBeDefined();
      expect(unsupported?.reason).toBe("unsupported_version");
      expect(unsupported?.version).toBe(99);

      // Recorded failing import as failure without throwing
      const failing = failures.find((f) => f.name === "failing-ext");
      expect(failing).toBeDefined();
      expect(failing?.reason).toBe("import_failed");

      // Verify getExtensionFailures() matches
      expect(getExtensionFailures()).toHaveLength(2);

      delete (globalThis as any).__testExtensionInit;
      consoleWarnSpy.mockRestore();
      consoleErrorSpy.mockRestore();
    });
  });

  describe("ExtensionBoundary", () => {
    it("boundary renders fallback when a component throws", () => {
      const consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      const ThrowingComponent = () => {
        throw new Error("Simulated extension component explosion");
      };

      render(
        <ExtensionBoundary extensionName="failing-plugin">
          <ThrowingComponent />
        </ExtensionBoundary>
      );

      // Fallback is rendered
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(
        screen.getByText(/Extension Component Error \(failing-plugin\)/)
      ).toBeInTheDocument();
      expect(
        screen.getByText("Simulated extension component explosion")
      ).toBeInTheDocument();

      consoleErrorSpy.mockRestore();
    });

    it("renders custom fallback if provided", () => {
      const consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

      const ThrowingComponent = () => {
        throw new Error("Custom error");
      };

      render(
        <ExtensionBoundary fallback={<div>Custom Fallback View</div>}>
          <ThrowingComponent />
        </ExtensionBoundary>
      );

      expect(screen.getByText("Custom Fallback View")).toBeInTheDocument();
      consoleErrorSpy.mockRestore();
    });
  });

  describe("Extension Error Banner", () => {
    it("renders non-blocking banner when failures exist", () => {
      const mockFailures = [
        {
          name: "crm-integration",
          version: 1,
          reason: "import_failed" as const,
          error: new Error("Failed to fetch"),
        },
        {
          name: "legacy-v0-extension",
          version: 0,
          reason: "unsupported_version" as const,
          error: new Error("Version 0 unsupported"),
        },
      ];

      const { rerender } = render(<ExtensionErrorBanner failures={mockFailures} />);

      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(
        screen.getByText(/crm-integration, legacy-v0-extension/)
      ).toBeInTheDocument();

      // Empty failures renders null
      rerender(<ExtensionErrorBanner failures={[]} />);
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("Page Registration and Routing", () => {
    it("page registered → route renders", () => {
      // 1. Register an extension page
      const ExtensionPage = () => (
        <div data-testid="extension-page-rendered">
          <h1>Demo Extension Overview</h1>
          <p>Welcome to the custom analytics extension.</p>
        </div>
      );

      ForgeAdmin.registerPage(
        {
          path: "overview",
          title: "Demo Overview",
          nav: { label: "Analytics Overview", order: 1 },
          extension: "demo-ext",
        },
        ExtensionPage
      );

      // Verify registry has fullPath /ext/demo-ext/overview
      const pages = getRegisteredPages();
      expect(pages).toHaveLength(1);
      expect(pages[0].fullPath).toBe("/ext/demo-ext/overview");
      expect(pages[0].title).toBe("Demo Overview");
      expect(pages[0].nav?.label).toBe("Analytics Overview");

      // 2. Render route under /ext/<name>/<path>
      mockLocation.pathname = "/ext/demo-ext/overview";
      mockParams._splat = "demo-ext/overview";

      render(<ExtensionPageRoute />);

      // The registered component renders within the shell
      expect(screen.getByTestId("extension-page-rendered")).toBeInTheDocument();
      expect(screen.getByText("Demo Extension Overview")).toBeInTheDocument();
      expect(
        screen.getByText("Welcome to the custom analytics extension.")
      ).toBeInTheDocument();
    });

    it("renders page not found for unregistered ext routes", () => {
      mockLocation.pathname = "/ext/unknown-ext/missing";
      mockParams._splat = "unknown-ext/missing";

      render(<ExtensionPageRoute />);

      expect(screen.getByText("Extension Page Not Found")).toBeInTheDocument();
      expect(
        screen.getByText(/No extension page is registered at \/ext\/unknown-ext\/missing/)
      ).toBeInTheDocument();
    });
  });
});
