import { render, screen, waitFor } from "@testing-library/react";

import { LayoutContent } from "@/components/layout-content";

const mockReplace = jest.fn();
const mockPathname = jest.fn();
const mockUseAuth = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: mockReplace,
    refresh: jest.fn(),
    back: jest.fn(),
    forward: jest.fn(),
    prefetch: jest.fn(),
  }),
  usePathname: () => mockPathname(),
}));

jest.mock("@/lib/auth-context", () => ({
  useAuth: () => mockUseAuth(),
}));

jest.mock("@/components/navigation/sidebar", () => ({
  Sidebar: () => <div data-testid="sidebar">Sidebar</div>,
}));

jest.mock("@/components/navigation/topbar", () => ({
  Topbar: () => <div data-testid="topbar">Topbar</div>,
}));

describe("LayoutContent auth guard", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockPathname.mockReturnValue("/");
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
    });
  });

  it("renders auth pages without redirect when unauthenticated", () => {
    mockPathname.mockReturnValue("/auth/login");

    render(
      <LayoutContent>
        <div data-testid="page-content">Auth Page</div>
      </LayoutContent>
    );

    expect(screen.getByTestId("page-content")).toBeInTheDocument();
    expect(mockReplace).not.toHaveBeenCalled();
  });

  it("redirects unauthenticated users on protected routes", async () => {
    mockPathname.mockReturnValue("/dashboard");
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
    });

    render(
      <LayoutContent>
        <div data-testid="page-content">Protected Page</div>
      </LayoutContent>
    );

    expect(screen.queryByTestId("page-content")).not.toBeInTheDocument();

    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith("/auth/login");
    });
  });

  it("shows loading state while auth is initializing", () => {
    mockPathname.mockReturnValue("/dashboard");
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      isLoading: true,
    });

    const { container } = render(
      <LayoutContent>
        <div data-testid="page-content">Protected Page</div>
      </LayoutContent>
    );

    expect(container.querySelector(".animate-spin")).toBeInTheDocument();
    expect(screen.queryByTestId("page-content")).not.toBeInTheDocument();
    expect(mockReplace).not.toHaveBeenCalled();
  });

  it("renders app shell for authenticated users", () => {
    mockPathname.mockReturnValue("/dashboard");
    mockUseAuth.mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
    });

    render(
      <LayoutContent>
        <div data-testid="page-content">Dashboard</div>
      </LayoutContent>
    );

    expect(screen.getByTestId("sidebar")).toBeInTheDocument();
    expect(screen.getByTestId("topbar")).toBeInTheDocument();
    expect(screen.getByTestId("page-content")).toBeInTheDocument();
    expect(mockReplace).not.toHaveBeenCalled();
  });
});
