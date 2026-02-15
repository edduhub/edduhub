import { ReactElement, ReactNode } from 'react';
import { render, RenderOptions } from '@testing-library/react';

interface MockAuthValue {
  user: unknown;
  session: unknown;
  isLoading: boolean;
  isAuthenticated: boolean;
  login?: () => Promise<void>;
  register?: () => Promise<void>;
  logout?: () => Promise<void>;
  refreshSession?: () => Promise<void>;
}

const defaultAuthValue = {
  user: null,
  session: null,
  isLoading: false,
  isAuthenticated: false,
};

const MockAuthContext = require('react').createContext(defaultAuthValue);

export function MockAuthProvider({ children, value = {} as MockAuthValue }: { children: ReactNode; value?: MockAuthValue }) {
  const contextValue = { ...defaultAuthValue, ...value };
  return (
    <MockAuthContext.Provider value={contextValue}>
      {children}
    </MockAuthContext.Provider>
  );
}

export function renderWithAuth(ui: ReactElement, authValue?: MockAuthValue, options?: Omit<RenderOptions, 'wrapper'>) {
  return render(ui, {
    wrapper: ({ children }) => MockAuthProvider({ children, value: authValue }),
    ...options,
  });
}

export * from '@testing-library/react';
