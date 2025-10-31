# Implementation Guide for Missing Features

**Date:** October 31, 2025
**Purpose:** Step-by-step guide to implement identified missing features
**Related Documents:** MISSING_FEATURES_ANALYSIS.md

---

## Quick Start

This guide provides practical steps to implement each missing feature identified in the codebase analysis.

---

## Priority 1: Course Materials Management System

### Step 1: Database Migration

Create migration file: `server/migrations/XXX_add_course_materials.up.sql`

```sql
-- Course Modules Table
CREATE TABLE IF NOT EXISTS course_modules (
    id SERIAL PRIMARY KEY,
    course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    display_order INT DEFAULT 0,
    is_published BOOLEAN DEFAULT false,
    college_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_course_modules_course ON course_modules(course_id);
CREATE INDEX idx_course_modules_published ON course_modules(is_published);

-- Course Materials Table
CREATE TABLE IF NOT EXISTS course_materials (
    id SERIAL PRIMARY KEY,
    course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    module_id INT REFERENCES course_modules(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('document', 'video', 'link', 'assignment', 'quiz')),
    file_id INT REFERENCES files(id) ON DELETE SET NULL,
    file_url TEXT,
    external_url TEXT,
    display_order INT DEFAULT 0,
    is_published BOOLEAN DEFAULT false,
    published_at TIMESTAMP,
    due_date TIMESTAMP,
    uploaded_by INT NOT NULL,
    college_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_course_materials_course ON course_materials(course_id);
CREATE INDEX idx_course_materials_module ON course_materials(module_id);
CREATE INDEX idx_course_materials_published ON course_materials(is_published);
CREATE INDEX idx_course_materials_type ON course_materials(type);

-- Material Access Logs Table
CREATE TABLE IF NOT EXISTS material_access_logs (
    id SERIAL PRIMARY KEY,
    material_id INT NOT NULL REFERENCES course_materials(id) ON DELETE CASCADE,
    student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    duration_seconds INT,
    completed BOOLEAN DEFAULT false
);

CREATE INDEX idx_material_access_material ON material_access_logs(material_id);
CREATE INDEX idx_material_access_student ON material_access_logs(student_id);
CREATE INDEX idx_material_access_date ON material_access_logs(accessed_at);
```

Down migration: `server/migrations/XXX_add_course_materials.down.sql`

```sql
DROP TABLE IF EXISTS material_access_logs;
DROP TABLE IF EXISTS course_materials;
DROP TABLE IF EXISTS course_modules;
```

### Step 2: Repository Implementation

Create: `server/internal/repository/course_material_repository.go`

```go
package repository

import (
	"context"
	"eduhub/server/internal/models"
)

type CourseMaterialRepository interface {
	// Module operations
	CreateModule(ctx context.Context, module *models.CourseModule) error
	GetModuleByID(ctx context.Context, collegeID, moduleID int) (*models.CourseModule, error)
	ListModulesByCourse(ctx context.Context, collegeID, courseID int) ([]*models.CourseModule, error)
	UpdateModule(ctx context.Context, module *models.CourseModule) error
	DeleteModule(ctx context.Context, collegeID, moduleID int) error

	// Material operations
	CreateMaterial(ctx context.Context, material *models.CourseMaterial) error
	GetMaterialByID(ctx context.Context, collegeID, materialID int) (*models.CourseMaterial, error)
	GetMaterialWithDetails(ctx context.Context, collegeID, materialID int) (*models.CourseMaterialWithDetails, error)
	ListMaterialsByCourse(ctx context.Context, collegeID, courseID int) ([]*models.CourseMaterial, error)
	ListMaterialsByCourseWithDetails(ctx context.Context, collegeID, courseID int, onlyPublished bool) ([]*models.CourseMaterialWithDetails, error)
	ListMaterialsByModule(ctx context.Context, collegeID, moduleID int) ([]*models.CourseMaterial, error)
	ListMaterialsByModuleWithDetails(ctx context.Context, collegeID, moduleID int, onlyPublished bool) ([]*models.CourseMaterialWithDetails, error)
	UpdateMaterial(ctx context.Context, material *models.CourseMaterial) error
	DeleteMaterial(ctx context.Context, collegeID, materialID int) error

	// Access tracking
	LogAccess(ctx context.Context, materialID, studentID int, durationSeconds int, completed bool) error
	GetAccessStats(ctx context.Context, materialID int) (map[string]interface{}, error)
	GetStudentProgress(ctx context.Context, courseID, studentID int) (map[string]interface{}, error)
}
```

