# EduHub Frontend Implementation Plan

## Executive Summary

This plan details the parallel execution strategy for completing the EduHub educational management platform frontend. Based on analysis of the existing codebase, the project is 90% complete with established patterns that need to be extended for the remaining 8 feature areas.

**Current State:**
- Next.js 15 + App Router + React Query + TypeScript + Shadcn/ui
- API endpoints exist for all missing features (backend is 95% complete)
- File upload patterns, React Query hooks, and component structure are established
- Authentication/authorization patterns are in place

**Missing Features:**
1. Course Materials Deep Integration (module management)
2. Exam Management Deep Features (creation, enrollment, hall tickets, results)
3. Role Permission Matrix (interactive permission assignment)
4. Reports (gradecard and transcript generation)
5. Faculty Dashboard (dedicated faculty view)
6. Charts/Visualizations (for analytics)
7. WebSocket Enhancements (real-time updates beyond notifications)

**Dependencies:**
- Add Recharts for data visualization
- Add socket.io-client for WebSocket (already using native WebSocket)

---

## Technical Research Summary

### Best Practices Identified via Context7

#### Next.js File Upload Patterns
- Use `FormData` with `fetch` API for multipart uploads
- Do NOT set `Content-Type` header manually (browser sets it with boundary)
- Track upload progress via `XMLHttpRequest` if needed
- Handle files via `FileList` from `<input type="file">`

**Reference Pattern (from `/client/src/app/files/page.tsx:126-159`):**
```typescript
const formData = new FormData();
formData.append('file', file);
formData.append('category', category);

const response = await fetch(`${API_BASE}/api/file-management/upload`, {
  method: 'POST',
  credentials: 'include',
  body: formData, // No Content-Type header!
});
```

#### React Query Mutation Patterns
- Use `useMutation` for POST/PUT/DELETE operations
- Always call `queryClient.invalidateQueries()` on success
- Handle loading states via `isPending` property
- Return typed data from mutationFn

**Reference Pattern (from `/client/src/lib/api-hooks.ts:111-123`):**
```typescript
export function useCreateStudent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (studentData: Partial<Student>) => {
      return api.post<Student>(endpoints.students.create, studentData);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.students });
    },
  });
}
```

#### WebSocket Integration in Next.js
- Use native `WebSocket` API in client components ("use client")
- Implement auto-reconnect with exponential backoff
- Clean up WebSocket on component unmount
- Handle connection lifecycle (onopen, onmessage, onclose, onerror)

**Reference Pattern (from `/client/src/app/notifications/page.tsx:61-92`):**
```typescript
const ws = new WebSocket(wsUrl);
ws.onmessage = (event) => {
  const newNotification = JSON.parse(event.data);
  // Update state
};
ws.onclose = () => {
  setTimeout(setupWebSocket, 3000); // Reconnect
};
return () => ws.close(); // Cleanup
```

#### Chart/Visualization Libraries
**Recharts** is the recommended library for React:
- Compose charts with `ComposedChart`, `Line`, `Bar`, `Area` components
- Wrap in `ResponsiveContainer` for responsive sizing
- Use `CartesianGrid`, `XAxis`, `YAxis`, `Tooltip`, `Legend` for polish

**Example Pattern:**
```typescript
import { ComposedChart, Line, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

<ResponsiveContainer width="100%" height={400}>
  <ComposedChart data={data}>
    <CartesianGrid stroke="#f5f5f5" />
    <XAxis dataKey="name" />
    <YAxis />
    <Tooltip />
    <Legend />
    <Bar dataKey="pv" barSize={20} fill="#413ea0" />
    <Line type="monotone" dataKey="uv" stroke="#ff7300" />
  </ComposedChart>
</ResponsiveContainer>
```

---

## Codebase Architecture Analysis

### Existing Structure
```
client/
├── src/
│   ├── app/                    # Next.js App Router pages
│   │   ├── students/page.tsx   # Reference implementation
│   │   ├── files/page.tsx      # File upload reference
│   │   ├── roles/page.tsx      # Basic roles UI (needs permission matrix)
│   │   ├── exams/page.tsx      # Basic exams UI (needs deep features)
│   │   ├── batch-operations/page.tsx  # Import/export reference
│   │   ├── notifications/page.tsx     # WebSocket reference
│   │   ├── advanced-analytics/page.tsx # Analytics reference
│   │   └── faculty-tools/page.tsx     # Basic faculty tools
│   ├── components/
│   │   └── ui/                 # Shadcn/ui components (card, button, table, etc.)
│   ├── lib/
│   │   ├── api-client.ts       # API client with retry logic
│   │   ├── api-hooks.ts        # React Query hooks
│   │   ├── types.ts            # TypeScript interfaces
│   │   └── auth-context.tsx    # Authentication context
│   └── ...
```

