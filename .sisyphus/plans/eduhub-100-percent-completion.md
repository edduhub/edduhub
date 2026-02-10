# EduHub 100% Completion Work Plan

## TL;DR

> **Quick Summary**: Complete the EduHub educational management platform by fixing 8 backend services/handlers and integrating 7 frontend pages with real APIs. Eliminate all placeholder code, mock services, and hardcoded data.
> 
> **Deliverables**:
> - 8 backend security fixes and service implementations
> - 12 new API endpoints
> - 7 frontend pages with full API integration
> - Comprehensive JWT test suite
> - Production-ready external service integrations (Push, SMS, Email)
> 
> **Estimated Effort**: Large (20-25 tasks across 7 waves)
> **Parallel Execution**: YES - 7 waves with up to 4 parallel tasks per wave
> **Critical Path**: Wave 1 (Security) → Wave 3 (Backend APIs) → Wave 4/5 (Frontend Integration)

---

## Context

### Original Request
Make EduHub educational management app 100% complete by:
1. Fixing backend services that return mock data or have security flaws
2. Connecting frontend pages to real APIs instead of local state
3. Eliminating all placeholder code, hardcoded values, and commented-out API calls

### Interview Summary

**Key Gaps Identified (Backend)**:
- **CRITICAL SECURITY**: Forum DeleteReply has no ownership verification
- **CRITICAL**: Email service silently fails when SMTP not configured
- Parent handler with TODOs and stub access verification
- Push/SMS services return mocks when not configured
- Analytics service with hardcoded probability 0.8
- Empty JWT test file
- Empty catch blocks in auth context

**Key Gaps Identified (Frontend)**:
- Settings page uses only local React state
- Self-service page has commented-out API calls
- Parent portal contact has commented API call
- Faculty tools with hardcoded rubrics and office hours
- Roles page with non-functional buttons
- System status with hardcoded metrics
- Topbar with hardcoded notification badge "3"

### Self-Review & Gap Analysis

**Assumptions Made**:
1. User has or will obtain external service credentials (FCM, Twilio, SMTP/SendGrid)
2. Database migrations can be added for device tokens, faculty tools tables
3. Test infrastructure should be included (JWT tests minimum)
4. Security fixes are highest priority (Wave 1)
5. Frontend integration requires corresponding backend APIs first

**Auto-Resolved** (minor gaps fixed in planning):
- Pattern for all fixes: Follow existing codebase patterns (student service, announcements, roles)
- Frontend API integration: Use existing React Query hooks pattern or direct API calls
- Test strategy: Add JWT tests and integration tests for critical paths

**Defaults Applied**:
- **External Services**: Plan assumes mock services stay in place until credentials configured
- **Database Changes**: Add migrations following existing pattern (000038, 000039, etc.)
- **Analytics**: Use improved statistical formulas (not full ML) - realistic for MVP
- **System Metrics**: Optional - include API but can be disabled if monitoring not needed

**Decisions Needed** (presented in summary below):
- External service credentials availability
- Test scope (minimum vs comprehensive)
- System metrics priority

---

## Work Objectives

### Core Objective
Eliminate all placeholder code from EduHub by implementing real API integrations, fixing security vulnerabilities, and connecting all frontend pages to backend services.

### Concrete Deliverables
- 12 new REST API endpoints (backend)
- 8 fixed services/handlers (backend)
- 7 integrated frontend pages
- 4+ database migrations
- JWT test suite with 4+ test cases
- Updated environment variable template

### Definition of Done
- [ ] All backend handlers return real data (no empty arrays, no TODOs)
- [ ] All frontend pages fetch from APIs (no hardcoded data, no local-only state)
- [ ] All security vulnerabilities fixed (ownership verification, access control)
- [ ] All external service integrations working (Push, SMS, Email) or gracefully degrading with errors
- [ ] JWT has comprehensive test coverage
- [ ] All commented-out API calls are active and working
- [ ] All non-functional buttons have working handlers

### Must Have
- Forum DeleteReply ownership verification
- Parent handler with real access control
- Email service error handling (no silent failures)
- Settings page API integration
- Self-service page working forms
- JWT tests

### Must NOT Have (Guardrails)
- No new frontend framework changes (keep Next.js + React Query)
- No major architecture changes (keep existing patterns)
- No breaking changes to existing working APIs
- No removal of existing features (only fix/add)

---

## Verification Strategy

### Test Infrastructure Assessment
- **Infrastructure exists**: YES - Go has testing support, frontend has TypeScript
- **User wants tests**: YES - Include for critical components (JWT, security fixes)
- **Approach**: Mix of TDD for new APIs and manual verification for frontend UI

### TDD for New Backend APIs
Each new API endpoint task includes:
1. **RED**: Write test expecting endpoint to exist (should fail initially)
2. **GREEN**: Implement endpoint to pass test
3. **REFACTOR**: Clean up while keeping tests green

### Manual Verification for Frontend
- Frontend changes use Playwright/automated browser testing where possible
- UI components verified through automated interaction tests

---

## Execution Strategy

