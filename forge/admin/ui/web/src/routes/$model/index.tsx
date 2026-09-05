import { createFileRoute } from '@tanstack/react-router'
import DynamicModelPage from '../../components/DynamicModelPage';

export const Route = createFileRoute('/$model/')({
  component: ModelRoute,
});

function ModelRoute() {
  return <DynamicModelPage mode="list" />;
}
