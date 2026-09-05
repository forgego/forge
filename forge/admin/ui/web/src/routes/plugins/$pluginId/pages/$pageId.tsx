import { createFileRoute } from "@tanstack/react-router";
import PluginPage from "@/pages/PluginPage";

// @ts-ignore
export const Route = createFileRoute("/plugins/$pluginId/pages/$pageId")({
  component: PluginPage,
});