### Key API Endpoints Available
- **Exams**: `/api/exams/*` - Full CRUD, enrollment, hall tickets, results
- **Roles**: `/api/roles/*`, `/api/permissions/*` - RBAC management
- **Reports**: `/api/reports/students/:id/gradecard`, `/api/reports/students/:id/transcript`
- **Analytics**: `/api/analytics/*` - Dashboard data
- **Files**: `/api/file-management/*` - Already implemented
- **Batch**: `/api/batch/*` - Already implemented

### Reusable Components Available
- **UI**: Card, Button, Input, Badge, Table, Dialog, Tabs, Select, etc.
- **Data Display**: Existing table patterns in students, files pages
- **Forms**: Dialog-based form patterns throughout
- **Charts**: None yet - need to add Recharts

---

## Task Dependency Graph

| Task | Depends On | Reason |
|------|------------|--------|
| **Task 1**: Install Recharts | None | Library installation, independent |
| **Task 2**: Course Materials Module UI | None | Can be developed in parallel |
| **Task 3**: Exam Creation Dialog | None | New component, independent |
| **Task 4**: Exam Enrollment UI | Task 3 | Needs exam data to exist |
| **Task 5**: Hall Ticket Generation | Task 3 | Needs exam data |
| **Task 6**: Exam Results Entry | Task 4 | Needs enrolled students |
| **Task 7**: Permission Matrix UI | None | Can be developed in parallel |
| **Task 8**: Reports (Gradecard/Transcript) | None | Can be developed in parallel |
| **Task 9**: Faculty Dashboard | Task 1 | Needs charts for analytics widgets |
| **Task 10**: Analytics Charts | Task 1 | Needs Recharts library |
| **Task 11**: WebSocket Context Provider | None | Can be developed in parallel |
| **Task 12**: Real-time Course Updates | Task 11 | Needs WebSocket provider |
| **Task 13**: Real-time Notification Badge | Task 11 | Needs WebSocket provider |

### Critical Path
Task 1 (Recharts) → Task 9 (Faculty Dashboard) + Task 10 (Analytics Charts)

---

## Parallel Execution Graph

### Wave 1: Foundation (Start Immediately)
These tasks have NO dependencies and can all start in parallel:

- **Task 1**: Install Recharts library for data visualization
- **Task 2**: Course Materials module management UI
- **Task 3**: Exam Creation Dialog
- **Task 7**: Permission Matrix UI for Roles
- **Task 8**: Reports (Gradecard/Transcript generation)
- **Task 11**: WebSocket Context Provider (real-time infrastructure)

**Parallel Group**: Tasks 1, 2, 3, 7, 8, 11

### Wave 2: Dependent Features (After Wave 1 completes)
These tasks depend on Wave 1 outputs:

- **Task 4**: Exam Enrollment UI (depends: Task 3)
- **Task 5**: Hall Ticket Generation (depends: Task 3)
- **Task 9**: Faculty Dashboard with charts (depends: Task 1)
- **Task 10**: Analytics Charts (depends: Task 1)
- **Task 12**: Real-time Course Updates (depends: Task 11)
- **Task 13**: Real-time Notification Badge (depends: Task 11)

**Parallel Groups:**
- Group A: Tasks 4, 5, 9, 10 (all depend on Wave 1, independent of each other)
- Group B: Tasks 12, 13 (both depend on Task 11, independent of each other)

### Wave 3: Final Integration (After Wave 2)
- **Task 6**: Exam Results Entry (depends: Task 4 - needs enrolled students)

**Parallel Speedup**: ~65% faster than sequential (13 tasks → 3 waves)

---

## Detailed Task Specifications

### Task 1: Install Recharts for Data Visualization
**Priority**: HIGH (blocks Task 9, 10)
**Estimated Effort**: 10 minutes
**Category**: `quick`
**Skills**: None required

**What to do:**
1. Install recharts package
2. Verify installation works with a test chart

**Implementation:**
```bash
cd client && npm install recharts
```

**Acceptance Criteria:**
- [ ] `recharts` added to package.json dependencies
- [ ] Test chart component renders without errors
- [ ] Build passes (`npm run build` succeeds)

**Files to Modify:**
- `client/package.json`

---

### Task 2: Course Materials Module Management UI
**Priority**: HIGH
**Estimated Effort**: 6 hours
**Category**: `unspecified-high`
**Skills**: `typescript-programmer`

**What to do:**
Create a comprehensive course materials module UI within course detail pages. This includes file organization by modules, upload, download, and version management.

