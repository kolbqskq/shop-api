package product

type DTOUpdateProduct struct {
	Name        *string
	Description *string
	Category    *string
	Price       *int64
	Stock       *int
	IsActive    *bool
}

type DTOCreateProduct struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Price       int64  `json:"price"`
	Stock       int    `json:"stock"`
	IsActive    bool   `json:"is_active"`
}

type DTOListFilters struct {
	Limit    *int    `form:"limit"`
	Offset   *int    `form:"offset"`
	SortBy   *string `form:"sort_by"`
	SortDesc *bool   `form:"sort_desc"`

	Category *string `form:"category"`
	MinPrice *int64  `form:"min_price"`
	MaxPrice *int64  `form:"max_price"`
	IsActive *bool   `form:"is_active"`
}

type DTOProduct struct {
	ID          string
	Name        string
	Description string
	Category    string
	Price       int64
	Available   int
}
