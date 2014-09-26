package jsonapi

import "encoding/json"

type ResourceLink struct {
	Href string `json:"href"`
	Kind string `json:"type"`
}

type Response struct {
	Links     map[string]ResourceLink
	Resources map[string][]interface{}

	primaryKind string
}

func NewResponse(kind string) *Response {
	return &Response{
		Links:       make(map[string]ResourceLink),
		Resources:   make(map[string][]interface{}),
		primaryKind: kind,
	}
}

func (r *Response) MarshalJSON() ([]byte, error) {
	res := make(map[string]interface{})

	if r.Links != nil && len(r.Links) != 0 {
		l := make(map[string]ResourceLink)
		for k, v := range r.Links {
			l[k] = v
		}
		res["links"] = l
	}

	if r.Resources != nil && len(r.Resources) != 0 {
		if primary, ok := r.Resources[r.primaryKind]; ok {
			if len(primary) == 1 {
				res[r.primaryKind] = primary[0]
			} else {
				res[r.primaryKind] = primary
			}
			delete(r.Resources, r.primaryKind)
		}

		hasLinked := false
		l := make(map[string]interface{})
		for k, v := range r.Resources {
			hasLinked = true
			if len(v) == 1 {
				l[k] = v[0]
			} else {
				l[k] = v
			}
		}
		if hasLinked {
			res["linked"] = l
		}
	}

	return json.Marshal(res)
}
