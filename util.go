package jsonapi

import "strings"

func sortDirection(field string) string {
	if field[0] == '-' && len(field) > 1 {
		return "DESC"
	} else {
		return "ASC"
	}
}

func processSortings(kind string, sortings string, dest map[string][][]string) {
	s := strings.Split(sortings, ",")
	dest[kind] = make([][]string, len(s))
	for i, v := range s {
		dest[kind][i] = make([]string, 2)
		dest[kind][i][0] = v
		dest[kind][i][1] = sortDirection(v)
	}
}

func processFields(kind, fields string, dest map[string][]string) {
	f := strings.Split(fields, ",")
	dest[kind] = make([]string, len(f))
	for i, v := range f {
		dest[kind][i] = v
	}
}
