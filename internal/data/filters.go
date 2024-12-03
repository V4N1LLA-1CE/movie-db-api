package data

import "github.com/V4N1LLA-1CE/movie-db-api/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// check that page and page_size values are valid
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page <= 10_000_000, "page", "must be a max of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "must be maximum on 100")

	// check sort param matches a value in safelist
	v.Check(validator.PermittedValue(f.Sort, f.SortSafeList...), "sort", "must be a valid value")
}