**Files to Create:**
1. `client/src/app/courses/[courseId]/materials/page.tsx` - Materials management page
2. `client/src/components/course-materials/module-list.tsx` - Module listing component
3. `client/src/components/course-materials/file-browser.tsx` - File browser component
4. `client/src/lib/api-hooks.ts` - Add course materials hooks (append to existing file)

**Files to Modify:**
1. `client/src/lib/api-client.ts` - Add materials endpoints (if missing)
2. `client/src/lib/types.ts` - Add CourseMaterial type (if missing)

**API Endpoints to Use:**
- GET `/api/courses/:courseId/materials`
- POST `/api/courses/:courseId/materials/upload`
- DELETE `/api/courses/:courseId/materials/:id`

**Reusable Patterns from:**
- `client/src/app/files/page.tsx` (file upload, folder management)
- `client/src/app/courses/page.tsx` (course listing patterns)

**Acceptance Criteria:**
- [ ] Can view materials organized by module
- [ ] Can upload files to specific modules
- [ ] Can download files
- [ ] Can delete materials (faculty/admin only)
- [ ] Shows file metadata (size, type, upload date)
- [ ] Uses existing file upload pattern from files/page.tsx

**Must NOT do:**
- Do NOT create new UI components that duplicate existing ones from shadcn/ui
- Do NOT implement server-side file storage (backend handles this)
- Do NOT create new auth patterns (use existing `useAuth`)

---

### Task 3: Exam Creation Dialog
**Priority**: HIGH
**Estimated Effort**: 4 hours
**Category**: `unspecified-high`
**Skills**: `typescript-programmer`

**What to do:**
Create a comprehensive exam creation dialog that allows faculty/admin to create new exams with all required metadata.

**Files to Create:**
1. `client/src/components/exams/exam-create-dialog.tsx` - Exam creation dialog
2. `client/src/components/exams/exam-form.tsx` - Reusable exam form

**Files to Modify:**
1. `client/src/app/exams/page.tsx` - Add "Create Exam" button that opens dialog
2. `client/src/lib/api-hooks.ts` - Add `useCreateExam` mutation hook
3. `client/src/lib/types.ts` - Add Exam type if missing

**API Endpoints to Use:**
- POST `/api/exams`
- GET `/api/courses` (for course selection dropdown)

**Reusable Patterns from:**
- `client/src/app/students/page.tsx` (create form pattern)
- `client/src/app/files/page.tsx` (dialog patterns)

**Form Fields Required:**
- Title (text input)
- Description (textarea)
- Course (select dropdown from existing courses)
- Exam Type (select: midterm, final, quiz)
- Start Time (datetime picker)
- End Time (datetime picker)
- Duration (number input, minutes)
- Total Marks (number)
- Passing Marks (number)
- Room Number (text)

**Acceptance Criteria:**
- [ ] Dialog opens from "Schedule Exam" button on exams page
- [ ] All form fields validate input
- [ ] Submitting creates exam via API
- [ ] Success shows toast/notification
- [ ] Table refreshes after creation
- [ ] Form resets after successful submission

---

### Task 4: Exam Enrollment UI
**Priority**: MEDIUM
**Estimated Effort**: 5 hours
**Category**: `unspecified-high`
**Skills**: `typescript-programmer`

**What to do:**
Create UI for managing exam enrollments - viewing enrolled students, adding/removing students from exams.

**Files to Create:**
1. `client/src/components/exams/exam-enrollment-dialog.tsx` - Enrollment management dialog
2. `client/src/components/exams/enrolled-students-table.tsx` - Table of enrolled students

**Files to Modify:**
1. `client/src/app/exams/page.tsx` - Add enrollment action to exam rows
2. `client/src/lib/api-hooks.ts` - Add enrollment hooks

**API Endpoints to Use:**
- GET `/api/exams/:examId/enrollments`
- POST `/api/exams/:examId/enroll`
- DELETE `/api/exams/:examId/enrollments/:studentId`
- GET `/api/students` (for student selection)

**Reusable Patterns from:**
- `client/src/app/students/page.tsx` (table patterns)
- `client/src/app/files/page.tsx` (dialog patterns)

**Acceptance Criteria:**
- [ ] Can view list of enrolled students per exam
- [ ] Can search and add students to exam
- [ ] Can remove students from exam
- [ ] Shows enrollment status
- [ ] Validates against max capacity if applicable

---

### Task 5: Hall Ticket Generation UI
**Priority**: MEDIUM
**Estimated Effort**: 3 hours
**Category**: `unspecified-high`
**Skills**: `typescript-programmer`

**What to do:**
Create UI for generating and downloading hall tickets for exams. Both bulk generation and individual student tickets.

