# Comprehensive Auth Testing Plan (Post-Ory Migration)

## Objective
To introduce comprehensive test coverage across the Go backend, Next.js frontend, and end-to-end (E2E) suites. This will ensure the stability and correctness of the authentication system following the complete removal of local JWTs and the transition to an exclusive Ory stack (Kratos, Hydra, Keto) for IAM features.

## Scope & Impact
- **Backend (Go)**: Fill gaps left by the removal of legacy JWT tests. Focus on `auth_service.go`, `auth.go` (handlers), and `auth.go` (middleware).
- **Frontend (Next.js)**: Update and expand unit tests for auth contexts, React queries, and routing guards to ensure they integrate correctly with the Ory session model.
- **End-to-End (Playwright)**: Add robust workflows in `client/tests/e2e/workflows` that validate the complete lifecycle of registration, login, token refresh, and authorization through the real Ory stack.

## Implementation Plan

### Phase 1: Backend (Go) Unit & Integration Tests
1. **Auth Service Tests** (`server/internal/services/auth/auth_service_test.go`):
   - Add tests to ensure `Login` and `CompleteRegistration` return a pure `*models.Identity` without generating local tokens.
   - Mock Kratos & Hydra clients to verify correct calls and proper error handling.
   - Verify that JIT provisioning (`resolveAndProvisionLocalIdentity`) still functions properly under the new paradigm.
2. **Auth Handler Tests** (`server/api/handler/auth_test.go`):
   - Add tests for `DirectLogin` to ensure it successfully delegates to Ory and omits local token issuance.
   - Add tests for `RefreshToken` to confirm it exclusively processes Hydra OAuth2 refresh tokens.
3. **Middleware Tests** (`server/internal/middleware/auth_test.go`):
   - Add tests to ensure `ValidateToken` strictly performs Hydra introspection without attempting any local JWT fallback.
   - Add tests to verify role-based access control and integration with Ory Keto.

### Phase 2: Frontend (Next.js) Component Tests
1. **Auth Guards** (`client/src/__tests__/layout-content-auth-guard.test.tsx`):
   - Verify the guard strictly relies on the Ory session state and correctly redirects unauthenticated users to the Kratos login UI.
   - Verify it permits access when a valid Ory session is present.
2. **Auth Context and API Handlers**:
   - Write tests for the hooks/contexts that manage user state, ensuring they accurately parse identity and session data from Ory.
   - Ensure the API client strictly passes Ory-issued cookies/tokens and does not attempt to attach legacy custom JWTs.

### Phase 3: End-to-End (Playwright) Workflow Tests
1. **Authentication Workflows** (`client/tests/e2e/workflows/auth.spec.ts`):
   - **Registration**: Automate the complete registration flow through the Ory Kratos UI, ensuring the backend successfully catches the webhook and provisions the local user identity.
   - **Login/Logout**: Automate the login and logout flows, validating the creation and destruction of the Kratos session cookie and Hydra tokens.
2. **Authorization Workflows** (`client/tests/e2e/workflows/rbac.spec.ts`):
   - Test role-based access control (RBAC) by attempting to access protected routes (e.g., student vs. faculty dashboards) and verifying that Ory Keto permissions enforce the correct access levels.

## Verification & Testing
1. **Backend**: Run `cd server && go test ./... -v -cover` to ensure increased code coverage and zero regressions.
2. **Frontend**: Run `cd client && bun test` to verify component stability.
3. **E2E**: Run `cd client && npx playwright test` to validate the full application workflow.

## Rollback & Alternatives
- Since these are purely additive test files, no migrations or rollbacks of application code are necessary.
- If E2E tests are consistently flaky due to external Ory service latencies, an alternative is to mock the Ory endpoints for the frontend tests while relying exclusively on the backend unit tests for logic verification.