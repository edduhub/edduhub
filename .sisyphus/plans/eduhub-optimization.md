# EduHub Comprehensive Production Optimization - Master Work Plan

## TL;DR

> **Objective**: Transform EduHub into a production-ready, fully functional educational platform with complete feature parity, zero TypeScript errors, optimized React performance, and fully operational backend APIs for all portal pages.
> 
> **Deliverables**:
> - **Frontend**: Zero `any` types (47 → 0), Zero console.logs (30 → 0), 130+ memoized functions, React Query integration (0% → 100%)
> - **Backend**: Complete Self-Service Portal APIs, Parent Portal APIs, Faculty Tools APIs, Swagger documentation (empty → full)
> - **Integration**: All portal pages fully functional end-to-end
> 
> **Estimated Effort**: Large (~2-3 weeks with parallel execution)
> **Parallel Execution**: YES - 5 major waves with 25+ parallel task groups
> **Critical Path**: Wave 1 (Types/Config) → Wave 2 (React Query) → Wave 3 (Performance) → Wave 4 (Backend APIs) → Wave 5 (Integration/Swagger)

---

## Context

### Original Request
> "Based on the extensive context gathered, create the FINAL comprehensive work plan integrating both frontend optimization AND backend API completion into ONE master plan."

### Interview Summary

**Key Decisions**:
1. **Plan File**: OVERWRITE `.sisyphus/plans/eduhub-optimization.md` with FINAL, EXECUTABLE version
2. **Backend Features**: INTEGRATE frontend optimization + backend API completion in ONE master plan
3. **Timeline**: 2-3 weeks is acceptable for comprehensive completion
4. **Verification**: Include BOTH simple verification (grep, type-check) AND full integration verification (E2E tests, build)

**Frontend Issues Identified**:
- **Console Cleanup**: 11 files with 30 console.log statements
- **Type Safety**: 16 files with 47 `any` type occurrences
- **Performance**: 14+ files with 130+ inline functions needing useCallback
- **Memory Leaks**: 5 files with missing cleanup functions (setInterval/setTimeout)
- **React Query**: 43 unoptimized useEffect hooks needing migration
- **Build**: next.config.ts needs optimization

**Backend Gaps Identified**:
- **Self-Service Portal**: Frontend exists, backend APIs MISSING
- **Parent Portal**: Frontend exists, backend APIs MISSING
- **Faculty Tools**: Frontend exists, backend APIs MISSING (rubrics, office hours)
- **Swagger Docs**: swagger.json is EMPTY `{}`

### Research Findings (Exact File Analysis)

**Console Cleanup - 11 Files, 30 Statements**:
| File | Count | Notes |
|------|-------|-------|
| `client/src/app/timetable/page.tsx` | 4 | Error handling console.error |
| `client/src/app/roles/page.tsx` | 3 | Debug logging |
| `client/src/app/self-service/page.tsx` | 3 | Error logging |
| `client/src/app/notifications/page.tsx` | 7 | WebSocket logging |
| `client/src/app/placements/page.tsx` | 3 | Debug logging |
| `client/src/app/forum/page.tsx` | 2 | Error handling |
| `client/src/app/analytics/page.tsx` | 1 | Debug logging |
| `client/src/app/parent-portal/page.tsx` | 1 | Error logging |
| `client/src/app/parent-portal/contact/page.tsx` | 1 | Debug logging |
| `client/src/lib/logger.ts` | 5 | Intentional - keep these |
| `client/src/components/error-boundary.tsx` | 1 | Legitimate error reporting - keep |

**Any Type Fixes - 16 Files, 47 Occurrences**:
| File | Count | Locations |
|------|-------|-----------|
| `client/src/app/page.tsx` | 2 | catch blocks, reduce callbacks |
| `client/src/app/timetable/page.tsx` | 1 | onValueChange handler |
| `client/src/app/quizzes/[quizId]/attempt/[attemptId]/page.tsx` | 2 | catch blocks (x2) |
| `client/src/app/quizzes/[quizId]/results/[attemptId]/page.tsx` | 1 | catch block |
| `client/src/app/quizzes/page.tsx` | 2 | catch block, questions array |
| `client/src/app/auth/login/page.tsx` | 1 | catch block |
| `client/src/app/system-status/page.tsx` | 1 | catch block |
| `client/src/app/files/page.tsx` | 7 | catch blocks (multiple) |
| `client/src/app/webhooks/page.tsx` | 5 | catch blocks |
| `client/src/app/batch-operations/page.tsx` | 4 | catch blocks |
| `client/src/app/assignments/page.tsx` | 1 | catch block |
| `client/src/app/grades/page.tsx` | 1 | catch block |
| `client/src/app/placements/page.tsx` | 1 | onValueChange handler |
| `client/src/app/fees/page.tsx` | 1 | Razorpay window |
| `client/src/app/audit-logs/page.tsx` | 1 | changes field |
| `client/src/components/navigation/sidebar.tsx` | 1 | icon prop |

**Performance Optimizations - 14 Files, 130+ Functions**:
| File | Functions to Memoize | Priority |
|------|---------------------|----------|
| `client/src/app/timetable/page.tsx` | 2 (.map functions) | HIGH |
| `client/src/app/students/page.tsx` | 2 functions | HIGH |
| `client/src/app/assignments/page.tsx` | 1 function | HIGH |
| `client/src/app/parent-portal/page.tsx` | 1 function | MEDIUM |
| `client/src/app/announcements/page.tsx` | 1 function | MEDIUM |
| `client/src/app/roles/page.tsx` | 2 functions | MEDIUM |
| `client/src/app/webhooks/page.tsx` | 3 functions | MEDIUM |
| `client/src/app/quizzes/page.tsx` | 1 function | MEDIUM |
| `client/src/app/placements/page.tsx` | 2 functions | MEDIUM |
| `client/src/app/files/page.tsx` | 7 functions | HIGH |
| `client/src/app/audit-logs/page.tsx` | 2 functions | MEDIUM |
| `client/src/app/users/page.tsx` | 3 functions | MEDIUM |
| `client/src/app/forum/page.tsx` | 1 function | LOW |
| `client/src/app/fees/page.tsx` | 1 function | LOW |

**Cleanup Fixes - 5 Files with Memory Leaks**:
| File | Issue | Timer Type |
|------|-------|------------|
| `client/src/app/system-status/page.tsx` | setInterval without cleanup | setInterval |
| `client/src/app/parent-portal/page.tsx` | setTimeout without cleanup | setTimeout |
| `client/src/app/notifications/page.tsx` | setTimeout without cleanup | setTimeout |
| `client/src/app/attendance/page.tsx` | setTimeout without cleanup | setTimeout |
| `client/src/app/fees/page.tsx` | setTimeout without cleanup | setTimeout |

**Backend APIs to Create**:

**Self-Service Portal APIs**:
- `POST /api/self-service/requests` - Submit enrollment/schedule/document requests
- `GET /api/self-service/requests` - Get user's requests
- `GET /api/self-service/requests/:id` - Get specific request details
- `PUT /api/self-service/requests/:id` - Update request (admin)

**Parent Portal APIs**:
- `GET /api/parent/children` - Get linked students for parent
- `GET /api/parent/children/:id/dashboard` - Get child's dashboard data
- `GET /api/parent/children/:id/attendance` - Get child's attendance
- `GET /api/parent/children/:id/grades` - Get child's grades
- `GET /api/parent/children/:id/assignments` - Get child's assignments

**Faculty Tools APIs**:
- `GET /api/faculty/rubrics` - List all grading rubrics
- `POST /api/faculty/rubrics` - Create new rubric
- `GET /api/faculty/rubrics/:id` - Get specific rubric
- `PUT /api/faculty/rubrics/:id` - Update rubric
- `DELETE /api/faculty/rubrics/:id` - Delete rubric
- `GET /api/faculty/office-hours` - Get office hours schedule
- `POST /api/faculty/office-hours` - Create office hours slot
- `PUT /api/faculty/office-hours/:id` - Update office hours
- `DELETE /api/faculty/office-hours/:id` - Delete office hours

---

## Work Objectives

### Core Objective
Transform EduHub codebase into a production-ready platform with complete type safety, optimized React performance, functional backend APIs for all portal pages, and professional code quality standards.

### Concrete Deliverables
1. **Frontend Type Safety**:
   - `/Users/kasyap/Documents/edduhub/client/src/types/index.ts` with all API types
   - `/Users/kasyap/Documents/edduhub/client/src/lib/errors.ts` with custom error classes
   - Zero `any` types in production code

2. **Frontend Performance**:
   - React Query integration for all data fetching
   - useCallback for 130+ inline functions
   - useMemo for expensive computations
   - React.memo for pure components
   - Fixed 5 memory leaks

3. **Frontend Build**:
   - Optimized next.config.ts with bundle analyzer
   - Zero console.log in production builds
   - Lighthouse score >90

4. **Backend APIs**:
   - 4 Self-Service Portal API endpoints
   - 5 Parent Portal API endpoints
   - 8 Faculty Tools API endpoints
   - Full swagger documentation

5. **Integration**:
   - All portal pages fully functional end-to-end
   - E2E tests passing
   - Build succeeds with all optimizations

### Definition of Done
```bash
# Type Safety
bun run type-check  # Zero errors

# Build Success
bun run build  # Completes with optimizations

# Lint Success
bun run lint  # Zero warnings

# Test Success
bun run test  # All unit tests pass
bun run test:e2e  # All E2E tests pass

# Metrics
- Zero `any` types in production code
- Zero console.log in production builds
- 100% React Query coverage for data fetching
- All 130+ inline functions use useCallback
- All 43 useEffect hooks optimized
- 5 memory leaks fixed
- Lighthouse performance score >90
- Swagger docs show all 30+ endpoints
- All portal pages functional end-to-end
```

### Must Have (Non-Negotiable)
- All 47 `any` types replaced with proper types
- All 30 console.logs removed from production
- React Query integrated for all data fetching (43 useEffects migrated)
- useCallback applied to all 130+ inline functions
- useMemo applied to all expensive computations
- 5 missing cleanup functions fixed
- 17 backend API endpoints created and functional
- Swagger documentation regenerated with all endpoints
- Self-service, parent-portal, faculty-tools fully functional
- Build succeeds with all optimizations
- All E2E tests pass

### Must NOT Have (Guardrails)
- NO `any` types remaining in production code
- NO console.log in production builds
- NO inline functions in JSX without useCallback
- NO raw useEffect for data fetching (use React Query)
- NO missing useEffect cleanup functions
- NO broken functionality during migration
- NO incomplete API endpoints
- NO swagger.json left as empty `{}`

---

## Task Dependency Graph

