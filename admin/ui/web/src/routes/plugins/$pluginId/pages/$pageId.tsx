import { createFileRoute } from "@tanstack/react-router";
import { lazy, Suspense } from "react";
import { Loader2 } from "lucide-react";

const PluginPage = lazy(() => import("@/pages/PluginPage"));

// @ts-ignore
export const Route = createFileRoute("/plugins/$pluginId/pages/$pageId")({
  component: () => (
    <Suspense
      fallback={
        <div className="flex min-h-[400px] w-full items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <PluginPage />
    </Suspense>
  ),
});
