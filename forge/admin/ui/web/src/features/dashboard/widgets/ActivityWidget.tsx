import type { WidgetProps } from '../../../lib/widgets';

interface ActivityItem {
  id: string;
  user: string;
  action: string;
  time: string;
  avatar?: string;
}

const defaultActivities: ActivityItem[] = [
  { id: '1', user: 'Hamid Rabedi', action: 'Created new product "Ultra Watch"', time: '2 mins ago' },
  { id: '2', user: 'Elena Smith', action: 'Updated user "John Doe" permissions', time: '15 mins ago' },
  { id: '3', user: 'System', action: 'Automated backup completed', time: '1 hour ago' },
  { id: '4', user: 'Alex Johnson', action: 'Deleted order #4521', time: '3 hours ago' },
];

export default function ActivityWidget({ config }: WidgetProps) {
  const activities = config.params?.activities || defaultActivities;

  return (
    <div className="space-y-4 mt-2">
      {activities.map((item: ActivityItem, idx: number) => (
        <div key={item.id} className="flex items-start gap-3 group">
          <div className="relative">
            <div className="w-8 h-8 rounded-full bg-accent flex items-center justify-center text-[10px] font-bold">
              {item.user.split(' ').map((n: string) => n[0]).join('')}
            </div>
            {idx !== activities.length - 1 && (
              <div className="absolute top-8 left-1/2 -translate-x-1/2 w-[1px] h-4 bg-border" />
            )}
          </div>
          <div className="flex-1 space-y-1">
            <p className="text-sm font-medium leading-none group-hover:text-primary transition-colors cursor-default">
              {item.user} <span className="text-muted-foreground font-normal">{item.action}</span>
            </p>
            <p className="text-[10px] text-muted-foreground">{item.time}</p>
          </div>
        </div>
      ))}
    </div>
  );
}
