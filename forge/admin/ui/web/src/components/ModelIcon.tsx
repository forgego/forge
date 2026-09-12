import {
  Activity,
  AlertTriangle,
  BarChart3,
  ClipboardList,
  CreditCard,
  FolderTree,
  Heart,
  HelpCircle,
  Image,
  Layers,
  MapPin,
  Package,
  Repeat,
  Settings,
  Shapes,
  ShoppingBag,
  ShoppingCart,
  Star,
  Tag,
  ThumbsUp,
  Ticket,
  TrendingUp,
  Truck,
  User,
  Users,
  Warehouse,
  FileBox,
  type LucideIcon,
} from "lucide-react";
import { cn } from "../lib/utils";

const ICONS: Record<string, LucideIcon> = {
  activity: Activity,
  alerttriangle: AlertTriangle,
  "alert-triangle": AlertTriangle,
  barchart3: BarChart3,
  "bar-chart-3": BarChart3,
  chart: BarChart3,
  clipboardlist: ClipboardList,
  "clipboard-list": ClipboardList,
  creditcard: CreditCard,
  "credit-card": CreditCard,
  foldertree: FolderTree,
  "folder-tree": FolderTree,
  heart: Heart,
  helpcircle: HelpCircle,
  "help-circle": HelpCircle,
  image: Image,
  layers: Layers,
  mappin: MapPin,
  "map-pin": MapPin,
  package: Package,
  repeat: Repeat,
  settings: Settings,
  shoppingbag: ShoppingBag,
  "shopping-bag": ShoppingBag,
  shoppingcart: ShoppingCart,
  "shopping-cart": ShoppingCart,
  star: Star,
  tag: Tag,
  thumbsup: ThumbsUp,
  "thumbs-up": ThumbsUp,
  ticket: Ticket,
  trendingup: TrendingUp,
  "trending-up": TrendingUp,
  truck: Truck,
  user: User,
  users: Users,
  warehouse: Warehouse,
  file: FileBox,
  filebox: FileBox,
};

function normalizeIconName(name?: string): string {
  return (name ?? "").trim().toLowerCase();
}

/** Resolves a backend icon name (e.g. "FolderTree") to a Lucide component. */
export function modelIconComponent(name?: string): LucideIcon {
  const key = normalizeIconName(name);
  if (ICONS[key]) return ICONS[key];
  const nospace = key.replace(/[_-\s]+/g, "");
  if (ICONS[nospace]) return ICONS[nospace];
  return Shapes;
}

export function ModelIcon({
  name,
  className,
}: {
  name?: string;
  className?: string;
}) {
  const Icon = modelIconComponent(name);
  // eslint-disable-next-line react-hooks/static-components -- pure registry lookup of a stable component.
  return <Icon className={cn("h-4 w-4", className)} aria-hidden />;
}
