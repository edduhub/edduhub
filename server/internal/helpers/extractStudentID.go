package helpers

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func ExtractStudentID(c echo.Context) (int, error) {
	studentID := c.Get("student_id")
	if studentID == nil {
		return 0, Error(c, "student authentication required - LoadStudentProfile middleware must be applied", 401)
	}

	studentIDInt, ok := studentID.(int)
	if !ok {
		return 0, Error(c, "invalid student_id format in context", 500)
	}

	return studentIDInt, nil
}

func GetIDFromParam(c echo.Context, paramName string) (int, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, Error(c, "missing path parameter", 400)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, Error(c, "Invalid path parameter", 400)
	}
	return id, nil
}
