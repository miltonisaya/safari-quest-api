package pagination

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const defaultPage = 1
const defaultPerPage = 15
const maxPerPage = 100

type Params struct {
	Page    int
	PerPage int
	Search  string
	SortBy  string
	SortDir string
}

type Meta struct {
	Total    int64  `json:"total"`
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
	LastPage int    `json:"last_page"`
	Search   string `json:"search,omitempty"`
	SortBy   string `json:"sort_by"`
	SortDir  string `json:"sort_dir"`
}

type Result[T any] struct {
	Items []T  `json:"items"`
	Meta  Meta `json:"meta"`
}

// FromContext reads all query params with safe defaults and validation.
func FromContext(c *gin.Context) Params {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = defaultPage
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "15"))
	if err != nil || perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	sortDir := strings.ToLower(c.DefaultQuery("sort_dir", "asc"))
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "asc"
	}

	return Params{
		Page:    page,
		PerPage: perPage,
		Search:  strings.TrimSpace(c.Query("search")),
		SortBy:  strings.TrimSpace(c.Query("sort_by")),
		SortDir: sortDir,
	}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// OrderClause returns a safe ORDER BY expression. allowed maps accepted API
// sort keys to their actual DB column names, preventing SQL injection.
// fallback is used when sort_by is absent or not in allowed.
func (p Params) OrderClause(allowed map[string]string, fallback string) string {
	col, ok := allowed[p.SortBy]
	if !ok {
		col = fallback
	}
	return col + " " + p.SortDir
}

func NewMeta(total int64, params Params) Meta {
	lastPage := int(math.Ceil(float64(total) / float64(params.PerPage)))
	if lastPage < 1 {
		lastPage = 1
	}
	return Meta{
		Total:    total,
		Page:     params.Page,
		PerPage:  params.PerPage,
		LastPage: lastPage,
		Search:   params.Search,
		SortBy:   params.SortBy,
		SortDir:  params.SortDir,
	}
}
