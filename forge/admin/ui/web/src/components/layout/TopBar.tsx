import { Button } from "../ui/button";
import { Menu, ChevronRight, Keyboard } from "lucide-react";
import { GlobalSearch } from "./GlobalSearch";
import { ThemeCustomizer } from "../../features/theme/ThemeCustomizer";

export interface TopBarProps {
  models: any[];
  userName: string;
  userRole: string;
  userInitials: string;
  onOpenMobileNav: () => void;
  onShowHelp: () => void;
  showPluginHeader: boolean;
  pluginLabel?: string;
  entryLabel?: string;
}

export function TopBar({
  models,
  userName,
  userRole,
  userInitials,
  onOpenMobileNav,
  onShowHelp,
  showPluginHeader,
  pluginLabel,
  entryLabel,
}: TopBarProps) {
  return (
        <header className="h-16 border-b border-border/50 flex items-center px-4 sm:px-6 bg-card/80 backdrop-blur-md sticky top-0 z-30 shrink-0 gap-3">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden text-muted-foreground shrink-0"
            onClick={onOpenMobileNav}
            aria-label="Open navigation"
          >
            <Menu className="h-5 w-5" />
          </Button>

          <div className="min-w-0 flex-1 max-w-md">
            <GlobalSearch models={models} />
          </div>

          {showPluginHeader && (
            <nav
              aria-label="Breadcrumb"
              className="hidden xl:flex items-center gap-1.5 text-xs text-muted-foreground min-w-0"
            >
              <span className="font-semibold uppercase tracking-wider">Plugin</span>
              <ChevronRight className="h-3 w-3 shrink-0" aria-hidden />
              <span className="text-foreground text-sm font-medium truncate">
                {pluginLabel}
              </span>
              {entryLabel && (
                <>
                  <ChevronRight className="h-3 w-3 shrink-0" aria-hidden />
                  <span className="text-sm truncate">{entryLabel}</span>
                </>
              )}
            </nav>
          )}

          <div className="ml-auto flex items-center gap-1 sm:gap-2 shrink-0">
            <Button
              variant="ghost"
              size="icon"
              className="text-muted-foreground hover:text-foreground"
              onClick={onShowHelp}
              title="Keyboard Shortcuts (?)"
              aria-label="Keyboard Shortcuts"
            >
              <Keyboard className="h-4 w-4" />
            </Button>
            <ThemeCustomizer />
            <div className="h-6 w-[1px] bg-border mx-1 sm:mx-2" aria-hidden />
            <div className="flex items-center gap-3 pl-1 sm:pl-2">
              <div className="text-right hidden md:block">
                <p className="text-sm font-semibold leading-none text-foreground">
                  {userName}
                </p>
                <p className="text-micro text-muted-foreground mt-1 uppercase font-bold tracking-wider">
                  {userRole}
                </p>
              </div>
              <div
                className="w-8 h-8 rounded-full bg-primary/10 border border-primary/20 flex items-center justify-center text-primary text-xs font-bold"
                title={`${userName} (${userRole})`}
                aria-hidden
              >
                {userInitials}
              </div>
            </div>
          </div>
        </header>
  );
}
