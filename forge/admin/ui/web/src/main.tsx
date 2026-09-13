import "./lib/admin-prefix";
import { StrictMode } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider, createRouter } from '@tanstack/react-router';
import { bootstrapAdmin } from './lib/boot';
import { ThemeProvider } from './components/providers/ThemeProvider';
import { TooltipProvider } from './components/ui/tooltip';
import '@fontsource/inter/400.css';
import '@fontsource/inter/500.css';
import '@fontsource/inter/600.css';
import '@fontsource/inter/700.css';
import '@fontsource/jetbrains-mono/400.css';
import '@fontsource/jetbrains-mono/700.css';
import './index.css';

// Bootstrap registries
bootstrapAdmin();

// Import the generated route tree
import { routeTree } from './routeTree.gen';
import { getAdminPrefix } from './lib/admin-prefix';

// Create a new router instance
const router = createRouter({ 
  routeTree,
  basepath: getAdminPrefix() || '/' 
});

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

// Render the app
const rootElement = document.getElementById('root')!;
if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <StrictMode>
      <ThemeProvider>
        <TooltipProvider delayDuration={300} skipDelayDuration={150}>
          <RouterProvider router={router} />
        </TooltipProvider>
      </ThemeProvider>
    </StrictMode>
  );
}

