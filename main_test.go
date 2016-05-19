package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arschles/assert"
)

type configureRoutesTestCase struct {
	postHandler   http.Handler
	deleteHandler http.Handler
	method        string
	path          string
	respCode      int
	respBody      string
}

func TestConfigureRoutes(t *testing.T) {
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler1"))
	})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
		w.Write([]byte("handler2"))
	})
	testCases := []configureRoutesTestCase{
		configureRoutesTestCase{
			postHandler:   handler1,
			deleteHandler: handler2,
			method:        "GET",
			path:          "/lease",
			respCode:      http.StatusNotFound,
			respBody:      "GET /lease not found",
		},
		configureRoutesTestCase{
			postHandler:   handler1,
			deleteHandler: handler2,
			method:        "POST",
			path:          "/lease",
			respCode:      http.StatusOK,
			respBody:      "handler1",
		},
		configureRoutesTestCase{
			postHandler:   handler1,
			deleteHandler: handler2,
			method:        "POST",
			path:          "/lease/abc",
			respCode:      http.StatusNotFound,
			respBody:      "POST /lease/abc not found",
		},
		configureRoutesTestCase{
			postHandler:   handler1,
			deleteHandler: handler2,
			method:        "DELETE",
			path:          "/lease",
			respCode:      http.StatusNotFound,
			respBody:      "DELETE /lease not found",
		},
		configureRoutesTestCase{
			postHandler:   handler1,
			deleteHandler: handler2,
			method:        "DELETE",
			path:          "/lease/abc",
			respCode:      http.StatusFound,
			respBody:      "handler2",
		},
	}

	for _, testCase := range testCases {
		mux := http.NewServeMux()
		configureRoutesWithAuth(mux, testCase.postHandler, testCase.deleteHandler, "auth token")
		req, err := http.NewRequest(testCase.method, testCase.path, nil)
		req.Header.Set("Authorization", "auth token")
		assert.NoErr(t, err)
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		assert.Equal(
			t,
			res.Code,
			testCase.respCode,
			fmt.Sprintf("response code for %s %s", testCase.method, testCase.path),
		)
		assert.Equal(
			t,
			strings.TrimSpace(string(res.Body.Bytes())),
			testCase.respBody,
			fmt.Sprintf("response body for %s %s", testCase.method, testCase.path),
		)
	}
}

func TestKubeNamespacesFromConfig(t *testing.T) {
	t.Skip("TODO")
}
