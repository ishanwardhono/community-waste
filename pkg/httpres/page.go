package httpres

import (
	"net/http"
	"strconv"
)

type Page struct {
	Page  int
	Limit int
}

func ParsePage(r *http.Request) Page {
	page := atoiOr(r.URL.Query().Get("page"), 1)
	limit := atoiOr(r.URL.Query().Get("limit"), 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return Page{Page: page, Limit: limit}
}

func (p Page) Offset() int { return (p.Page - 1) * p.Limit }

func (p Page) Meta(total int64) Meta {
	return Meta{Page: p.Page, Limit: p.Limit, Total: total}
}

func atoiOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