**Files to Create:**
1. `client/src/components/exams/hall-ticket-dialog.tsx` - Hall ticket generation dialog

**Files to Modify:**
1. `client/src/app/exams/page.tsx` - Add hall ticket action

**API Endpoints to Use:**
- GET `/api/exams/:examId/hall-tickets` (list)
- POST `/api/exams/:examId/hall-tickets/generate` (bulk generate)
- GET `/api/exams/:examId/hall-tickets/:studentId/download` (download)

**Acceptance Criteria:**
- [ ] Can generate hall tickets for all enrolled students
- [ ] Can download individual hall tickets
- [ ] Shows hall ticket status (generated/pending)
- [ ] Preview hall ticket before download
- [ ] Batch download all hall tickets as ZIP

---

### Task 6: Exam Results Entry UI
**Priority**: MEDIUM
**Estimated Effort**: 6 hours
**Category**: `unspecified-high`
**Skills**: `typescript-programmer`

**What to do:**
Create UI for entering and managing exam results. Faculty should be able to enter marks for all enrolled students efficiently.

**Files to Create:**
1. `client/src/app/exams/[examId]/results/page.tsx` - Results entry page
2. `client/src/components/exams/results-entry-table.tsx` - Results entry table with inline editing

**Files to Modify:**
1. `client/src/lib/api-hooks.ts` - Add results mutation hooks

**API Endpoints to Use:**
- GET `/api/exams/:examId/results`
- POST `/api/exams/:examId/results` (bulk update)
- PUT `/api/exams/:examId/results/:studentId` (individual update)
- GET `/api/exams/:examId/stats` (result statistics)

**Acceptance Criteria:**
- [ ] Table shows all enrolled students
- [ ] Can enter marks per student inline
- [ ] Shows total marks and auto-calculates percentage
- [ ] Validates marks don't exceed total marks
- [ ] Shows pass/fail status based on passing marks
- [ ] Can save results (individual or bulk)
- [ ] Shows result statistics (class average, highest, lowest)
- [ ] Import results from CSV (optional, reuse batch-operations pattern)

---

### Task 7: Permission Matrix UI for Roles
**Priority**: HIGH
**Estimated Effort**: 6 hours
**Category**: `unspecified-high`
**Skills**: `typescript-programmer`

**What to do:**
Create an interactive permission matrix interface for the roles page. Allow admins to assign/remove permissions from roles via a matrix/grid UI.

**Files to Create:**
1. `client/src/components/roles/permission-matrix.tsx` - Permission matrix component
2. `client/src/components/roles/permission-cell.tsx` - Individual permission toggle cell

**Files to Modify:**
1. `client/src/app/roles/page.tsx` - Replace "Manage Permissions" button with actual matrix

**API Endpoints to Use:**
- GET `/api/roles/:roleId/permissions` (current permissions)
- PUT `/api/roles/:roleId/permissions` (update permissions)
- GET `/api/permissions` (all available permissions)

**Reusable Patterns from:**
- `client/src/app/roles/page.tsx` (existing roles structure)
- `client/src/app/files/page.tsx` (table/grid patterns)

**UI Design:**
- Matrix with resources as rows, actions as columns
- Each cell is a checkbox/toggle
- Resources: students, courses, exams, grades, etc.
- Actions: create, read, update, delete
- Bulk actions (select all/none per row/column)
- Search/filter permissions

**Acceptance Criteria:**
- [ ] Shows matrix of all permissions vs roles
- [ ] Can toggle individual permissions on/off
- [ ] Can bulk select/deselect per resource or action
- [ ] Shows current permission state clearly
- [ ] Save changes via API
- [ ] Shows loading state during save
- [ ] Success/error feedback

---

### Task 8: Reports (Gradecard and Transcript Generation)
**Priority**: HIGH
**Estimated Effort**: 5 hours
**Category**: `unspecified-high`
**Skills**: `typescript-programmer`

**What to do:**
Create UI for generating and downloading student gradecards and transcripts with PDF download capability.

**Files to Create:**
1. `client/src/app/reports/page.tsx` - Reports dashboard
2. `client/src/components/reports/gradecard-viewer.tsx` - Gradecard display component
3. `client/src/components/reports/transcript-viewer.tsx` - Transcript display component

**Files to Modify:**
1. `client/src/lib/api-hooks.ts` - Add report hooks

**API Endpoints to Use:**
- GET `/api/reports/students/:id/gradecard`
- GET `/api/reports/students/:id/transcript`
- GET `/api/reports/students/:id/gradecard/download` (PDF)
- GET `/api/reports/students/:id/transcript/download` (PDF)

