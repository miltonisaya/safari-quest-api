package authority

import (
	"net/http"
	"strings"

	"github.com/jinzhu/inflection"
)

// DeriveCode builds the authority code from a route method and its full path
// pattern (e.g. "/api/v1/roles/:uuid"). The code follows the convention
// {resource}_{action} where resource is the singularised URL segment and action
// is resolved from the HTTP method and whether a path parameter is present.
//
// Returns an empty string if the code cannot be determined (e.g. unknown method).
func DeriveCode(method, fullPath string) string {
	if fullPath == "" {
		return ""
	}

	segments := strings.Split(strings.Trim(fullPath, "/"), "/")
	hasParam := len(segments) > 0 && strings.HasPrefix(segments[len(segments)-1], ":")

	resourceIdx := len(segments) - 1
	if hasParam {
		resourceIdx = len(segments) - 2
	}
	if resourceIdx < 0 {
		return ""
	}

	resource := inflection.Singular(segments[resourceIdx])
	action := resolveAction(method, hasParam)
	if action == "" {
		return ""
	}

	return resource + "_" + action
}

func resolveAction(method string, hasParam bool) string {
	switch method {
	case http.MethodGet:
		if hasParam {
			return "show"
		}
		return "index"
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return ""
	}
}
