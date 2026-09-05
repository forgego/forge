import { useState } from "react";
import { Moon, Sun, Settings, X } from "lucide-react";
import { useThemeStore } from "@/store/themeStore";
import { primaries } from "@/lib/themes";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export function ThemeCustomizer() {
  const [isOpen, setIsOpen] = useState(false);
  const { theme, setTheme, primary, setPrimary, radius, setRadius } =
    useThemeStore();

  return (
    <>
      <Button variant="ghost" size="icon" onClick={() => setIsOpen(true)}>
        <Settings className="h-4 w-4" />
      </Button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm"
            onClick={() => setIsOpen(false)}
          />
          <div className="fixed inset-y-0 right-0 z-50 w-80 bg-background border-l shadow-2xl p-6 overflow-y-auto animate-in slide-in-from-right duration-300">
            {/* Header */}
            <div className="flex items-center justify-between mb-8">
              <h2 className="font-semibold text-lg">Theme Settings</h2>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setIsOpen(false)}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            {/* Mode */}
            <div className="mb-8">
              <h3 className="font-medium mb-4">Mode</h3>
              <div className="grid grid-cols-2 gap-2">
                <Button
                  variant={theme === "light" ? "default" : "outline"}
                  onClick={() => setTheme("light")}
                  className="justify-start"
                >
                  <Sun className="mr-2 h-4 w-4" /> Light
                </Button>
                <Button
                  variant={theme === "dark" ? "default" : "outline"}
                  onClick={() => setTheme("dark")}
                  className="justify-start"
                >
                  <Moon className="mr-2 h-4 w-4" /> Dark
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
              <h3 className="font-medium mb-4">Radius</h3>
              <div className="grid grid-cols-5 gap-2">
                {[0, 0.3, 0.5, 0.75, 1.0].map((r) => (
                  <Button
                    key={r}
                    variant="outline"
                    size="sm"
                    className={cn(
                      radius === r && "border-primary ring-1 ring-primary"
                    )}
                    onClick={() => setRadius(r)}
                  >
                    {r}
                  </Button>
                ))}
              </div>
            </div>
          </div>
        </>
      )}
    </>
  );
}
