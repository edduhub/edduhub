import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render as rtlRender, RenderOptions } from '@testing-library/react';
import { ReactNode } from 'react';

const defaultQueryClientOptions = {
  defaultOptions: {
    queries: {
      retry: false,
      gcTime: 0,
    },
    mutations: {
      retry: false,
    },
  },
};

export function createTestQueryClient() {
  return new QueryClient(defaultQueryClientOptions);
}

interface WrapperProps {
  children: ReactNode;
}

function QueryProviderWrapper({ children }: WrapperProps) {
  const queryClient = createTestQueryClient();
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}

export function renderWithQuery(ui: React.ReactElement, options?: Omit<RenderOptions, 'wrapper'>) {
  return rtlRender(ui, { wrapper: QueryProviderWrapper, ...options });
}

export * from '@testing-library/react';
export { defaultQueryClientOptions };
