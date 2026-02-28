import { render, screen } from "@testing-library/react";

import LoginPage from "@/app/auth/login/page";
import RegisterPage from "@/app/auth/register/page";

const mockUseAuth = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({
    push: jest.fn(),
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

describe("Auth pages autocomplete attributes", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({
      login: jest.fn(),
      register: jest.fn(),
    });
  });

  it("sets expected autocomplete values on login fields", () => {
    render(<LoginPage />);

    expect(screen.getByLabelText("Email")).toHaveAttribute("autocomplete", "email");
    expect(screen.getByLabelText("Password")).toHaveAttribute("autocomplete", "current-password");
  });

  it("sets expected autocomplete values on register fields", () => {
    render(<RegisterPage />);

    expect(screen.getByLabelText("First Name")).toHaveAttribute("autocomplete", "given-name");
    expect(screen.getByLabelText("Last Name")).toHaveAttribute("autocomplete", "family-name");
    expect(screen.getByLabelText("Email")).toHaveAttribute("autocomplete", "email");
    expect(screen.getByLabelText("Password")).toHaveAttribute("autocomplete", "new-password");
    expect(screen.getByLabelText("Confirm Password")).toHaveAttribute("autocomplete", "new-password");
  });
});
