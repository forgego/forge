import { createFileRoute } from "@tanstack/react-router";
import DynamicModelPage from "../../components/DynamicModelPage";

export const Route = createFileRoute("/$model/$id/view")({
  component: () => <DynamicModelPage mode="detail" />,
});
