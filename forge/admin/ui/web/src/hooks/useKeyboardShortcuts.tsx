import { useEffect, useRef } from 'react';
import { useNavigate } from '@tanstack/react-router';

interface KeyboardShortcut {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  action: () => void;
  description?: string;
}

interface UseKeyboardShortcutsOptions {
  enabled?: boolean;
  global?: boolean;
}

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
      // Don't trigger shortcuts when typing in input fields
      const target = e.target as HTMLElement;
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.isContentEditable
      ) {
        return;
      }

      for (const shortcut of shortcutsRef.current) {
        const ctrlMatch = shortcut.ctrl ? e.ctrlKey || e.metaKey : !e.ctrlKey && !e.metaKey;
        const shiftMatch = shortcut.shift ? e.shiftKey : !e.shiftMatch;
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

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [enabled, global]);
}

// Predefined shortcuts for admin
export function useAdminShortcuts() {
  const navigate = useNavigate();

  const shortcuts: KeyboardShortcut[] = [
    {
      key: 'h',
      ctrl: true,
      action: () => navigate({ to: '/' }),
      description: 'Go to Dashboard',
    },
    {
      key: 'n',
      ctrl: true,
      shift: true,
      action: () => {
        const path = window.location.pathname;
        if (!path.includes('/new')) {
          navigate({ to: `${path}/new` });
        }
      },
      description: 'Create new object',
    },
    {
      key: 'Escape',
      action: () => {
        navigate({ to: '/' });
      },
      description: 'Go back to Dashboard',
    },
    {
      key: '/',
      ctrl: true,
      action: () => {
        document.querySelector<HTMLInputElement>('[data-testid="search-input"]')?.focus();
      },
      description: 'Focus search',
    },
    {
      key: 's',
      ctrl: true,
      action: () => {
        // Trigger save if in form
        const saveButton = document.querySelector<HTMLButtonElement>('[data-testid="save-button"]');
        saveButton?.click();
      },
      description: 'Save form',
    },
  ];

  useKeyboardShortcuts(shortcuts, { global: true });
}

// Keyboard shortcut help dialog
export function ShortcutHelpDialog({
  open,
  onOpenChange,
  shortcuts,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  shortcuts: KeyboardShortcut[];
}) {
  const formatKey = (shortcut: KeyboardShortcut) => {
    const parts: string[] = [];
    if (shortcut.ctrl) parts.push('Ctrl');
    if (shortcut.alt) parts.push('Alt');
    if (shortcut.shift) parts.push('Shift');
    parts.push(shortcut.key.toUpperCase());
    return parts.join(' + ');
  };

  return (
    <div
      className={`fixed inset-0 z-50 flex items-center justify-center bg-black/50 transition-opacity ${
        open ? 'opacity-100' : 'opacity-0 pointer-events-none'
      }`}
      onClick={() => onOpenChange(false)}
    >
      <div
        className="bg-background rounded-lg p-6 w-full max-w-md shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold mb-4">Keyboard Shortcuts</h2>
        <div className="space-y-2">
          {shortcuts.map((shortcut, index) => (
            <div
              key={index}
              className="flex items-center justify-between py-2 border-b last:border-0"
            >
              <span className="text-sm text-muted-foreground">
                {shortcut.description}
              </span>
              <kbd className="px-2 py-1 bg-muted rounded text-sm font-mono">
                {formatKey(shortcut)}
              </kbd>
            </div>
          ))}
        </div>
        <div className="mt-4 text-xs text-muted-foreground text-center">
          Press ? to show this dialog
        </div>
      </div>
    </div>
  );
}

// Hook for showing keyboard shortcuts dialog
export function useShortcutHelp() {
  const [showHelp, setShowHelp] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === '?' && !e.target) {
        setShowHelp(true);
      }
      if (e.key === 'Escape') {
        setShowHelp(false);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return { showHelp, setShowHelp };
}

import { useState } from 'react';
