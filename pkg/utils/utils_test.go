// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestStartHttpReadinessProbe(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	go StartHttpReadinessProbe(":8080", ts.URL, "/healthz", "v1/templates")

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", resp.StatusCode)
	}
}

func checkFilters(t *testing.T, filters []*Filter, wantedFieldList string, wantedValuesList string) {
	if len(filters) == 0 && wantedValuesList == "" {
		return
	}
	wantedFields := strings.Split(wantedFieldList, ",")
	wantedValues := strings.Split(wantedValuesList, ",")

	assert.Len(t, filters, len(wantedFields))
	for i := range wantedFields {
		assert.Equal(t, wantedFields[i], filters[i].Name)
		assert.Equal(t, wantedValues[i], filters[i].Value)
	}
}

func checkOrders(t *testing.T, orders []*OrderBy, wantedFieldList string, wantedOrderList []bool) {
	if len(orders) == 0 && len(wantedOrderList) == 0 {
		return
	}
	wantedFields := strings.Split(wantedFieldList, ",")

	assert.Len(t, orders, len(wantedFields))
	for i := range wantedFields {
		assert.Equal(t, wantedFields[i], orders[i].Name)
		assert.Equal(t, wantedOrderList[i], orders[i].IsDesc)
	}
}

func TestFilterPredicates(t *testing.T) {
	type args struct {
		filters []*Filter
		columns map[string]string
	}
	tests := map[string]struct {
		args             args
		wantNoOfSelector int
		wantErr          bool
	}{
		"success-one-wildcard": {
			args: args{
				filters: []*Filter{
					{
						Name:  "name",
						Value: "fo*",
					},
				},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfSelector: 1,
			wantErr:          false,
		},
		"success-multiple-wildcard": {
			args: args{
				filters: []*Filter{
					{
						Name:  "name",
						Value: "*o*",
					},
				},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfSelector: 1,
			wantErr:          false,
		},
		"success-substring": {
			args: args{
				filters: []*Filter{
					{
						Name:  "name",
						Value: "fo",
					},
				},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfSelector: 1,
			wantErr:          false,
		},
		"success-multiple-selectors": {
			args: args{
				filters: []*Filter{
					{
						Name:  "name",
						Value: "fo",
					},
					{
						Name:  "version",
						Value: "v1*",
					},
				},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfSelector: 2,
			wantErr:          false,
		},
		"failing-non-existent-column": {
			args: args{
				filters: []*Filter{
					{
						Name:  "non-existent-column",
						Value: "fo",
					},
				},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfSelector: 0,
			wantErr:          true,
		},
		"success-no-filters": {
			args: args{
				filters: []*Filter{},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfSelector: 0,
			wantErr:          false,
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := FilterPredicates(testCase.args.filters, testCase.args.columns)
			if (err != nil) != testCase.wantErr {
				t.Errorf("got err %v wantErr %v", err, testCase.wantErr)
				return
			}
			if err == nil && len(resp) != testCase.wantNoOfSelector {
				t.Errorf("received number of selectors not same received. want: %v, got: %v", testCase.wantNoOfSelector, len(resp))
				return
			}
		})
	}
}

func TestFiltersParsing(t *testing.T) {
	tests := map[string]struct {
		filter           string
		wantedFieldList  string
		wantedValuesList string
		expectedError    string
	}{
		"none":            {filter: "", wantedValuesList: "", wantedFieldList: ""},
		"single":          {filter: "field1=value1", wantedFieldList: "field1", wantedValuesList: "value1"},
		"double":          {filter: "name=acme OR description=widget company", wantedFieldList: "name,description", wantedValuesList: "acme,widget company"},
		"triple":          {filter: "f1=v1 OR f2=v2 OR f3=v3", wantedFieldList: "f1,f2,f3", wantedValuesList: "v1,v2,v3"},
		"equals error":    {filter: "=", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"two equals":      {filter: "= =", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"no field":        {filter: "=v1", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"no value":        {filter: "f1=", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"no equals":       {filter: "f1 v1", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"just OR":         {filter: "OR", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"hanging OR":      {filter: "f1=v1 OR f2=v2 OR", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"OR no left side": {filter: "OR f2=v2", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := ParseFilter(testCase.filter)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			} else {
				checkFilters(t, resp, testCase.wantedFieldList, testCase.wantedValuesList)
			}
		})
	}
}

func TestOrderByOptions(t *testing.T) {
	type args struct {
		orderBys []*OrderBy
		columns  map[string]string
	}
	tests := map[string]struct {
		args            args
		wantNoOfOrderBy int
		wantErr         bool
	}{
		"success-desc": {
			args: args{
				orderBys: []*OrderBy{
					{
						Name:   "name",
						IsDesc: true,
					},
				},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfOrderBy: 1,
			wantErr:         false,
		},
		"success-asc": {
			args: args{
				orderBys: []*OrderBy{
					{
						Name:   "name",
						IsDesc: false,
					},
				},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfOrderBy: 1,
			wantErr:         false,
		},
		"failure-non-existent-column": {
			args: args{
				orderBys: []*OrderBy{
					{
						Name:   "non-existent-column",
						IsDesc: true,
					},
				},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfOrderBy: 0,
			wantErr:         true,
		},
		"success-multiple-orderby": {
			args: args{
				orderBys: []*OrderBy{
					{
						Name:   "name",
						IsDesc: true,
					},
					{
						Name:   "version",
						IsDesc: false,
					},
				},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfOrderBy: 2,
			wantErr:         false,
		},
		"success-no-orderby": {
			args: args{
				orderBys: []*OrderBy{},
				columns: map[string]string{
					"name":    "foo",
					"version": "v1.0",
				},
			},
			wantNoOfOrderBy: 0,
			wantErr:         false,
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := OrderByOptions(testCase.args.orderBys, testCase.args.columns)
			if (err != nil) != testCase.wantErr {
				t.Errorf("got err %v wantErr %v", err, testCase.wantErr)
				return
			}
			if err == nil && len(resp) != testCase.wantNoOfOrderBy {
				t.Errorf("received number of orderBy not same received. want: %v, got: %v", testCase.wantNoOfOrderBy, len(resp))
				return
			}
		})
	}
}

func TestParseOrderBy(t *testing.T) {
	tests := map[string]struct {
		orderBy         string
		wantedFieldList string
		WantedOrderList []bool
		expectedError   string
	}{
		"none": {orderBy: "", wantedFieldList: "", WantedOrderList: []bool{}},
		"single, no order specified, defaults to asc": {orderBy: "name", wantedFieldList: "name", WantedOrderList: []bool{false}},
		"single, desc order specified":                {orderBy: "name desc", wantedFieldList: "name", WantedOrderList: []bool{true}},
		"single, asc order specified":                 {orderBy: "name asc", wantedFieldList: "name", WantedOrderList: []bool{false}},
		"double, both asc":                            {orderBy: "name1 asc, name2 asc", wantedFieldList: "name1,name2", WantedOrderList: []bool{false, false}},
		"double, both desc":                           {orderBy: "name1 desc, name2 desc", wantedFieldList: "name1,name2", WantedOrderList: []bool{true, true}},
		"double, asc and desc":                        {orderBy: "name1 asc, name2 desc", wantedFieldList: "name1,name2", WantedOrderList: []bool{false, true}},
		"double, desc order and missing order":        {orderBy: "name1 desc, name2", wantedFieldList: "name1,name2", WantedOrderList: []bool{true, false}},
		"double, invalid order type1":                 {orderBy: "name1 something, name2=desc", wantedFieldList: "", WantedOrderList: []bool{}, expectedError: "invalid order by"},
		"single, multiple orders":                     {orderBy: "name1 asc desc", wantedFieldList: "", WantedOrderList: []bool{}, expectedError: "invalid order by"},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := ParseOrderBy(testCase.orderBy)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			} else {
				checkOrders(t, resp, testCase.wantedFieldList, testCase.WantedOrderList)
			}
		})
	}
}

func TestComputePageRange(t *testing.T) {
	tests := map[string]struct {
		pageSize      int32
		offset        int32
		totalCount    int
		expectedStart int
		expectedEnd   int
	}{
		"zeros":                                    {pageSize: 0, offset: 0, totalCount: 0, expectedStart: 0, expectedEnd: -1},
		"whole array no page":                      {pageSize: 0, offset: 0, totalCount: 10, expectedStart: 0, expectedEnd: 9},
		"whole array":                              {pageSize: 10, offset: 0, totalCount: 10, expectedStart: 0, expectedEnd: 9},
		"page larger than array":                   {pageSize: 10, offset: 0, totalCount: 5, expectedStart: 0, expectedEnd: 4},
		"first page":                               {pageSize: 10, offset: 0, totalCount: 35, expectedStart: 0, expectedEnd: 9},
		"second page":                              {pageSize: 10, offset: 10, totalCount: 35, expectedStart: 10, expectedEnd: 19},
		"last page":                                {pageSize: 10, offset: 30, totalCount: 35, expectedStart: 30, expectedEnd: 34},
		"page size zero with non-zero offset":      {pageSize: 0, offset: 2, totalCount: 10, expectedStart: 0, expectedEnd: -1},
		"offset greater than total count":          {pageSize: 10, offset: 40, totalCount: 35, expectedStart: 0, expectedEnd: -1},
		"negative page size":                       {pageSize: -10, offset: 0, totalCount: 35, expectedStart: 0, expectedEnd: -1},
		"negative offset":                          {pageSize: 10, offset: -10, totalCount: 35, expectedStart: 0, expectedEnd: -1},
		"total count zero with non-zero page size": {pageSize: 10, offset: 0, totalCount: 0, expectedStart: 0, expectedEnd: -1},
		"total count zero with non-zero offset":    {pageSize: 0, offset: 10, totalCount: 0, expectedStart: 0, expectedEnd: -1},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			start, end := ComputePageRange(testCase.pageSize, testCase.offset, testCase.totalCount)
			assert.Equal(t, testCase.expectedStart, start)
			assert.Equal(t, testCase.expectedEnd, end)
		})
	}
}

func TestHttpToGrpcStatusCode(t *testing.T) {
	type args struct {
		httpStatusCode int
	}
	tests := []struct {
		name string
		args args
		want codes.Code
	}{
		{
			name: "http status ok",
			args: args{
				httpStatusCode: http.StatusOK,
			},
			want: codes.OK,
		},
		{
			name: "http status created",
			args: args{
				httpStatusCode: http.StatusCreated,
			},
			want: codes.OK,
		},
		{
			name: "http status accepted",
			args: args{
				httpStatusCode: http.StatusAccepted,
			},
			want: codes.OK,
		},
		{
			name: "http status bad request",
			args: args{
				httpStatusCode: http.StatusBadRequest,
			},
			want: codes.InvalidArgument,
		},
		{
			name: "http status StatusUnauthorized",
			args: args{
				httpStatusCode: http.StatusUnauthorized,
			},
			want: codes.Unauthenticated,
		},
		{
			name: "http status  StatusForbidden",
			args: args{
				httpStatusCode: http.StatusForbidden,
			},
			want: codes.PermissionDenied,
		},
		{
			name: "http status  StatusNotFound",
			args: args{
				httpStatusCode: http.StatusNotFound,
			},
			want: codes.NotFound,
		},
		{
			name: "http status  StatusMethodNotAllowed",
			args: args{
				httpStatusCode: http.StatusMethodNotAllowed,
			},
			want: codes.Unimplemented,
		},
		{
			name: "http status  StatusRequestTimeout",
			args: args{
				httpStatusCode: http.StatusRequestTimeout,
			},
			want: codes.DeadlineExceeded,
		},
		{
			name: "http status  StatusInternalServerError",
			args: args{
				httpStatusCode: http.StatusInternalServerError,
			},
			want: codes.Unknown,
		},
		{
			name: "http status  StatusServiceUnavailable (unmapped to known gRPC code, default case)",
			args: args{
				httpStatusCode: http.StatusServiceUnavailable,
			},
			want: codes.Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, HttpToGrpcStatusCode(tt.args.httpStatusCode), "HttpToGrpcStatusCode(%v)", tt.args.httpStatusCode)
		})
	}
}

func TestGrpcToHttpStatusCode(t *testing.T) {
	type args struct {
		grpcStatus codes.Code
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "codes.OK",
			args: args{
				grpcStatus: codes.OK,
			},
			want: http.StatusOK,
		},
		{
			name: "codes.Canceled",
			args: args{
				grpcStatus: codes.Canceled,
			},
			want: http.StatusRequestTimeout,
		},
		{
			name: "codes.Unknown",
			args: args{
				grpcStatus: codes.Unknown,
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "codes.InvalidArgument",
			args: args{
				grpcStatus: codes.InvalidArgument,
			},
			want: http.StatusBadRequest,
		},
		{
			name: "codes.DeadlineExceeded",
			args: args{
				grpcStatus: codes.DeadlineExceeded,
			},
			want: http.StatusGatewayTimeout,
		},
		{
			name: "codes.NotFound",
			args: args{
				grpcStatus: codes.NotFound,
			},
			want: http.StatusNotFound,
		},
		{
			name: "codes.AlreadyExists",
			args: args{
				grpcStatus: codes.AlreadyExists,
			},
			want: http.StatusConflict,
		},
		{
			name: "codes.PermissionDenied",
			args: args{
				grpcStatus: codes.PermissionDenied,
			},
			want: http.StatusForbidden,
		},
		{
			name: "codes.Unauthenticated",
			args: args{
				grpcStatus: codes.Unauthenticated,
			},
			want: http.StatusUnauthorized,
		},
		{
			name: "codes.ResourceExhausted",
			args: args{
				grpcStatus: codes.ResourceExhausted,
			},
			want: http.StatusTooManyRequests,
		},
		{
			name: "codes.FailedPrecondition",
			args: args{
				grpcStatus: codes.FailedPrecondition,
			},
			want: http.StatusBadRequest,
		},
		{
			name: "codes.Aborted",
			args: args{
				grpcStatus: codes.Aborted,
			},
			want: http.StatusConflict,
		},
		{
			name: "codes.OutOfRange",
			args: args{
				grpcStatus: codes.OutOfRange,
			},
			want: http.StatusBadRequest,
		},
		{
			name: "codes.Unimplemented",
			args: args{
				grpcStatus: codes.Unimplemented,
			},
			want: http.StatusNotImplemented,
		},
		{
			name: "codes.Internal",
			args: args{
				grpcStatus: codes.Internal,
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "codes.Unavailable",
			args: args{
				grpcStatus: codes.Unavailable,
			},
			want: http.StatusServiceUnavailable,
		},
		{
			name: "codes.DataLoss",
			args: args{
				grpcStatus: codes.DataLoss,
			},
			want: http.StatusInternalServerError,
		},
		{
			name: "Unknown",
			args: args{
				grpcStatus: 12123, // some unknown code
			},
			want: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GrpcToHttpStatusCode(tt.args.grpcStatus), "GrpcToHttpStatusCode(%v)", tt.args.grpcStatus)
		})
	}
}
