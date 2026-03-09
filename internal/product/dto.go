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
	IsActive    bool   `json:"isActive"`
}

type DTOListFilters struct {
	Limit    *int
	Offset   *int
	SortBy   *string
	SortDesc *bool

	Category *string
	MinPrice *int64
	MaxPrice *int64
	IsActive *bool
}

type DTOProduct struct {
	ID          string
	Name        string
	Description string
	Category    string
	Price       int64
	Available   int
}
