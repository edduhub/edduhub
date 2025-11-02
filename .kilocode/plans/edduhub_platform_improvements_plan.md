# EdduHub Platform Improvements - Architecture & Implementation Plan

**Status:** DRAFT
**Created:** 2025-10-31

## 1. Architecture Overview

This document outlines a comprehensive plan to address critical security vulnerabilities, complete partially implemented features, and build new functionality for the EdduHub platform. The plan is divided into four phases, prioritizing security and stability first.

### 1.1. Security First (RBAC & Multi-Tenancy)

The current implementation has critical security flaws. This plan addresses them with:

*   **RBAC with Ory Keto:** We will enforce fine-grained permissions for every API endpoint using a new middleware that leverages the existing `RequirePermission` function. This will replace the inadequate `RequireRole` checks.
*   **Multi-Tenancy Isolation:** We will enforce strict data isolation between colleges at the database query level. This will be achieved by creating a new `db` package that ensures every query has a `college_id` filter.
*   **Input Validation:** We will introduce a validation middleware that uses the `go-playground/validator` library to automatically validate all incoming request bodies.

### 1.2. Feature Completion & Development

The plan includes detailed tasks to complete the QR attendance, assignment submission, and grades viewing features. It also includes a plan to build the dynamic timetable system from scratch.

## 2. Technical Decisions

*   **RBAC Middleware:** A new middleware will be created in `server/internal/middleware/permission.go` that will be used to protect all API endpoints.
*   **Ory Keto Integration:** We will define namespaces and relationship tuples in `auth/keto/keto.yml` to model the application's authorization policies.
*   **Database Query Patterns:** A new `db` package will be created to abstract database operations and enforce multi-tenancy.
*   **Input Validation:** We will use the `go-playground/validator` library and create a custom validation middleware in `server/internal/middleware/validation.go`.

## 3. Implementation Phases

### Phase 1: Critical Security (2 weeks)

*   **Task 1: Implement RBAC Middleware & Permissions**
    *   **Files to create:**
        *   `server/internal/middleware/permission.go`: New RBAC middleware.
        *   `auth/keto/namespaces.yml`: Ory Keto namespace definitions.
        *   `auth/keto/relationships.yml`: Ory Keto relationship tuples.
    *   **Files to modify:**
        *   `server/api/handler/router.go`: Replace `RequireRole` with new RBAC middleware.
        *   `auth/keto/keto.yml`: Update to use new namespace and relationship files.
    *   **Changes:**
        *   Implement the `permission.go` middleware to check for permissions using Ory Keto.
        *   Define namespaces for all resources (courses, students, etc.) in `namespaces.yml`.
        *   Define relationship tuples for all resources in `relationships.yml`.
        *   Update `router.go` to use the new middleware for all routes.
*   **Task 2: Implement Multi-Tenancy Isolation**
    *   **Files to create:**
        *   `server/internal/db/db.go`: New database package to enforce multi-tenancy.
    *   **Files to modify:**
        *   All repository files in `server/internal/repository/`: Update to use the new `db` package.
    *   **Changes:**
        *   Implement the `db.go` package to ensure all queries have a `college_id` filter.
        *   Update all repository files to use the new `db` package for all database operations.
*   **Task 3: Add Comprehensive Input Validation**
    *   **Files to create:**
        *   `server/internal/middleware/validation.go`: New validation middleware.
    *   **Files to modify:**
        *   `server/api/handler/router.go`: Add the new validation middleware to all routes with request bodies.
        *   All model files in `server/internal/models/`: Add validation tags to all struct fields.
    *   **Changes:**
        *   Implement the `validation.go` middleware to automatically validate request bodies.
        *   Add validation tags to all struct fields in the model files.

### Phase 2: Complete Partially Implemented Features (3 weeks)

*   **Task 4: Complete QR Attendance System**
    *   **Files to modify:**
        *   `server/api/handler/attendance_handler.go`: Implement QR code scanning and validation.
        *   `client/src/app/attendance/page.tsx`: Add QR code scanning functionality.
    *   **Changes:**
        *   Implement the `ProcessAttendance` handler to validate QR codes and mark attendance.
        *   Add a QR code scanner to the frontend.
*   **Task 5: Complete Assignment Submission**
    *   **Files to modify:**
        *   `server/api/handler/assignment_handler.go`: Implement file upload and submission tracking.
        *   `client/src/app/assignments/page.tsx`: Add file upload functionality.
    *   **Changes:**
        *   Implement the `SubmitAssignment` handler to handle file uploads and track submissions.
        *   Add a file upload component to the frontend.
*   **Task 6: Complete Grades Viewing with RBAC**
    *   **Files to modify:**
        *   `server/api/handler/grade_handler.go`: Add RBAC checks to ensure students only see their own grades.
        *   `client/src/app/grades/page.tsx`: Update to handle access denied errors.
    *   **Changes:**
        *   Add `RequirePermission` middleware to the `GetStudentGrades` handler.
        *   Update the frontend to gracefully handle cases where a user is not authorized to view grades.

### Phase 3: Build Missing Features (4 weeks)

*   **Task 7: Implement Dynamic Timetable System**
    *   **Files to create:**
        *   `server/internal/models/timetable.go`: Timetable model.
        *   `server/internal/repository/timetable_repository.go`: Timetable repository.
        *   `server/internal/services/timetable/timetable_service.go`: Timetable service.
        *   `server/api/handler/timetable_handler.go`: Timetable handler.
        *   `client/src/app/timetable/page.tsx`: Timetable page.
    *   **Files to modify:**
        *   `server/api/handler/router.go`: Add routes for the timetable feature.
    *   **Changes:**
        *   Implement the full CRUD functionality for the timetable feature.
        *   Create a new page on the frontend to display the timetable.

### Phase 4: Testing and Optimization (2 weeks)

*   **Task 8: Add Comprehensive Tests**
    *   **Files to create:**
        *   Unit and integration tests for all new and modified code.
    *   **Changes:**
        *   Write tests to ensure the quality and correctness of the implementation.
*   **Task 9: Performance Optimization**
    *   **Files to modify:**
        *   `server/internal/repository/`: Add caching to frequently accessed queries.
    *   **Changes:**
        *   Use a caching library like `go-redis` to cache database queries and improve performance.

## 4. Database Schema Changes

*   **New Tables:**
    *   `timetables`
*   **New Columns:**
    *   Add `college_id` to all tables that don't have it.

## 5. API Endpoint Security

All API endpoints will be protected with the new RBAC middleware. The following is a sample of the permissions that will be defined:

*   `courses:create`
*   `courses:read`
*   `courses:update`
*   `courses:delete`
*   `students:create`
*   `students:read`
*   `students:update`
*   `students:delete`

## 6. Edge Cases and Error Handling

*   **Permission Denied:** The API will return a `403 Forbidden` error.
*   **Cross-College Access:** The API will return a `403 Forbidden` error.
*   **File Uploads:** The API will have limits on file size and type.
*   **QR Code Expiration:** QR codes will have a short expiration time to prevent misuse.

## 7. Testing Strategy

*   **Unit Tests:** All new services and helpers will have 100% unit test coverage.
*   **Integration Tests:** We will add integration tests for all API endpoints.
*   **E2E Tests:** We will create E2E tests for the main user flows.