### Parallel Execution Waves

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ WAVE 1: Security & Critical Infrastructure (Sequential, No Dependencies)   │
│ Duration: ~2-3 days                                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│ 1.1  Forum DeleteReply ownership verification                               │
│ 1.2  Parent handler access control implementation                           │
│ 1.3  Email service error handling (no silent failures)                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ WAVE 2: Communication Services (Parallel)                                  │
│ Duration: ~2-3 days                                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│ 2.1  Push notification service with device token storage                    │
│ 2.2  SMS service with Twilio integration                                    │
│ 2.3  Auth context error handling (empty catch blocks)                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ WAVE 3: Core Backend APIs (Parallel, depends on Wave 1)                    │
│ Duration: ~3-4 days                                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│ 3.1  Settings API (GET/PUT /api/settings)                                   │
│ 3.2  Parent contact API (POST /api/parent/contact)                          │
│ 3.3  Faculty tools APIs (office-hours, rubrics CRUD)                        │
│ 3.4  System metrics API (GET /api/system/metrics) - optional                │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ WAVE 4: Frontend Integration Part 1 (Parallel, depends on Wave 3)          │
│ Duration: ~2-3 days                                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│ 4.1  Settings page API integration                                          │
│ 4.2  Self-service page form submissions                                     │
│ 4.3  Parent portal contact form                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ WAVE 5: Frontend Integration Part 2 (Parallel, depends on Wave 3)          │
│ Duration: ~2-3 days                                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│ 5.1  Faculty tools integration (office hours, rubrics)                      │
│ 5.2  Topbar notifications badge                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ WAVE 6: Roles & System Status (Parallel, partial Wave 3 dependency)        │
│ Duration: ~2 days                                                            │
├─────────────────────────────────────────────────────────────────────────────┤
│ 6.1  Roles page button functionality (permissions, modify role)             │
│ 6.2  System status real metrics (if 3.4 done)                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ WAVE 7: Testing & Analytics (Parallel, depends on all above)               │
│ Duration: ~2-3 days                                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│ 7.1  JWT test suite                                                         │
│ 7.2  Advanced analytics improvements                                        │
│ 7.3  Integration tests for critical paths                                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Dependency Matrix

| Task | Wave | Depends On | Blocks | Can Parallelize With |
|------|------|------------|--------|---------------------|
| 1.1 Forum ownership | 1 | None | None | 1.2, 1.3, 2.x |
| 1.2 Parent handler | 1 | None | 3.2 | 1.1, 1.3, 2.x |
| 1.3 Email errors | 1 | None | None | 1.1, 1.2, 2.x |
| 2.1 Push service | 2 | None | None | 2.2, 2.3, 1.x |
| 2.2 SMS service | 2 | None | None | 2.1, 2.3, 1.x |
| 2.3 Auth errors | 2 | None | None | 2.1, 2.2, 1.x |
| 3.1 Settings API | 3 | None | 4.1 | 3.2, 3.3, 3.4 |
| 3.2 Parent contact | 3 | 1.2 | 4.3 | 3.1, 3.3, 3.4 |
| 3.3 Faculty APIs | 3 | None | 5.1 | 3.1, 3.2, 3.4 |
| 3.4 System metrics | 3 | None | 6.2 | 3.1, 3.2, 3.3 |
| 4.1 Settings page | 4 | 3.1 | None | 4.2, 4.3 |
| 4.2 Self-service | 4 | None | None | 4.1, 4.3 |
| 4.3 Parent contact | 4 | 3.2 | None | 4.1, 4.2 |
| 5.1 Faculty tools | 5 | 3.3 | None | 5.2 |
| 5.2 Topbar | 5 | None | None | 5.1 |
| 6.1 Roles page | 6 | None | None | 6.2 |
| 6.2 System status | 6 | 3.4 | None | 6.1 |
| 7.1 JWT tests | 7 | All | None | 7.2, 7.3 |
| 7.2 Analytics | 7 | All | None | 7.1, 7.3 |
| 7.3 Integration | 7 | All | None | 7.1, 7.2 |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Profile | Skills |
|------|-------|-------------------|--------|
| 1 | 1.1-1.3 | backend-go | security, golang |
| 2 | 2.1-2.3 | backend-go | golang, third-party-apis |
| 3 | 3.1-3.4 | backend-go | golang, sql, api-design |
| 4 | 4.1-4.3 | fullstack-ts-react | react, nextjs, api-integration |
| 5 | 5.1-5.2 | fullstack-ts-react | react, nextjs |
| 6 | 6.1-6.2 | fullstack-ts-react | react, nextjs |
| 7 | 7.1-7.3 | backend-go + testing | golang, testing |

---

## TODOs

### WAVE 1: Security & Critical Infrastructure

---

- [ ] **1.1 Fix Forum DeleteReply ownership verification**

  **What to do**:
  - Add ownership verification to `DeleteReply` method in `/server/internal/services/forum/forum_service.go`
  - Fetch reply first, check if `reply.AuthorID == userID` or role is admin
  - Return error if unauthorized
  - Follow pattern from `DeleteThread` (lines 72-86)

  **Must NOT do**:
  - Don't change other forum methods (already working)
  - Don't modify database schema

  **Recommended Agent Profile**:
  - **Category**: `quick` (small, focused fix)
  - **Skills**: `go`, `security`
  - **Domain**: Go backend service layer

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with 1.2, 1.3)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `/server/internal/services/forum/forum_service.go` line 100-103 - Current DeleteReply (no verification)
  - `/server/internal/services/forum/forum_service.go` line 72-86 - DeleteThread pattern to copy
  - `/server/internal/repository/forum_repository.go` - GetReply method needed

  **Acceptance Criteria**:
  - [ ] Reply author can delete their own reply
  - [ ] Admin/super_admin can delete any reply
  - [ ] Non-author, non-admin gets "unauthorized" error
  - [ ] Non-existent reply returns "not found" error
  - [ ] Test: `go test ./internal/services/forum/... -run TestDeleteReply` passes

  **Commit**: YES
  - Message: `fix(forum): add ownership verification to DeleteReply`
  - Files: `server/internal/services/forum/forum_service.go`

