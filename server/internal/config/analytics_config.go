package config

import (
	"os"
	"strconv"
)

type AnalyticsConfig struct {
	RiskWeightGradeVeryLow       float64
	RiskWeightGradeLow           float64
	RiskWeightGradeMedium        float64
	RiskWeightAttendanceVeryLow  float64
	RiskWeightAttendanceLow      float64
	RiskWeightAttendanceMedium   float64
	RiskWeightNoRecentGrades     float64
	RiskWeightFewRecentGrades    float64
	RiskWeightNoRecentAttendance float64
	RiskLevelHighThreshold       float64
	RiskLevelLowThreshold        float64
	RiskMinScore                 float64
	RiskMaxScore                 float64
}

func LoadAnalyticsConfig() *AnalyticsConfig {
	return &AnalyticsConfig{
		RiskWeightGradeVeryLow:       getEnvFloat("ANALYTICS_RISK_GRADE_VERY_LOW_WEIGHT", 0.45),
		RiskWeightGradeLow:           getEnvFloat("ANALYTICS_RISK_GRADE_LOW_WEIGHT", 0.30),
		RiskWeightGradeMedium:        getEnvFloat("ANALYTICS_RISK_GRADE_MEDIUM_WEIGHT", 0.15),
		RiskWeightAttendanceVeryLow:  getEnvFloat("ANALYTICS_RISK_ATTENDANCE_VERY_LOW_WEIGHT", 0.35),
		RiskWeightAttendanceLow:      getEnvFloat("ANALYTICS_RISK_ATTENDANCE_LOW_WEIGHT", 0.25),
		RiskWeightAttendanceMedium:   getEnvFloat("ANALYTICS_RISK_ATTENDANCE_MEDIUM_WEIGHT", 0.10),
		RiskWeightNoRecentGrades:     getEnvFloat("ANALYTICS_RISK_NO_RECENT_GRADES_WEIGHT", 0.20),
		RiskWeightFewRecentGrades:    getEnvFloat("ANALYTICS_RISK_FEW_RECENT_GRADES_WEIGHT", 0.10),
		RiskWeightNoRecentAttendance: getEnvFloat("ANALYTICS_RISK_NO_RECENT_ATTENDANCE_WEIGHT", 0.10),
		RiskLevelHighThreshold:       getEnvFloat("ANALYTICS_RISK_HIGH_THRESHOLD", 0.75),
		RiskLevelLowThreshold:        getEnvFloat("ANALYTICS_RISK_LOW_THRESHOLD", 0.45),
		RiskMinScore:                 getEnvFloat("ANALYTICS_RISK_MIN_SCORE", 0.05),
		RiskMaxScore:                 getEnvFloat("ANALYTICS_RISK_MAX_SCORE", 0.99),
	}
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}
