import { describe, expect, it, vi } from "vitest";
import { render, screen, act } from "@testing-library/react";
import { useDebouncedValue } from "./useDebouncedValue";

function Probe({ value, delay }: { value: string; delay?: number }) {
  const debounced = useDebouncedValue(value, delay);
  return <span data-testid="out">{debounced}</span>;
}

describe("useDebouncedValue", () => {
  it("returns the value immediately on first render and debounces updates", () => {
    vi.useFakeTimers();
    try {
      const { rerender } = render(<Probe value="a" delay={200} />);
      expect(screen.getByTestId("out")).toHaveTextContent("a");

      rerender(<Probe value="ab" delay={200} />);
      // Not yet: still the old value.
      expect(screen.getByTestId("out")).toHaveTextContent("a");

      act(() => {
        vi.advanceTimersByTime(200);
      });
      expect(screen.getByTestId("out")).toHaveTextContent("ab");
    } finally {
      vi.useRealTimers();
    }
  });
});