### Step 3: API Handler

Create: `server/api/handler/course_material_handler.go`

Key endpoints:
- `POST /api/courses/:courseID/modules`
- `GET /api/courses/:courseID/modules`
- `POST /api/courses/:courseID/materials`
- `GET /api/courses/:courseID/materials`
- `POST /api/materials/:materialID/access` (log student access)

### Step 4: Frontend Implementation

Create: `client/src/app/courses/[courseId]/materials/page.tsx`

Components needed:
- Material browser with folder/module view
- Upload interface for faculty
- Material preview (PDF, video player)
- Progress tracking for students
- Module organization drag-and-drop

### Step 5: Testing

Create tests:
- `server/internal/services/course_material/course_material_service_test.go`
- `server/api/handler/course_material_handler_test.go`
- `client/tests/course-materials.spec.ts`

---

## Priority 2: Timetable/Schedule System

### Step 1: Database Schema

```sql
CREATE TABLE timetable_slots (
    id SERIAL PRIMARY KEY,
    course_id INT NOT NULL REFERENCES courses(id),
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 1 AND 7),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    room_number VARCHAR(50),
    building VARCHAR(100),
    faculty_id INT REFERENCES users(id),
    type VARCHAR(50) CHECK (type IN ('lecture', 'lab', 'tutorial', 'practical')),
    is_active BOOLEAN DEFAULT true,
    semester VARCHAR(50),
    college_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT no_time_overlap UNIQUE(course_id, day_of_week, start_time, semester)
);

CREATE INDEX idx_timetable_course ON timetable_slots(course_id);
CREATE INDEX idx_timetable_faculty ON timetable_slots(faculty_id);
CREATE INDEX idx_timetable_day ON timetable_slots(day_of_week);
CREATE INDEX idx_timetable_active ON timetable_slots(is_active);
```

### Step 2: Conflict Detection Logic

```go
func (s *timetableService) CheckConflicts(ctx context.Context, slot *TimetableSlot) ([]Conflict, error) {
    var conflicts []Conflict

    // Check faculty conflicts
    facultySlots, err := s.repo.GetFacultySlots(ctx, slot.FacultyID, slot.DayOfWeek, slot.Semester)
    if err != nil {
        return nil, err
    }

    for _, existing := range facultySlots {
        if timeOverlaps(slot.StartTime, slot.EndTime, existing.StartTime, existing.EndTime) {
            conflicts = append(conflicts, Conflict{
                Type: "faculty",
                Message: fmt.Sprintf("Faculty has another class at this time"),
                ExistingSlot: existing,
            })
        }
    }

    // Check room conflicts
    roomSlots, err := s.repo.GetRoomSlots(ctx, slot.RoomNumber, slot.DayOfWeek, slot.Semester)
    if err != nil {
        return nil, err
    }

    for _, existing := range roomSlots {
        if timeOverlaps(slot.StartTime, slot.EndTime, existing.StartTime, existing.EndTime) {
            conflicts = append(conflicts, Conflict{
                Type: "room",
                Message: fmt.Sprintf("Room is occupied at this time"),
                ExistingSlot: existing,
            })
        }
    }

    return conflicts, nil
}
```

### Step 3: Frontend Timetable Component

```typescript
// Weekly calendar view
const TimetableView = ({ slots }: { slots: TimetableSlot[] }) => {
  const days = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday'];
  const hours = Array.from({ length: 12 }, (_, i) => i + 8); // 8 AM to 8 PM

  return (
    <div className="timetable-grid">
      {days.map(day => (
        <div key={day} className="day-column">
          <h3>{day}</h3>
          {hours.map(hour => (
            <div key={hour} className="hour-slot">
              {renderSlotsForTime(slots, day, hour)}
            </div>
          ))}
        </div>
      ))}
    </div>
  );
};
```

---

## Priority 3: Advanced Search System

### Implementation Pattern