---

- [ ] **1.2 Implement parent handler access control**

  **What to do**:
  - Fix `verifyParentAccess()` in `/server/api/handler/parent_handler.go` (line 249-262)
  - Query `parent_student_relationships` table to verify parent-student link
  - Fix `GetLinkedChildren` to return actual linked students (line 62-65)
  - Fix `GetChildAssignments` to return real assignments (line 239)
  - Create parent service if needed in `/server/internal/services/parent/`

  **Must NOT do**:
  - Don't modify existing working endpoints (GetChildDashboard, GetChildAttendance, GetChildGrades)
  - Don't create new tables (parent_student_relationships already exists - migration 0000XX)

  **Recommended Agent Profile**:
  - **Category**: `quick` 
  - **Skills**: `go`, `sql`
  - **Domain**: Go handler and service layer

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: 3.2 (Parent contact API), 4.3 (Parent portal contact page)
  - **Blocked By**: None

  **References**:
  - `/server/api/handler/parent_handler.go` line 249-262 - verifyParentAccess stub
  - `/server/api/handler/parent_handler.go` line 62-65 - GetLinkedChildren empty return
  - `/server/internal/services/student/student_service.go` - Service pattern to follow
  - `/server/db/migrations/0000XX_create_parent_student_relationships_table.up.sql` - Existing table schema

  **Acceptance Criteria**:
  - [ ] `verifyParentAccess()` checks parent_student_relationships table
  - [ ] Parent can only access their linked children's data
  - [ ] `GetLinkedChildren` returns actual linked students
  - [ ] `GetChildAssignments` returns real assignments from DB
  - [ ] Unauthorized parent gets 403 error
  - [ ] Test: `go test ./api/handler/... -run TestParentHandler` passes

  **Commit**: YES
  - Message: `fix(parent): implement access verification and real data fetching`
  - Files: `server/api/handler/parent_handler.go`, `server/internal/services/parent/` (new)

---

- [ ] **1.3 Fix email service silent failures**

  **What to do**:
  - Modify `/server/internal/services/email/email_service.go` line 87-90
  - Return error when SMTP not configured instead of returning nil
  - Add error logging
  - Update all methods (SendWelcomeEmail, SendPasswordReset, etc.) to propagate errors

  **Must NOT do**:
  - Don't change the email template system (already working when SMTP configured)
  - Don't switch to SendGrid/AWS SES (out of scope - just fix error handling)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `go`, `error-handling`
  - **Domain**: Go service layer

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `/server/internal/services/email/email_service.go` line 87-90 - Silent nil return
  - `/server/internal/services/push/push_service.go` - Good error handling pattern

  **Acceptance Criteria**:
  - [ ] Returns error when SMTP not configured
  - [ ] Error message indicates "SMTP not configured"
  - [ ] Logs error at WARN level
  - [ ] All email methods propagate errors properly
  - [ ] Test: `go test ./internal/services/email/...` passes with error scenarios

  **Commit**: YES
  - Message: `fix(email): return errors instead of silent failures when SMTP not configured`
  - Files: `server/internal/services/email/email_service.go`

---

### WAVE 2: Communication Services

---

- [ ] **2.1 Implement push notification service with device token storage**

  **What to do**:
  - Create device token storage (new migration: device_tokens table)
  - Implement device token CRUD in push service
  - Add methods: `RegisterDevice`, `UnregisterDevice`, `GetUserDevices`
  - Connect to FCM for actual push delivery
  - Handle FCM token refresh

  **Must NOT do**:
  - Don't remove existing mock implementation (keep as fallback)
  - Don't implement iOS APNS separately (FCM handles both)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high` (complex integration)
  - **Skills**: `go`, `firebase`, `sql`
  - **Domain**: Go service layer with external API

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `/server/internal/services/push/push_service.go` line 87-375 - FCM implementation exists
  - `/server/internal/services/push/push_service.go` line 377-422 - Mock implementation
  - `/server/internal/services/notification/notification_service.go` - WebSocket integration pattern
  - `/server/db/migrations/` - Migration pattern

  **Acceptance Criteria**:
  - [ ] Migration 000038_create_device_tokens_table.up.sql created
  - [ ] `RegisterDevice` stores token in DB
  - [ ] `GetUserDevices` retrieves tokens for user
  - [ ] `UnregisterDevice` removes token
  - [ ] Push notification sends to all user devices
  - [ ] Graceful degradation when FCM_SERVER_KEY not set
  - [ ] Test: Device token CRUD operations work

  **Commit**: YES
  - Message: `feat(push): implement device token storage and FCM integration`
  - Files: `server/internal/services/push/push_service.go`, `server/db/migrations/000038_*.sql`

---

