import { afterEach, describe, expect, it } from "vitest";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { ThemeProvider } from "./ThemeProvider";
import { ThemeCustomizer } from "../../features/theme/ThemeCustomizer";
import { useThemeStore } from "../../store/themeStore";

describe("Admin theme", () => {
  afterEach(() => {
    act(() => useThemeStore.getState().setTheme("system"));
    document.documentElement.classList.remove("light", "dark");
  });

  it("switches the admin between dark and light themes", () => {
    render(
      <ThemeProvider>
        <ThemeCustomizer />
      </ThemeProvider>
    );

    fireEvent.click(screen.getByRole("button", { name: "Open theme settings" }));
    fireEvent.click(screen.getByRole("button", { name: "Dark" }));
    expect(document.documentElement).toHaveClass("dark");
    expect(screen.getByRole("button", { name: "Dark" })).toHaveAttribute("aria-pressed", "true");

    fireEvent.click(screen.getByRole("button", { name: "Light" }));
    expect(document.documentElement).toHaveClass("light");
    expect(document.documentElement).not.toHaveClass("dark");
    expect(screen.getByRole("button", { name: "Light" })).toHaveAttribute("aria-pressed", "true");
  });
});
