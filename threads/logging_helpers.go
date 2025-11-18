package threads

import "fmt"

func coloredMetric(label string, value interface{}, accentColor, baseColor string) string {
	if accentColor == "" {
		return fmt.Sprintf("%s=%v", label, value)
	}
	return fmt.Sprintf("%s%s=%v%s", accentColor, label, value, baseColor)
}
