package output

import (
	"fmt"
	"io"
	"os"

	"github.com/lissto-dev/cli/pkg/client"
)

const notAvailable = "N/A"

// PrintImagePreview prints a preview of resolved images in table format
func PrintImagePreview(w io.Writer, images []client.DetailedImageResolutionInfo, exposed []client.ExposedServiceInfo) {
	// Create URL map for quick lookup
	urlMap := make(map[string]string)
	for _, exp := range exposed {
		urlMap[exp.Service] = exp.URL
	}

	// Print header
	fmt.Fprintln(w, "\n🔍 Image Preview:")
	fmt.Fprintln(w, "")

	headers := []string{"SERVICE", "STATUS", "IMAGE", "URL"}
	var rows [][]string

	for _, img := range images {
		status := "✅ Resolved"
		image := img.Image
		if image == "" {
			image = img.Digest
		}

		// Check if image is missing
		if img.Digest == "" || img.Digest == notAvailable {
			status = "❌ Missing"

			// Show what was attempted - check candidates for image URLs
			if len(img.Candidates) > 0 {
				// Show the first candidate that was tried
				image = img.Candidates[0].ImageURL
			} else if image == "" || image == notAvailable {
				// Fallback: try to construct from registry/imageName
				if img.Registry != "" && img.ImageName != "" {
					image = fmt.Sprintf("%s/%s", img.Registry, img.ImageName)
				} else if img.ImageName != "" {
					image = img.ImageName
				} else {
					image = "(no image specified)"
				}
			}
		}

		// Get URL if exposed
		url := ""
		if img.Exposed && img.URL != "" {
			url = fmt.Sprintf("https://%s", img.URL)
		} else if exposedURL, ok := urlMap[img.Service]; ok {
			url = fmt.Sprintf("https://%s", exposedURL)
		}

		rows = append(rows, []string{img.Service, status, image, url})
	}

	PrintTable(w, headers, rows)
	fmt.Fprintln(w, "")
}

// PrintImagePreviewJSON prints image preview in JSON format
func PrintImagePreviewJSON(w io.Writer, response *client.PrepareStackResponse) error {
	return PrintJSON(w, response)
}

// PrintImagePreviewYAML prints image preview in YAML format
func PrintImagePreviewYAML(w io.Writer, response *client.PrepareStackResponse) error {
	return PrintYAML(w, response)
}

// PrintImagePreviewWithFormat prints preview in the specified format
func PrintImagePreviewWithFormat(format string, response *client.PrepareStackResponse) error {
	switch format {
	case "json":
		return PrintImagePreviewJSON(os.Stdout, response)
	case "yaml":
		return PrintImagePreviewYAML(os.Stdout, response)
	default:
		PrintImagePreview(os.Stdout, response.Images, response.Exposed)
		return nil
	}
}

// HasMissingImages checks if any images are missing
func HasMissingImages(images []client.DetailedImageResolutionInfo) bool {
	for _, img := range images {
		if img.Digest == "" || img.Digest == "N/A" {
			return true
		}
	}
	return false
}