- [ ] **2.2 Implement SMS service with Twilio integration**

  **What to do**:
  - Integrate Twilio HTTP API for SMS sending
  - Add phone number validation (E.164 format)
  - Implement SMS delivery tracking (status callbacks)
  - Add rate limiting (max 5 SMS per user per hour)
  - Add SMS templates: fee_reminder, attendance_alert, grade_notification

  **Must NOT do**:
  - Don't remove mock implementation (keep as fallback)
  - Don't add fallback providers (AWS SNS, etc.) - out of scope

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `go`, `twilio`, `validation`
  - **Domain**: Go service layer with external API

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `/server/internal/services/sms/sms_service.go` line 48-244 - Twilio implementation exists
  - `/server/internal/services/sms/sms_service.go` line 247-284 - Mock implementation
  - `/server/internal/middleware/rate_limiter.go` - Rate limiting pattern

  **Acceptance Criteria**:
  - [ ] SMS sends via Twilio API when configured
  - [ ] Phone number validation (E.164 format)
  - [ ] Rate limiting: max 5 per user per hour
  - [ ] Delivery status tracking
  - [ ] Templates work: fee_reminder, attendance_alert, grade_notification
  - [ ] Graceful degradation when Twilio not configured
  - [ ] Test: SMS sending with mock Twilio client

  **Commit**: YES
  - Message: `feat(sms): implement Twilio integration with validation and rate limiting`
  - Files: `server/internal/services/sms/sms_service.go`, `server/internal/middleware/sms_rate_limiter.go` (new)

---

- [ ] **2.3 Fix auth context error handling**

  **What to do**:
  - Add error logging to 3 empty catch blocks in `/client/src/lib/auth-context.tsx`
  - Line 77: bootstrap catch block
  - Line 177: login catch block  
  - Line 200: register catch block
  - Use logger utility (see line 95-96 for pattern)

  **Must NOT do**:
  - Don't change authentication logic (just add logging)
  - Don't add new error states to UI (logging only)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `typescript`, `react`, `error-handling`
  - **Domain**: Frontend React context

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `/client/src/lib/auth-context.tsx` line 77 - Empty catch block
  - `/client/src/lib/auth-context.tsx` line 177 - Empty catch block
  - `/client/src/lib/auth-context.tsx` line 200 - Empty catch block
  - `/client/src/lib/auth-context.tsx` line 95-96 - Good error handling pattern

  **Acceptance Criteria**:
  - [ ] All 3 catch blocks have error logging
  - [ ] Uses logger utility (not console.log)
  - [ ] Error messages are descriptive
  - [ ] No functionality changes (only logging added)
  - [ ] Test: Simulate errors and verify logs

  **Commit**: YES
  - Message: `fix(auth): add error logging to empty catch blocks`
  - Files: `client/src/lib/auth-context.tsx`

---

### WAVE 3: Core Backend APIs

---

- [ ] **3.1 Create Settings API (GET/PUT /api/settings)**

  **What to do**:
  - Add migration for user_settings table (or add columns to users table)
  - Create settings repository in `/server/internal/repository/settings_repository.go`
  - Create settings service in `/server/internal/services/settings/settings_service.go`
  - Add settings handler in `/server/api/handler/settings_handler.go`
  - Add routes: GET /api/settings, PUT /api/settings
  - Settings fields: email_notifications, push_notifications, theme, language

  **Must NOT do**:
  - Don't add complex settings (keep to 4-5 core preferences)
  - Don't break existing user table schema

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `go`, `sql`, `api-design`
  - **Domain**: Full backend stack (handler → service → repository → db)

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: 4.1 (Settings page)
  - **Blocked By**: None

  **References**:
  - `/server/internal/services/student/student_service.go` - Service pattern
  - `/server/internal/repository/student_repository.go` - Repository pattern
  - `/server/api/handler/student_handler.go` - Handler pattern
  - `/server/api/handler/router.go` - Route registration pattern

  **Acceptance Criteria**:
  - [ ] Migration 000039_add_user_settings_columns.sql (or new table)
  - [ ] GET /api/settings returns current user settings
  - [ ] PUT /api/settings updates settings
  - [ ] Settings persisted in database
  - [ ] Response format: `{ "data": { "email_notifications": true, ... } }`
  - [ ] Test: `curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/settings`

  **Commit**: YES
  - Message: `feat(settings): add settings API with GET/PUT endpoints`
  - Files: `server/api/handler/settings_handler.go`, `server/internal/services/settings/`, `server/internal/repository/settings_repository.go`, `server/db/migrations/000039_*.sql`

---

- [ ] **3.2 Create Parent Contact API (POST /api/parent/contact)**

  **What to do**:
  - Create endpoint for parents to send messages to faculty/admin
  - Add parent_contact_requests table (migration)
  - Create service to handle contact requests
  - Integrate with notification service to alert recipients
  - Add handler with route POST /api/parent/contact

  **Must NOT do**:
  - Don't modify parent handler (separate feature)
  - Don't add real-time chat (just contact form submission)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `go`, `sql`, `api-design`
  - **Domain**: Backend API

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: 4.3 (Parent portal contact page)
  - **Blocked By**: 1.2 (Parent handler fix)

  **References**:
  - `/server/internal/services/notification/notification_service.go` - Notification integration
  - `/server/internal/services/announcement/announcement_service.go` - Similar message pattern

  **Acceptance Criteria**:
  - [ ] Migration 000040_create_parent_contact_requests_table.sql
  - [ ] POST /api/parent/contact accepts { subject, message, recipient_id }
  - [ ] Contact request stored in database
  - [ ] Notification sent to recipient
  - [ ] Returns 201 Created with request ID
  - [ ] Test: Submit contact form, verify DB entry and notification

  **Commit**: YES
  - Message: `feat(parent): add contact API for parent-faculty messaging`
  - Files: `server/api/handler/parent_contact_handler.go`, `server/internal/services/parent/contact_service.go`, `server/db/migrations/000040_*.sql`

---

