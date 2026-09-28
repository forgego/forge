import type { WidgetProps } from '../../../lib/widgets';
import { ArrowDownRight, ArrowUpRight } from 'lucide-react';
import { ModelIcon } from '../../../components/ModelIcon';
import { cn } from '../../../lib/utils';

export default function StatsWidget({ config }: WidgetProps) {
  const { value, trend, icon } = config.params || {};
  const isNegative = trend?.startsWith('-');

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between">
        <div className="text-2xl font-bold tracking-tight">{value}</div>
        {icon && (
          <div className="p-2 bg-primary/10 rounded-lg">
            <ModelIcon name={icon} className="h-4 w-4 text-primary" />
          </div>
        )}
      </div>
      {trend && (
        <div className={cn(
          "flex items-center text-xs font-semibold px-2 py-0.5 rounded-full w-fit",
          isNegative ? "bg-destructive/10 text-destructive" : "bg-success-surface text-success"
        )}>
          {isNegative ? (
            <ArrowDownRight className="h-3 w-3 mr-1" />
          ) : (
            <ArrowUpRight className="h-3 w-3 mr-1" />
          )}
          {trend}
        </div>
      )}
    </div>
  );
}
