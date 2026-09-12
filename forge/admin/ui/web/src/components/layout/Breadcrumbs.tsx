import { useNavigate, useParams, useLocation } from '@tanstack/react-router';
import { ChevronRight, Home, Database, Plus, Edit, Eye, Sparkles, Layers } from 'lucide-react';
import { Button } from '../ui/button';
import { cn } from '../../lib/utils';

export interface BreadcrumbItem {
  label: string;
  path?: string;
  icon?: React.ReactNode;
  action?: () => void;
}

export interface BreadcrumbsProps {
  model?: string;
  modelLabel?: string;
  mode?: 'list' | 'create' | 'edit' | 'detail';
  customItems?: BreadcrumbItem[];
  className?: string;
}

export function Breadcrumbs({
  model: propModel,
  modelLabel: propModelLabel,
  mode: propMode,
  customItems,
  className,
}: BreadcrumbsProps) {
  const navigate = useNavigate();
  const params = useParams({ strict: false }) as Record<string, string>;
  const location = useLocation();

  const model = propModel || params?.model;
  const modelLabel =
    propModelLabel ||
    (model ? model.charAt(0).toUpperCase() + model.slice(1) : undefined);

  const pathname = location?.pathname || '';

  const derivedMode: 'list' | 'create' | 'edit' | 'detail' = (() => {
    if (propMode) return propMode;
    if (pathname.endsWith('/create') || pathname.endsWith('/new')) return 'create';
    if (pathname.endsWith('/view')) return 'detail';
    if (params?.id) return 'edit';
    return 'list';
  })();

  const getModeIcon = (m: string) => {
    switch (m) {
      case 'create':
        return <Plus className="h-3.5 w-3.5" />;
      case 'edit':
        return <Edit className="h-3.5 w-3.5" />;
      case 'detail':
        return <Eye className="h-3.5 w-3.5" />;
      default:
        return <Database className="h-3.5 w-3.5" />;
    }
  };

  const getModeLabel = (m: string) => {
    switch (m) {
      case 'create':
        return 'Create';
      case 'edit':
        return 'Edit';
      case 'detail':
        return 'View';
      default:
        return '';
    }
  };

  const buildBreadcrumbs = (): BreadcrumbItem[] => {
    if (customItems) return customItems;

    const items: BreadcrumbItem[] = [
      {
        label: 'Home',
        path: '/',
        icon: <Home className="h-3.5 w-3.5" />,
      },
    ];

    if (model) {
      const isList = derivedMode === 'list';
      items.push({
        label: modelLabel || model,
        path: isList ? undefined : `/${model}`,
        icon: <Database className="h-3.5 w-3.5" />,
      });

      if (!isList) {
        items.push({
          label: getModeLabel(derivedMode),
          icon: getModeIcon(derivedMode),
        });
      }
    } else if (pathname === '/style-guide') {
      items.push({
        label: 'Design System',
        icon: <Sparkles className="h-3.5 w-3.5" />,
      });
    } else if (pathname === '/form-playground') {
      items.push({
        label: 'Form Playground',
        icon: <Layers className="h-3.5 w-3.5" />,
      });
    }

    return items;
  };

  const breadcrumbs = buildBreadcrumbs();

  return (
    <nav aria-label="Breadcrumbs" className={cn("flex items-center gap-1 text-xs", className)}>
      {breadcrumbs.map((item, index) => {
        const isLast = index === breadcrumbs.length - 1;

        return (
          <div key={index} className="flex items-center gap-1">
            {index > 0 && (
              <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60 shrink-0" />
            )}
            {isLast || !item.path ? (
              <span className="flex items-center gap-1.5 font-semibold text-foreground">
                {item.icon}
                <span>{item.label}</span>
              </span>
            ) : (
              <Button
                variant="ghost"
                size="sm"
                className="h-6 px-1.5 text-xs text-muted-foreground hover:text-foreground font-normal"
                onClick={() => {
                  if (item.action) {
                    item.action();
                  } else if (item.path) {
                    navigate({ to: item.path });
                  }
                }}
              >
                {item.icon}
                <span className="ml-1">{item.label}</span>
              </Button>
            )}
          </div>
        );
      })}
    </nav>
  );
}

// Hook for programmatic breadcrumbs
export function useBreadcrumbs() {
  const navigate = useNavigate();

  const setBreadcrumb = (items: BreadcrumbItem[]) => {
    console.log('Breadcrumbs set:', items);
  };

  const navigateTo = (path: string) => {
    navigate({ to: path });
  };

  return {
    setBreadcrumb,
    navigateTo,
  };
}