**Acceptance Criteria:**
- [ ] Can search and select students
- [ ] Shows gradecard with all courses and grades
- [ ] Shows transcript with academic history
- [ ] Can download as PDF
- [ ] Print-friendly view
- [ ] Shows GPA calculation
- [ ] Shows academic standing

---

### Task 9: Faculty Dashboard
**Priority**: HIGH
**Estimated Effort**: 8 hours
**Category**: `visual-engineering`
**Skills**: `typescript-programmer`, `frontend-ui-ux`

**What to do:**
Create a dedicated faculty dashboard at `/faculty-dashboard` with role-specific widgets, charts, and quick actions.

**Files to Create:**
1. `client/src/app/faculty-dashboard/page.tsx` - Main dashboard page
2. `client/src/components/faculty/stats-cards.tsx` - Statistics cards
3. `client/src/components/faculty/quick-actions.tsx` - Quick action buttons
4. `client/src/components/faculty/course-performance-chart.tsx` - Performance charts using Recharts
5. `client/src/components/faculty/recent-submissions.tsx` - Recent assignment submissions
6. `client/src/components/faculty/upcoming-lectures.tsx` - Upcoming schedule

**API Endpoints to Use:**
- GET `/api/faculty/dashboard` (dashboard stats)
- GET `/api/faculty/courses` (faculty's courses)
- GET `/api/faculty/assignments/pending` (pending grading)
- GET `/api/faculty/analytics` (course analytics)

**Dashboard Widgets:**
1. **Stats Row**: Active courses, total students, pending grades, upcoming exams
2. **Course Performance Chart**: Recharts bar chart of course averages
3. **Quick Actions**: Create assignment, schedule exam, send announcement, grade submissions
4. **Recent Submissions**: Table of latest assignment submissions needing grading
5. **Upcoming Lectures**: Schedule for next 7 days
6. **Notifications**: Recent notifications (reuse existing component)

**Acceptance Criteria:**
- [ ] Shows faculty-specific overview
- [ ] Interactive charts using Recharts
- [ ] Quick action buttons work
- [ ] Recent submissions show actionable items
- [ ] Responsive layout
- [ ] Data updates in real-time (optional, use existing patterns)

**Must NOT do:**
- Do NOT duplicate admin dashboard widgets
- Do NOT show data for courses the faculty doesn't teach

---

### Task 10: Analytics Charts
**Priority**: MEDIUM
**Estimated Effort**: 4 hours
**Category**: `visual-engineering`
**Skills**: `typescript-programmer`

**What to do:**
Add Recharts-based charts to the existing analytics page at `/advanced-analytics`.

**Files to Create:**
1. `client/src/components/analytics/grade-distribution-chart.tsx` - Grade distribution histogram
2. `client/src/components/analytics/attendance-trend-chart.tsx` - Attendance trend line chart
3. `client/src/components/analytics/course-performance-chart.tsx` - Course performance bar chart

**Files to Modify:**
1. `client/src/app/advanced-analytics/page.tsx` - Add chart components to tabs

**API Endpoints to Use:**
- GET `/api/analytics/courses/:id/grades/distribution`
- GET `/api/analytics/attendance/trends`
- GET `/api/analytics/courses/:id/analytics`

**Chart Types:**
1. **Grade Distribution**: Bar chart showing count of students per grade range
2. **Attendance Trends**: Line chart showing attendance % over time
3. **Course Performance**: Composed chart with grades, attendance, participation

**Acceptance Criteria:**
- [ ] Charts render using Recharts
- [ ] Responsive sizing
- [ ] Interactive tooltips
- [ ] Legend visible
- [ ] Updates when selecting different students/courses
- [ ] Loading states handled

---

### Task 11: WebSocket Context Provider
**Priority**: HIGH
**Estimated Effort**: 3 hours
**Category**: `unspecified-high`
**Skills**: `typescript-programmer`

**What to do:**
Create a reusable WebSocket context provider that can be used across the app for real-time updates. This centralizes WebSocket management.

**Files to Create:**
1. `client/src/lib/websocket-context.tsx` - WebSocket context and provider

**Reusable Patterns from:**
- `client/src/app/notifications/page.tsx:61-92` (existing WebSocket implementation)
- `client/src/lib/auth-context.tsx` (context pattern)

**Features:**
- Auto-connect on auth
- Auto-reconnect with backoff
- Subscribe/unsubscribe to channels
- Typed message handlers
- Connection status tracking
- Centralized error handling

**API:**
```typescript
const { 
  isConnected, 
  subscribe, 
  unsubscribe, 
  send 
} = useWebSocket();
```

**Acceptance Criteria:**
- [ ] Context provider wraps app in layout
- [ ] Auto-connects when user is authenticated
- [ ] Reconnects on disconnect
- [ ] Provides subscribe/unsubscribe methods
- [ ] Tracks connection status
- [ ] TypeScript types for messages

---

### Task 12: Real-time Course Updates
**Priority**: LOW
**Estimated Effort**: 3 hours
**Category**: `unspecified-high`
**Skills**: `typescript-programmer`

**What to do:**
Implement real-time updates for course pages using WebSocket. When a course is modified, all viewers see updates immediately.

**Files to Modify:**
1. `client/src/app/courses/page.tsx` - Add WebSocket subscription
2. `client/src/app/courses/[courseId]/page.tsx` - Add real-time updates

**WebSocket Channels:**
- `course:${courseId}` - Specific course updates
- `courses` - General course list updates

**Acceptance Criteria:**
- [ ] Subscribes to course channel on mount
- [ ] Receives real-time updates when course changes
- [ ] Updates UI without refresh
- [ ] Unsubscribes on unmount

---

### Task 13: Real-time Notification Badge
**Priority**: LOW
**Estimated Effort**: 2 hours
**Category**: `quick`
**Skills**: `typescript-programmer`

**What to do:**
Add real-time notification badge updates in the top navigation bar using WebSocket.

**Files to Create/Modify:**
1. `client/src/components/navigation/topbar.tsx` - Add notification badge with WebSocket

**Reusable Patterns from:**
- `client/src/app/notifications/page.tsx` (existing WebSocket code)

**Acceptance Criteria:**
- [ ] Shows badge with unread count
- [ ] Updates in real-time when new notifications arrive
- [ ] Links to notifications page
- [ ] Clears/hides when count is 0

---

## Commit Strategy

| After Task | Commit Message | Files | Verification |
|------------|----------------|-------|--------------|
| Task 1 | `chore(deps): add recharts for data visualization` | package.json, package-lock.json | Build passes |
| Task 2 | `feat(materials): add course materials module UI` | materials/*, api-hooks.ts | Test upload/download |
| Task 3 | `feat(exams): add exam creation dialog` | exam-create-dialog.tsx, exam-form.tsx, exams/page.tsx | Create exam works |
| Task 4 | `feat(exams): add exam enrollment management` | exam-enrollment-dialog.tsx, exams/page.tsx | Enroll students works |
| Task 5 | `feat(exams): add hall ticket generation` | hall-ticket-dialog.tsx | Generate/download tickets |
| Task 6 | `feat(exams): add results entry UI` | results/page.tsx, results-entry-table.tsx | Enter/save results |
| Task 7 | `feat(roles): add permission matrix UI` | permission-matrix.tsx, roles/page.tsx | Assign/remove permissions |
| Task 8 | `feat(reports): add gradecard and transcript UI` | reports/page.tsx, gradecard-viewer.tsx, transcript-viewer.tsx | Generate/download reports |
| Task 9 | `feat(dashboard): add faculty dashboard` | faculty-dashboard/page.tsx, faculty/* | Dashboard renders |
| Task 10 | `feat(analytics): add charts to analytics page` | analytics/*, advanced-analytics/page.tsx | Charts render |
| Task 11 | `feat(websocket): add WebSocket context provider` | websocket-context.tsx | Connection works |
| Task 12 | `feat(realtime): add real-time course updates` | courses/page.tsx | Updates received |
| Task 13 | `feat(realtime): add real-time notification badge` | topbar.tsx | Badge updates |

---

## Success Criteria

### Overall Completion Criteria
- [ ] All 13 tasks completed and committed
- [ ] No TypeScript errors (`npm run build` passes)
- [ ] All API integrations tested
- [ ] UI follows existing design patterns
- [ ] Responsive on desktop and tablet

### Per-Feature Verification

**Course Materials:**
- Upload file to course module ✓
- View module files ✓
- Download file ✓
- Delete file (faculty only) ✓

**Exam Management:**
- Create new exam ✓
- Enroll students ✓
- Generate hall tickets ✓
- Enter results ✓
- View results ✓

**Role Management:**
- View permission matrix ✓
- Toggle permissions ✓
- Save changes ✓

**Reports:**
- View gradecard ✓
- View transcript ✓
- Download PDF ✓

**Faculty Dashboard:**
- Dashboard loads ✓
- Charts display ✓
- Quick actions work ✓
- Shows faculty-specific data ✓

**Analytics:**
- Charts render ✓
- Data updates on selection ✓
- Interactive tooltips ✓

**Real-time:**
- WebSocket connects ✓
- Notifications update badge ✓
- Course updates reflect ✓

---

## Risk Mitigation

### Identified Risks

1. **Backend API Changes**: APIs are 95% complete but may need minor adjustments
   - **Mitigation**: Test each API endpoint before implementing frontend
   - **Contingency**: Mock data for development if API not ready

2. **Recharts Bundle Size**: Charts library may increase bundle size
   - **Mitigation**: Import only needed components (tree-shaking)
   - **Contingency**: Use dynamic imports for chart components

3. **WebSocket Scaling**: Real-time features may not scale with many users
   - **Mitigation**: Limit to high-value features (notifications, critical updates)
   - **Contingency**: Fallback to polling for less critical updates

4. **Permission Complexity**: RBAC matrix may be complex to implement
   - **Mitigation**: Start with simple grid, iterate based on feedback
   - **Contingency**: Use simpler list-based UI if matrix is too complex

### Guardrails (Must NOT Have)

- **No new authentication patterns** - Use existing `useAuth` context
- **No new API client patterns** - Use existing `api-client.ts`
- **No new component libraries** - Use existing shadcn/ui
- **No breaking changes to existing pages** - Add features, don't modify existing functionality
- **No server-side storage implementation** - Backend handles all persistence
- **No complex state management** - React Query is sufficient

---

## Shared Components Reference

### Existing Components to Reuse

**From shadcn/ui:**
- `Card`, `CardHeader`, `CardTitle`, `CardContent`, `CardDescription`
- `Button` (all variants)
- `Input`, `Textarea`, `Label`
- `Select`, `SelectContent`, `SelectItem`, `SelectTrigger`, `SelectValue`
- `Dialog`, `DialogContent`, `DialogHeader`, `DialogTitle`, `DialogFooter`
- `Table`, `TableHeader`, `TableBody`, `TableRow`, `TableCell`, `TableHead`
- `Tabs`, `TabsList`, `TabsContent`, `TabsTrigger`
- `Badge`
- `Progress`
- `Avatar`, `AvatarImage`, `AvatarFallback`
- `DropdownMenu`, `DropdownMenuContent`, `DropdownMenuItem`, `DropdownMenuTrigger`

**Custom Components:**
- `ProtectedRoute` - For auth-required pages
- `ErrorBoundary` - For error handling

### Patterns to Follow

**Page Structure:**
```typescript
"use client";
import { useState } from "react";
import { useAuth } from "@/lib/auth-context";
// ... other imports

export default function PageName() {
  const { user } = useAuth();
  const [state, setState] = useState();
  
  // React Query hooks
  // Local state
  // Handlers
  
  return (
    <div className="space-y-6">
      {/* Header with title and actions */}
      {/* Stats cards (if applicable) */}
      {/* Main content */}
    </div>
  );
}
```

**Dialog Pattern:**
```typescript
<Dialog open={open} onOpenChange={setOpen}>
  <DialogTrigger asChild>
    <Button>Open Dialog</Button>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Title</DialogTitle>
      <DialogDescription>Description</DialogDescription>
    </DialogHeader>
    {/* Form or content */}
    <DialogFooter>
      <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
      <Button onClick={handleSubmit}>Submit</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

