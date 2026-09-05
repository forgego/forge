import { createFileRoute } from '@tanstack/react-router';
import DynamicModelPage from '../../components/DynamicModelPage';

export const Route = createFileRoute('/$model/$id')({
  component: () => <DynamicModelPage mode="edit" />,
});
