package authority

import (
	"net/http"
	"strings"

	"github.com/jinzhu/inflection"
)

// DeriveCode builds the authority code from a route method and its full path
// pattern (e.g. "/api/v1/roles/:uuid"). The code follows the convention
// {resource}_{verb} where resource is the singularised URL segment and verb
// is the lowercase HTTP method (get, post, put, delete).
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
	verb := resolveVerb(method)
	if verb == "" {
		return ""
	}

	return resource + "_" + verb
}

func resolveVerb(method string) string {
	switch method {
	case http.MethodGet:
		return "get"
	case http.MethodPost:
		return "post"
	case http.MethodPut, http.MethodPatch:
		return "put"
	case http.MethodDelete:
		return "delete"
	default:
		return ""
	}
}
