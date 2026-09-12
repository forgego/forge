import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { renderHook, act } from "@testing-library/react";
import {
  ShortcutHelpDialog,
  useShortcutHelp,
  DEFAULT_ADMIN_SHORTCUTS,
} from "./useKeyboardShortcuts";

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => vi.fn(),
}));

describe("useShortcutHelp", () => {
  it("opens help dialog when ? is pressed outside input", () => {
    const { result } = renderHook(() => useShortcutHelp());

    expect(result.current.showHelp).toBe(false);

    act(() => {
      window.dispatchEvent(new KeyboardEvent("keydown", { key: "?" }));
    });

    expect(result.current.showHelp).toBe(true);

    act(() => {
      window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    });

    expect(result.current.showHelp).toBe(false);
  });
});

describe("ShortcutHelpDialog", () => {
  it("renders shortcuts dialog with shortcut keys", () => {
    const onOpenChange = vi.fn();
    render(
      <ShortcutHelpDialog
        open={true}
        onOpenChange={onOpenChange}
        shortcuts={DEFAULT_ADMIN_SHORTCUTS}
      />
    );

    expect(screen.getByText("Keyboard Shortcuts")).toBeInTheDocument();
    expect(screen.getByText("Go to Dashboard")).toBeInTheDocument();
    expect(screen.getByText("Quick Search")).toBeInTheDocument();
  });
});
