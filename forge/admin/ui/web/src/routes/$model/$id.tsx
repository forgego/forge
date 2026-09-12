import { Outlet, createFileRoute } from '@tanstack/react-router';

// Layout route: renders the child (edit form or detail view) via Outlet.
// The edit form lives in ./$id/index.tsx, the detail view in ./$id.view.tsx.
export const Route = createFileRoute('/$model/$id')({
  component: () => <Outlet />,
});
