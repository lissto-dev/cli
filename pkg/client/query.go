package client

import (
	"fmt"
	"strings"
)

// buildQueryParams builds a query string from scope, env, and repository parameters
func buildQueryParams(scope, env, repository string) string {
	params := []string{}

	if scope != "" {
		params = append(params, fmt.Sprintf("scope=%s", scope))
	}
	if env != "" {
		params = append(params, fmt.Sprintf("env=%s", env))
	}
	if repository != "" {
		params = append(params, fmt.Sprintf("repository=%s", repository))
	}

	if len(params) > 0 {
		return "?" + strings.Join(params, "&")
	}
	return ""
}

// buildResourcePath builds a full resource path with query parameters
func buildResourcePath(base, id, scope, env, repository string) string {
	path := fmt.Sprintf("%s/%s", base, id)
	return path + buildQueryParams(scope, env, repository)
}
