import { createFileRoute } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';
import { Loader2 } from 'lucide-react';

const ModelRegistryPage = lazy(() => import('../pages/ModelRegistryPage'));

export const Route = createFileRoute('/registry')({
  component: () => (
    <Suspense
      fallback={
        <div className="flex min-h-[400px] w-full items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <ModelRegistryPage />
    </Suspense>
  ),
});
