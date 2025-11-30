package output

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
)

func Gray(s string) string {
	return ColorGray + s + ColorReset
}

func Yellow(s string) string {
	return ColorYellow + s + ColorReset
}

func Green(s string) string {
	return ColorGreen + s + ColorReset
}

func Bold(s string) string {
	return ColorBold + s + ColorReset
}

func GreenCheck() string {
	return Green("✓")
}
