import { useLocation, useNavigate } from '@tanstack/react-router';
import { ChevronRight, Home, Database, Plus, Edit, Eye, Trash2 } from 'lucide-react';
import { Button } from '../ui/button';
import { cn } from '../../lib/utils';

interface BreadcrumbItem {
  label: string;
  path?: string;
  icon?: React.ReactNode;
  action?: () => void;
}

interface BreadcrumbsProps {
  model?: string;
  modelLabel?: string;
  mode?: 'list' | 'create' | 'edit' | 'detail';
  customItems?: BreadcrumbItem[];
  className?: string;
}

export function Breadcrumbs({
  model,
  modelLabel,
  mode,
  customItems,
  className,
}: BreadcrumbsProps) {
  const location = useLocation();
  const navigate = useNavigate();

  const getModeIcon = (mode: string) => {
    switch (mode) {
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

  const getModeLabel = (mode: string) => {
    switch (mode) {
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
      items.push({
        label: modelLabel || model.charAt(0).toUpperCase() + model.slice(1),
        path: `/${model}`,
        icon: <Database className="h-3.5 w-3.5" />,
      });

      if (mode && mode !== 'list') {
        items.push({
          label: getModeLabel(mode),
          icon: getModeIcon(mode),
        });
      }
    }

    return items;
  };

  const breadcrumbs = buildBreadcrumbs();

  return (
    <nav className={cn("flex items-center gap-1 text-sm", className)}>
      {breadcrumbs.map((item, index) => {
        const isLast = index === breadcrumbs.length - 1;

        return (
          <div key={index} className="flex items-center gap-1">
            {index > 0 && (
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
            )}
            {isLast ? (
              <span className="flex items-center gap-1.5 font-medium text-foreground">
                {item.icon}
                {item.label}
              </span>
            ) : item.path ? (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 px-1.5 text-muted-foreground hover:text-foreground"
                onClick={() => navigate({ to: item.path })}
              >
                {item.icon}
                <span className="ml-1">{item.label}</span>
              </Button>
            ) : (
              <span className="flex items-center gap-1.5 text-muted-foreground">
                {item.icon}
                <span className="ml-1">{item.label}</span>
              </span>
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
    // This could be used with a store if needed
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