```go
type SearchService struct {
    db *pgxpool.Pool
}

func (s *SearchService) GlobalSearch(ctx context.Context, query string, types []string, collegeID int) (*SearchResults, error) {
    results := &SearchResults{}

    // Use PostgreSQL full-text search
    for _, entityType := range types {
        switch entityType {
        case "students":
            students, err := s.searchStudents(ctx, query, collegeID)
            if err == nil {
                results.Students = students
            }
        case "courses":
            courses, err := s.searchCourses(ctx, query, collegeID)
            if err == nil {
                results.Courses = courses
            }
        case "materials":
            materials, err := s.searchMaterials(ctx, query, collegeID)
            if err == nil {
                results.Materials = materials
            }
        }
    }

    return results, nil
}

func (s *SearchService) searchStudents(ctx context.Context, query string, collegeID int) ([]*Student, error) {
    sql := `
        SELECT * FROM students
        WHERE college_id = $1
        AND (
            to_tsvector('english', first_name || ' ' || last_name) @@ plainto_tsquery('english', $2)
            OR roll_no ILIKE $3
            OR email ILIKE $3
        )
        ORDER BY
            ts_rank(to_tsvector('english', first_name || ' ' || last_name), plainto_tsquery('english', $2)) DESC
        LIMIT 20
    `

    var students []*Student
    err := s.db.Select(ctx, &students, sql, collegeID, query, "%"+query+"%")
    return students, err
}
```

---

## Quick Wins Implementation

### 1. Email Templates (1-2 days)

Create: `server/internal/services/email/templates.go`

```go
package email

var templates = map[string]string{
    "assignment_due": `
        <h2>Assignment Due Reminder</h2>
        <p>Dear {{.StudentName}},</p>
        <p>This is a reminder that the assignment <strong>{{.AssignmentTitle}}</strong>
        for course {{.CourseName}} is due on {{.DueDate}}.</p>
        <p>Please submit your work before the deadline.</p>
    `,
    "grade_posted": `
        <h2>New Grade Posted</h2>
        <p>Dear {{.StudentName}},</p>
        <p>Your grade for {{.AssessmentName}} in {{.CourseName}} has been posted.</p>
        <p>Score: {{.Score}}/{{.TotalMarks}} ({{.Percentage}}%)</p>
    `,
    "new_material": `
        <h2>New Learning Material Available</h2>
        <p>Dear {{.StudentName}},</p>
        <p>New learning material "{{.MaterialTitle}}" has been posted in {{.CourseName}}.</p>
        <p><a href="{{.MaterialURL}}">View Material</a></p>
    `,
}

func (s *emailService) SendFromTemplate(to, templateName string, data interface{}) error {
    tmpl, exists := templates[templateName]
    if !exists {
        return fmt.Errorf("template not found: %s", templateName)
    }

    t := template.Must(template.New(templateName).Parse(tmpl))
    var body bytes.Buffer
    if err := t.Execute(&body, data); err != nil {
        return err
    }

    return s.Send(to, getSubjectForTemplate(templateName), body.String())
}
```

### 2. Dashboard Widgets (3-4 days)

```typescript
// Widget system
interface Widget {
  id: string;
  type: string;
  title: string;
  size: 'small' | 'medium' | 'large';
  position: { x: number; y: number };
}

const DashboardWithWidgets = () => {
  const [widgets, setWidgets] = useState<Widget[]>([
    { id: '1', type: 'grades', title: 'Recent Grades', size: 'medium', position: { x: 0, y: 0 } },
    { id: '2', type: 'attendance', title: 'Attendance', size: 'small', position: { x: 2, y: 0 } },
    // ... more widgets
  ]);

  return (
    <GridLayout cols={3} rowHeight={200}>
      {widgets.map(widget => (
        <div key={widget.id} data-grid={widget.position}>
          <WidgetComponent type={widget.type} data={widget} />
        </div>
      ))}
    </GridLayout>
  );
};
```

---

## Testing Strategy

### Unit Tests Template

