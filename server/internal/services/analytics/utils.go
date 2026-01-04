package analytics

import "math"

// PercentageToGPA converts a percentage score to a GPA on a 0.0-4.0 scale.
// Uses standard academic grade boundaries aligned with dashboard_handler.go.
// Exported to allow other packages to use the same calculation for consistency.
func PercentageToGPA(percentage float64) float64 {
	switch {
	case percentage >= 90:
		return 4.0 // A
	case percentage >= 85:
		return 3.7 // A-
	case percentage >= 80:
		return 3.3 // B+
	case percentage >= 75:
		return 3.0 // B
	case percentage >= 70:
		return 2.7 // B-
	case percentage >= 65:
		return 2.3 // C+
	case percentage >= 60:
		return 2.0 // C
	case percentage >= 55:
		return 1.7 // C-
	case percentage >= 50:
		return 1.0 // D
	default:
		return 0.0 // F
	}
}

// roundFloat rounds a float to the specified number of decimal places.
func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
