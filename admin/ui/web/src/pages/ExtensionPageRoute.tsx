import { useLocation, useParams } from "@tanstack/react-router";
import AdminLayout from "../components/layout/AdminLayout";
import { getRegisteredPage } from "../lib/extensions/api";

export default function ExtensionPageRoute() {
  const location = useLocation();
  const params = useParams({ strict: false }) as Record<string, string | undefined>;

  const page =
    getRegisteredPage(location.pathname) ||
    (params?._splat ? getRegisteredPage(params._splat) : undefined);

  if (!page) {
    return (
      <AdminLayout>
        <div className="flex min-h-[400px] flex-col items-center justify-center p-8 text-center">
          <h2 className="text-xl font-bold text-foreground">
            Extension Page Not Found
          </h2>
          <p className="mt-2 text-sm text-muted-foreground">
            No extension page is registered at {location.pathname}.
          </p>
        </div>
      </AdminLayout>
    );
  }

  const Component = page.component;

  return (
    <AdminLayout>
      <Component />
    </AdminLayout>
  );
}