**Table Pattern:**
```typescript
<Table>
  <TableHeader>
    <TableRow>
      <TableHead>Column 1</TableHead>
      <TableHead>Column 2</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    {data.map((item) => (
      <TableRow key={item.id}>
        <TableCell>{item.value1}</TableCell>
        <TableCell>{item.value2}</TableCell>
      </TableRow>
    ))}
  </TableBody>
</Table>
```

**Mutation Hook Pattern:**
```typescript
export function useCreateX() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: CreateXInput) => {
      return api.post<X>(endpoints.x.create, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.x });
    },
  });
}
```

---

## Appendix: Existing Endpoint Reference

**Key endpoints available (from `/client/src/lib/api-client.ts`):**

```typescript
// Exams (missing deep integration)
exams: {
  list: '/api/exams',
  get: (id: number) => `/api/exams/${id}`,
  create: '/api/exams',
  update: (id: number) => `/api/exams/${id}`,
  delete: (id: number) => `/api/exams/${id}`,
}

// Roles (missing permission matrix)
roles: {
  list: '/api/roles',
  get: (id: number) => `/api/roles/${id}`,
  create: '/api/roles',
  update: (id: number) => `/api/roles/${id}`,
  delete: (id: number) => `/api/roles/${id}`,
  assignPermissions: (id: number) => `/api/roles/${id}/permissions`,
}
permissions: {
  list: '/api/permissions',
}

// Reports (missing UI)
reports: {
  gradeCard: (studentId: number) => `/api/reports/students/${studentId}/gradecard`,
  transcript: (studentId: number) => `/api/reports/students/${studentId}/transcript`,
}

// Analytics (missing charts)
analytics: {
  collegeDashboard: '/api/analytics/dashboard',
  courseAnalytics: (courseId: number) => `/api/analytics/courses/${courseId}/analytics`,
  gradeDistribution: (courseId: number) => `/api/analytics/courses/${courseId}/grades/distribution`,
  studentPerformance: (studentId: number) => `/api/analytics/students/${studentId}/performance`,
  attendanceTrends: '/api/analytics/attendance/trends',
}

// Files (already implemented - reference pattern)
files: {
  upload: '/api/files/upload',
  delete: (key: string) => `/api/files/${key}`,
  getUrl: (key: string) => `/api/files/${key}/url`,
}
```

