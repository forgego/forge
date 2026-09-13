import { createFileRoute } from "@tanstack/react-router";
import { lazy, Suspense } from "react";
import { Loader2 } from "lucide-react";
import AdminLayout from "../components/layout/AdminLayout";

const StyleGuidePage = lazy(() => import("../pages/StyleGuidePage"));

// @ts-ignore
export const Route = createFileRoute("/style-guide")({
  component: () => (
    <AdminLayout>
      <Suspense
        fallback={
          <div className="flex min-h-[400px] w-full items-center justify-center">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        }
      >
        <StyleGuidePage />
      </Suspense>
    </AdminLayout>
  ),
});
