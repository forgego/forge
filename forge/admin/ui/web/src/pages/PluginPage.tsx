import { useParams } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import AdminLayout from "../components/layout/AdminLayout";
import { SDUIRoot } from "../components/sdui/DynamicRenderer";
import { Loader2 } from "lucide-react";
import apiClient from "../api/client";

export default function PluginPage() {
  const params: any = useParams({
    from: "/admin/plugins/$pluginId/pages/$pageId",
  } as any);
  const { pluginId, pageId } = params;

  const {
    data: pageComponent,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["admin", "plugins", pluginId, "pages", pageId],
    queryFn: async () => {
      const response = await apiClient.get(
        `/plugins/${pluginId}/pages/${pageId}`
      );
      return response.data;
    },
  });

  if (isLoading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      </AdminLayout>
    );
  }

  if (error || !pageComponent) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-destructive">Error</h2>
            <p className="text-muted-foreground mt-2">
              Failed to load plugin page.
            </p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <SDUIRoot component={pageComponent} />
    </AdminLayout>
  );
}
