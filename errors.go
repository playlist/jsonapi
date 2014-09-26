package jsonapi

import "errors"

var ErrMissingFetchIDs = errors.New("missing FetchIDs func")
var ErrMismatchedFieldsParams = errors.New("mismatched fields param, got fields and fields[kind] - use one format or the other")
var ErrMismatchedSortParams = errors.New("mismatched sort param, got sort and sort[kind] - use one format or the other")
