import { createFileRoute } from '@tanstack/react-router';
import ModelRegistryPage from '../pages/ModelRegistryPage';

export const Route = createFileRoute('/registry')({
  component: ModelRegistryPage,
});
