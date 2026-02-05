import { useModelMetadata } from '../api/hooks/adminHooks';
import ModelListPage from '../pages/ModelListPage';
import ModelUpsertPage from '../pages/ModelUpsertPage';
import { useParams } from '@tanstack/react-router';

interface DynamicModelPageProps {
  mode: 'list' | 'create' | 'edit' | 'detail';
}

export default function DynamicModelPage({ mode }: DynamicModelPageProps) {
  const { model } = useParams({ strict: false }) as any;
  const { data: metadata, isLoading, error } = useModelMetadata(model);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error || !metadata) {
    return (
      <div className="p-8 text-center text-red-500">
        Error loading metadata for "{model}"
      </div>
    );
  }

  // Layout logic based on metadata (could be used for more complex switching)
  
  if (mode === 'list') {
    return <ModelListPage />;
  }
  
  if (mode === 'create' || mode === 'edit') {
    return <ModelUpsertPage mode={mode as any} />;
  }

  // Detail view (TODO: implement ModelViewPage)
  return <ModelListPage />; 
}
