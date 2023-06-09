package contracts

import "strconv"

type PaginatedRequest struct {
	Page int `json:"page" query:"page"`
	Size int `json:"size" query:"size"`
}

type PaginatedResponse[T any] struct {
	Page  int  `json:"page" validate:"min=0"`
	Size  int  `json:"size" validate:"min=0"`
	Total int  `json:"total"`
	Items []*T `json:"items"`
}

func (req *PaginatedRequest) ToQueryParams() map[string]string {
	params := make(map[string]string, 2)
	if req.Page > 0 {
		params["page"] = strconv.Itoa(req.Page)
	}
	if req.Size > 0 {
		params["size"] = strconv.Itoa(req.Size)
	}
	return params
}
