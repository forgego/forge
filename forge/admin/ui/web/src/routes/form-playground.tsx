import { createFileRoute } from "@tanstack/react-router";
import FormPlaygroundPage from "../pages/FormPlaygroundPage";

export const Route = createFileRoute("/form-playground")({
  component: FormPlaygroundPage,
});
