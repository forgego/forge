import type { WidgetProps } from '../../../lib/widgets';
import * as LucideIcons from 'lucide-react';
import { cn } from '../../../lib/utils';

export default function StatsWidget({ config }: WidgetProps) {
  const { value, trend, icon } = config.params || {};
  const isNegative = trend?.startsWith('-');
  const IconComponent = icon ? (LucideIcons as any)[icon] : null;

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between">
        <div className="text-2xl font-bold tracking-tight">{value}</div>
        {IconComponent && (
          <div className="p-2 bg-primary/10 rounded-lg">
            <IconComponent className="h-4 w-4 text-primary" />
          </div>
        )}
      </div>
      {trend && (
        <div className={cn(
          "flex items-center text-xs font-semibold px-2 py-0.5 rounded-full w-fit",
          isNegative ? "bg-destructive/10 text-destructive" : "bg-green-500/10 text-green-600 dark:text-green-400"
        )}>
          {isNegative ? (
            <LucideIcons.ArrowDownRight className="h-3 w-3 mr-1" />
          ) : (
            <LucideIcons.ArrowUpRight className="h-3 w-3 mr-1" />
          )}
          {trend}
        </div>
      )}
    </div>
  );
}
