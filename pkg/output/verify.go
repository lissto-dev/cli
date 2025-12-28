package output

import (
	"io"
	"os"
	"text/template"

	apicompose "github.com/lissto-dev/api/pkg/compose"
)

const verifyTemplate = `{{if .Valid -}}
{{if .Warnings -}}
⚠️  Compose file is valid but warnings were found
{{if not .Verbose -}}
Found {{.WarningCount}} warning(s). Run with --verbose for details.
{{end -}}
{{else -}}
✅ Compose file is valid
{{end -}}
{{if .Verbose -}}
{{if .Metadata.Title}}
Title: {{.Metadata.Title}}
{{end -}}
{{if .Metadata.Services.Services}}
📦 Services:
{{range .Metadata.Services.Services}}  - {{.}}
{{end -}}
{{end -}}
{{if .Metadata.Services.Infra}}
🗄️ Infrastructure:
{{range .Metadata.Services.Infra}}  - {{.}}
{{end -}}
{{end -}}
{{if .Metadata.Volumes}}
💾 Volumes:
{{range .Metadata.Volumes}}  - {{.}}
{{end -}}
{{end -}}
{{if .Metadata.Networks}}
🌐 Networks:
{{range .Metadata.Networks}}  - {{.}}
{{end -}}
{{end -}}
{{if .Warnings}}
⚠️  Warnings ({{.WarningCount}}):
{{range .Warnings}}  - {{.}}
{{end -}}
{{end -}}
{{end -}}
{{else -}}
❌ Compose file is invalid
{{if .Errors}}
Errors:
{{range .Errors}}  - {{.}}
{{end -}}
{{end -}}
{{if .Warnings}}
Warnings:
{{range .Warnings}}  - {{.}}
{{end -}}
{{end -}}
{{end -}}
`

// VerifyTemplateData contains the data for verification output templates
type VerifyTemplateData struct {
	Valid        bool
	Verbose      bool
	Metadata     *apicompose.BlueprintMetadata
	Errors       []string
	Warnings     []string
	WarningCount int
}

// PrintVerificationResult renders the verification result using templates
// and writes it to the provided writer
func PrintVerificationResult(result *VerifyTemplateData, writer io.Writer) error {
	tmpl, err := template.New("verify").Parse(verifyTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(writer, result)
}

// PrintVerificationResultToStdout is a helper function that writes
// the verification result to stdout
func PrintVerificationResultToStdout(result *VerifyTemplateData) error {
	return PrintVerificationResult(result, os.Stdout)
}
