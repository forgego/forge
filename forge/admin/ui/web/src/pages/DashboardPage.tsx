import AdminLayout from "../components/layout/AdminLayout";
import { useConfig } from "../api/hooks/adminHooks";
import { SDUIRoot } from "../components/sdui/DynamicRenderer";
import { Loader2 } from "lucide-react";

export default function DashboardPage() {
  const { data: config, isLoading, error } = useConfig();

  if (isLoading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      </AdminLayout>
    );
  }

  if (error) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-destructive">Error</h2>
            <p className="text-muted-foreground mt-2">Failed to load dashboard configuration.</p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  const dashboardConfig = config?.dashboard;

  if (!dashboardConfig || !dashboardConfig.Layout) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <h2 className="text-2xl font-bold">Welcome to Forge Admin</h2>
            <p className="text-muted-foreground mt-2">No dashboard layout configured.</p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <SDUIRoot component={dashboardConfig.Layout} />
    </AdminLayout>
  );
}
