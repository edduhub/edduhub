package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- DefaultRedisConfig ---

func TestDefaultRedisConfig(t *testing.T) {
	cfg := DefaultRedisConfig()

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "6379", cfg.Port)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, 0, cfg.DB)
	assert.Equal(t, "eduhub:", cfg.Prefix)
	assert.Equal(t, 10, cfg.PoolSize)
	assert.Equal(t, 2, cfg.MinIdleConns)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 5*time.Second, cfg.DialTimeout)
	assert.Equal(t, 3*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 3*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 4*time.Second, cfg.PoolTimeout)
}

// --- BuildStudentKey ---

func TestBuildStudentKey(t *testing.T) {
	t.Run("standard key", func(t *testing.T) {
		assert.Equal(t, "student:1:42", BuildStudentKey(1, 42))
	})

	t.Run("zero values", func(t *testing.T) {
		assert.Equal(t, "student:0:0", BuildStudentKey(0, 0))
	})

	t.Run("large values", func(t *testing.T) {
		assert.Equal(t, "student:99999:88888", BuildStudentKey(99999, 88888))
	})
}

// --- BuildCourseKey ---

func TestBuildCourseKey(t *testing.T) {
	t.Run("standard key", func(t *testing.T) {
		assert.Equal(t, "course:1:100", BuildCourseKey(1, 100))
	})

	t.Run("zero values", func(t *testing.T) {
		assert.Equal(t, "course:0:0", BuildCourseKey(0, 0))
	})
}

// --- BuildCourseListKey ---

func TestBuildCourseListKey(t *testing.T) {
	t.Run("standard key", func(t *testing.T) {
		assert.Equal(t, "course:list:1:p1:l20", BuildCourseListKey(1, 1, 20))
	})

	t.Run("different pagination", func(t *testing.T) {
		assert.Equal(t, "course:list:5:p3:l50", BuildCourseListKey(5, 3, 50))
	})
}

// --- BuildAttendanceKey ---

func TestBuildAttendanceKey(t *testing.T) {
	t.Run("standard key", func(t *testing.T) {
		assert.Equal(t, "attendance:1:42:100", BuildAttendanceKey(1, 42, 100))
	})

	t.Run("zero values", func(t *testing.T) {
		assert.Equal(t, "attendance:0:0:0", BuildAttendanceKey(0, 0, 0))
	})
}

// --- BuildCalendarKey ---

func TestBuildCalendarKey(t *testing.T) {
	t.Run("standard key", func(t *testing.T) {
		assert.Equal(t, "calendar:1:2024-01-01:2024-01-31", BuildCalendarKey(1, "2024-01-01", "2024-01-31"))
	})

	t.Run("empty dates", func(t *testing.T) {
		assert.Equal(t, "calendar:1::", BuildCalendarKey(1, "", ""))
	})
}

// --- BuildSessionKey ---

func TestBuildSessionKey(t *testing.T) {
	t.Run("standard key", func(t *testing.T) {
		assert.Equal(t, "session:abc-123-def", BuildSessionKey("abc-123-def"))
	})

	t.Run("empty session", func(t *testing.T) {
		assert.Equal(t, "session:", BuildSessionKey(""))
	})
}

// --- CacheKey ---

func TestCacheKey(t *testing.T) {
	t.Run("single component", func(t *testing.T) {
		key := CacheKey("student")
		assert.NotEmpty(t, key)
	})

	t.Run("multiple components", func(t *testing.T) {
		key := CacheKey("student", 1, "grades")
		assert.NotEmpty(t, key)
	})

	t.Run("different components produce different keys", func(t *testing.T) {
		key1 := CacheKey("student", 1)
		key2 := CacheKey("student", 2)
		assert.NotEqual(t, key1, key2)
	})

	t.Run("same components produce same key", func(t *testing.T) {
		key1 := CacheKey("course", 5, "list")
		key2 := CacheKey("course", 5, "list")
		assert.Equal(t, key1, key2)
	})
}

// --- TTL constants ---

func TestTTLConstants(t *testing.T) {
	assert.Equal(t, 5*time.Minute, TTLShort)
	assert.Equal(t, 30*time.Minute, TTLMedium)
	assert.Equal(t, 2*time.Hour, TTLLong)
	assert.Equal(t, 24*time.Hour, TTLDay)
}

// --- Prefix constants ---

func TestPrefixConstants(t *testing.T) {
	assert.Equal(t, "student:", PrefixStudent)
	assert.Equal(t, "course:", PrefixCourse)
	assert.Equal(t, "lecture:", PrefixLecture)
	assert.Equal(t, "attendance:", PrefixAttendance)
	assert.Equal(t, "grade:", PrefixGrade)
	assert.Equal(t, "calendar:", PrefixCalendar)
	assert.Equal(t, "department:", PrefixDepartment)
	assert.Equal(t, "college:", PrefixCollege)
	assert.Equal(t, "user:", PrefixUser)
	assert.Equal(t, "session:", PrefixSession)
}

// --- buildKey ---

func TestRedisCache_BuildKey(t *testing.T) {
	t.Run("with prefix", func(t *testing.T) {
		rc := &RedisCache{prefix: "eduhub:"}
		assert.Equal(t, "eduhub:mykey", rc.buildKey("mykey"))
	})

	t.Run("without prefix", func(t *testing.T) {
		rc := &RedisCache{prefix: ""}
		assert.Equal(t, "mykey", rc.buildKey("mykey"))
	})

	t.Run("empty key with prefix", func(t *testing.T) {
		rc := &RedisCache{prefix: "app:"}
		assert.Equal(t, "app:", rc.buildKey(""))
	})
}
