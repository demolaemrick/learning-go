package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Pagination struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
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

	tags := queryParams.Get("tags")
	if tags != "" {
		q.Tags = strings.Split(tags, ",")
	} else {
		q.Tags = []string{}
	}

	search := queryParams.Get("search")
	if search != "" {
		q.Search = search
	}

	since := queryParams.Get("since")
	if since != "" {
		q.Since = parseTime(since)
	}

	until := queryParams.Get("until")
	if until != "" {
		q.Until = parseTime(until)
	}

	return q, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}

	return t.Format(time.DateTime)
}