- [ ] **3.3 Create Faculty Tools APIs (office-hours, rubrics)**

  **What to do**:
  - Create migrations: office_hours table, rubrics table
  - Create faculty repository and service
  - Create handler with CRUD endpoints:
    - GET/POST/PUT/DELETE /api/faculty/office-hours
    - GET/POST/PUT/DELETE /api/faculty/rubrics
  - Office hours fields: faculty_id, day, start_time, end_time, location, is_available
  - Rubrics fields: faculty_id, name, criteria (JSON), max_score, course_id

  **Must NOT do**:
  - Don't add complex rubric UI logic (backend only)
  - Don't link to existing tables yet (faculty_id can reference users.id)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `go`, `sql`, `api-design`, `json`
  - **Domain**: Backend API with JSON data

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: 5.1 (Faculty tools page)
  - **Blocked By**: None

  **References**:
  - `/server/internal/models/assignment.go` - JSON field pattern
  - `/server/api/handler/course_handler.go` - CRUD pattern

  **Acceptance Criteria**:
  - [ ] Migrations: 000041_create_office_hours_table.sql, 000042_create_rubrics_table.sql
  - [ ] All CRUD endpoints work for office-hours
  - [ ] All CRUD endpoints work for rubrics
  - [ ] Rubrics criteria stored as JSON
  - [ ] Proper authorization (faculty can only modify own data, admin can modify all)
  - [ ] Test: Full CRUD cycle for both resources

  **Commit**: YES
  - Message: `feat(faculty): add office hours and rubrics APIs`
  - Files: `server/api/handler/faculty_handler.go` (extended), `server/internal/services/faculty/`, `server/db/migrations/000041_*.sql`, `server/db/migrations/000042_*.sql`

---

- [ ] **3.4 Create System Metrics API (GET /api/system/metrics)**

  **What to do**:
  - Create endpoint to expose real-time system metrics
  - Metrics: CPU usage %, memory usage %, response time avg, uptime %
  - Use system monitoring (e.g., gopsutil library or system calls)
  - Cache metrics for 30 seconds to avoid overhead
  - Add route GET /api/system/metrics (admin only)

  **Must NOT do**:
  - Don't store metrics history (just current snapshot)
  - Don't add complex monitoring (keep simple)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low` (optional but included)
  - **Skills**: `go`, `system-calls`
  - **Domain**: Backend API with system monitoring

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: 6.2 (System status page)
  - **Blocked By**: None

  **References**:
  - `/server/api/handler/system_handler.go` - System endpoints pattern
  - `/server/internal/services/analytics/` - Service pattern

  **Acceptance Criteria**:
  - [ ] GET /api/system/metrics returns current metrics
  - [ ] CPU usage percentage (accurate)
  - [ ] Memory usage percentage (accurate)
  - [ ] Average response time (last 100 requests or similar)
  - [ ] Uptime percentage (accurate)
  - [ ] Metrics cached for 30 seconds
  - [ ] Admin-only access
  - [ ] Test: `curl -H "Authorization: Bearer $ADMIN_TOKEN" http://localhost:8080/api/system/metrics`

  **Commit**: YES
  - Message: `feat(system): add metrics API for real-time monitoring`
  - Files: `server/api/handler/system_handler.go` (extended), `server/internal/services/system/metrics_service.go` (new)

---

### WAVE 4: Frontend Integration Part 1

---

- [ ] **4.1 Integrate Settings page with API**

  **What to do**:
  - Add endpoints to `/client/src/lib/api-client.ts`:
    - `endpoints.settings.get: '/api/settings'`
    - `endpoints.settings.update: '/api/settings'`
  - Add hooks to `/client/src/lib/api-hooks.ts` (optional but recommended):
    - `useSettings()` - React Query hook
    - `useUpdateSettings()` - Mutation hook
  - Update `/client/src/app/settings/page.tsx`:
    - Fetch settings on load
    - Replace `alert()` with API call on save
    - Add loading states
    - Add error handling with toast notifications

  **Must NOT do**:
  - Don't change the UI design (keep existing form)
  - Don't add new settings fields (use what backend provides)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `typescript`, `react`, `nextjs`, `api-integration`
  - **Domain**: Frontend Next.js with React Query

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: None
  - **Blocked By**: 3.1 (Settings API)

  **References**:
  - `/client/src/app/announcements/page.tsx` - API integration pattern
  - `/client/src/lib/api-hooks.ts` line 88-152 - React Query hook pattern
  - `/client/src/lib/api-client.ts` line 201-438 - Endpoints pattern

  **Acceptance Criteria**:
  - [ ] Settings load from API on page mount
  - [ ] "Save" button calls PUT /api/settings
  - [ ] Loading spinner during API calls
  - [ ] Success toast on save
  - [ ] Error toast on failure
  - [ ] Test: Change setting, save, refresh page - change persists

  **Commit**: YES
  - Message: `feat(settings): connect settings page to API`
  - Files: `client/src/app/settings/page.tsx`, `client/src/lib/api-client.ts`, `client/src/lib/api-hooks.ts`

---

