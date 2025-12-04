package store

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=20"`
	Offset int    `json:"offset" validate:"gte=0"`
	Sort   string `json:"sort" validate:"oneof=asc desc"`
}

func (q Pagination) ParsePagination(r *http.Request) (Pagination, error) {

	queryParams := r.URL.Query()

	limit := queryParams.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return q, nil
		}
		q.Limit = l
	}
	offset := queryParams.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return q, nil
		}
		q.Offset = o
	}
	sort := queryParams.Get("sort")
	if sort != "" {
		q.Sort = sort
	}

	return q, nil
}