---

## Execution Command

To execute this plan with delegate_task agents:

```typescript
// Wave 1 - Start all immediately
delegate_task({
  category: 'quick',
  description: 'Task 1: Install Recharts',
  prompt: 'Install recharts package in client directory. Update package.json, ensure build passes.',
  run_in_background: true
});

delegate_task({
  category: 'unspecified-high',
  description: 'Task 2: Course Materials Module UI',
  prompt: 'Create course materials module UI with file upload/download. Follow patterns from files/page.tsx. Create: materials page, module list, file browser. Use existing API endpoints.',
  run_in_background: true
});

delegate_task({
  category: 'unspecified-high',
  description: 'Task 3: Exam Creation Dialog',
  prompt: 'Create exam creation dialog with form fields for title, description, course, type, dates, marks, room. Add to exams/page.tsx. Follow student creation pattern from students/page.tsx.',
  run_in_background: true
});

delegate_task({
  category: 'unspecified-high',
  description: 'Task 7: Permission Matrix UI',
  prompt: 'Create interactive permission matrix for roles page. Grid UI with resources as rows, actions as columns. Toggle permissions on/off. Save via API.',
  run_in_background: true
});

delegate_task({
  category: 'unspecified-high',
  description: 'Task 8: Reports UI',
  prompt: 'Create reports page for gradecard and transcript generation. Search students, view reports, download PDF.',
  run_in_background: true
});

delegate_task({
  category: 'unspecified-high',
  description: 'Task 11: WebSocket Context Provider',
  prompt: 'Create WebSocket context provider for real-time updates. Auto-connect, reconnect, subscribe/unsubscribe methods. Follow notification WebSocket pattern.',
  run_in_background: true
});

// Wave 2 - After Wave 1 completes
delegate_task({
  category: 'unspecified-high',
  description: 'Task 4: Exam Enrollment UI',
  prompt: 'Create exam enrollment UI. Dialog to view/add/remove students from exams. Depends on Task 3.',
  run_in_background: true
});

delegate_task({
  category: 'unspecified-high',
  description: 'Task 5: Hall Ticket Generation',
  prompt: 'Create hall ticket generation UI. Generate tickets for enrolled students, download individual/bulk. Depends on Task 3.',
  run_in_background: true
});

delegate_task({
  category: 'visual-engineering',
  description: 'Task 9: Faculty Dashboard',
  prompt: 'Create faculty dashboard page with charts, stats, quick actions. Use Recharts for visualizations. Widgets: stats, course performance chart, quick actions, recent submissions, upcoming lectures.',
  run_in_background: true
});

delegate_task({
  category: 'visual-engineering',
  description: 'Task 10: Analytics Charts',
  prompt: 'Add Recharts charts to advanced-analytics page. Grade distribution histogram, attendance trends, course performance. Use Recharts library.',
  run_in_background: true
});

delegate_task({
  category: 'unspecified-high',
  description: 'Task 12: Real-time Course Updates',
  prompt: 'Add real-time updates to course pages using WebSocket context. Subscribe to course channels, update UI on changes. Depends on Task 11.',
  run_in_background: true
});

delegate_task({
  category: 'quick',
  description: 'Task 13: Real-time Notification Badge',
  prompt: 'Add real-time notification badge to topbar using WebSocket context. Updates when new notifications arrive. Depends on Task 11.',
  run_in_background: true
});

// Wave 3 - Final integration
delegate_task({
  category: 'unspecified-high',
  description: 'Task 6: Exam Results Entry',
  prompt: 'Create exam results entry page. Table for entering marks per student. Inline editing, bulk save. Show statistics. Depends on Task 4 (enrollment).',
  run_in_background: true
});
```

---

**Plan Version**: 1.0
**Created**: 2026-02-01
**Total Tasks**: 13
**Estimated Total Effort**: 55 hours
**Parallel Waves**: 3
**Time Savings**: ~65% vs sequential execution