- [ ] **4.2 Integrate Self-Service page form submissions**

  **What to do**:
  - Update `/client/src/app/self-service/page.tsx`:
    - Import and use `useCreateSelfServiceRequest()` hook (already imported!)
    - Remove commented-out API calls
    - Replace `alert()` calls with actual mutation
    - Handle loading and error states
  - Forms: enrollment request, schedule change, document request

  **Must NOT do**:
  - Don't modify the form UI (keep existing design)
  - Don't add new request types (use existing)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `typescript`, `react`, `nextjs`, `api-integration`
  - **Domain**: Frontend Next.js

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: None
  - **Blocked By**: None (hooks already exist!)

  **References**:
  - `/client/src/app/self-service/page.tsx` - Current commented code
  - `/client/src/lib/api-hooks.ts` - useCreateSelfServiceRequest hook
  - `/client/src/app/announcements/page.tsx` - Form submission pattern

  **Acceptance Criteria**:
  - [ ] All 3 forms use useCreateSelfServiceRequest hook
  - [ ] No more commented-out code
  - [ ] No more alert() calls
  - [ ] Loading state during submission
  - [ ] Success toast on submission
  - [ ] Error handling with toast
  - [ ] Test: Submit each form type, verify DB entry created

  **Commit**: YES
  - Message: `feat(self-service): connect forms to API using existing hooks`
  - Files: `client/src/app/self-service/page.tsx`

---

- [ ] **4.3 Integrate Parent Portal Contact form**

  **What to do**:
  - Add endpoint to `/client/src/lib/api-client.ts`:
    - `endpoints.parent.contact: '/api/parent/contact'`
  - Update `/client/src/app/parent-portal/contact/page.tsx`:
    - Remove commented-out API call
    - Implement actual API submission
    - Add loading and error states

  **Must NOT do**:
  - Don't change form fields (keep existing)
  - Don't add file attachments (out of scope)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `typescript`, `react`, `nextjs`, `api-integration`
  - **Domain**: Frontend Next.js

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: None
  - **Blocked By**: 3.2 (Parent contact API), 1.2 (Parent handler fix)

  **References**:
  - `/client/src/app/parent-portal/contact/page.tsx` - Current commented code
  - `/client/src/app/announcements/page.tsx` - API submission pattern

  **Acceptance Criteria**:
  - [ ] Contact form uses real API
  - [ ] No more commented-out code
  - [ ] Loading state during submission
  - [ ] Success/error states handled
  - [ ] Test: Submit contact form, verify API call succeeds

  **Commit**: YES
  - Message: `feat(parent-portal): connect contact form to API`
  - Files: `client/src/app/parent-portal/contact/page.tsx`, `client/src/lib/api-client.ts`

---

### WAVE 5: Frontend Integration Part 2

---

- [ ] **5.1 Integrate Faculty Tools with APIs**

  **What to do**:
  - Update `/client/src/app/faculty-tools/page.tsx`:
    - Replace hardcoded officeHours array with API call to `/api/faculty/office-hours`
    - Replace hardcoded rubrics array with API call to `/api/faculty/rubrics`
    - Connect announcement form to existing announcements API
    - Add CRUD operations for office hours and rubrics (add, edit, delete)
  - Add endpoints to `/client/src/lib/api-client.ts` for faculty tools

  **Must NOT do**:
  - Don't change the UI design (keep existing tabs/forms)
  - Don't add complex rubric editing UI (simple JSON editor or form is fine)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `typescript`, `react`, `nextjs`, `api-integration`, `json`
  - **Domain**: Frontend with complex data management

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5
  - **Blocks**: None
  - **Blocked By**: 3.3 (Faculty tools APIs)

  **References**:
  - `/client/src/app/announcements/page.tsx` - CRUD pattern
  - `/client/src/app/students/page.tsx` - Complex data management
  - `/client/src/lib/api-client.ts` - Endpoints pattern

  **Acceptance Criteria**:
  - [ ] Office hours fetched from API on load
  - [ ] Rubrics fetched from API on load
  - [ ] Can add new office hours
  - [ ] Can edit office hours
  - [ ] Can delete office hours
  - [ ] Can add new rubrics
  - [ ] Can edit rubrics
  - [ ] Can delete rubrics
  - [ ] Announcements submitted to correct API
  - [ ] Test: Full CRUD cycle for office hours and rubrics

  **Commit**: YES
  - Message: `feat(faculty-tools): connect to real APIs for office hours and rubrics`
  - Files: `client/src/app/faculty-tools/page.tsx`, `client/src/lib/api-client.ts`

---

- [ ] **5.2 Integrate Topbar notification badge**

  **What to do**:
  - Update `/client/src/components/navigation/topbar.tsx`:
    - Replace hardcoded "3" with `useUnreadCount()` hook
    - Fetch actual notifications list for dropdown
    - Add real-time updates via WebSocket (optional but recommended)
  - Badge should show unread notification count

  **Must NOT do**:
  - Don't redesign the topbar (keep existing)
  - Don't add notification management UI (just badge and list)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `typescript`, `react`, `api-integration`
  - **Domain**: Frontend React component

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5
  - **Blocks**: None
  - **Blocked By**: None (hooks already exist!)

  **References**:
  - `/client/src/components/navigation/topbar.tsx` - Current hardcoded badge
  - `/client/src/lib/api-hooks.ts` - useUnreadCount hook (already exists!)
  - `/client/src/app/notifications/page.tsx` - Notifications API usage

  **Acceptance Criteria**:
  - [ ] Badge shows actual unread count from API
  - [ ] Dropdown shows real notifications
  - [ ] Badge updates when new notification arrives
  - [ ] Clicking notification marks as read
  - [ ] Test: Create notification, verify badge increments

  **Commit**: YES
  - Message: `feat(topbar): connect notification badge to real API`
  - Files: `client/src/components/navigation/topbar.tsx`

---

### WAVE 6: Roles & System Status

---

