package utils

import (
	"net/http"
	"strings"
)

func BuildSort(r *http.Request, allowed map[string]bool) string {
	sortParams := r.URL.Query()["sortBy"]

	var parts []string
	for _, p := range sortParams {
		s := strings.Split(p, ":")
		if len(s) != 2 { continue }

		field, order := s[0], s[1]

		if !allowed[field] { continue }
		if order != "asc" && order != "desc" { continue }

		parts = append(parts, field+" "+order)
	}

	return strings.Join(parts, ", ")
}
