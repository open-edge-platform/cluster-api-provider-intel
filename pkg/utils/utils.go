// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/bnkamalesh/errors"
	"google.golang.org/grpc/codes"
)

const (
	baseURLIP = "0.0.0.0"
	localIP   = "127.0.0.1"
)

type Filter struct {
	Name  string
	Value string
}

type OrderBy struct {
	Name   string
	IsDesc bool
}

func StartHttpReadinessProbe(readinessProbe string, baseURL string, pattern string, uri string) {
	url := strings.ReplaceAll(baseURL, baseURLIP, localIP)
	http.HandleFunc(pattern, func(w http.ResponseWriter, _ *http.Request) {
		resp, err := http.Get(fmt.Sprintf("http://%s/%s", url, uri)) //nolint: noctx // FIXME: This is not the right way per linter issue. Check more
		// here https://github.com/sonatard/noctx to fix it. More effort and change in this PR and hence skipping this.
		if err != nil {
			log.Error().Msgf("readiness probe failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		defer CloseHttpClient(resp)
		if resp != nil {
			w.WriteHeader(resp.StatusCode)
		}
	},
	)
	server := &http.Server{
		Addr:              readinessProbe,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Error().Msgf("Start readiness probe failed: %v", err)
	}
}

// GrpcToHttpStatusCode coverts grpc status code to http status code
func GrpcToHttpStatusCode(grpcStatus codes.Code) int {
	var httpStatus int
	switch grpcStatus {
	case codes.OK:
		httpStatus = http.StatusOK
	case codes.Canceled:
		httpStatus = http.StatusRequestTimeout
	case codes.Unknown:
		httpStatus = http.StatusInternalServerError
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.DeadlineExceeded:
		httpStatus = http.StatusGatewayTimeout
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.ResourceExhausted:
		httpStatus = http.StatusTooManyRequests
	case codes.FailedPrecondition:
		httpStatus = http.StatusBadRequest
	case codes.Aborted:
		httpStatus = http.StatusConflict
	case codes.OutOfRange:
		httpStatus = http.StatusBadRequest
	case codes.Unimplemented:
		httpStatus = http.StatusNotImplemented
	case codes.Internal:
		httpStatus = http.StatusInternalServerError
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
	case codes.DataLoss:
		httpStatus = http.StatusInternalServerError
	default:
		httpStatus = http.StatusInternalServerError
	}
	return httpStatus
}

// HttpToGrpcStatusCode coverts http status code to grpc status code
func HttpToGrpcStatusCode(httpStatusCode int) codes.Code {
	var code codes.Code
	switch httpStatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		code = codes.OK
	case http.StatusBadRequest:
		code = codes.InvalidArgument
	case http.StatusUnauthorized:
		code = codes.Unauthenticated
	case http.StatusForbidden:
		code = codes.PermissionDenied
	case http.StatusNotFound:
		code = codes.NotFound
	case http.StatusMethodNotAllowed:
		code = codes.Unimplemented
	case http.StatusRequestTimeout:
		code = codes.DeadlineExceeded
	case http.StatusInternalServerError:
		code = codes.Unknown
	default:
		code = codes.Unknown
	}
	return code
}

// ParseFilter parses the given filter string and returns a list of Filter
// If any error is encountered, an nil Filter slice and non-nil error is returned
func ParseFilter(filterParameter string) ([]*Filter, error) {
	if filterParameter == "" {
		return nil, nil
	}
	// The string should contain '=' with one more space or tabs on either of it.
	normalizeEqualsRe := regexp.MustCompile("[ \t]*=[ \t]*")
	// Replace the matched pattern in regexp 'normalizeEqualsRe' with just '=' (basically the spaces and tabs are removed)
	normalizedFilterParameter := normalizeEqualsRe.ReplaceAllString(filterParameter, "=")

	// Now split the string with space as delimiter. Note that there could be a 'OR' predicate with on or
	// more space on either side of it
	// Consider this example 'f1=v1 OR f2=v2 OR f3=v3'.
	// After below step, the 'elements' contains ["f1=v1", "OR",  "f2=v2", "OR", "f3=v3"]
	elements := strings.Split(normalizedFilterParameter, " ")

	var filters []*Filter
	var currentFilter *Filter

	// Now parse each element and make a list of all 'name=value' filters
	for index, element := range elements {
		if strings.Contains(element, "=") { //nolint: gocritic // not easy to convert to switch-case.
			selectors := strings.Split(element, "=")
			if currentFilter != nil || len(selectors) != 2 || selectors[0] == "" || selectors[1] == "" {
				// Error condition - too many equals
				return nil, errors.Validationf("filter: invalid filter request: %s", elements)
			}
			currentFilter = &Filter{}
			// This is the start of a selector. Grab the name and the value
			currentFilter.Name = selectors[0]
			currentFilter.Value = selectors[1]
		} else if element == "OR" {
			if currentFilter == nil || index == len(elements)-1 {
				//  Error condition - OR with no other term
				return nil, errors.Validationf("filter: invalid filter request: %s", elements)
			}
			filters = append(filters, currentFilter)
			currentFilter = nil
			continue
		} else {
			if currentFilter == nil {
				// Error condition - missing an =
				return nil, errors.Validationf("filter: invalid filter request: %s", elements)
			}
			currentFilter.Value = currentFilter.Value + " " + element
		}
	}
	if currentFilter != nil {
		filters = append(filters, currentFilter)
	}

	return filters, nil
}

// findColumnName checks if the given attributeName is in the map of allowed columnNames
// Returns a valid columnName if found in map with nil error, else an "" column name and an error.
func findColumnName(attributeName string, columns map[string]string, operation string) (string, error) {
	columnName, ok := columns[attributeName]
	if !ok {
		return "", errors.Validationf("%s: no such attribute: %s", operation, attributeName)
	}
	if columnName == "" {
		return "", errors.Validationf("%s: cannot sort on attribute: %s", operation, attributeName)
	}
	return columnName, nil
}

// caseInsensitiveLike generates a case-insensitive SQL predicate for a given column name and the value to be searched on that column
// Below generates a predicate equivalent to this ==>>  'WHERE LOWER(col) LIKE LOWER(val);'
// The crux of the logic here is to convert the column values and the given value to lower case and then do the LIKE comparison.
func caseInsensitiveLike(col, val string) func(s *sql.Selector) {
	return func(s *sql.Selector) {
		s.Where(sql.P(func(b *sql.Builder) {
			b.WriteString("LOWER(").Ident(col).WriteByte(')').WriteOp(sql.OpLike).WriteString("LOWER(").Arg(val).WriteByte(')')
		}))
	}
}

// FilterPredicates creates SQL predicates based on provided Filter configuration.
func FilterPredicates(filters []*Filter, columns map[string]string) ([]func(s *sql.Selector), error) {
	var preds []func(s *sql.Selector)
	if len(filters) != 0 {
		for _, f := range filters {
			column, err := findColumnName(f.Name, columns, "filter")
			if err != nil {
				return nil, err
			}
			// The filter could contain a wild-card match or substring/full match.
			// The substring and full-match get same treatment to generate its respective predicate, because full-string is substring of itself
			//  Based on the type of match conveyed in the
			// 'value' field of the Filter configuration, generate appropriate SQL predicate.
			if strings.Contains(f.Value, "*") {
				// In SQL schematics, '%' is used to signify 'any match' unlike '*' in regular bash like convention.
				// So we need to replace '*' with '%' when converting to SQL commands.
				likeValue := strings.ReplaceAll(f.Value, "*", "%")
				// we need to support case-insensitive search, so generate such a predicate
				preds = append(preds, caseInsensitiveLike(column, likeValue))
			} else {
				// We need to support sub-string search and also be case-insensitive
				// Prefix and suffix with '%' to the value represent wildcard search, and generate a case-insensitive predicate
				preds = append(preds, caseInsensitiveLike(column, "%"+f.Value+"%"))
			}
		}
	}
	return preds, nil
}

// ParseOrderBy parses the incoming orderBy query string
// Below is a sample orderBy query specifying that the results should be sorted
// that name is ascending and create_time should be descending
//
//	/books?orderBy="name asc, create_time desc"
func ParseOrderBy(orderByParameter string) ([]*OrderBy, error) {
	if orderByParameter == "" {
		return nil, nil
	}
	// orderBy commands should be separated by ',' if there are more than one.
	// Split them by ',' delimiter.
	elements := strings.Split(orderByParameter, ",")
	var orderBys []*OrderBy
	for _, element := range elements {
		descending := false
		// Parse each orderBy command to extract the field name and the command (asc or desc)
		direction := strings.Split(strings.Trim(element, " "), " ")
		// Do some validations to ensure we have the right format and right command
		if len(direction) > 2 {
			return nil, errors.Validationf("invalid order by: %s", element)
		}
		if len(direction) == 2 {
			switch direction[1] {
			case "asc":
				descending = false
			case "desc":
				descending = true
			default:
				return nil, errors.Validationf("invalid order by: %s", element)
			}
		}
		orderBys = append(orderBys, &OrderBy{
			Name:   direction[0],
			IsDesc: descending,
		})
	}
	return orderBys, nil
}

// orderByDirection creates the SQL Ordering Term Option
func (o *OrderBy) orderByDirection() sql.OrderTermOption {
	orderMap := map[bool]sql.OrderTermOption{
		true:  sql.OrderDesc(),
		false: sql.OrderAsc(),
	}
	return orderMap[o.IsDesc]
}

// OrderByOptions generates SQL selector based on the Ordering configuration and Valid Columns on which ordering is allowed
func OrderByOptions(orderBys []*OrderBy, columns map[string]string) ([]func(selector *sql.Selector), error) {
	var options []func(s *sql.Selector)

	if len(orderBys) != 0 {
		for _, o := range orderBys {
			// If the passed field name is not in the valid list of allowed columns, return error
			columnName, err := findColumnName(o.Name, columns, "orderBy")
			if err != nil {
				return nil, err
			}
			// Generate the SQL Ordering Term Option
			orderTermOption := o.orderByDirection()
			options = append(options, sql.OrderByField(columnName, orderTermOption).ToFunc())
		}
	}
	return options, nil
}

// ComputePageRange computes the startIndex and endIndex based on the provided pageSize, offset and totalCount.
// It returns -1 for the endIndex if there are no items to paginate.
// This function is used during the handling of pageSize/Offset in a HTTP Query
func ComputePageRange(pageSize int32, offset int32, totalCount int) (int, int) {
	if offset < 0 || // Invalid offset
		pageSize < 0 || // Invalid pageSize
		totalCount <= 0 || // Invalid totalCount
		totalCount > math.MaxInt32 || // totalCount exceeds int32 range
		offset >= int32(totalCount) || // Offset out of bounds
		(pageSize == 0 && offset != 0) { // Invalid combination of pageSize and offset
		return 0, -1 // -1 to indicate that there are no items to paginate
	}

	startIndex := int(offset)
	endIndex := startIndex + int(pageSize) - 1
	if pageSize == 0 && offset == 0 || endIndex >= totalCount {
		endIndex = totalCount - 1
	}

	return startIndex, endIndex
}