| Task ID | Task Name | Depends On | Blocks | Reason |
|---------|-----------|------------|--------|--------|
| 1.1.1 | Setup Jest Testing | None | 1.1.2, 1.1.3 | Infrastructure must be ready first |
| 1.1.2 | Create Type Definitions | None | 2.1, 2.2, 2.3 | Types needed for all migrations |
| 1.1.3 | Create Custom Errors | None | 2.1 | Error types needed for catch migration |
| 1.2 | Optimize Build Config | None | 3.5 | Bundle analysis depends on config |
| 1.3 | Console Cleanup | None | None | Independent cleanup task |
| 1.4 | Type API Client | 1.1.2 | 2.2, 2.3 | Typed client needed for React Query |
| 2.1 | Error Handling Migration | 1.1.3 | 2.2, 2.3 | Error types must be defined first |
| 2.2 | React Query Setup | 1.1.2, 1.4 | 2.4, 2.5, 3.x | Infrastructure before migrations |
| 2.3 | Create Query Hooks | 1.1.2, 1.4, 2.2 | 2.4, 2.5 | Hooks needed before page migration |
| 2.4 | Migrate Pages A-M | 2.2, 2.3 | 3.1, 3.2, 3.4 | Performance optimization needs migrated pages |
| 2.5 | Migrate Pages N-Z | 2.2, 2.3 | 3.1, 3.2, 3.4 | Performance optimization needs migrated pages |
| 3.1 | Add useMemo | 2.4, 2.5 | None | Can optimize after migration |
| 3.2 | Add useCallback | 2.4, 2.5 | None | Can optimize after migration |
| 3.3 | Add React.memo | 2.4, 2.5 | None | Can optimize after migration |
| 3.4 | Fix useEffect Cleanup | 2.4, 2.5 | None | Can fix after migration |
| 4.1 | Self-Service APIs | None | 4.4 | Backend can start in parallel with frontend |
| 4.2 | Parent Portal APIs | None | 4.4 | Backend can start in parallel with frontend |
| 4.3 | Faculty Tools APIs | None | 4.4 | Backend can start in parallel with frontend |
| 4.4 | Backend Integration Tests | 4.1, 4.2, 4.3 | 5.1 | APIs must be complete before integration |
| 5.1 | Swagger Documentation | 4.4 | 5.2 | Swagger needs all APIs defined |
| 5.2 | Final Integration & E2E | ALL | None | Final verification step |

---

## Parallel Execution Graph

### Wave 1: Foundation (Start Immediately - No Dependencies)
```
Wave 1 (All can start in parallel):
├── Group 1.1: Infrastructure Setup
│   ├── Task 1.1.1: Setup Jest + React Testing Library (quick)
│   ├── Task 1.1.2: Create Type Definitions (ultrabrain)
│   └── Task 1.1.3: Create Custom Error Classes (quick)
├── Group 1.2: Build Configuration
│   └── Task 1.2: Optimize Next.js Build (quick)
├── Group 1.3: Code Cleanup
│   └── Task 1.3: Remove Console.log Statements (ultrabrain)
└── Group 1.4: API Client
    └── Task 1.4: Add Types to API Client (ultrabrain)
```

### Wave 2: Core Refactoring (Depends on Wave 1)
```
Wave 2 (After Wave 1 Groups 1.1 and 1.4 complete):
├── Group 2.1: Error Handling
│   └── Task 2.1: Migrate Error Handling (catch any → unknown) (ultrabrain)
├── Group 2.2: React Query Infrastructure
│   ├── Task 2.2: Setup React Query Infrastructure (ultrabrain)
│   └── Task 2.3: Create React Query Custom Hooks (ultrabrain)
├── Group 2.3: Page Migration A-M (19 parallel tasks)
│   ├── Task 2.4.1: Migrate analytics/page.tsx (quick)
│   ├── Task 2.4.2: Migrate announcements/page.tsx (quick)
│   ├── Task 2.4.3: Migrate assignments/page.tsx (quick)
│   ├── Task 2.4.4: Migrate attendance/page.tsx (quick)
│   ├── Task 2.4.5: Migrate audit-logs/page.tsx (quick)
│   ├── Task 2.4.6: Migrate auth/login/page.tsx (quick)
│   ├── Task 2.4.7: Migrate batch-operations/page.tsx (quick)
│   ├── Task 2.4.8: Migrate calendar/page.tsx (quick)
│   ├── Task 2.4.9: Migrate courses/page.tsx (quick)
│   ├── Task 2.4.10: Migrate departments/page.tsx (quick)
│   ├── Task 2.4.11: Migrate exams/page.tsx (quick)
│   ├── Task 2.4.12: Migrate faculty-tools/page.tsx (quick)
│   ├── Task 2.4.13: Migrate fees/page.tsx (quick)
│   ├── Task 2.4.14: Migrate files/page.tsx (quick)
│   ├── Task 2.4.15: Migrate forum/page.tsx (quick)
│   ├── Task 2.4.16: Migrate grades/page.tsx (quick)
│   ├── Task 2.4.17: Migrate layout.tsx (quick)
│   ├── Task 2.4.18: Migrate notifications/page.tsx (quick)
│   └── Task 2.4.19: Migrate page.tsx (dashboard) (quick)
└── Group 2.4: Page Migration N-Z (16 parallel tasks)
    ├── Task 2.5.1: Migrate parent-portal/page.tsx (quick)
    ├── Task 2.5.2: Migrate parent-portal/contact/page.tsx (quick)
    ├── Task 2.5.3: Migrate placements/page.tsx (quick)
    ├── Task 2.5.4: Migrate profile/page.tsx (quick)
    ├── Task 2.5.5: Migrate quizzes/page.tsx (quick)
    ├── Task 2.5.6: Migrate quizzes/[quizId]/attempt/[attemptId]/page.tsx (quick)
    ├── Task 2.5.7: Migrate quizzes/[quizId]/results/[attemptId]/page.tsx (quick)
    ├── Task 2.5.8: Migrate roles/page.tsx (quick)
    ├── Task 2.5.9: Migrate self-service/page.tsx (quick)
    ├── Task 2.5.10: Migrate settings/page.tsx (quick)
    ├── Task 2.5.11: Migrate student-dashboard/page.tsx (quick)
    ├── Task 2.5.12: Migrate students/page.tsx (quick)
    ├── Task 2.5.13: Migrate system-status/page.tsx (quick)
    ├── Task 2.5.14: Migrate timetable/page.tsx (quick)
    └── Task 2.5.15: Migrate users/page.tsx (quick)
```

### Wave 3: Performance Optimization (Depends on Wave 2)
```
Wave 3 (After Wave 2 completes):
├── Group 3.1: Computations
│   └── Task 3.1: Add useMemo for Expensive Computations (ultrabrain)
├── Group 3.2: Callbacks
│   └── Task 3.2: Add useCallback for Inline Functions (ultrabrain)
├── Group 3.3: Component Memoization
│   └── Task 3.3: Add React.memo for Pure Components (visual-engineering)
└── Group 3.4: Memory Leak Fixes
    └── Task 3.4: Fix useEffect Cleanup Functions (ultrabrain)
```

### Wave 4: Backend API Development (Can Run Parallel with Waves 1-3)
```
Wave 4 (Can start with Wave 1, independent):
├── Group 4.1: Self-Service Portal APIs
│   ├── Task 4.1.1: Create Self-Service Request Model (quick)
│   ├── Task 4.1.2: Create Self-Service Handler - POST /requests (quick)
│   ├── Task 4.1.3: Create Self-Service Handler - GET /requests (quick)
│   └── Task 4.1.4: Create Self-Service Handler - GET /requests/:id (quick)
├── Group 4.2: Parent Portal APIs
│   ├── Task 4.2.1: Create Parent-Student Link Model (quick)
│   ├── Task 4.2.2: Create Parent Handler - GET /children (quick)
│   ├── Task 4.2.3: Create Parent Handler - GET /children/:id/dashboard (quick)
│   ├── Task 4.2.4: Create Parent Handler - GET /children/:id/attendance (quick)
│   └── Task 4.2.5: Create Parent Handler - GET /children/:id/grades (quick)
└── Group 4.3: Faculty Tools APIs
    ├── Task 4.3.1: Create Rubric Model (quick)
    ├── Task 4.3.2: Create Rubric Handlers - CRUD (quick)
    ├── Task 4.3.3: Create Office Hours Model (quick)
    └── Task 4.3.4: Create Office Hours Handlers - CRUD (quick)

Wave 4 Integration (After Groups 4.1-4.3):
└── Group 4.4: Backend Testing
    └── Task 4.4: Backend Integration Tests (quick)
```

### Wave 5: Integration & Documentation (Depends on ALL Previous Waves)
```
Wave 5 (After ALL previous waves complete):
├── Group 5.1: Documentation
│   └── Task 5.1: Generate Swagger Documentation (quick)
└── Group 5.2: Final Verification
    └── Task 5.2: Final Integration & E2E Testing (visual-engineering)
```

**Critical Path**: Wave 1 → Wave 2 → Wave 3 → Wave 5 (Frontend track)
**Parallel Track**: Wave 4 (Backend) can run concurrently with Waves 1-3
**Estimated Parallel Speedup**: ~60% faster than sequential (backend runs parallel)

---

## Category + Skills Recommendations

### For Each Wave:

**Wave 1 (Foundation)**:
- **Category**: `quick` for setup/config tasks, `ultrabrain` for type definitions
- **Skills**: `frontend-ui-ux`, `git-master` (for setup), `typescript-programmer`
- **Reason**: Setup tasks are well-defined; type definitions require deep TypeScript knowledge

**Wave 2 (Core Refactoring)**:
- **Category**: `ultrabrain` for all tasks
- **Skills**: `frontend-ui-ux`, `typescript-programmer`, `git-master`
- **Reason**: Complex refactoring requiring React Query expertise and careful type handling

**Wave 3 (Performance)**:
- **Category**: `ultrabrain` for useMemo/useCallback, `visual-engineering` for React.memo
- **Skills**: `frontend-ui-ux`, `typescript-programmer`
- **Reason**: Performance optimization requires profiling judgment; component memoization is visual

