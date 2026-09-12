import { useState, useEffect } from "react";
import { Moon, Sun, Monitor, Settings, X, RotateCcw } from "lucide-react";
import { useThemeStore, DEFAULT_THEME } from "@/store/themeStore";
import { primaries } from "@/lib/themes";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const RADIUS_OPTIONS = [0, 0.3, 0.5, 0.625, 0.75, 1.0];

export function ThemeCustomizer() {
  const [isOpen, setIsOpen] = useState(false);
  const { theme, setTheme, primary, setPrimary, radius, setRadius, resetTheme } =
    useThemeStore();

  useEffect(() => {
    if (!isOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setIsOpen(false);
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [isOpen]);

  return (
    <>
      <Button
        variant="ghost"
        size="icon"
        data-testid="theme-settings-trigger"
        onClick={() => setIsOpen(true)}
        aria-label="Open theme settings"
        title="Theme settings"
        className="text-muted-foreground hover:text-foreground"
      >
        <Settings className="h-4 w-4" />
      </Button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm"
            onClick={() => setIsOpen(false)}
          />
          <div
            role="dialog"
            aria-modal="true"
            aria-label="Theme settings"
            data-testid="theme-settings"
            className="fixed inset-y-0 right-0 z-50 w-80 bg-background border-l shadow-2xl p-6 overflow-y-auto animate-in slide-in-from-right duration-300"
          >
            {/* Header */}
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-semibold text-lg">Theme Settings</h2>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setIsOpen(false)}
                aria-label="Close theme settings"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            {/* Mode */}
            <div className="mb-8">
              <h3 className="font-medium mb-4">Mode</h3>
              <div className="grid grid-cols-3 gap-2">
                <Button
                  variant={theme === "light" ? "default" : "outline"}
                  onClick={() => setTheme("light")}
                  className="justify-start"
                  aria-pressed={theme === "light"}
                >
                  <Sun className="mr-2 h-4 w-4" /> Light
                </Button>
                <Button
                  variant={theme === "dark" ? "default" : "outline"}
                  onClick={() => setTheme("dark")}
                  className="justify-start"
                  aria-pressed={theme === "dark"}
                >
                  <Moon className="mr-2 h-4 w-4" /> Dark
                </Button>
                <Button
                  variant={theme === "system" ? "default" : "outline"}
                  onClick={() => setTheme("system")}
                  className="justify-start"
                  aria-pressed={theme === "system"}
                >
                  <Monitor className="mr-2 h-4 w-4" /> Auto
                </Button>
              </div>
            </div>

            {/* Colors */}
            <div className="mb-8">
              <h3 className="font-medium mb-4">Color</h3>
              <div className="grid grid-cols-2 gap-2">
                {primaries.map((p) => (
                  <Button
                    key={p.name}
                    variant="outline"
                    className={cn(
                      "justify-start px-2",
                      primary === p.name && "border-primary ring-1 ring-primary"
                    )}
                    onClick={() => setPrimary(p.name)}
                    aria-pressed={primary === p.name}
                  >
                    <span
                      className="h-4 w-4 rounded-full mr-2 shrink-0"
                      style={{ backgroundColor: `hsl(${p.active})` }}
                    />
                    <span className="text-xs">{p.label}</span>
                  </Button>
                ))}
              </div>
            </div>

            {/* Radius */}
            <div className="mb-8">
              <h3 className="font-medium mb-4">Corner radius</h3>
              <div className="grid grid-cols-6 gap-2" role="group" aria-label="Corner radius">
                {RADIUS_OPTIONS.map((r) => (
                  <Button
                    key={r}
                    variant="outline"
                    size="sm"
                    data-testid={`radius-${r}`}
                    className={cn(
                      radius === r && "border-primary ring-1 ring-primary"
                    )}
                    onClick={() => setRadius(r)}
                    aria-pressed={radius === r}
                    title={`${r}rem`}
                  >
                    {r}
                  </Button>
                ))}
              </div>
              <p className="mt-2 text-[11px] text-muted-foreground">
                Radius in rem units. Current: {radius}rem.
              </p>
            </div>

            <Button
              variant="ghost"
              className="w-full gap-2 text-muted-foreground"
              onClick={() => resetTheme()}
              title="Reset to defaults"
            >
              <RotateCcw className="h-4 w-4" />
              Reset to defaults ({DEFAULT_THEME.theme}, {DEFAULT_THEME.primary},{" "}
              {DEFAULT_THEME.radius}rem)
            </Button>
          </div>
        </>
      )}
    </>
  );
}
