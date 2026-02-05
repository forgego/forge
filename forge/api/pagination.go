package api

import (
	"net/http"
	"strconv"

	forgehttp "github.com/forgego/forge/server"
)

// Pagination provides pagination functionality for API responses
type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	TotalCount int  `json:"total_count"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_previous"`
	NextPage   *int `json:"next_page,omitempty"`
	PrevPage   *int `json:"previous_page,omitempty"`
}

// PaginatedResponse wraps data with pagination info
type PaginatedResponse struct {
	Count    int         `json:"count"`
	Next     *string     `json:"next,omitempty"`
	Previous *string     `json:"previous,omitempty"`
	Results  interface{} `json:"results"`
}

// ParsePaginationParams extracts pagination parameters from an HTTP request.
// It returns page number, page size, and calculated offset.
// Invalid values are replaced with defaults.
func ParsePaginationParams(r *http.Request, defaultPageSize int) (page, pageSize, offset int) {
	page = forgehttp.GetQueryInt(r, "page", 1)
	pageSize = forgehttp.GetQueryInt(r, "page_size", defaultPageSize)

	// Validate page
	if page < 1 {
		page = 1
	}

	// Validate page size
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > 100 {
		pageSize = 100 // Max page size
	}

	offset = (page - 1) * pageSize
	return page, pageSize, offset
}

// GetPaginationParams extracts pagination parameters from request.
//
// Deprecated: Use ParsePaginationParams() for clarity (parsing vs getting).
// GetPaginationParams will be removed in v3.0.
// Migration:
//   // Old
//   page, size, offset := api.GetPaginationParams(r, 20)
//   // New
//   page, size, offset := api.ParsePaginationParams(r, 20)
func GetPaginationParams(r *http.Request, defaultPageSize int) (page, pageSize, offset int) {
	return ParsePaginationParams(r, defaultPageSize)
}

// NewPagination creates a new Pagination instance
func NewPagination(page, pageSize, totalCount int) *Pagination {
	totalPages := (totalCount + pageSize - 1) / pageSize // Ceiling division

	var nextPage, prevPage *int
	hasNext := page < totalPages
	hasPrev := page > 1

	if hasNext {
		next := page + 1
		nextPage = &next
	}
	if hasPrev {
		prev := page - 1
		prevPage = &prev
	}

	return &Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		NextPage:   nextPage,
		PrevPage:   prevPage,
	}
}

// BuildPaginatedResponse creates a paginated response
func BuildPaginatedResponse(r *http.Request, results interface{}, totalCount, page, pageSize int) *PaginatedResponse {
	baseURL := r.URL.Scheme + "://" + r.URL.Host + r.URL.Path

	var next, previous *string
	pagination := NewPagination(page, pageSize, totalCount)

	if pagination.HasNext {
		nextURL := baseURL + "?page=" + strconv.Itoa(*pagination.NextPage) + "&page_size=" + strconv.Itoa(pageSize)
		next = &nextURL
	}
	if pagination.HasPrev {
		prevURL := baseURL + "?page=" + strconv.Itoa(*pagination.PrevPage) + "&page_size=" + strconv.Itoa(pageSize)
		previous = &prevURL
	}

	return &PaginatedResponse{
		Count:    totalCount,
		Next:     next,
		Previous: previous,
		Results:  results,
	}
}

// SendPaginatedResponse sends a paginated JSON response
func SendPaginatedResponse(w http.ResponseWriter, r *http.Request, results interface{}, totalCount, page, pageSize int) error {
	response := BuildPaginatedResponse(r, results, totalCount, page, pageSize)
	return forgehttp.SendJSON(w, http.StatusOK, response)
}