- [ ] **6.1 Implement Roles page button functionality**

  **What to do**:
  - Update `/client/src/app/roles/page.tsx`:
    - Add "Manage Permissions" modal/dialog
      - Show all permissions with checkboxes
      - Save permission assignments
    - Add "Modify Role" modal/dialog
      - Show users with this role
      - Allow adding/removing users from role
    - Connect buttons to open modals

  **Must NOT do**:
  - Don't change the role CRUD (already working)
  - Don't add new role management features (just these two buttons)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `typescript`, `react`, `nextjs`
  - **Domain**: Frontend UI components

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 6
  - **Blocks**: None
  - **Blocked By**: None (roles API already exists!)

  **References**:
  - `/client/src/app/roles/page.tsx` - Current button code (line 218-220, 263-266)
  - `/client/src/components/ui/dialog.tsx` - Dialog component
  - `/client/src/app/announcements/page.tsx` - Modal pattern

  **Acceptance Criteria**:
  - [ ] "Manage Permissions" button opens permission dialog
  - [ ] Can toggle permissions for role
  - [ ] Changes persist to backend
  - [ ] "Modify Role" button opens user assignment dialog
  - [ ] Can add users to role
  - [ ] Can remove users from role
  - [ ] Test: Assign permissions, assign users, verify changes

  **Commit**: YES
  - Message: `feat(roles): implement manage permissions and modify role buttons`
  - Files: `client/src/app/roles/page.tsx`

---

- [ ] **6.2 Integrate System Status with real metrics**

  **What to do**:
  - Update `/client/src/app/system-status/page.tsx`:
    - Replace hardcoded metrics with API call to `/api/system/metrics`
    - Add loading state while fetching
    - Add auto-refresh every 30 seconds
    - Keep existing health status fetching (already working)

  **Must NOT do**:
  - Don't change the UI design (keep existing cards)
  - Don't add charts/graphs (simple numbers are fine)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `typescript`, `react`, `nextjs`, `api-integration`
  - **Domain**: Frontend Next.js

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 6
  - **Blocks**: None
  - **Blocked By**: 3.4 (System metrics API - optional)

  **References**:
  - `/client/src/app/system-status/page.tsx` - Current hardcoded values (line 176-211)
  - `/client/src/app/system-status/page.tsx` - Existing health status fetch pattern

  **Acceptance Criteria**:
  - [ ] Metrics fetched from API (if 3.4 done)
  - [ ] Loading state during fetch
  - [ ] Auto-refresh every 30 seconds
  - [ ] Real values shown (not 99.9%, 32%, etc.)
  - [ ] OR: If 3.4 not done, show "Metrics unavailable" message instead of fake numbers
  - [ ] Test: View system status page, verify real-time metrics

  **Commit**: YES
  - Message: `feat(system-status): display real metrics from API`
  - Files: `client/src/app/system-status/page.tsx`

---

### WAVE 7: Testing & Analytics

---

- [ ] **7.1 Write comprehensive JWT test suite**

  **What to do**:
  - Write tests in `/server/pkg/jwt/jwt_test.go`:
    - `TestJWTManager_GenerateAndVerify` - Test successful generation and verification
    - `TestJWTManager_Verify_ExpiredToken` - Test expired token rejection
    - `TestJWTManager_Verify_InvalidSignature` - Test invalid signature detection
    - `TestJWTManager_Verify_MalformedToken` - Test malformed token handling
    - `TestJWTManager_ClaimsExtraction` - Test extracting claims from token
  - Use table-driven tests
  - Mock time for expiration tests

  **Must NOT do**:
  - Don't test external JWT library (test our wrapper only)
  - Don't add integration tests (unit tests only)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `go`, `testing`, `jwt`
  - **Domain**: Go unit tests

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 7
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `/server/pkg/jwt/jwt.go` - JWT implementation to test
  - `/server/tests/security_test.go` line 49 - Existing JWT security test
  - `/server/internal/config/config_test.go` - Test pattern

  **Acceptance Criteria**:
  - [ ] All 5+ test cases pass
  - [ ] Table-driven test structure
  - [ ] Coverage for: success, expired, invalid signature, malformed
  - [ ] Run: `go test ./pkg/jwt/... -v` passes
  - [ ] Run: `go test ./pkg/jwt/... -cover` shows >80% coverage

  **Commit**: YES
  - Message: `test(jwt): add comprehensive test suite`
  - Files: `server/pkg/jwt/jwt_test.go`

---

- [ ] **7.2 Improve Advanced Analytics**

  **What to do**:
  - Update `/server/internal/services/analytics/advanced_analytics_service.go`:
    - Replace hardcoded probability 0.8 (line 661) with actual calculation
    - Improve prediction formulas (line 701) using statistical methods
    - Complete `getSkillDevelopment` implementation
    - Add Redis caching for expensive analytics queries
    - Add data export functionality (CSV, PDF)

  **Must NOT do**:
  - Don't add full ML pipeline (complex)
  - Don't change existing working queries

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `go`, `statistics`, `redis`, `analytics`
  - **Domain**: Backend analytics with caching

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 7
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `/server/internal/services/analytics/advanced_analytics_service.go` - Current implementation
  - `/server/internal/cache/` - Redis caching pattern

  **Acceptance Criteria**:
  - [ ] Hardcoded 0.8 replaced with calculated probability
  - [ ] Prediction formulas use statistical methods (not just * 1.1)
  - [ ] getSkillDevelopment returns real data
  - [ ] Redis caching added for expensive queries
  - [ ] Cache TTL: 5 minutes for real-time metrics, 1 hour for historical
  - [ ] Test: Analytics endpoints return accurate predictions

  **Commit**: YES
  - Message: `feat(analytics): improve predictions and add caching`
  - Files: `server/internal/services/analytics/advanced_analytics_service.go`, `server/internal/cache/analytics_cache.go` (new)