```go
func TestCourseMaterialService_CreateMaterial(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*mock.Mock)
        req     *CreateMaterialRequest
        wantErr bool
    }{
        {
            name: "success",
            setup: func(m *mock.Mock) {
                m.On("GetCourseByID", mock.Anything, 1, 1).Return(&Course{ID: 1}, nil)
                m.On("CreateMaterial", mock.Anything, mock.Anything).Return(nil)
            },
            req: &CreateMaterialRequest{
                Title: "Test Material",
                Type: "document",
            },
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockRepository)
            tt.setup(&mockRepo.Mock)

            svc := NewCourseMaterialService(mockRepo)
            _, err := svc.CreateMaterial(context.Background(), 1, 1, 1, tt.req)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### E2E Tests Template

```typescript
test('should create and view course material', async ({ page }) => {
  // Navigate to course materials
  await page.goto('/courses/1/materials');

  // Click add material button
  await page.click('button:has-text("Add Material")');

  // Fill in the form
  await page.fill('input[name="title"]', 'Test Material');
  await page.selectOption('select[name="type"]', 'document');

  // Upload file
  await page.setInputFiles('input[type="file"]', 'test-file.pdf');

  // Submit
  await page.click('button:has-text("Save")');

  // Verify material appears in list
  await expect(page.locator('text=Test Material')).toBeVisible();
});
```

---

## Deployment Checklist

### Before Deploying Any New Feature

- [ ] Run database migrations
- [ ] Update API documentation
- [ ] Run all tests (unit + integration + E2E)
- [ ] Performance testing completed
- [ ] Security review done
- [ ] Backup database
- [ ] Update environment variables if needed
- [ ] Deploy to staging first
- [ ] Smoke test on staging
- [ ] Monitor logs during deployment
- [ ] Have rollback plan ready

---

## Performance Optimization Tips

### 1. Database Query Optimization

```go
// Bad: N+1 query problem
for _, student := range students {
    grades, _ := repo.GetGrades(student.ID)
    // process grades
}

// Good: Use joins or batch queries
grades, _ := repo.GetGradesForStudents(studentIDs)
gradesMap := groupByStudentID(grades)
for _, student := range students {
    studentGrades := gradesMap[student.ID]
    // process grades
}
```

### 2. Caching Strategy

```go
// Cache frequently accessed data
func (s *service) GetCourse(ctx context.Context, courseID int) (*Course, error) {
    cacheKey := fmt.Sprintf("course:%d", courseID)

    // Try cache first
    if cached, err := s.cache.Get(cacheKey); err == nil {
        return cached.(*Course), nil
    }

    // Fetch from database
    course, err := s.repo.GetCourse(ctx, courseID)
    if err != nil {
        return nil, err
    }

    // Store in cache
    s.cache.Set(cacheKey, course, 15*time.Minute)

    return course, nil
}
```

### 3. API Response Pagination

```go
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Page       int         `json:"page"`
    PageSize   int         `json:"pageSize"`
    TotalItems int         `json:"totalItems"`
    TotalPages int         `json:"totalPages"`
}

// Always paginate list endpoints
GET /api/students?page=1&pageSize=20
GET /api/courses?page=1&pageSize=50
```

---

## Common Pitfalls to Avoid

### 1. Security Issues

```go
// ❌ Bad: SQL injection vulnerability
query := fmt.Sprintf("SELECT * FROM students WHERE name = '%s'", userInput)

// ✅ Good: Use parameterized queries
query := "SELECT * FROM students WHERE name = $1"
rows, err := db.Query(ctx, query, userInput)
```

### 2. Missing Authorization Checks

```go
// ❌ Bad: Anyone can access
func GetGrades(c echo.Context) error {
    studentID, _ := strconv.Atoi(c.Param("studentID"))
    grades := getGrades(studentID)
    return c.JSON(200, grades)
}

// ✅ Good: Check authorization
func GetGrades(c echo.Context) error {
    studentID, _ := strconv.Atoi(c.Param("studentID"))
    currentUserID := getUserID(c)

    // Students can only see their own grades
    if !isAdmin(currentUserID) && currentUserID != studentID {
        return c.JSON(403, "Forbidden")
    }

    grades := getGrades(studentID)
    return c.JSON(200, grades)
}
```

### 3. Not Handling Errors Properly

```go
// ❌ Bad: Swallow errors
result, _ := someOperation()

// ✅ Good: Handle errors
result, err := someOperation()
if err != nil {
    log.Error("Operation failed", "error", err)
    return fmt.Errorf("failed to perform operation: %w", err)
}
```

---

## Next Steps

1. **Review and Prioritize**: Discuss with stakeholders which features to implement first
2. **Resource Allocation**: Assign developers to each feature
3. **Set Milestones**: Create sprint plans with deliverable dates
4. **Start with Quick Wins**: Build momentum with easy features
5. **Iterative Development**: Release features incrementally
6. **Gather Feedback**: Collect user feedback after each release
7. **Iterate and Improve**: Refine based on usage patterns

---

## Need Help?

- Check existing implementations in the codebase
- Review test files for usage examples
- Refer to MISSING_FEATURES_ANALYSIS.md for detailed specifications
- Follow existing code patterns and conventions

---

*This guide provides a practical roadmap for implementing missing features in the edduhub platform.*
