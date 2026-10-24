import { createFileRoute } from "@tanstack/react-router";
import ExtensionPageRoute from "../../pages/ExtensionPageRoute";

export const Route = createFileRoute("/ext/$")({
  component: ExtensionPageRoute,
});