---

- [ ] **7.3 Write integration tests for critical paths**

  **What to do**:
  - Write integration tests covering:
    - Settings API roundtrip (GET/PUT)
    - Parent contact submission flow
    - Faculty tools CRUD operations
    - Forum ownership verification
  - Use test database (not production)
  - Use test HTTP server

  **Must NOT do**:
  - Don't test all endpoints (focus on critical paths)
  - Don't use production database

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: `go`, `testing`, `integration-testing`
  - **Domain**: Go integration tests

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 7
  - **Blocks**: None
  - **Blocked By**: All previous waves (tests dependent on implementation)

  **References**:
  - `/server/tests/` - Integration test pattern
  - `/server/internal/services/fee/fee_service_test.go` - Service test pattern

  **Acceptance Criteria**:
  - [ ] 4+ critical paths tested
  - [ ] Tests use isolated test database
  - [ ] Tests clean up after themselves
  - [ ] Run: `go test ./tests/... -v` passes
  - [ ] Tests cover: settings, parent contact, faculty tools, forum

  **Commit**: YES
  - Message: `test(integration): add tests for critical paths`
  - Files: `server/tests/settings_integration_test.go`, `server/tests/parent_integration_test.go`, etc.

---

## Database Migrations Summary

| Migration | Table | Purpose | Wave |
|-----------|-------|---------|------|
| 000038 | device_tokens | Store push notification device tokens | 2.1 |
| 000039 | user_settings | Store user preferences | 3.1 |
| 000040 | parent_contact_requests | Store parent-faculty messages | 3.2 |
| 000041 | office_hours | Store faculty office hours | 3.3 |
| 000042 | rubrics | Store grading rubrics | 3.3 |

---

## Environment Variables to Add

Add to `.env.example`:

```bash
# Push Notifications (Firebase Cloud Messaging)
FCM_SERVER_KEY=your_fcm_server_key_here
FCM_PROJECT_ID=your_fcm_project_id_here
PUSH_ENABLED=true

# SMS (Twilio)
TWILIO_ACCOUNT_SID=your_twilio_account_sid_here
TWILIO_AUTH_TOKEN=your_twilio_auth_token_here
TWILIO_PHONE_NUMBER=+1234567890
SMS_ENABLED=true

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password_here
EMAIL_FROM=noreply@eduhub.com
SMTP_STARTTLS=true

# Analytics Cache
REDIS_URL=redis://localhost:6379
```

---

## Commit Strategy

| After Task | Message | Files |
|------------|---------|-------|
| 1.1 | `fix(forum): add ownership verification to DeleteReply` | `forum_service.go` |
| 1.2 | `fix(parent): implement access verification` | `parent_handler.go`, `parent/` |
| 1.3 | `fix(email): return errors instead of silent failures` | `email_service.go` |
| 2.1 | `feat(push): implement device token storage` | `push_service.go`, migration |
| 2.2 | `feat(sms): implement Twilio integration` | `sms_service.go` |
| 2.3 | `fix(auth): add error logging to catch blocks` | `auth-context.tsx` |
| 3.1 | `feat(settings): add settings API` | `settings_handler.go`, migration |
| 3.2 | `feat(parent): add contact API` | `parent_contact_handler.go`, migration |
| 3.3 | `feat(faculty): add office hours and rubrics APIs` | `faculty_handler.go`, 2 migrations |
| 3.4 | `feat(system): add metrics API` | `system_handler.go`, `metrics_service.go` |
| 4.1 | `feat(settings): connect settings page to API` | `settings/page.tsx`, `api-client.ts` |
| 4.2 | `feat(self-service): connect forms to API` | `self-service/page.tsx` |
| 4.3 | `feat(parent-portal): connect contact form to API` | `contact/page.tsx`, `api-client.ts` |
| 5.1 | `feat(faculty-tools): connect to APIs` | `faculty-tools/page.tsx`, `api-client.ts` |
| 5.2 | `feat(topbar): connect notification badge to API` | `topbar.tsx` |
| 6.1 | `feat(roles): implement permission and user management` | `roles/page.tsx` |
| 6.2 | `feat(system-status): display real metrics` | `system-status/page.tsx` |
| 7.1 | `test(jwt): add comprehensive test suite` | `jwt_test.go` |
| 7.2 | `feat(analytics): improve predictions and caching` | `advanced_analytics_service.go` |
| 7.3 | `test(integration): add tests for critical paths` | `tests/*_integration_test.go` |

---

## Success Criteria

### Verification Commands
```bash
# Backend tests
cd server && go test ./... -v

# Frontend build
cd client && npm run build

# API health check
curl http://localhost:8080/health

# Test settings API
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/settings

# Test forum security (should fail without auth)
curl -X DELETE http://localhost:8080/api/forum/replies/1
```

### Final Checklist
- [ ] All 19 TODOs completed
- [ ] All backend security vulnerabilities fixed
- [ ] All frontend pages connected to APIs
- [ ] All commented-out code activated
- [ ] No hardcoded data remaining (except sample data for demo)
- [ ] JWT tests passing
- [ ] Integration tests passing
- [ ] Environment variables documented
- [ ] Database migrations applied
- [ ] External service integrations working (or gracefully degrading)
