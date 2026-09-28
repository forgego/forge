import { useEffect, useRef, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Keyboard } from "lucide-react";
import {
  getAdminPrefix,
  stripAdminPrefix,
  withAdminPrefix,
} from "@/lib/admin-prefix";

export interface KeyboardShortcut {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  action?: () => void;
  description: string;
}

interface UseKeyboardShortcutsOptions {
  enabled?: boolean;
  global?: boolean;
}

export const DEFAULT_ADMIN_SHORTCUTS: KeyboardShortcut[] = [
  {
    key: "k",
    ctrl: true,
    description: "Quick Search",
  },
  {
    key: "/",
    description: "Focus Search (outside inputs)",
  },
  {
    key: "h",
    ctrl: true,
    description: "Go to Dashboard",
  },
  {
    key: "n",
    ctrl: true,
    shift: true,
    description: "Create New Record",
  },
  {
    key: "s",
    ctrl: true,
    description: "Save Active Form",
  },
  {
    key: "?",
    description: "Show Keyboard Shortcuts",
  },
  {
    key: "Escape",
    description: "Close Modal / Cancel",
  },
];

export function useKeyboardShortcuts(
  shortcuts: KeyboardShortcut[],
  options: UseKeyboardShortcutsOptions = {}
) {
  const { enabled = true, global = true } = options;
  const shortcutsRef = useRef(shortcuts);

  useEffect(() => {
    shortcutsRef.current = shortcuts;
  }, [shortcuts]);

  useEffect(() => {
    if (!enabled || !global) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Plain (non-modifier) shortcuts don't fire while typing; Ctrl/Cmd
      // combos (save, search) still work from inside inputs.
      const target = e.target as HTMLElement;
      const isTyping =
        target?.tagName === "INPUT" ||
        target?.tagName === "TEXTAREA" ||
        target?.isContentEditable;
      if (isTyping && !e.ctrlKey && !e.metaKey) {
        return;
      }

      for (const shortcut of shortcutsRef.current) {
        if (!shortcut.action) continue;
        const ctrlMatch = shortcut.ctrl
          ? e.ctrlKey || e.metaKey
          : !e.ctrlKey && !e.metaKey;
        const shiftMatch = shortcut.shift ? e.shiftKey : !e.shiftKey;
        const altMatch = shortcut.alt ? e.altKey : !e.altKey;

        if (
          e.key.toLowerCase() === shortcut.key.toLowerCase() &&
          ctrlMatch &&
          shiftMatch &&
          altMatch
        ) {
          e.preventDefault();
          shortcut.action();
          return;
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [enabled, global]);
}

// Predefined shortcuts for admin
export function useAdminShortcuts() {
  const navigate = useNavigate();

  const shortcuts: KeyboardShortcut[] = [
    {
      key: "h",
      ctrl: true,
      action: () => navigate({ to: "/" }),
      description: "Go to Dashboard",
    },
    {
      key: "n",
      ctrl: true,
      shift: true,
      action: () => {
        const path = window.location.pathname;
        const prefix = getAdminPrefix();
        const isInsideAdmin =
          stripAdminPrefix(path, prefix) !== path ||
          path === withAdminPrefix("/", prefix) ||
          prefix === "";
        // Never hijack the browser's new-window shortcut outside the admin.
        if (!isInsideAdmin) return;
        if (!path.includes("/create")) {
          navigate({ to: `${path}/create` });
        }
      },
      description: "Create New Record",
    },
    {
      key: "Escape",
      description: "Close Modal / Cancel",
    },
    {
      key: "/",
      ctrl: true,
      action: () => {
        document
          .querySelector<HTMLInputElement>('[data-testid="search-input"]')
          ?.focus();
      },
      description: "Focus Search",
    },
    {
      key: "s",
      ctrl: true,
      action: () => {
        // Skip when a modal is open: Ctrl+S there must not submit the
        // form behind the dialog.
        if (document.querySelector('[role="dialog"], [role="alertdialog"]')) {
          return;
        }
        const saveButton = document.querySelector<HTMLButtonElement>(
          '[data-testid="save-button"], [data-testid="submit-button"]'
        );
        saveButton?.click();
      },
      description: "Save Form",
    },
  ];

  useKeyboardShortcuts(shortcuts, { global: true });
}

// Keyboard shortcut help dialog
export function ShortcutHelpDialog({
  open,
  onOpenChange,
  shortcuts = DEFAULT_ADMIN_SHORTCUTS,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  shortcuts?: KeyboardShortcut[];
}) {
  const formatKey = (shortcut: KeyboardShortcut) => {
    const parts: string[] = [];
    if (shortcut.ctrl) parts.push("⌘/Ctrl");
    if (shortcut.alt) parts.push("Alt");
    if (shortcut.shift) parts.push("Shift");
    parts.push(shortcut.key.toUpperCase());
    return parts.join(" + ");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <Keyboard className="h-5 w-5 text-primary" />
            <DialogTitle>Keyboard Shortcuts</DialogTitle>
          </div>
          <DialogDescription>
            Speed up your administrative workflows with instant keyboard commands.
          </DialogDescription>
        </DialogHeader>

        <div className="divide-y divide-border/40 py-2">
          {shortcuts.map((shortcut, index) => (
            <div
              key={index}
              className="flex items-center justify-between py-2.5 text-xs"
            >
              <span className="text-muted-foreground font-medium">
                {shortcut.description}
              </span>
              <kbd className="px-2 py-1 bg-muted/80 border border-border/60 rounded text-micro font-mono font-semibold text-foreground shadow-xs">
                {formatKey(shortcut)}
              </kbd>
            </div>
          ))}
        </div>

        <div className="text-micro text-muted-foreground text-center pt-2 border-t border-border/40">
          Press <kbd className="px-1.5 py-0.5 bg-muted rounded font-mono">?</kbd> anywhere to open this cheat sheet
        </div>
      </DialogContent>
    </Dialog>
  );
}

// Hook for showing keyboard shortcuts dialog
export function useShortcutHelp() {
  const [showHelp, setShowHelp] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const isInput =
        target?.tagName === "INPUT" ||
        target?.tagName === "TEXTAREA" ||
        target?.isContentEditable;

      if (!isInput) {
        if (e.key === "?" || (e.key === "/" && (e.ctrlKey || e.metaKey))) {
          e.preventDefault();
          setShowHelp((prev) => !prev);
        }
      }

      if (e.key === "Escape") {
        setShowHelp(false);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  return { showHelp, setShowHelp };
}
