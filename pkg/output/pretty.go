package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lissto-dev/cli/pkg/k8s"
)

// PrettyPrinter handles pretty formatted output
type PrettyPrinter struct {
	writer io.Writer
}

// NewPrettyPrinter creates a new pretty printer
func NewPrettyPrinter(w io.Writer) *PrettyPrinter {
	return &PrettyPrinter{writer: w}
}

// PrintHeader prints a section header
func (p *PrettyPrinter) PrintHeader(text string) {
	separator := strings.Repeat("━", 50)
	fmt.Fprintf(p.writer, "\n%s\n%s\n", text, separator)
}

// PrintField prints a labeled field
func (p *PrettyPrinter) PrintField(label, value string) {
	fmt.Fprintf(p.writer, "%s: %s\n", label, value)
}

// PrintSubSection prints a subsection with indentation
func (p *PrettyPrinter) PrintSubSection(emoji, title string) {
	fmt.Fprintf(p.writer, "\n  %s %s\n", emoji, title)
}

// PrintIndentedLine prints an indented line
func (p *PrettyPrinter) PrintIndentedLine(indent int, text string) {
	spaces := strings.Repeat(" ", indent*2)
	fmt.Fprintf(p.writer, "%s%s\n", spaces, text)
}

// PrintBullet prints a bullet point
func (p *PrettyPrinter) PrintBullet(indent int, text string) {
	spaces := strings.Repeat(" ", indent*2)
	fmt.Fprintf(p.writer, "%s• %s\n", spaces, text)
}

// PrintDivider prints a visual divider
func (p *PrettyPrinter) PrintDivider() {
	fmt.Fprintf(p.writer, "\n%s\n", strings.Repeat("─", 50))
}

// PrintNewline prints a newline
func (p *PrettyPrinter) PrintNewline() {
	fmt.Fprintln(p.writer)
}

// FormatTimestamp formats a timestamp into a human-readable format with "ago" suffix
func FormatTimestamp(t time.Time) (string, string) {
	formatted := t.UTC().Format("2006-01-02 15:04 MST")
	
	// Calculate time ago
	now := time.Now().UTC()
	diff := now.Sub(t)
	
	var timeAgo string
	seconds := int(diff.Seconds())
	if seconds < 60 {
		timeAgo = fmt.Sprintf("%ds ago", seconds)
	} else if seconds < 3600 {
		minutes := seconds / 60
		timeAgo = fmt.Sprintf("%dm ago", minutes)
	} else if seconds < 86400 {
		hours := seconds / 3600
		timeAgo = fmt.Sprintf("%dh ago", hours)
	} else {
		days := seconds / 86400
		timeAgo = fmt.Sprintf("%dd ago", days)
	}
	
	return formatted, timeAgo
}

// ExtractBlueprintAge extracts the timestamp from blueprint ID and calculates age
// ID format: scope/YYYYMMDD-HHMMSS-hash
func ExtractBlueprintAge(id string) string {
	// Split by / to get the name part
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "unknown"
	}

	// Extract timestamp from name (format: YYYYMMDD-HHMMSS-hash)
	nameParts := strings.Split(parts[1], "-")
	if len(nameParts) < 2 {
		return "unknown"
	}

	// Parse YYYYMMDD-HHMMSS
	timestampStr := nameParts[0] + nameParts[1]
	timestamp, err := time.Parse("20060102150405", timestampStr)
	if err != nil {
		return "unknown"
	}

	// Calculate and format age using shared k8s.FormatAge function
	duration := time.Since(timestamp)
	return k8s.FormatAge(duration)
}
