import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';
import { describe, it, jest } from '@jest/globals';
import { ReactQueryProvider } from '@/lib/react-query-provider';

// Mock the auth context
jest.mock('@/lib/auth-context', () => ({
  useAuth: () => ({
    user: null,
    isLoading: false,
  }),
  AuthProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

describe('ReactQueryProvider', () => {
  it('renders children without crashing', () => {
    render(
      <ReactQueryProvider>
        <div data-testid="test-child">Test Content</div>
      </ReactQueryProvider>
    );
    
    expect(screen.getByTestId('test-child')).toBeInTheDocument();
    expect(screen.getByText('Test Content')).toBeInTheDocument();
  });
});
