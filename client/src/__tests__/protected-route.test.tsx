import { render, screen, waitFor } from "@testing-library/react";

import { ProtectedRoute } from "@/components/auth/protected-route";

const mockPush = jest.fn();
const mockUseAuth = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mockPush,
    replace: jest.fn(),
    refresh: jest.fn(),
    back: jest.fn(),
    forward: jest.fn(),
    prefetch: jest.fn(),
  }),
}));

jest.mock("@/lib/auth-context", () => ({
  useAuth: () => mockUseAuth(),
}));

describe("ProtectedRoute", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({
      user: null,
      isLoading: false,
    });
  });

  it("shows loading spinner while auth is loading", () => {
    mockUseAuth.mockReturnValue({
      user: null,
      isLoading: true,
    });

    const { container } = render(
      <ProtectedRoute>
        <div data-testid="protected-content">Secret</div>
      </ProtectedRoute>
    );

    expect(container.querySelector(".animate-spin")).toBeInTheDocument();
    expect(screen.queryByTestId("protected-content")).not.toBeInTheDocument();
    expect(mockPush).not.toHaveBeenCalled();
  });

  it("redirects unauthenticated users to login", async () => {
    mockUseAuth.mockReturnValue({
      user: null,
      isLoading: false,
    });

    render(
      <ProtectedRoute>
        <div data-testid="protected-content">Secret</div>
      </ProtectedRoute>
    );

    expect(screen.queryByTestId("protected-content")).not.toBeInTheDocument();

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/auth/login");
    });
  });

  it("redirects users without required role", async () => {
    mockUseAuth.mockReturnValue({
      user: {
        role: "student",
      },
      isLoading: false,
    });

    render(
      <ProtectedRoute allowedRoles={["admin"]}>
        <div data-testid="protected-content">Admin Secret</div>
      </ProtectedRoute>
    );

    expect(screen.queryByTestId("protected-content")).not.toBeInTheDocument();

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/");
    });
  });

  it("renders children when user is authorized", () => {
    mockUseAuth.mockReturnValue({
      user: {
        role: "admin",
      },
      isLoading: false,
    });

    render(
      <ProtectedRoute allowedRoles={["admin"]}>
        <div data-testid="protected-content">Admin Secret</div>
      </ProtectedRoute>
    );

    expect(screen.getByTestId("protected-content")).toBeInTheDocument();
    expect(mockPush).not.toHaveBeenCalled();
  });
});
