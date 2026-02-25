# AGENTS.md - Agentic Coding Guidelines for edduhub

This file provides guidelines for AI agents working on the edduhub codebase.

## Project Overview

- **Type**: Full-stack monorepo (Next.js frontend + Go backend)
- **Frontend**: Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS 4
- **Backend**: Go 1.25, Echo framework, PostgreSQL, Redis
- **Auth**: Ory Kratos (identity), Ory Keto (authorization)
- **Storage**: MinIO (S3-compatible)

## Directory Structure

```
/client                 # Next.js frontend
  /src/app            # Next.js App Router pages
  /src/components     # UI components (ui/, navigation/, auth/)
  /src/lib           # Utilities, hooks, providers, types
  /src/__tests__     # Jest tests + mocks
/server                 # Go backend
  /api/handler       # HTTP handlers
  /api/app           # Application setup
  /internal          # Business logic (repository, services, models)
  /db/migrations    # PostgreSQL migrations
  /pkg              # Shared packages (jwt, logger)
```

---

## Commands

### Backend (Go) - Use `task` CLI from root

| Command | Description |
|---------|-------------|
| `task` | Run all tests (unit + integration) |
| `task test:unit` | Run unit tests only |
| `task test:integration` | Run integration tests (requires Docker) |
| `task lint` | Run golangci-lint |
| `task fmt` | Format code (gofumpt + goimports) |
| `task build` | Build binary to `bin/edduhub` |
| `task dev` | Start server + client (hot reload) |
| `task dev:server` | Server only with air hot reload |
| `task dev:client` | Client only |
| `task swagger` | Generate OpenAPI docs |
| `task db:migrate` | Run database migrations |
| `task db:start` | Start Docker database services |
| `task bootstrap` | Full setup (deps, env, db, migrations) |

**Single Test (Backend)**:
```bash
go test -v ./internal/services/college -run TestCreateCollege_Success
go test -tags=unit -run "TestName" ./...
```

### Frontend (Client) - Use npm in `/client`

| Command | Description |
|---------|-------------|
| `npm run dev` | Start dev server (port 3000) |
| `npm run build` | Production build |
| `npm run lint` | ESLint check |
| `npm run lint:fix` | ESLint fix |
| `npm run test` | Run Jest tests |
| `npm run test:watch` | Watch mode |
| `npm run test:coverage` | Coverage report |

**Single Test (Frontend)**:
```bash
npm test -- button.test.tsx
npm test -- --testNamePattern="renders"
```

---

## Code Style Guidelines

### TypeScript / React (Client)

- **Strict TypeScript**: All strict options enabled. Never use `any` unless absolutely necessary.
- **Imports**: Use path alias `@/` for internal imports (e.g., `@/lib/utils`). Use `import type` for types only.
- **Component Pattern**: Use `React.memo()` + `React.forwardRef()` + `cva` for variants.
- **Naming**: Components PascalCase, hooks `use*`, utils/functions camelCase, types PascalCase.

```tsx
import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva("base-classes", {
  variants: {
    variant: { default: "default", secondary: "secondary" },
    size: { default: "default", sm: "sm" },
  },
  defaultVariants: { variant: "default", size: "default" },
});

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> {}

const Button = React.memo(React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return <button ref={ref} className={cn(buttonVariants({ variant, size, className }))} {...props} />;
  }
));
Button.displayName = "Button";
export { Button, buttonVariants };
```

**Error Handling**: Use custom error classes from `@/lib/errors.ts`:
- `APIError`, `NetworkError`, `AuthenticationError`, `NotFoundError` with type guards (`isAPIError()`, etc.)

---

### Go (Server)

- **Formatting**: Run `task fmt` before committing (gofumpt + goimports)
- **Architecture**: Handlers → Services → Repositories (clean architecture)
- **Error Handling**: Use `fmt.Errorf("failed to %s: %w", operation, err)`
- **Naming**: Files snake_case, exported PascalCase, unexported camelCase
- **Test Tags**: No tag for unit, `//go:build integration` for integration tests

```go
// Unit test
func TestCreateUser_Success(t *testing.T) {}

// Integration test
//go:build integration
func TestCreateUser_Integration(t *testing.T) {}
```

---

## Testing Guidelines

### Frontend (Jest)
- **Config**: `client/jest.config.js` - jsdom environment, 70% coverage threshold
- **Location**: `client/src/__tests__/**/*.test.{ts,tsx}`
- **Mocks**: `__tests__/mocks/` - `renderWithAuth()`, `mockFetchSuccess()`, etc.
- **Libraries**: `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom`

### Backend (Go)
- **Unit tests**: No build tag (default)
- **Integration tests**: `//go:build integration` tag
- **Libraries**: `github.com/stretchr/testify/assert`, `require`, `mock`
- **Mocks**: Manual struct mocks or `task mocks` (mockery)

---

## Linting & Configuration

### Client
- **ESLint**: `client/eslint.config.mjs` - extends Next.js core-web-vitals + TypeScript (any allowed)
- **TypeScript**: `client/tsconfig.json` - strict mode, path alias `@/*` → `./src/*`

### Server
- **Linter**: golangci-lint
- **Formatter**: gofumpt + goimports (`task fmt`)
- **API Docs**: Swagger at `http://localhost:8080/swagger` (`task swagger`)

---

## Important Notes

1. **Environment Setup**: Copy `.env.example` to `.env.local` before running
2. **Docker Required**: Integration tests and `task dev` require Docker running
3. **No Commit Rules**: Do not commit without explicit user request
4. **Hot Reload**: Backend uses `air`, frontend uses Next.js dev server
5. **Breaking Changes**: Always verify changes don't break existing functionality