**Wave 4 (Backend)**:
- **Category**: `unspecified-high` for API development
- **Skills**: None (Go backend doesn't match available skills)
- **Reason**: Backend Go development - skills are frontend-focused, need manual expertise

**Wave 5 (Integration)**:
- **Category**: `visual-engineering`
- **Skills**: `frontend-ui-ux`, `agent-browser` (for E2E testing)
- **Reason**: Integration and E2E testing requires browser automation

---

## TODOs

---

### WAVE 1: FOUNDATION (Independent - Can Start Immediately)

---

#### **TASK 1.1.1: Setup Jest + React Testing Library**

**What to do**:
1. Install Jest, React Testing Library, and types
2. Create `jest.config.ts` with Next.js support
3. Create `jest.setup.ts` with testing-library/jest-dom
4. Add test scripts to package.json
5. Create example test to verify setup works

**Must NOT do**:
- Don't write tests for all components yet (just setup infrastructure)
- Don't modify source code

**Recommended Agent Profile**:
- **Category**: `quick` (setup task, well-defined)
- **Skills**: `frontend-ui-ux`, `git-master`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: React Testing Library configuration requires frontend expertise
  - ✅ INCLUDED `git-master`: Setup involves package.json changes, needs atomic commits
  - ❌ OMITTED `typescript-programmer`: Setup is configuration, not complex TypeScript logic

**Parallelization**:
- **Can Run In Parallel**: YES
- **Parallel Group**: Wave 1 - Group 1.1
- **Blocks**: 1.1.2, 1.1.3
- **Blocked By**: None

**References**:
- `/Users/kasyap/Documents/edduhub/client/package.json` - Check existing scripts and dependencies
- Next.js testing documentation - Official Jest setup guide
- React Testing Library docs - Best practices

**Acceptance Criteria**:
```bash
# Verification commands:
bun add -d jest @testing-library/react @testing-library/jest-dom @testing-library/user-event jest-environment-jsdom @types/jest

# After setup:
bun run test -- --listTests  # Should show example test

# Example test passes:
bun run test  # Output: 1 test passed

# Evidence:
ls client/jest.config.ts  # File exists
ls client/jest.setup.ts   # File exists
grep "test" client/package.json  # Shows test scripts
```

**Commit**: YES
- Message: `chore(test): setup jest and react testing library`
- Files: `jest.config.ts`, `jest.setup.ts`, `package.json`, `src/__tests__/example.test.tsx`

---

#### **TASK 1.1.2: Create Type Definitions File**

**What to do**:
1. Create `/Users/kasyap/Documents/edduhub/client/src/types/index.ts`
2. Define all API response types based on backend models
3. Define form data types
4. Define component prop types
5. Export all types from index

**Must NOT do**:
- Don't import these types yet (that comes in Wave 2)
- Don't use `any` anywhere in this file

**Recommended Agent Profile**:
- **Category**: `ultrabrain` (requires understanding backend Go types)
- **Skills**: `typescript-programmer`, `frontend-ui-ux`
- **Skills Evaluation**:
  - ✅ INCLUDED `typescript-programmer`: Complex type definitions require TypeScript expertise
  - ✅ INCLUDED `frontend-ui-ux`: Types are for frontend consumption
  - ❌ OMITTED `git-master`: Single file creation, simple commit

**Parallelization**:
- **Can Run In Parallel**: YES (with 1.1.1)
- **Parallel Group**: Wave 1 - Group 1.1
- **Blocks**: 2.1, 2.2, 2.3 (all type migrations)
- **Blocked By**: None

**References**:
- Backend models: `/Users/kasyap/Documents/edduhub/server/internal/models/` (check all .go files)
- API responses in: `/Users/kasyap/Documents/edduhub/server/api/handler/` (check response structs)
- Current `any` usage locations (from analysis)
- Go struct to TypeScript mapping patterns

**Type Definitions to Create**:
```typescript
// User types
export interface User {
  id: string;
  email: string;
  name: { first: string; last: string };
  role: 'student' | 'faculty' | 'admin';
  collegeId: string;
  // ... map from backend
}

// Dashboard types
export interface DashboardData {
  courseGrades: CourseGrade[];
  attendance: AttendanceStats[];
  assignments: Assignment[];
  recentGrades: Grade[];
}

export interface CourseGrade {
  id: string;
  name: string;
  credits: number;
  grade: string;
  points: number;
}

export interface AttendanceStats {
  courseId: string;
  totalSessions: number;
  presentCount: number;
  percentage: number;
}

// Self-Service types
export interface SelfServiceRequest {
  id: string;
  type: 'enrollment' | 'schedule_change' | 'document';
  status: 'pending' | 'approved' | 'rejected';
  description: string;
  createdAt: string;
  updatedAt: string;
}

// Parent Portal types
export interface ParentChild {
  id: string;
  name: string;
  email: string;
  dashboard: StudentDashboard;
}

// Faculty Tools types
export interface GradingRubric {
  id: string;
  name: string;
  criteria: RubricCriteria[];
  maxScore: number;
}

export interface OfficeHoursSlot {
  id: string;
  day: string;
  startTime: string;
  endTime: string;
  location: string;
}

// ... all other types
```

**Acceptance Criteria**:
```bash
# File exists and exports types:
ls /Users/kasyap/Documents/edduhub/client/src/types/index.ts

# No any types in the file:
grep -n "any" /Users/kasyap/Documents/edduhub/client/src/types/index.ts || echo "No any types found - PASS"

# TypeScript compiles:
bun run type-check  # Should pass

# Evidence:
wc -l client/src/types/index.ts  # Should be substantial (>200 lines)
grep "export interface" client/src/types/index.ts | wc -l  # Should show many interfaces
```

**Commit**: YES
- Message: `feat(types): add comprehensive TypeScript type definitions`
- Files: `src/types/index.ts`

---

#### **TASK 1.1.3: Create Custom Error Classes**

**What to do**:
1. Create `/Users/kasyap/Documents/edduhub/client/src/lib/errors.ts`
2. Define custom error classes: APIError, ValidationError, NetworkError, AuthError
3. Add error type guards: `isAPIError()`, `isValidationError()`, etc.
4. Add error message extraction helpers

**Must NOT do**:
- Don't use these errors yet (that comes in Wave 2)
- Don't import into existing files

**Recommended Agent Profile**:
- **Category**: `quick` (well-defined utility)
- **Skills**: `frontend-ui-ux`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: Error handling is frontend concern
  - ❌ OMITTED `typescript-programmer`: Simple class definitions
  - ❌ OMITTED `git-master`: Single file utility

**Parallelization**:
- **Can Run In Parallel**: YES (with 1.1.1, 1.1.2)
- **Parallel Group**: Wave 1 - Group 1.1
- **Blocks**: 2.1 (error handling migration)
- **Blocked By**: None

**Error Classes to Create**:
```typescript
// Base API Error
export class APIError extends Error {
  constructor(
    message: string,
    public statusCode: number,
    public code: string,
    public validationErrors?: Record<string, string[]>
  ) {
    super(message);
    this.name = 'APIError';
  }
}

// Specific error types
export class ValidationError extends APIError {
  constructor(message: string, validationErrors: Record<string, string[]>) {
    super(message, 400, 'VALIDATION_ERROR', validationErrors);
    this.name = 'ValidationError';
  }
}

export class NetworkError extends Error {
  constructor(message: string = 'Network request failed') {
    super(message);
    this.name = 'NetworkError';
  }
}

export class AuthError extends APIError {
  constructor(message: string = 'Authentication required') {
    super(message, 401, 'AUTH_ERROR');
    this.name = 'AuthError';
  }
}

// Type guards
export function isAPIError(error: unknown): error is APIError {
  return error instanceof APIError;
}

export function isValidationError(error: unknown): error is ValidationError {
  return error instanceof ValidationError;
}

export function isNetworkError(error: unknown): error is NetworkError {
  return error instanceof NetworkError;
}

// Message extraction
export function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}
```

**Acceptance Criteria**:
```bash
# File exists:
ls /Users/kasyap/Documents/edduhub/client/src/lib/errors.ts

# TypeScript compiles:
bun run type-check

# Can import and instantiate:
bun -e "import { APIError, isAPIError } from './src/lib/errors'; const e = new APIError('test', 500, 'TEST'); console.log(isAPIError(e));"  # Should output: true

# Evidence:
grep "export class" client/src/lib/errors.ts | wc -l  # Should be 4 classes
grep "export function" client/src/lib/errors.ts | wc -l  # Should be 4+ functions
```

**Commit**: YES
- Message: `feat(errors): add custom error classes and type guards`
- Files: `src/lib/errors.ts`

---

#### **TASK 1.2: Optimize Next.js Build Configuration**

**What to do**:
1. Update `/Users/kasyap/Documents/edduhub/client/next.config.ts`
2. Add `compiler.removeConsole` for production builds
3. Add `experimental.optimizePackageImports` for heavy packages
4. Add webpack bundle analyzer (optional, dev only)
5. Keep existing security headers

**Must NOT do**:
- Don't remove security headers (keep existing)
- Don't change output or distDir settings
- Don't add experimental features that could break build

**Recommended Agent Profile**:
- **Category**: `quick` (configuration change)
- **Skills**: `frontend-ui-ux`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: Next.js configuration is frontend domain
  - ❌ OMITTED `typescript-programmer`: Configuration, not TypeScript logic
  - ❌ OMITTED `git-master`: Simple config change

**Parallelization**:
- **Can Run In Parallel**: YES
- **Parallel Group**: Wave 1 - Group 1.2
- **Blocks**: 3.5 (bundle analysis requires config)
- **Blocked By**: None

**Config Changes**:
```typescript
import type { NextConfig } from "next";
import withBundleAnalyzer from '@next/bundle-analyzer';

const nextConfig: NextConfig = {
  // Remove console logs in production
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production' ? {
      exclude: ['error'], // Keep console.error for production debugging
    } : false,
  },
  
  // Optimize heavy package imports
  experimental: {
    optimizePackageImports: [
      'lucide-react',
      'date-fns',
      'recharts',  // if used
    ],
  },
  
  // Existing security headers (KEEP THESE)
  async headers() {
    return [/* existing headers */];
  },
};

// Enable bundle analyzer in ANALYZE mode
const withAnalyzer = withBundleAnalyzer({
  enabled: process.env.ANALYZE === 'true',
});

export default withAnalyzer(nextConfig);
```

**Package Installation**:
```bash
bun add -d @next/bundle-analyzer
```

**Acceptance Criteria**:
```bash
# Build succeeds:
bun run build

# No console.log in production build (check compiled output):
grep -r "console.log" .next/static/ || echo "No console.log in build - PASS"

# Bundle analyzer works:
ANALYZE=true bun run build  # Generates bundle analysis

# Evidence:
ls .next/analyze/  # Should exist after ANALYZE build
```

**Commit**: YES
- Message: `perf(config): add build optimizations and bundle analyzer`
- Files: `next.config.ts`, `package.json`

---

#### **TASK 1.3: Remove Console.log Statements**

**What to do**:
1. Remove or replace all 30 console.log/error statements in production code
2. Use the logger utility (`/Users/kasyap/Documents/edduhub/client/src/lib/logger.ts`) for legitimate error logging
3. Remove debug console.logs entirely
4. Keep error-boundary.tsx console.error (legitimate error reporting)

**Files to Modify** (11 files, 30 statements total):
| File | Line Numbers | Count | Action |
|------|--------------|-------|--------|
| `client/src/app/timetable/page.tsx` | Lines TBD | 4 | Replace with logger |
| `client/src/app/roles/page.tsx` | Lines TBD | 3 | Remove debug logs |
| `client/src/app/self-service/page.tsx` | Lines TBD | 3 | Replace with logger |
| `client/src/app/notifications/page.tsx` | Lines TBD | 7 | Replace WebSocket logs |
| `client/src/app/placements/page.tsx` | Lines TBD | 3 | Remove debug logs |
| `client/src/app/forum/page.tsx` | Lines TBD | 2 | Replace with logger |
| `client/src/app/analytics/page.tsx` | Lines TBD | 1 | Remove debug log |
| `client/src/app/parent-portal/page.tsx` | Lines TBD | 1 | Replace with logger |
| `client/src/app/parent-portal/contact/page.tsx` | Lines TBD | 1 | Remove debug log |
| `client/src/lib/logger.ts` | Lines TBD | 5 | KEEP THESE (intentional) |
| `client/src/components/error-boundary.tsx` | Lines TBD | 1 | KEEP THIS (legitimate) |

**Must NOT do**:
- Don't remove logger.ts console statements (those are intentional)
- Don't remove error-boundary.tsx console.error (legitimate)
- Don't add new console statements

**Recommended Agent Profile**:
- **Category**: `ultrabrain` (requires judgment on which to remove vs replace)
- **Skills**: `frontend-ui-ux`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: Understanding error handling context requires frontend expertise
  - ❌ OMITTED `typescript-programmer`: Logic is simple, judgment is key
  - ❌ OMITTED `git-master`: Multiple file changes, but straightforward

**Parallelization**:
- **Can Run In Parallel**: YES
- **Parallel Group**: Wave 1 - Group 1.3
- **Blocks**: None
- **Blocked By**: None

**Changes per File**:

**timetable/page.tsx**:
```typescript
// BEFORE:
} catch (error) {
  console.error('Failed to fetch timetable:', error);
}

// AFTER:
import { logger } from '@/lib/logger';
import { getErrorMessage } from '@/lib/errors';

} catch (error: unknown) {
  logger.error('Failed to fetch timetable', { error: getErrorMessage(error) });
}
```

**notifications/page.tsx** (WebSocket console.log):
```typescript
// BEFORE:
console.log('WebSocket connection closed. Reconnecting...');

// AFTER (remove entirely - use logger.debug if needed):
import { logger } from '@/lib/logger';
logger.debug('WebSocket connection closed. Reconnecting...');
```

**Acceptance Criteria**:
```bash
# Verify no console.log in production files (excluding tests and logger):
grep -r "console\.log" client/src/app/ client/src/components/ --include="*.tsx" | grep -v "error-boundary" | wc -l  # Should be 0

# Verify console.error only in error-boundary and logger:
grep -r "console\.error" client/src/app/ client/src/components/ --include="*.tsx" | grep -v "error-boundary" | grep -v "logger.ts" | wc -l  # Should be 0

# Build succeeds:
bun run build

# Evidence:
echo "Console statements eliminated: $(grep -r 'console\\.log' client/src/app/ client/src/components/ --include='*.tsx' | wc -l)"
```

**Commit**: YES
- Message: `refactor(console): replace console statements with logger utility`
- Files: 10 modified files (list in commit body)

---

#### **TASK 1.4: Add Types to API Client**

**What to do**:
1. Update `/Users/kasyap/Documents/edduhub/client/src/lib/api-client.ts`
2. Add generic type parameter to api.get/post/put/delete methods
3. Replace `any` type assertions with proper types
4. Update error handling to use custom error types

**Must NOT do**:
- Don't change API endpoint URLs
- Don't change request/response structure
- Don't break existing usage (just add types)

**Recommended Agent Profile**:
- **Category**: `ultrabrain` (requires understanding current API client structure)
- **Skills**: `frontend-ui-ux`, `typescript-programmer`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: API client is core frontend infrastructure
  - ✅ INCLUDED `typescript-programmer`: Generic types and error handling require TypeScript expertise
  - ❌ OMITTED `git-master`: Single file modification

**Parallelization**:
- **Can Run In Parallel**: YES
- **Parallel Group**: Wave 1 - Group 1.4
- **Blocks**: 2.2, 2.3 (React Query migration needs typed API)
- **Blocked By**: 1.1.2 (needs types defined)

**API Client Changes**:
```typescript
// BEFORE:
async get<T>(url: string): Promise<T> {
  const response = await fetch(url);
  return response.json() as any;  // ❌ any
}

// AFTER:
import { APIError } from '@/lib/errors';

async get<T>(url: string): Promise<T> {
  const response = await fetch(url);
  if (!response.ok) {
    throw new APIError(
      `HTTP ${response.status}`,
      response.status,
      'HTTP_ERROR'
    );
  }
  return response.json() as T;  // ✅ Generic type
}
```

**Acceptance Criteria**:
```bash
# TypeScript compiles:
bun run type-check

# API client can be used with types:
bun -e "import { api } from './src/lib/api-client'; import { User } from './src/types'; api.get<User>('/api/user')"

# No any types in api-client.ts:
grep -n "any" client/src/lib/api-client.ts || echo "No any types - PASS"

# Evidence:
grep "async get<T>" client/src/lib/api-client.ts  # Should show generic method
grep "async post<T>" client/src/lib/api-client.ts  # Should show generic method
```

**Commit**: YES
- Message: `feat(api): add TypeScript generics to api client`
- Files: `src/lib/api-client.ts`

---

### WAVE 2: CORE REFACTORING (Depends on Wave 1)

---

#### **TASK 2.1: Migrate Error Handling (catch any → unknown)**

**What to do**:
1. Update 26 files to replace `catch (e: any)` with `catch (error: unknown)`
2. Add type guards using custom error classes
3. Use `getErrorMessage()` helper for consistent error messages
4. Update error handling logic to check error types

**Files to Modify** (26 files with catch blocks):
| File | Count | Priority |
|------|-------|----------|
| `client/src/app/page.tsx` | 2 | HIGH |
| `client/src/app/grades/page.tsx` | 1 | MEDIUM |
| `client/src/app/assignments/page.tsx` | 1 | MEDIUM |
| `client/src/app/quizzes/[quizId]/attempt/[attemptId]/page.tsx` | 2 | MEDIUM |
| `client/src/app/quizzes/[quizId]/results/[attemptId]/page.tsx` | 1 | MEDIUM |
| `client/src/app/quizzes/page.tsx` | 2 | MEDIUM |
| `client/src/app/auth/login/page.tsx` | 1 | HIGH |
| `client/src/app/system-status/page.tsx` | 1 | MEDIUM |
| `client/src/app/files/page.tsx` | 7 | HIGH |
| `client/src/app/webhooks/page.tsx` | 5 | MEDIUM |
| `client/src/app/batch-operations/page.tsx` | 4 | MEDIUM |

**Must NOT do**:
- Don't change error handling logic, only the typing
- Don't remove error handling, only improve typing

**Recommended Agent Profile**:
- **Category**: `ultrabrain` (complex refactoring across many files)
- **Skills**: `frontend-ui-ux`, `typescript-programmer`, `git-master`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: Error handling is frontend concern
  - ✅ INCLUDED `typescript-programmer`: Type guards and unknown types require expertise
  - ✅ INCLUDED `git-master`: 26 files need atomic commits

**Parallelization**:
- **Can Run In Parallel**: NO (depends on 1.1.3)
- **Parallel Group**: Wave 2 - Group 2.1
- **Blocks**: 2.2, 2.3 (React Query migration)
- **Blocked By**: 1.1.3 (error types)

**Change Pattern**:
```typescript
// BEFORE:
try {
  await api.post('/api/data', payload);
} catch (e: any) {
  console.error('Failed:', e.message);
  setError(e.message);
}

// AFTER:
import { getErrorMessage, isAPIError } from '@/lib/errors';

try {
  await api.post('/api/data', payload);
} catch (error: unknown) {
  logger.error('Failed to save data', { error: getErrorMessage(error) });
  
  if (isAPIError(error) && error.validationErrors) {
    setValidationErrors(error.validationErrors);
  } else {
    setError(getErrorMessage(error));
  }
}
```

**Acceptance Criteria**:
```bash
# No catch (e: any) patterns remain:
grep -r "catch.*any" client/src/app/ --include="*.tsx" | wc -l  # Should be 0

# TypeScript compiles:
bun run type-check

# Build succeeds:
bun run build

# Evidence:
grep -r "catch.*unknown" client/src/app/ --include="*.tsx" | wc -l  # Should be 26+ (one per catch)
```

**Commit**: YES
- Message: `refactor(errors): migrate catch any to catch unknown with type guards`
- Files: 26 modified files

---

#### **TASK 2.2: Setup React Query Infrastructure**

**What to do**:
1. Install React Query (@tanstack/react-query)
2. Create QueryClient provider component
3. Add QueryClientProvider to root layout.tsx
4. Create custom hooks for common queries
5. Add React Query Devtools (dev only)

**Must NOT do**:
- Don't migrate pages yet (that comes in 2.3 and 2.4)
- Don't remove existing useEffect code yet
- Don't break existing functionality

**Recommended Agent Profile**:
- **Category**: `ultrabrain` (setup requires understanding React Query patterns)
- **Skills**: `frontend-ui-ux`, `typescript-programmer`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: React Query is frontend infrastructure
  - ✅ INCLUDED `typescript-programmer`: Provider setup requires TypeScript
  - ❌ OMITTED `git-master`: Infrastructure setup, single commit

**Parallelization**:
- **Can Run In Parallel**: YES (with 2.1)
- **Parallel Group**: Wave 2 - Group 2.2
- **Blocks**: 2.3, 2.4 (React Query migrations)
- **Blocked By**: 1.1.2 (types), 1.4 (typed API client)

**Setup Steps**:
```bash
bun add @tanstack/react-query @tanstack/react-query-devtools
```

**Create `/Users/kasyap/Documents/edduhub/client/src/providers/query-provider.tsx`**:
```typescript
'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { useState } from 'react';

export function QueryProvider({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 1000 * 60 * 5, // 5 minutes
        gcTime: 1000 * 60 * 30,   // 30 minutes
        refetchOnWindowFocus: false,
        retry: 1,
      },
    },
  }));

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}
```

**Update `/Users/kasyap/Documents/edduhub/client/src/app/layout.tsx`**:
```typescript
import { QueryProvider } from '@/providers/query-provider';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        <QueryProvider>
          {children}
        </QueryProvider>
      </body>
    </html>
  );
}
```

**Acceptance Criteria**:
```bash
# React Query installed:
bun list @tanstack/react-query

# Provider component exists:
ls client/src/providers/query-provider.tsx

# TypeScript compiles:
bun run type-check

# App still runs:
bun run dev  # Should start without errors

# Evidence:
grep "QueryProvider" client/src/app/layout.tsx  # Should show import and usage
```

**Commit**: YES
- Message: `feat(react-query): setup React Query infrastructure`
- Files: `src/providers/query-provider.tsx`, `src/app/layout.tsx`, `package.json`

---

#### **TASK 2.3: Create React Query Custom Hooks**

**What to do**:
1. Create `/Users/kasyap/Documents/edduhub/client/src/hooks/queries/` directory
2. Create useDashboard, useStudents, useCourses, etc. hooks
3. Each hook wraps api calls with useQuery/useMutation
4. Add proper typing with types from 1.1.2
5. Add error handling with custom errors

**Must NOT do**:
- Don't migrate pages to use these hooks yet
- Don't remove existing fetch functions yet
- Don't change API endpoints

**Recommended Agent Profile**:
- **Category**: `ultrabrain` (requires understanding all data fetching patterns)
- **Skills**: `frontend-ui-ux`, `typescript-programmer`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: React Query hooks are core frontend patterns
  - ✅ INCLUDED `typescript-programmer`: Hook typing requires TypeScript expertise
  - ❌ OMITTED `git-master`: Multiple files but simple structure

**Parallelization**:
- **Can Run In Parallel**: YES (with 2.1, 2.2)
- **Parallel Group**: Wave 2 - Group 2.3
- **Blocks**: 2.4 (page migrations)
- **Blocked By**: 1.1.2 (types), 1.4 (typed API), 2.2 (React Query setup)

**Hooks to Create** (one file per domain):

**`/Users/kasyap/Documents/edduhub/client/src/hooks/queries/useDashboard.ts`**:
```typescript
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api-client';
import { DashboardData } from '@/types';

export function useDashboard() {
  return useQuery({
    queryKey: ['dashboard'],
    queryFn: async () => {
      const { data } = await api.get<DashboardData>('/api/dashboard');
      return data;
    },
  });
}
```

**`/Users/kasyap/Documents/edduhub/client/src/hooks/queries/useStudents.ts`**:
```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api-client';
import { Student, CreateStudentInput } from '@/types';

export function useStudents() {
  return useQuery({
    queryKey: ['students'],
    queryFn: async () => {
      const { data } = await api.get<Student[]>('/api/students');
      return data;
    },
  });
}

export function useCreateStudent() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (input: CreateStudentInput) => {
      const { data } = await api.post<Student>('/api/students', input);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['students'] });
    },
  });
}
```

**Additional hooks needed**:
- useCourses.ts
- useAssignments.ts
- useQuizzes.ts
- useGrades.ts
- useAttendance.ts
- useFaculty.ts
- useSelfService.ts (new - for self-service portal)
- useParent.ts (new - for parent portal)
- useFacultyTools.ts (new - for faculty rubrics/office hours)

**Acceptance Criteria**:
```bash
# Hooks directory exists:
ls client/src/hooks/queries/

# All hooks compile:
bun run type-check

# Can import and use hooks:
bun -e "import { useDashboard } from './src/hooks/queries/useDashboard'; console.log('Hook imported successfully')"

# Evidence:
ls client/src/hooks/queries/*.ts | wc -l  # Should show 10+ hook files
grep "useQuery" client/src/hooks/queries/*.ts | wc -l  # Should show multiple useQuery calls
```

**Commit**: YES
- Message: `feat(hooks): add React Query custom hooks for all data fetching`
- Files: `src/hooks/queries/*.ts`

---

#### **TASK 2.4.x: Migrate Pages to React Query (A-M - 19 Parallel Tasks)**

**What to do** (for each page):
1. Replace useEffect + useState patterns with useQuery
2. Replace manual mutations with useMutation
3. Remove fetch functions (now handled by React Query)
4. Update loading states to use isLoading/isPending

**Pages to Migrate** (19 pages, each as separate parallel task):

**TASK 2.4.1: analytics/page.tsx**
**TASK 2.4.2: announcements/page.tsx**
**TASK 2.4.3: assignments/page.tsx**
**TASK 2.4.4: attendance/page.tsx**
**TASK 2.4.5: audit-logs/page.tsx**
**TASK 2.4.6: auth/login/page.tsx**
**TASK 2.4.7: batch-operations/page.tsx**
**TASK 2.4.8: calendar/page.tsx**
**TASK 2.4.9: courses/page.tsx**
**TASK 2.4.10: departments/page.tsx**
**TASK 2.4.11: exams/page.tsx**
**TASK 2.4.12: faculty-tools/page.tsx**
**TASK 2.4.13: fees/page.tsx**
**TASK 2.4.14: files/page.tsx**
**TASK 2.4.15: forum/page.tsx**
**TASK 2.4.16: grades/page.tsx**
**TASK 2.4.17: layout.tsx**
**TASK 2.4.18: notifications/page.tsx**
**TASK 2.4.19: page.tsx (dashboard)**

**Must NOT do**:
- Don't break existing functionality
- Don't change UI/UX behavior
- Don't remove error handling

**Recommended Agent Profile** (for each task):
- **Category**: `quick` (each page is well-defined)
- **Skills**: `frontend-ui-ux`, `typescript-programmer`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: Page migration requires React expertise
  - ✅ INCLUDED `typescript-programmer`: Hook typing and data typing require expertise
  - ❌ OMITTED `git-master`: Single file per task

**Parallelization**:
- **Can Run In Parallel**: YES (each page is independent)
- **Parallel Group**: Wave 2 - Group 2.4 (all 19 tasks parallel)
- **Blocks**: 3.1, 3.2, 3.4 (optimization tasks)
- **Blocked By**: 2.2, 2.3 (React Query setup and hooks)

**Migration Pattern**:
```typescript
// BEFORE:
const [data, setData] = useState([]);
const [loading, setLoading] = useState(false);
const [error, setError] = useState(null);

useEffect(() => {
  const fetchData = async () => {
    setLoading(true);
    try {
      const { data } = await api.get('/api/students');
      setData(data);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };
  fetchData();
}, []);

// AFTER:
import { useStudents } from '@/hooks/queries/useStudents';

const { data, isLoading, error } = useStudents();

if (isLoading) return <Loading />;
if (error) return <Error message={error.message} />;
```

**Acceptance Criteria per Page**:
```bash
# Page has no useEffect for data fetching:
grep -n "useEffect.*fetch" client/src/app/PAGE/page.tsx || echo "No fetch useEffect - PASS"

# Page uses React Query:
grep -n "useQuery\|useMutation" client/src/app/PAGE/page.tsx  # Should find usage

# TypeScript compiles:
bun run type-check

# Evidence:
echo "Pages migrated to React Query: $(grep -r 'useQuery' client/src/app/ --include='*.tsx' | wc -l)"
```

**Commit**: YES (one commit per page)
- Message: `refactor(PAGE): migrate to React Query`
- Files: `src/app/PAGE/page.tsx`

---

#### **TASK 2.5.x: Migrate Pages to React Query (N-Z - 16 Parallel Tasks)**

**Pages to Migrate**:
**TASK 2.5.1: parent-portal/page.tsx**
**TASK 2.5.2: parent-portal/contact/page.tsx**
**TASK 2.5.3: placements/page.tsx**
**TASK 2.5.4: profile/page.tsx**
**TASK 2.5.5: quizzes/page.tsx**
**TASK 2.5.6: quizzes/[quizId]/attempt/[attemptId]/page.tsx**
**TASK 2.5.7: quizzes/[quizId]/results/[attemptId]/page.tsx**
**TASK 2.5.8: roles/page.tsx**
**TASK 2.5.9: self-service/page.tsx**
**TASK 2.5.10: settings/page.tsx**
**TASK 2.5.11: student-dashboard/page.tsx**
**TASK 2.5.12: students/page.tsx**
**TASK 2.5.13: system-status/page.tsx**
**TASK 2.5.14: timetable/page.tsx**
**TASK 2.5.15: users/page.tsx**
**TASK 2.5.16: webhooks/page.tsx**

**Acceptance Criteria**: Same as Task 2.4.x

**Commit**: YES (one commit per page or logical group)

---

### WAVE 3: PERFORMANCE OPTIMIZATIONS (Depends on Wave 2)

---

#### **TASK 3.1: Add useMemo for Expensive Computations**

**What to do**:
1. Add useMemo for GPA calculations in page.tsx
2. Add useMemo for attendance rate calculations
3. Add useMemo for filtered data (notifications, forum threads)
4. Add useMemo for mapped arrays in render

**Files to Optimize**:
| File | Computations to Memoize | Lines |
|------|------------------------|-------|
| `client/src/app/page.tsx` | GPA, attendance, pending tasks | TBD |
| `client/src/app/notifications/page.tsx` | filteredNotifications | TBD |
| `client/src/app/forum/page.tsx` | filteredThreads | TBD |
| `client/src/app/student-dashboard/page.tsx` | Multiple computations | TBD |
| `client/src/app/analytics/page.tsx` | Chart data | TBD |
| `client/src/app/advanced-analytics/page.tsx` | Student/course lists | TBD |

**Must NOT do**:
- Don't memoize simple calculations (not worth overhead)
- Don't change logic, only wrap with useMemo
- Don't forget dependency arrays

**Recommended Agent Profile**:
- **Category**: `ultrabrain` (requires understanding which computations are expensive)
- **Skills**: `frontend-ui-ux`, `typescript-programmer`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: Performance optimization requires React profiling judgment
  - ✅ INCLUDED `typescript-programmer`: useMemo typing and dependencies require expertise
  - ❌ OMITTED `git-master`: Multiple files but straightforward

**Parallelization**:
- **Can Run In Parallel**: YES (after Wave 2)
- **Parallel Group**: Wave 3 - Group 3.1
- **Blocks**: None
- **Blocked By**: 2.4, 2.5 (pages must be migrated first)

**Example Optimization**:
```typescript
// BEFORE:
const totalCredits = courseGrades.reduce((acc, c) => acc + (c.credits || 3), 0);
const totalPoints = courseGrades.reduce((acc, c) => acc + (c.points || 0), 0);
const gpa = totalPoints / totalCredits;

// AFTER:
const { gpa, totalCredits } = useMemo(() => {
  const totalCredits = courseGrades.reduce((acc, c) => acc + (c.credits || 3), 0);
  const totalPoints = courseGrades.reduce((acc, c) => acc + (c.points || 0), 0);
  return {
    gpa: totalCredits > 0 ? totalPoints / totalCredits : 0,
    totalCredits,
  };
}, [courseGrades]);
```

**Acceptance Criteria**:
```bash
# Computations wrapped in useMemo:
grep -n "useMemo" client/src/app/page.tsx  # Should find GPA, attendance calculations

# TypeScript compiles:
bun run type-check

# No performance regressions (build succeeds):
bun run build

# Evidence:
grep -r "useMemo" client/src/app/ --include="*.tsx" | wc -l  # Should show 15+ uses
```

**Commit**: YES
- Message: `perf(components): add useMemo for expensive computations`
- Files: 6-8 modified files

---

#### **TASK 3.2: Add useCallback for Inline Functions**

**What to do**:
1. Add useCallback for all inline functions inside .map() (27 critical locations)
2. Add useCallback for other inline handlers (130+ locations)
3. Use proper dependency arrays
4. Focus on functions passed as props to child components

**Files with Critical .map() Inline Functions**:
| File | Functions to Memoize | Priority |
|------|---------------------|----------|
| `client/src/app/timetable/page.tsx` | 2 functions | HIGH |
| `client/src/app/students/page.tsx` | 2 functions | HIGH |
| `client/src/app/assignments/page.tsx` | 1 function | HIGH |
| `client/src/app/parent-portal/page.tsx` | 1 function | MEDIUM |
| `client/src/app/announcements/page.tsx` | 1 function | MEDIUM |
| `client/src/app/roles/page.tsx` | 2 functions | MEDIUM |
| `client/src/app/webhooks/page.tsx` | 3 functions | MEDIUM |
| `client/src/app/quizzes/page.tsx` | 1 function | MEDIUM |
| `client/src/app/placements/page.tsx` | 2 functions | MEDIUM |
| `client/src/app/files/page.tsx` | 7 functions | HIGH |
| `client/src/app/audit-logs/page.tsx` | 2 functions | MEDIUM |
| `client/src/app/users/page.tsx` | 3 functions | MEDIUM |
| `client/src/app/forum/page.tsx` | 1 function | LOW |
| `client/src/app/fees/page.tsx` | 1 function | LOW |

**Must NOT do**:
- Don't use useCallback for functions not passed as props
- Don't forget dependency arrays
- Don't break existing functionality

**Recommended Agent Profile**:
- **Category**: `ultrabrain` (massive refactoring task)
- **Skills**: `frontend-ui-ux`, `typescript-programmer`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: useCallback optimization requires React expertise
  - ✅ INCLUDED `typescript-programmer`: 130+ functions need proper typing and dependencies
  - ❌ OMITTED `git-master`: Many files but straightforward pattern

**Parallelization**:
- **Can Run In Parallel**: YES (after Wave 2)
- **Parallel Group**: Wave 3 - Group 3.2
- **Blocks**: None
- **Blocked By**: 2.4, 2.5

**Example Optimization**:
```typescript
// BEFORE (inside .map()):
{blocks.map(block => (
  <div key={block.id}>
    <button onClick={() => openEditDialog(block)}>Edit</button>
    <button onClick={() => handleDelete(block.id)}>Delete</button>
  </div>
))}

// AFTER:
const handleEdit = useCallback((block: TimetableBlock) => {
  openEditDialog(block);
}, [openEditDialog]);

const handleDelete = useCallback((blockId: string) => {
  handleDeleteBlock(blockId);
}, [handleDeleteBlock]);

{blocks.map(block => (
  <div key={block.id}>
    <button onClick={() => handleEdit(block)}>Edit</button>
    <button onClick={() => handleDelete(block.id)}>Delete</button>
  </div>
))}
```

**Acceptance Criteria**:
```bash
# Critical .map() functions use useCallback:
grep -n "useCallback" client/src/app/timetable/page.tsx  # Should be present

# TypeScript compiles:
bun run type-check

# Build succeeds:
bun run build

# Evidence:
grep -r "useCallback" client/src/app/ --include="*.tsx" | wc -l  # Should be 130+
```

**Commit**: YES
- Message: `perf(components): add useCallback for inline functions`
- Files: 14+ modified files

---

#### **TASK 3.3: Add React.memo for Pure Components**

**What to do**:
1. Identify pure components (no side effects, props-only)
2. Add React.memo wrapper
3. Focus on reusable components in `/components`
4. Add displayName for debugging

**Candidates for Memoization**:
| Component | File | Reason |
|-----------|------|--------|
| Button | `client/src/components/ui/button.tsx` | Reusable, props-only |
| Card | `client/src/components/ui/card.tsx` | Reusable, props-only |
| Input | `client/src/components/ui/input.tsx` | Reusable, props-only |
| Select | `client/src/components/ui/select.tsx` | Reusable, props-only |
| Sidebar | `client/src/components/navigation/sidebar.tsx` | Props-only |
| Topbar | `client/src/components/navigation/topbar.tsx` | Props-only |
| List items | Various | Rendered in .map() loops |

**Must NOT do**:
- Don't memoize components with side effects
- Don't memoize components that use context (unless stable)
- Don't change component logic

**Recommended Agent Profile**:
- **Category**: `visual-engineering` (component optimization)
- **Skills**: `frontend-ui-ux`, `typescript-programmer`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: Component memoization is visual optimization
  - ✅ INCLUDED `typescript-programmer`: React.memo typing
  - ❌ OMITTED `git-master`: Multiple components but simple changes

**Parallelization**:
- **Can Run In Parallel**: YES (after 2.4, 2.5)
- **Parallel Group**: Wave 3 - Group 3.3
- **Blocks**: None
- **Blocked By**: 2.4, 2.5

**Example**:
```typescript
// BEFORE:
export function Button({ children, onClick }) {
  return <button onClick={onClick}>{children}</button>;
}

// AFTER:
import { memo } from 'react';

export const Button = memo(function Button({ children, onClick }) {
  return <button onClick={onClick}>{children}</button>;
});
```

**Acceptance Criteria**:
```bash
# Memoized components use memo:
grep -r "memo(" client/src/components/  # Should find multiple uses

# TypeScript compiles:
bun run type-check

# Build succeeds:
bun run build

# Evidence:
grep -r "React.memo\|memo(" client/src/components/ --include="*.tsx" | wc -l  # Should be 10+
```

**Commit**: YES
- Message: `perf(components): add React.memo for pure components`
- Files: `src/components/ui/*.tsx`, `src/components/navigation/*.tsx`

---

#### **TASK 3.4: Fix useEffect Cleanup Functions**

**What to do**:
1. Add cleanup functions for setInterval in system-status/page.tsx
2. Add cleanup for setTimeout in parent-portal/page.tsx
3. Add cleanup for setTimeout in notifications/page.tsx
4. Add cleanup for setTimeout in attendance/page.tsx
5. Add cleanup for setTimeout in fees/page.tsx

**Files to Fix**:
| File | Line Numbers | Issue |
|------|--------------|-------|
| `client/src/app/system-status/page.tsx` | TBD | setInterval without cleanup |
| `client/src/app/parent-portal/page.tsx` | TBD | setTimeout without cleanup |
| `client/src/app/notifications/page.tsx` | TBD | setTimeout without cleanup |
| `client/src/app/attendance/page.tsx` | TBD | setTimeout without cleanup |
| `client/src/app/fees/page.tsx` | TBD | setTimeout without cleanup |

**Must NOT do**:
- Don't change timer logic, only add cleanup
- Don't break timer functionality

**Recommended Agent Profile**:
- **Category**: `ultrabrain` (requires understanding cleanup patterns)
- **Skills**: `frontend-ui-ux`, `typescript-programmer`
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: useEffect cleanup is React-specific
  - ✅ INCLUDED `typescript-programmer`: Cleanup function typing
  - ❌ OMITTED `git-master`: 5 files, straightforward

**Parallelization**:
- **Can Run In Parallel**: YES (after Wave 2)
- **Parallel Group**: Wave 3 - Group 3.4
- **Blocks**: None
- **Blocked By**: 2.4, 2.5

**Example Fix**:
```typescript
// BEFORE:
useEffect(() => {
  const interval = setInterval(() => {
    fetchStatus();
  }, 5000);
  // ❌ Missing cleanup
}, []);

// AFTER:
useEffect(() => {
  const interval = setInterval(() => {
    fetchStatus();
  }, 5000);
  
  return () => clearInterval(interval);  // ✅ Cleanup
}, [fetchStatus]);
```

**Acceptance Criteria**:
```bash
# All useEffect with timers have cleanup:
grep -A5 "setInterval\|setTimeout" client/src/app/system-status/page.tsx | grep "return.*clear"

# TypeScript compiles:
bun run type-check

# Build succeeds:
bun run build

# Evidence:
grep -r "return.*clearInterval\|return.*clearTimeout" client/src/app/ --include="*.tsx" | wc -l  # Should be 5+
```

**Commit**: YES
- Message: `fix(effects): add missing cleanup functions for timers`
- Files: 5 modified files

---

### WAVE 4: BACKEND API DEVELOPMENT (Can Run Parallel with Waves 1-3)

---

#### **TASK 4.1.x: Self-Service Portal APIs**

**What to do**: Create backend APIs for Self-Service Portal

**TASK 4.1.1: Create Self-Service Request Model**
- File: `server/internal/models/selfservice.go`
- Define: SelfServiceRequest struct with fields (id, type, status, description, createdAt, updatedAt)

**TASK 4.1.2: Create POST /api/self-service/requests Handler**
- File: `server/api/handler/selfservice.go`
- Accept: enrollment, schedule_change, document request types
- Validate request and save to database

**TASK 4.1.3: Create GET /api/self-service/requests Handler**
- File: `server/api/handler/selfservice.go`
- Return: List of requests for authenticated user
- Support: Filtering by status, type

**TASK 4.1.4: Create GET /api/self-service/requests/:id Handler**
- File: `server/api/handler/selfservice.go`
- Return: Specific request details
- Include: Request history, status updates

**Recommended Agent Profile**:
- **Category**: `unspecified-high` (Go backend development)
- **Skills**: None (Go backend doesn't match available skills)
- **Skills Evaluation**:
  - ❌ NO SKILLS MATCH: Available skills are frontend-focused
  - Backend Go development requires manual expertise

**Parallelization**:
- **Can Run In Parallel**: YES (with Waves 1-3)
- **Parallel Group**: Wave 4 - Group 4.1
- **Blocks**: 4.4
- **Blocked By**: None

**Acceptance Criteria**:
```bash
# Server builds successfully:
cd server && go build ./...

# APIs respond correctly:
curl -X POST http://localhost:8080/api/self-service/requests \
  -H "Content-Type: application/json" \
  -d '{"type":"enrollment","description":"Test request"}' \
  | jq '.id'  # Should return UUID

curl http://localhost:8080/api/self-service/requests | jq '.length'  # Should return count

curl http://localhost:8080/api/self-service/requests/{id} | jq '.id'  # Should return same ID

# Evidence:
ls server/internal/models/selfservice.go  # File exists
ls server/api/handler/selfservice.go  # File exists
grep "func.*SelfService" server/api/handler/selfservice.go | wc -l  # Should be 3+ handlers
```

**Commit**: YES
- Message: `feat(api): add self-service portal endpoints`
- Files: `server/internal/models/selfservice.go`, `server/api/handler/selfservice.go`

---

#### **TASK 4.2.x: Parent Portal APIs**

**What to do**: Create backend APIs for Parent Portal

**TASK 4.2.1: Create Parent-Student Link Model**
- File: `server/internal/models/parent.go`
- Define: ParentChildLink struct with fields (parentId, studentId, relationship, status)

**TASK 4.2.2: Create GET /api/parent/children Handler**
- File: `server/api/handler/parent.go`
- Return: List of linked students for authenticated parent

**TASK 4.2.3: Create GET /api/parent/children/:id/dashboard Handler**
- File: `server/api/handler/parent.go`
- Return: Dashboard data for specific child

**TASK 4.2.4: Create GET /api/parent/children/:id/attendance Handler**
- File: `server/api/handler/parent.go`
- Return: Attendance records for specific child

**TASK 4.2.5: Create GET /api/parent/children/:id/grades Handler**
- File: `server/api/handler/parent.go`
- Return: Grades for specific child

**Recommended Agent Profile**:
- **Category**: `unspecified-high` (Go backend development)
- **Skills**: None

**Parallelization**:
- **Can Run In Parallel**: YES (with Waves 1-3)
- **Parallel Group**: Wave 4 - Group 4.2
- **Blocks**: 4.4
- **Blocked By**: None

**Acceptance Criteria**:
```bash
# Server builds:
cd server && go build ./...

# APIs respond:
curl http://localhost:8080/api/parent/children | jq '.length'  # Should return count

curl http://localhost:8080/api/parent/children/{id}/dashboard | jq '.studentId'  # Should match

curl http://localhost:8080/api/parent/children/{id}/attendance | jq '.percentage'

curl http://localhost:8080/api/parent/children/{id}/grades | jq '.gpa'

# Evidence:
ls server/internal/models/parent.go
ls server/api/handler/parent.go
grep "func.*Parent" server/api/handler/parent.go | wc -l  # Should be 5+ handlers
```

**Commit**: YES
- Message: `feat(api): add parent portal endpoints`
- Files: `server/internal/models/parent.go`, `server/api/handler/parent.go`

---

#### **TASK 4.3.x: Faculty Tools APIs**

**What to do**: Create backend APIs for Faculty Tools

**TASK 4.3.1: Create Rubric Model**
- File: `server/internal/models/rubric.go`
- Define: GradingRubric struct with fields (id, name, criteria[], maxScore)

**TASK 4.3.2: Create Rubric CRUD Handlers**
- File: `server/api/handler/rubric.go`
- Implement: GET /api/faculty/rubrics, POST /api/faculty/rubrics
- Implement: GET /api/faculty/rubrics/:id, PUT /api/faculty/rubrics/:id, DELETE /api/faculty/rubrics/:id

**TASK 4.3.3: Create Office Hours Model**
- File: `server/internal/models/officehours.go`
- Define: OfficeHoursSlot struct with fields (id, day, startTime, endTime, location)

**TASK 4.3.4: Create Office Hours CRUD Handlers**
- File: `server/api/handler/officehours.go`
- Implement: GET /api/faculty/office-hours, POST /api/faculty/office-hours
- Implement: PUT /api/faculty/office-hours/:id, DELETE /api/faculty/office-hours/:id

**Recommended Agent Profile**:
- **Category**: `unspecified-high` (Go backend development)
- **Skills**: None

**Parallelization**:
- **Can Run In Parallel**: YES (with Waves 1-3)
- **Parallel Group**: Wave 4 - Group 4.3
- **Blocks**: 4.4
- **Blocked By**: None

**Acceptance Criteria**:
```bash
# Server builds:
cd server && go build ./...

# Rubric APIs:
curl http://localhost:8080/api/faculty/rubrics | jq '.length'
curl -X POST http://localhost:8080/api/faculty/rubrics -H "Content-Type: application/json" -d '{"name":"Test","criteria":[],"maxScore":100}' | jq '.id'

# Office Hours APIs:
curl http://localhost:8080/api/faculty/office-hours | jq '.length'
curl -X POST http://localhost:8080/api/faculty/office-hours -H "Content-Type: application/json" -d '{"day":"Monday","startTime":"09:00","endTime":"11:00","location":"Office 101"}' | jq '.id'

# Evidence:
ls server/internal/models/rubric.go
ls server/internal/models/officehours.go
ls server/api/handler/rubric.go
ls server/api/handler/officehours.go
grep "func.*Rubric" server/api/handler/rubric.go | wc -l  # Should be 5+ handlers
grep "func.*OfficeHours" server/api/handler/officehours.go | wc -l  # Should be 4+ handlers
```

**Commit**: YES
- Message: `feat(api): add faculty tools endpoints (rubrics and office hours)`
- Files: `server/internal/models/rubric.go`, `server/internal/models/officehours.go`, `server/api/handler/rubric.go`, `server/api/handler/officehours.go`

---

#### **TASK 4.4: Backend Integration Tests**

**What to do**:
1. Create integration tests for all new APIs
2. Test CRUD operations
3. Test authentication/authorization
4. Test edge cases

**Recommended Agent Profile**:
- **Category**: `unspecified-high` (Go testing)
- **Skills**: None

**Parallelization**:
- **Can Run In Parallel**: NO (after 4.1, 4.2, 4.3)
- **Parallel Group**: Wave 4 - Group 4.4
- **Blocks**: 5.1
- **Blocked By**: 4.1, 4.2, 4.3

**Acceptance Criteria**:
```bash
# Run tests:
cd server && go test ./... -v

# Evidence:
cd server && go test ./api/handler -run "SelfService|Parent|Rubric|OfficeHours" -v  # Should pass
```

**Commit**: YES
- Message: `test(api): add integration tests for new endpoints`
- Files: `server/api/handler/*_test.go`

---

### WAVE 5: INTEGRATION & DOCUMENTATION (Depends on ALL Previous Waves)

---

#### **TASK 5.1: Generate Swagger Documentation**

**What to do**:
1. Install/update swagger generation tools
2. Add swagger annotations to all new handlers
3. Regenerate swagger.json
4. Verify all 30+ endpoints are documented

**Files to Modify**:
- `server/docs/swagger.json` (currently empty `{}`)
- All handler files need swagger annotations

**Must NOT do**:
- Don't leave swagger.json as empty `{}`
- Don't skip any endpoints

**Recommended Agent Profile**:
- **Category**: `unspecified-high` (Go tooling)
- **Skills**: None

**Parallelization**:
- **Can Run In Parallel**: NO (after ALL previous waves)
- **Parallel Group**: Wave 5 - Group 5.1
- **Blocks**: 5.2
- **Blocked By**: ALL

**Steps**:
```bash
# Install swag if not present
cd server && go install github.com/swaggo/swag/cmd/swag@latest

# Add annotations to handlers
# Example annotation:
// @Summary Get self-service requests
// @Tags self-service
// @Produce json
// @Success 200 {array} models.SelfServiceRequest
// @Router /api/self-service/requests [get]

# Generate swagger
cd server && swag init

# Verify
cat server/docs/swagger.json | jq '.paths | keys | length'  # Should be 30+
```

**Acceptance Criteria**:
```bash
# swagger.json is not empty:
cat server/docs/swagger.json | jq '.' | head -20  # Should show actual content, not {}

# All endpoints documented:
cat server/docs/swagger.json | jq '.paths | keys'  # Should list all 30+ endpoints

# Evidence:
cat server/docs/swagger.json | jq '.paths | keys | length'  # Should be >= 30
cat server/docs/swagger.json | jq '.paths."/api/self-service/requests"'  # Should exist
cat server/docs/swagger.json | jq '.paths."/api/parent/children"'  # Should exist
cat server/docs/swagger.json | jq '.paths."/api/faculty/rubrics"'  # Should exist
```

**Commit**: YES
- Message: `docs(api): generate swagger documentation for all endpoints`
- Files: `server/docs/swagger.json`, handler files with annotations

---

#### **TASK 5.2: Final Integration & E2E Testing**

**What to do**:
1. Run full TypeScript check on frontend
2. Run Go build and tests on backend
3. Run full E2E test suite with Playwright
4. Verify all portal pages are functional end-to-end
5. Run Lighthouse performance audit
6. Create summary report

**Must NOT do**:
- Don't skip any verification steps
- Don't ignore warnings
- Don't skip broken tests

**Recommended Agent Profile**:
- **Category**: `visual-engineering` (integration and testing)
- **Skills**: `frontend-ui-ux`, `agent-browser` (for E2E)
- **Skills Evaluation**:
  - ✅ INCLUDED `frontend-ui-ux`: Final integration requires frontend expertise
  - ✅ INCLUDED `agent-browser`: E2E testing needs browser automation
  - ❌ OMITTED `typescript-programmer`: Testing is execution, not coding

**Parallelization**:
- **Can Run In Parallel**: NO (final step)
- **Parallel Group**: Wave 5 - Group 5.2
- **Blocks**: None
- **Blocked By**: ALL previous tasks

**Verification Commands**:
```bash
# Frontend Type Checking
cd client && bun run type-check

# Frontend Linting
cd client && bun run lint

# Frontend Build
cd client && bun run build

# Frontend Unit Tests
cd client && bun run test

# Backend Build
cd server && go build ./...

# Backend Tests
cd server && go test ./...

# E2E Tests
cd client && bun run test:e2e

# Bundle Analysis
cd client && ANALYZE=true bun run build

# Lighthouse (requires server running)
cd client && bun run build && npx lighthouse http://localhost:3000 --output=json
```

**Portal Page Verification** (via Playwright):
```bash
# Self-Service Portal
curl http://localhost:3000/self-service  # Should load without errors

# Parent Portal
curl http://localhost:3000/parent-portal  # Should load without errors

# Faculty Tools
curl http://localhost:3000/faculty-tools  # Should load without errors
```

**Acceptance Criteria**:
```bash
# Type Safety:
grep -r ": any" client/src/ --include="*.tsx" --include="*.ts" | grep -v node_modules | wc -l  # Should be 0

# Console Cleanup:
grep -r "console\.log" client/src/app/ client/src/components/ --include="*.tsx" | wc -l  # Should be 0

# React Query Coverage:
grep -r "useEffect.*fetch" client/src/app/ --include="*.tsx" | wc -l  # Should be 0 (or very few legitimate)

# useCallback Coverage:
grep -r "useCallback" client/src/app/ --include="*.tsx" | wc -l  # Should be 130+

# Cleanup Functions:
grep -r "setInterval\|setTimeout" client/src/app/ --include="*.tsx" | grep -v "clearInterval\|clearTimeout" | wc -l  # Should show cleanup pattern

# Build Success:
cd client && ls .next/  # Should exist

# Swagger:
cat server/docs/swagger.json | jq '.paths | keys | length'  # Should be >= 30

# Portal Functionality:
curl -s http://localhost:3000/api/self-service/requests | jq '.'  # Should work
curl -s http://localhost:3000/api/parent/children | jq '.'  # Should work
curl -s http://localhost:3000/api/faculty/rubrics | jq '.'  # Should work
```

**Commit**: YES
- Message: `chore(release): final integration and optimizations`
- Files: Any final fixes

---

## Commit Strategy

| Task | Commit Message | Files | Verification |
|------|----------------|-------|--------------|
| 1.1.1 | `chore(test): setup jest and react testing library` | jest.config.ts, jest.setup.ts, package.json | `bun run test --listTests` |
| 1.1.2 | `feat(types): add comprehensive TypeScript type definitions` | src/types/index.ts | `bun run type-check` |
| 1.1.3 | `feat(errors): add custom error classes and type guards` | src/lib/errors.ts | `bun run type-check` |
| 1.2 | `perf(config): add build optimizations and bundle analyzer` | next.config.ts, package.json | `bun run build` |
| 1.3 | `refactor(console): replace console statements with logger` | 10 files | `grep -r console.log` |
| 1.4 | `feat(api): add TypeScript generics to api client` | src/lib/api-client.ts | `bun run type-check` |
| 2.1 | `refactor(errors): migrate catch any to catch unknown` | 26 files | `grep -r "catch.*any"` |
| 2.2 | `feat(react-query): setup React Query infrastructure` | src/providers/, layout.tsx | `bun run type-check` |
| 2.3 | `feat(hooks): add React Query custom hooks` | src/hooks/queries/*.ts | `bun run type-check` |
| 2.4.x | `refactor(PAGE): migrate to React Query` | src/app/PAGE/page.tsx | `grep "useQuery\|useMutation"` |
| 2.5.x | `refactor(PAGE): migrate to React Query` | src/app/PAGE/page.tsx | `grep "useQuery\|useMutation"` |
| 3.1 | `perf(components): add useMemo for expensive computations` | 6-8 files | `bun run build` |
| 3.2 | `perf(components): add useCallback for inline functions` | 14+ files | `bun run build` |
| 3.3 | `perf(components): add React.memo for pure components` | components/*.tsx | `bun run build` |
| 3.4 | `fix(effects): add missing cleanup functions` | 5 files | `bun run build` |
| 4.1.x | `feat(api): add self-service portal endpoints` | server/models/, server/handlers/ | `go test ./...` |
| 4.2.x | `feat(api): add parent portal endpoints` | server/models/, server/handlers/ | `go test ./...` |
| 4.3.x | `feat(api): add faculty tools endpoints` | server/models/, server/handlers/ | `go test ./...` |
| 4.4 | `test(api): add integration tests for new endpoints` | server/api/handler/*_test.go | `go test ./...` |
| 5.1 | `docs(api): generate swagger documentation` | server/docs/swagger.json | `cat swagger.json` |
| 5.2 | `chore(release): final integration and testing` | Any fixes | `bun run test:e2e` |

---

## Success Criteria

### Final Metrics (Before → After)

| Metric | Before | After (Target) | Verification Command |
|--------|--------|----------------|---------------------|
| `any` types | 47 | 0 | `grep -r ": any" client/src/ --include="*.tsx" \| wc -l` |
| console.log | 30 | 0 | `grep -r "console\.log" client/src/app/ --include="*.tsx" \| wc -l` |
| Inline functions (unmemoized) | 130+ | 0 | `grep -r "useCallback" client/src/app/ --include="*.tsx" \| wc -l` |
| Unoptimized useEffect | 43 | 0 (React Query) | `grep -r "useEffect.*fetch" client/src/app/ --include="*.tsx" \| wc -l` |
| Missing cleanup | 5 | 0 | `grep -r "return.*clear" client/src/app/ --include="*.tsx" \| wc -l` |
| React Query coverage | 0% | 100% | `ls client/src/hooks/queries/*.ts \| wc -l` |
| Backend APIs (Self-Service) | 0 | 4 | `grep "self-service" server/docs/swagger.json \| wc -l` |
| Backend APIs (Parent Portal) | 0 | 5 | `grep "parent" server/docs/swagger.json \| wc -l` |
| Backend APIs (Faculty Tools) | 0 | 8 | `grep "faculty" server/docs/swagger.json \| wc -l` |
| Swagger endpoints | 0 (empty) | 30+ | `cat server/docs/swagger.json \| jq '.paths \| keys \| length'` |
| Type coverage | ~70% | 100% | `bun run type-check` |
| Build optimization | None | Full | `ANALYZE=true bun run build` |
| Lighthouse score | Unknown | >90 | `lighthouse http://localhost:3000` |

### Verification Commands

```bash
# =====================================
# FRONTEND VERIFICATION
# =====================================

# Type Safety - Zero any types
echo "=== Type Safety Check ==="
grep -r ": any" client/src/ --include="*.tsx" --include="*.ts" | grep -v node_modules | wc -l
cd client && bun run type-check

# Console Cleanup - Zero console.log in production
echo "=== Console Cleanup Check ==="
grep -r "console\.log" client/src/app/ client/src/components/ --include="*.tsx" | grep -v "error-boundary\|logger.ts" | wc -l

# React Query Migration
echo "=== React Query Check ==="
grep -r "useQuery\|useMutation" client/src/hooks/queries/*.ts | wc -l
grep -r "useEffect.*fetch" client/src/app/ --include="*.tsx" | wc -l  # Should be ~0

# Performance Optimizations
echo "=== Performance Check ==="
grep -r "useCallback" client/src/app/ --include="*.tsx" | wc -l  # Should be 130+
grep -r "useMemo" client/src/app/ --include="*.tsx" | wc -l  # Should be 15+
grep -r "React.memo\|memo(" client/src/components/ --include="*.tsx" | wc -l  # Should be 10+

# Memory Leak Fixes
echo "=== Cleanup Check ==="
grep -r "return.*clearInterval\|return.*clearTimeout" client/src/app/ --include="*.tsx" | wc -l  # Should be 5+

# Build
echo "=== Build Check ==="
cd client && bun run build

# =====================================
# BACKEND VERIFICATION
# =====================================

# Server Build
echo "=== Server Build Check ==="
cd server && go build ./...

# Tests
echo "=== Server Tests ==="
cd server && go test ./...

# Swagger
echo "=== Swagger Check ==="
cat server/docs/swagger.json | jq '.paths | keys | length'  # Should be >= 30
cat server/docs/swagger.json | jq '.paths."/api/self-service/requests"'  # Should exist
cat server/docs/swagger.json | jq '.paths."/api/parent/children"'  # Should exist
cat server/docs/swagger.json | jq '.paths."/api/faculty/rubrics"'  # Should exist

# =====================================
# END-TO-END VERIFICATION
# =====================================

# Portal Pages Functional
echo "=== Portal Pages Check ==="
curl -s http://localhost:3000/api/self-service/requests | jq '. | length'  # Should work
curl -s http://localhost:3000/api/parent/children | jq '. | length'  # Should work
curl -s http://localhost:3000/api/faculty/rubrics | jq '. | length'  # Should work

# E2E Tests
echo "=== E2E Tests ==="
cd client && bun run test:e2e
```

### Final Checklist

**Frontend**:
- [ ] Zero `any` types in production code (47 → 0)
- [ ] Zero console.log in production builds (30 → 0)
- [ ] All data fetching uses React Query (43 useEffects migrated)
- [ ] All 130+ inline functions use useCallback
- [ ] All expensive computations use useMemo
- [ ] All 5 useEffect timers have cleanup
- [ ] Build succeeds with optimizations
- [ ] All tests pass
- [ ] Lighthouse performance score >90
- [ ] Bundle size optimized

**Backend**:
- [ ] 4 Self-Service Portal API endpoints created
- [ ] 5 Parent Portal API endpoints created
- [ ] 8 Faculty Tools API endpoints created
- [ ] All backend tests pass
- [ ] Server builds successfully

**Integration**:
- [ ] Self-service portal fully functional end-to-end
- [ ] Parent portal fully functional end-to-end
- [ ] Faculty tools fully functional end-to-end
- [ ] Swagger documentation shows all 30+ endpoints
- [ ] E2E tests pass for all critical flows

---

## Summary Statistics

| Metric | Value |
|--------|-------|
| **Total Tasks** | 50+ tasks across 5 waves |
| **Estimated Duration** | 2-3 weeks with parallel execution |
| **Frontend Files Modified** | 60+ files |
| **Backend Files Created** | 8+ files |
| **Lines Changed** | 3000+ lines |
| **API Endpoints Created** | 17 new endpoints |
| **Parallel Speedup** | ~60% faster than sequential |

### Wave Breakdown

| Wave | Tasks | Duration | Dependencies |
|------|-------|----------|--------------|
| Wave 1 | 6 tasks | 2-3 days | None |
| Wave 2 | 38 tasks | 5-7 days | Wave 1 |
| Wave 3 | 4 tasks | 2-3 days | Wave 2 |
| Wave 4 | 12 tasks | 5-7 days | None (parallel) |
| Wave 5 | 2 tasks | 1-2 days | All Waves |

---

## Next Steps

1. **Run `/start-work`** to begin execution
2. **Start with Wave 1** tasks (all independent, can run in parallel)
3. **Parallel Track**: Wave 4 (Backend) can run concurrently with Waves 1-3
4. **Progress tracking**: Each task has clear acceptance criteria
5. **Final verification**: Run all verification commands after Wave 5

This master plan transforms EduHub into a production-grade, fully optimized educational platform with complete feature parity, professional code quality standards, and fully functional backend APIs for all portal pages.

---

## Plan Metadata

**Created**: 2026-02-01
**Version**: 2.0 (FINAL)
**Status**: Ready for Execution
**Plan File**: `.sisyphus/plans/eduhub-optimization.md`
