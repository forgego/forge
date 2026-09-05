import { createFileRoute } from "@tanstack/react-router";
import StyleGuidePage from "../pages/StyleGuidePage";
import AdminLayout from "../components/layout/AdminLayout";

// @ts-ignore
export const Route = createFileRoute("/style-guide")({
  component: () => (
    <AdminLayout>
      <StyleGuidePage />
    </AdminLayout>
  ),
});
