package jsonapi

import (
	"regexp"
	"strings"
)

var fieldsRegex = regexp.MustCompile(`^fields\[([^\]]+)\]$`)
var sortRegex = regexp.MustCompile(`^sort\[([^\]]+)\]$`)

type Query struct {
	FetchIDs       func() ([]string, error)   // returns an array of IDs for the primary collection
	DefaultFields  func(kind string) []string // (optional) returns the default fields to fetch for a kind
	FilterAllowed  func(filter string) bool   // (optional) returns true if the filter is a valid field
	FetchResources func(
		kind string,
		ids []string,
		fields []string,
		filters map[string][]string,
		sorting [][]string) []interface{} // returns resource objects for a kind and a set of IDs
	ResolveLink      func(link string, r *Response) ResourceLink                                       // returns a resource link for a requested link, provided with the in-progress response
	ResolveLinkedIDs func(link string, resources map[string][]interface{}) (kind string, ids []string) // returns a set of IDs given the link and the available resources
	kind             string
	primaryIDs       []string

	includes []string
	sortings map[string][][]string
	fields   map[string][]string
	filters  map[string][]string
}

func NewQuery(kind string) Query {
	return Query{
		kind: kind,
	}
}

// Dump returns a debug dump of the query internals
func (q *Query) Dump() map[string]interface{} {
	return map[string]interface{}{
		"kind":       q.kind,
		"primaryIDs": q.primaryIDs,
		"includes":   q.includes,
		"sortings":   q.sortings,
		"fields":     q.fields,
		"filters":    q.filters,
	}
}

func (q *Query) Parse(params map[string][]string) error {
	// return an error if the FetchIDs func is not set
	if q.FetchIDs == nil {
		return ErrMissingFetchIDs
	}

	// fetch the primary resource IDs
	var err error
	q.primaryIDs, err = q.FetchIDs()
	if err != nil {
		return err
	}

	// parse the include param into the includes field
	if inc, ok := params["include"]; ok {
		inc := strings.Split(inc[0], ",")
		q.includes = make([]string, len(inc))
		for i, v := range inc {
			q.includes[i] = v
		}
		delete(params, "include")
	}

	// initialize the fields map
	q.fields = make(map[string][]string)
	fieldsProcessed := false

	// parse the root fields param
	if f, ok := params["fields"]; ok {
		fieldsProcessed = true
		processFields(q.kind, f[0], q.fields)
		delete(params, "fields")
	}

	// initialize the sortings map
	q.sortings = make(map[string][][]string)
	sortProcessed := false

	// parse the root sort param
	if s, ok := params["sort"]; ok {
		sortProcessed = true
		processSortings(q.kind, s[0], q.sortings)
		delete(params, "sort")
	}

	// initialize the filters array
	q.filters = make(map[string][]string)

	// parse the remaining params
	for k, v := range params {
		if fieldsRegex.MatchString(k) {
			if fieldsProcessed {
				return ErrMismatchedFieldsParams
			} else {
				r := fieldsRegex.FindStringSubmatch(k)
				processFields(r[1], v[0], q.fields)
			}
		} else if sortRegex.MatchString(k) {
			if sortProcessed {
				return ErrMismatchedSortParams
			} else {
				r := sortRegex.FindStringSubmatch(k)
				processSortings(r[1], v[0], q.sortings)
			}
		} else if q.FilterAllowed == nil || q.FilterAllowed(k) {
			q.filters[k] = v
		}
	}

	// add the default fields
	if q.DefaultFields != nil {
		for k, v := range q.sortings {
			q.sortings[k] = append(v, q.DefaultFields(k))
		}
	}

	return nil
}

func (q *Query) Execute() (*Response, error) {
	if q.FetchResources == nil {
		return nil, ErrMissingFetchResources
	}

	r := NewResponse(q.kind)

	if len(q.includes) > 0 && q.ResolveLink != nil {
		for _, v := range q.includes {
			r.Links[v] = q.ResolveLink(v, r)
		}
	}

	if len(q.primaryIDs) > 0 {
		r.Resources[q.kind] = q.FetchResources(q.kind, q.primaryIDs, q.fields[q.kind], q.filters, q.sortings[q.kind])
	}

	if len(q.includes) > 0 && q.ResolveLinkedIDs != nil {
		var links [][]string

		dotRegex := regexp.MustCompile(`\.`)
		for _, v := range q.includes {
			count := len(dotRegex.FindAllString(v, -1))
			if count+1 > len(links) {
				links = append(links, make([]string, count-len(links)))
			}

			links[count] = append(links[count], v)
		}

		for _, l := range links {
			for _, v := range l {
				kind, ids := q.ResolveLinkedIDs(v, r.Resources)
				r.Resources[kind] = append(r.Resources[kind], q.FetchResources(kind, ids, q.fields[kind], nil, q.sortings[kind]))
			}
		}
	}

	return r, nil
}
